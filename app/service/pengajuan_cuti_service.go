package service

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"cutidesk/app/model"
	"cutidesk/app/repository"
	"cutidesk/helper"

	"github.com/gofiber/fiber/v2"
)

// =====================================================================
// Struct & Constructor
// =====================================================================

type PengajuanCutiService struct {
	pengajuanRepo repository.PengajuanCutiRepository
	jenisCutiRepo repository.JenisCutiRepository
	riwayatRepo   repository.RiwayatStatusRepository
	logger        *slog.Logger
}

func BuatPengajuanCutiService(
	pengajuanRepo repository.PengajuanCutiRepository,
	jenisCutiRepo repository.JenisCutiRepository,
	riwayatRepo repository.RiwayatStatusRepository,
	logger *slog.Logger,
) *PengajuanCutiService {
	return &PengajuanCutiService{
		pengajuanRepo: pengajuanRepo,
		jenisCutiRepo: jenisCutiRepo,
		riwayatRepo:   riwayatRepo,
		logger:        logger,
	}
}

// =====================================================================
// Handler: POST /api/pengajuan-cuti — Employee mengajukan cuti
// =====================================================================

func (s *PengajuanCutiService) Buat(c *fiber.Ctx) error {
	var req model.CreatePengajuanCutiRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	// Business rule: validasi tanggal.
	if pesan := validasiTanggalCuti(req.TanggalMulai, req.TanggalSelesai); pesan != "" {
		return helper.BadRequest(pesan)
	}

	authUser := helper.AmbilAuthUser(c)

	ctx, batal := helper.RequestContext()
	defer batal()

	// Validasi jenis cuti ada.
	_, err := s.jenisCutiRepo.CariByID(ctx, req.JenisCutiID)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.BadRequest("Jenis cuti tidak ditemukan")
		}
		s.logger.Error("Gagal mencari jenis cuti", "id", req.JenisCutiID, "error", err)
		return helper.Internal("Gagal memproses pengajuan", err)
	}

	// Buat pengajuan (status default: pending, dari database).
	pengajuan, err := s.pengajuanRepo.Buat(ctx, authUser.ID, req)
	if err != nil {
		s.logger.Error("Gagal membuat pengajuan cuti", "error", err)
		return helper.Internal("Gagal membuat pengajuan cuti", err)
	}

	// Catat riwayat status awal: nil → "pending".
	_, err = s.riwayatRepo.Buat(ctx, pengajuan.ID, authUser.ID, nil, "pending", "Pengajuan cuti dibuat")
	if err != nil {
		s.logger.Error("Gagal mencatat riwayat status awal", "error", err)
		// Tidak gagalkan request — pengajuan sudah tersimpan.
	}

	lokasi := "/api/pengajuan-cuti/" + strconv.Itoa(pengajuan.ID)
	return helper.Created(c, "Pengajuan cuti berhasil dibuat", pengajuan, lokasi)
}

// =====================================================================
// Handler: GET /api/pengajuan-cuti — Daftar pengajuan cuti
// Employee: hanya miliknya. HR/Admin: semua (dengan filter opsional).
// =====================================================================

func (s *PengajuanCutiService) Daftar(c *fiber.Ctx) error {
	format, err := helper.Negotiate(c)
	if err != nil {
		return err
	}

	cursorQuery := helper.ParseCursorQuery(c)

	var cursorParam *model.CursorParam
	if cursorQuery.Cursor != "" {
		decoded, err := helper.DecodeCursor(cursorQuery.Cursor)
		if err != nil {
			return helper.BadRequest("Format cursor tidak valid")
		}
		if decoded != nil {
			cursorParam = &model.CursorParam{
				WaktuDibuat: decoded.WaktuDibuat,
				ID:          decoded.ID,
			}
		}
	}

	authUser := helper.AmbilAuthUser(c)

	// Bangun filter berdasarkan role.
	filter := repository.PengajuanFilter{
		Status: c.Query("status"),
		Search: cursorQuery.Search,
	}

	// Filter jenis_cuti_id (opsional).
	if jcIDStr := c.Query("jenis_cuti_id"); jcIDStr != "" {
		jcID, err := strconv.Atoi(jcIDStr)
		if err != nil || jcID < 1 {
			return helper.BadRequest("Parameter jenis_cuti_id harus berupa angka positif")
		}
		filter.JenisCutiID = &jcID
	}

	// Data-dependent authz: employee hanya lihat miliknya sendiri.
	if authUser.Role == "employee" {
		filter.UserID = &authUser.ID
	} else if userIDStr := c.Query("user_id"); userIDStr != "" {
		// HR/Admin bisa filter per user (opsional).
		uid, err := strconv.Atoi(userIDStr)
		if err != nil || uid < 1 {
			return helper.BadRequest("Parameter user_id harus berupa angka positif")
		}
		filter.UserID = &uid
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	daftar, adaLagi, err := s.pengajuanRepo.Daftar(ctx, cursorQuery.Limit, cursorParam, filter)
	if err != nil {
		s.logger.Error("Gagal mengambil daftar pengajuan", "error", err)
		return helper.Internal("Gagal mengambil daftar pengajuan cuti", err)
	}

	if format == "csv" {
		return s.tulisCSVPengajuan(c, daftar)
	}

	var nextCursor string
	if adaLagi && len(daftar) > 0 {
		terakhir := daftar[len(daftar)-1]
		nextCursor = helper.EncodeCursor(terakhir.CreatedAt, terakhir.ID)
	}

	meta := model.CursorMeta{
		Limit:      cursorQuery.Limit,
		NextCursor: nextCursor,
		HasMore:    adaLagi,
	}

	return helper.SuccessCursor(c, "Daftar pengajuan cuti berhasil diambil", daftar, meta)
}

// =====================================================================
// Handler: GET /api/pengajuan-cuti/:id — Detail pengajuan cuti
// =====================================================================

func (s *PengajuanCutiService) Detail(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	authUser := helper.AmbilAuthUser(c)

	ctx, batal := helper.RequestContext()
	defer batal()

	pengajuan, err := s.pengajuanRepo.CariByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("Pengajuan cuti tidak ditemukan")
		}
		s.logger.Error("Gagal mencari pengajuan", "id", id, "error", err)
		return helper.Internal("Gagal mengambil data pengajuan cuti", err)
	}

	// Data-dependent authz: employee hanya boleh lihat miliknya.
	if !bolehLihatPengajuan(authUser.Role, authUser.ID, pengajuan.UserID) {
		return helper.Forbidden("Anda tidak memiliki akses ke pengajuan ini")
	}

	return helper.Success(c, "Detail pengajuan cuti berhasil diambil", pengajuan)
}

// =====================================================================
// Handler: PATCH /api/pengajuan-cuti/:id/status — HR approve/reject
// =====================================================================

func (s *PengajuanCutiService) UbahStatus(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.UbahStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	// Business rule: status baru harus approved atau rejected.
	if !statusValidUntukPerubahan(req.Status) {
		return helper.BadRequest("Status hanya boleh diubah menjadi approved atau rejected")
	}

	authUser := helper.AmbilAuthUser(c)

	ctx, batal := helper.RequestContext()
	defer batal()

	// Ambil pengajuan saat ini.
	pengajuan, err := s.pengajuanRepo.CariByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("Pengajuan cuti tidak ditemukan")
		}
		s.logger.Error("Gagal mencari pengajuan", "id", id, "error", err)
		return helper.Internal("Gagal memproses perubahan status", err)
	}

	// Business rule: hanya pengajuan berstatus "pending" yang bisa diubah.
	if !apakahStatusBolehDiubah(pengajuan.Status) {
		return helper.BadRequest("Hanya pengajuan berstatus 'pending' yang dapat diubah")
	}

	statusSebelum := pengajuan.Status

	// Update status di database.
	pengajuanBaru, err := s.pengajuanRepo.UbahStatus(ctx, id, req.Status)
	if err != nil {
		s.logger.Error("Gagal mengubah status pengajuan", "id", id, "error", err)
		return helper.Internal("Gagal mengubah status pengajuan", err)
	}

	// Catat riwayat perubahan status.
	catatan := req.Catatan
	if catatan == nil || *catatan == "" {
		if req.Status == "approved" {
			defaultSetuju := "Pengajuan disetujui"
			catatan = &defaultSetuju
		} else {
			defaultTolak := "Pengajuan ditolak"
			catatan = &defaultTolak
		}
	}

	_, err = s.riwayatRepo.Buat(ctx, id, authUser.ID, &statusSebelum, req.Status, *catatan)
	if err != nil {
		s.logger.Error("Gagal mencatat riwayat status", "error", err)
		// Tidak gagalkan request — status sudah berubah.
	}

	return helper.Success(c, "Status pengajuan berhasil diubah", pengajuanBaru)
}

// =====================================================================
// Handler: GET /api/pengajuan-cuti/:id/riwayat — Riwayat perubahan status
// =====================================================================

func (s *PengajuanCutiService) RiwayatStatus(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	authUser := helper.AmbilAuthUser(c)

	ctx, batal := helper.RequestContext()
	defer batal()

	// Pastikan pengajuan ada dan user boleh melihatnya.
	pengajuan, err := s.pengajuanRepo.CariByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("Pengajuan cuti tidak ditemukan")
		}
		s.logger.Error("Gagal mencari pengajuan", "id", id, "error", err)
		return helper.Internal("Gagal mengambil riwayat status", err)
	}

	// Data-dependent authz.
	if !bolehLihatPengajuan(authUser.Role, authUser.ID, pengajuan.UserID) {
		return helper.Forbidden("Anda tidak memiliki akses ke pengajuan ini")
	}

	riwayat, err := s.riwayatRepo.DaftarByPengajuan(ctx, id)
	if err != nil {
		s.logger.Error("Gagal mengambil riwayat status", "pengajuan_id", id, "error", err)
		return helper.Internal("Gagal mengambil riwayat status", err)
	}

	return helper.Success(c, "Riwayat status berhasil diambil", riwayat)
}

// =====================================================================
// Helper internal: CSV export
// =====================================================================

func (s *PengajuanCutiService) tulisCSVPengajuan(c *fiber.Ctx, daftar []model.PengajuanCuti) error {
	header := []string{"ID", "Pengaju", "Jenis Cuti", "Tanggal Mulai", "Tanggal Selesai", "Alasan", "Status", "Dibuat"}

	baris := make([][]string, len(daftar))
	for i, pc := range daftar {
		baris[i] = []string{
			strconv.Itoa(pc.ID),
			pc.NamaPengaju,
			pc.NamaJenisCuti,
			pc.TanggalMulai,
			pc.TanggalSelesai,
			pc.Alasan,
			pc.Status,
			pc.CreatedAt.Format(time.RFC3339),
		}
	}

	return helper.TulisCSV(c, "pengajuan_cuti.csv", header, baris)
}