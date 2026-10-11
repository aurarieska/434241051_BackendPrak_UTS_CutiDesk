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

type UserService struct {
	userRepo      repository.UserRepository
	roleRepo      repository.RoleRepository
	setPermission *helper.SetPermission
	logger        *slog.Logger
}

func BuatUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	setPermission *helper.SetPermission,
	logger *slog.Logger,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		setPermission: setPermission,
		logger:        logger,
	}
}

// =====================================================================
// Handler: GET /api/users — Daftar user dengan cursor pagination
// =====================================================================

func (s *UserService) Daftar(c *fiber.Ctx) error {
	// Content negotiation: JSON (default) atau CSV.
	format, err := helper.Negotiate(c)
	if err != nil {
		return err // sudah berupa AppError 406
	}

	// Parse query parameter: limit, cursor, search.
	cursorQuery := helper.ParseCursorQuery(c)

	// Decode cursor (jika ada).
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

	daftar, adaLagi, err := s.userRepo.DaftarUser(ctx, cursorQuery.Limit, cursorParam, cursorQuery.Search)
	if err != nil {
		s.logger.Error("Gagal mengambil daftar user", "error", err)
		return helper.Internal("Gagal mengambil daftar user", err)
	}

	// Jika format CSV, kembalikan sebagai file CSV.
	if format == "csv" {
		return s.tulisCSVUser(c, daftar)
	}

	// Format JSON: pakai cursor pagination response.
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

	return helper.SuccessCursor(c, "Daftar user berhasil diambil", daftar, meta)
}

// =====================================================================
// Handler: GET /api/users/:id — Detail user
// =====================================================================

func (s *UserService) Detail(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err // sudah berupa AppError
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	user, err := s.userRepo.CariByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("User tidak ditemukan")
		}
		s.logger.Error("Gagal mencari user", "id", id, "error", err)
		return helper.Internal("Gagal mengambil data user", err)
	}

	return helper.Success(c, "Detail user berhasil diambil", user)
}

// =====================================================================
// Handler: POST /api/users — Admin membuat user baru
// =====================================================================

func (s *UserService) Buat(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	// Ambil role_id.
	roleID, err := s.roleRepo.CariRoleID(ctx, req.Role)
	if err != nil {
		s.logger.Error("Gagal mencari role_id", "role", req.Role, "error", err)
		return helper.BadRequest("Role tidak valid: " + req.Role)
	}

	// Hash password.
	hashPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Gagal hash password", "error", err)
		return helper.Internal("Gagal membuat user", err)
	}

	user, err := s.userRepo.BuatUser(ctx, req, hashPassword, roleID)
	if err != nil {
		if errors.Is(err, model.ErrDuplikat) {
			return helper.Conflict("Username atau email sudah digunakan")
		}
		s.logger.Error("Gagal membuat user", "error", err)
		return helper.Internal("Gagal membuat user", err)
	}

	lokasi := "/api/users/" + strconv.Itoa(user.ID)
	return helper.Created(c, "User berhasil dibuat", user, lokasi)
}

// =====================================================================
// Handler: PUT /api/users/:id — Update seluruh field user
// =====================================================================

func (s *UserService) UpdateFull(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	// Business rule: admin tidak boleh menonaktifkan dirinya sendiri.
	authUser := helper.AmbilAuthUser(c)
	if tidakBolehNonaktifkanDiriSendiri(authUser.ID, id, req.IsActive) {
		return helper.BadRequest("Anda tidak dapat menonaktifkan akun Anda sendiri")
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	user, err := s.userRepo.UpdateUser(ctx, id, req)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("User tidak ditemukan")
		}
		if errors.Is(err, model.ErrDuplikat) {
			return helper.Conflict("Email sudah digunakan")
		}
		s.logger.Error("Gagal update user", "id", id, "error", err)
		return helper.Internal("Gagal memperbarui user", err)
	}

	return helper.Success(c, "User berhasil diperbarui", user)
}

// =====================================================================
// Handler: PATCH /api/users/:id — Update sebagian field user
// =====================================================================

func (s *UserService) UpdateParsial(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	// Business rule: admin tidak boleh menonaktifkan dirinya sendiri.
	authUser := helper.AmbilAuthUser(c)
	if tidakBolehNonaktifkanDiriSendiri(authUser.ID, id, req.IsActive) {
		return helper.BadRequest("Anda tidak dapat menonaktifkan akun Anda sendiri")
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	user, err := s.userRepo.PatchUser(ctx, id, req)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("User tidak ditemukan")
		}
		if errors.Is(err, model.ErrDuplikat) {
			return helper.Conflict("Email sudah digunakan")
		}
		s.logger.Error("Gagal patch user", "id", id, "error", err)
		return helper.Internal("Gagal memperbarui user", err)
	}

	return helper.Success(c, "User berhasil diperbarui", user)
}

// =====================================================================
// Handler: PATCH /api/users/:id/role — Ubah role user
// =====================================================================

func (s *UserService) UbahRole(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	// Business rule: admin tidak boleh mengubah role-nya sendiri.
	authUser := helper.AmbilAuthUser(c)
	if tidakBolehUbahRoleSendiri(authUser.ID, id) {
		return helper.BadRequest("Anda tidak dapat mengubah role Anda sendiri")
	}

	// Validasi role dikenal.
	if !s.setPermission.RoleDikenal(req.Role) {
		return helper.BadRequest("Role tidak valid: " + req.Role)
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	user, err := s.userRepo.AssignRole(ctx, id, req.Role)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.NotFound("User tidak ditemukan")
		}
		s.logger.Error("Gagal ubah role", "id", id, "role", req.Role, "error", err)
		return helper.Internal("Gagal mengubah role user", err)
	}

	return helper.Success(c, "Role user berhasil diubah", user)
}

// =====================================================================
// Helper internal: tulis daftar user sebagai CSV
// =====================================================================

func (s *UserService) tulisCSVUser(c *fiber.Ctx, daftar []model.User) error {
	header := []string{"ID", "Username", "Email", "Nama Lengkap", "Role", "Aktif", "Dibuat"}

	baris := make([][]string, len(daftar))
	for i, u := range daftar {
		aktif := "Ya"
		if !u.IsActive {
			aktif = "Tidak"
		}
		baris[i] = []string{
			strconv.Itoa(u.ID),
			u.Username,
			u.Email,
			u.NamaLengkap,
			u.Role,
			aktif,
			u.CreatedAt.Format(time.RFC3339),
		}
	}

	return helper.TulisCSV(c, "users.csv", header, baris)
}