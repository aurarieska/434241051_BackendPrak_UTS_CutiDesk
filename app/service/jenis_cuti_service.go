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

type JenisCutiService struct {
	jenisCutiRepo repository.JenisCutiRepository
	logger        *slog.Logger
}

func BuatJenisCutiService(
	jenisCutiRepo repository.JenisCutiRepository,
	logger *slog.Logger,
) *JenisCutiService {
	return &JenisCutiService{
		jenisCutiRepo: jenisCutiRepo,
		logger:        logger,
	}
}

// =====================================================================
// Handler: GET /api/jenis-cuti — Daftar jenis cuti
// =====================================================================

func (s *JenisCutiService) Daftar(c *fiber.Ctx) error {
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

	ctx, batal := helper.RequestContext()
	defer batal()

	daftar, adaLagi, err := s.jenisCutiRepo.Daftar(ctx, cursorQuery.Limit, cursorParam, cursorQuery.Search)
	if err != nil {
		s.logger.Error("Gagal mengambil daftar jenis cuti", "error", err)
		return helper.Internal("Gagal mengambil daftar jenis cuti", err)
	}

	if format == "csv" {
		return s.tulisCSVJenisCuti(c, daftar)
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

	return helper.SuccessCursor(c, "Daftar jenis cuti berhasil diambil", daftar, meta)
}

// =====================================================================
// Handler: GET /api/jenis-cuti/:id — Detail jenis cuti
// =====================================================================

func (s *JenisCutiService) Detail(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	jc, err := s.jenisCutiRepo.CariByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("Jenis cuti tidak ditemukan")
		}
		s.logger.Error("Gagal mencari jenis cuti", "id", id, "error", err)
		return helper.Internal("Gagal mengambil data jenis cuti", err)
	}

	return helper.Success(c, "Detail jenis cuti berhasil diambil", jc)
}

// =====================================================================
// Handler: POST /api/jenis-cuti — Buat jenis cuti baru
// =====================================================================

func (s *JenisCutiService) Buat(c *fiber.Ctx) error {
	var req model.CreateJenisCutiRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	jc, err := s.jenisCutiRepo.Buat(ctx, req)
	if err != nil {
		if errors.Is(err, model.ErrDuplikat) {
			return helper.Conflict("Nama jenis cuti sudah digunakan")
		}
		s.logger.Error("Gagal membuat jenis cuti", "error", err)
		return helper.Internal("Gagal membuat jenis cuti", err)
	}

	lokasi := "/api/jenis-cuti/" + strconv.Itoa(jc.ID)
	return helper.Created(c, "Jenis cuti berhasil dibuat", jc, lokasi)
}

// =====================================================================
// Handler: PUT /api/jenis-cuti/:id — Update seluruh field
// =====================================================================

func (s *JenisCutiService) UpdateFull(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.UpdateJenisCutiRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	jc, err := s.jenisCutiRepo.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("Jenis cuti tidak ditemukan")
		}
		if errors.Is(err, model.ErrDuplikat) {
			return helper.Conflict("Nama jenis cuti sudah digunakan")
		}
		s.logger.Error("Gagal update jenis cuti", "id", id, "error", err)
		return helper.Internal("Gagal memperbarui jenis cuti", err)
	}

	return helper.Success(c, "Jenis cuti berhasil diperbarui", jc)
}

// =====================================================================
// Handler: PATCH /api/jenis-cuti/:id — Update sebagian field
// =====================================================================

func (s *JenisCutiService) UpdateParsial(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.PatchJenisCutiRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	// Business rule: nama tidak boleh diubah menjadi string kosong.
	if namaJenisCutiKosong(req.Nama) {
		return helper.BadRequest("Nama jenis cuti tidak boleh kosong")
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	jc, err := s.jenisCutiRepo.Patch(ctx, id, req)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("Jenis cuti tidak ditemukan")
		}
		if errors.Is(err, model.ErrDuplikat) {
			return helper.Conflict("Nama jenis cuti sudah digunakan")
		}
		s.logger.Error("Gagal patch jenis cuti", "id", id, "error", err)
		return helper.Internal("Gagal memperbarui jenis cuti", err)
	}

	return helper.Success(c, "Jenis cuti berhasil diperbarui", jc)
}

// =====================================================================
// Handler: DELETE /api/jenis-cuti/:id — Hapus jenis cuti
// =====================================================================

func (s *JenisCutiService) Hapus(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	if err := s.jenisCutiRepo.Hapus(ctx, id); err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("Jenis cuti tidak ditemukan")
		}
		s.logger.Error("Gagal menghapus jenis cuti", "id", id, "error", err)
		return helper.Internal("Gagal menghapus jenis cuti", err)
	}

	return helper.NoContent(c)
}

// =====================================================================
// Helper internal: CSV export
// =====================================================================

func (s *JenisCutiService) tulisCSVJenisCuti(c *fiber.Ctx, daftar []model.JenisCuti) error {
	header := []string{"ID", "Nama", "Deskripsi", "Dibuat"}

	baris := make([][]string, len(daftar))
	for i, jc := range daftar {
		deskripsi := ""
		if jc.Deskripsi != nil {
			deskripsi = *jc.Deskripsi
		}
		baris[i] = []string{
			strconv.Itoa(jc.ID),
			jc.Nama,
			deskripsi,
			jc.CreatedAt.Format(time.RFC3339),
		}
	}

	return helper.TulisCSV(c, "jenis_cuti.csv", header, baris)
}