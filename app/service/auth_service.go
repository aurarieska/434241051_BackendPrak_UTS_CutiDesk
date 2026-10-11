package service

import (
	"context"
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

type AuthService struct {
	userRepo           repository.UserRepository
	tokenRepo          repository.TokenRepository
	roleRepo           repository.RoleRepository
	jwtManager         *helper.JWTManager
	setPermission      *helper.SetPermission
	logger             *slog.Logger
	durasiRefreshToken time.Duration
}

func BuatAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	roleRepo repository.RoleRepository,
	jwtManager *helper.JWTManager,
	setPermission *helper.SetPermission,
	logger *slog.Logger,
	durasiRefreshHari int,
) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		tokenRepo:          tokenRepo,
		roleRepo:           roleRepo,
		jwtManager:         jwtManager,
		setPermission:      setPermission,
		logger:             logger,
		durasiRefreshToken: time.Duration(durasiRefreshHari) * 24 * time.Hour,
	}
}

// =====================================================================
// Response struct (lokal, tidak diekspor)
// =====================================================================

type authResponse struct {
	User   model.User      `json:"user"`
	Tokens model.TokenPair `json:"tokens"`
}

// =====================================================================
// Handler: POST /api/auth/register
// =====================================================================

func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	// Validasi input (username, email, password kuat, dll).
	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	// Tentukan role default (employee) — mass assignment prevention.
	role := roleDefaultRegister()

	// Ambil role_id dari tabel roles.
	roleID, err := s.roleRepo.CariRoleID(ctx, role)
	if err != nil {
		s.logger.Error("Gagal mencari role_id", "role", role, "error", err)
		return helper.Internal("Gagal memproses registrasi", err)
	}

	// Hash password dengan bcrypt cost 12.
	hashPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Gagal hash password", "error", err)
		return helper.Internal("Gagal memproses registrasi", err)
	}

	// Buat CreateUserRequest dari RegisterRequest + role server.
	createReq := model.CreateUserRequest{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password, // tidak dipakai oleh repo (repo pakai hashPassword)
		NamaLengkap: req.NamaLengkap,
		Role:        role,
	}

	// Simpan user ke database.
	user, err := s.userRepo.BuatUser(ctx, createReq, hashPassword, roleID)
	if err != nil {
		if errors.Is(err, model.ErrDuplikat) {
			return helper.Conflict("Username atau email sudah digunakan")
		}
		s.logger.Error("Gagal membuat user", "error", err)
		return helper.Internal("Gagal memproses registrasi", err)
	}

	// Generate token pair (access + refresh).
	tokens, err := s.buatTokenPair(ctx, user)
	if err != nil {
		s.logger.Error("Gagal membuat token", "error", err)
		return helper.Internal("Gagal memproses registrasi", err)
	}

	lokasi := "/api/users/" + strconv.Itoa(user.ID)
	return helper.Created(c, "Registrasi berhasil", authResponse{
		User:   *user,
		Tokens: *tokens,
	}, lokasi)
}

// =====================================================================
// Handler: POST /api/auth/login
// =====================================================================

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	// Cari user berdasarkan username.
	user, err := s.userRepo.CariByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			// Anti timing attack: tetap jalankan bcrypt meski user tidak ditemukan.
			helper.VerifyDummyPassword()
			return helper.Unauthorized("Username atau password salah")
		}
		s.logger.Error("Gagal mencari user", "error", err)
		return helper.Internal("Gagal memproses login", err)
	}

	// Verifikasi password.
	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized("Username atau password salah")
	}

	// Cek apakah user masih aktif.
	if !apakahUserAktif(user) {
		return helper.Forbidden("Akun Anda telah dinonaktifkan, hubungi admin")
	}

	// Generate token pair.
	tokens, err := s.buatTokenPair(ctx, user)
	if err != nil {
		s.logger.Error("Gagal membuat token", "error", err)
		return helper.Internal("Gagal memproses login", err)
	}

	return helper.Success(c, "Login berhasil", authResponse{
		User:   *user,
		Tokens: *tokens,
	})
}

// =====================================================================
// Handler: POST /api/auth/refresh
// =====================================================================

func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	// Hash token yang dikirim client.
	tokenHash := helper.SHA256Hex(req.RefreshToken)

	// Cari token di database.
	tokenData, err := s.tokenRepo.CariByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.Unauthorized("Refresh token tidak valid")
		}
		s.logger.Error("Gagal mencari refresh token", "error", err)
		return helper.Internal("Gagal memproses refresh token", err)
	}

	// Cek apakah token sudah dicabut.
	if apakahTokenDicabut(tokenData) {
		return helper.Unauthorized("Refresh token sudah dicabut")
	}

	// Cek apakah token sudah kedaluwarsa.
	if apakahTokenKedaluwarsa(tokenData) {
		return helper.Unauthorized("Refresh token sudah kedaluwarsa")
	}

	// Cari user pemilik token.
	user, err := s.userRepo.CariByID(ctx, tokenData.UserID)
	if err != nil {
		if errors.Is(err, model.ErrTidakDitemukan) {
			return helper.Unauthorized("User tidak ditemukan")
		}
		s.logger.Error("Gagal mencari user untuk refresh", "error", err)
		return helper.Internal("Gagal memproses refresh token", err)
	}

	// Cek apakah user masih aktif.
	if !apakahUserAktif(user) {
		return helper.Forbidden("Akun telah dinonaktifkan")
	}

	// Rotation: cabut token lama.
	if err := s.tokenRepo.CabutByHash(ctx, tokenHash); err != nil {
		s.logger.Error("Gagal mencabut token lama", "error", err)
		return helper.Internal("Gagal memproses refresh token", err)
	}

	// Generate token pair baru.
	tokens, err := s.buatTokenPair(ctx, user)
	if err != nil {
		s.logger.Error("Gagal membuat token baru", "error", err)
		return helper.Internal("Gagal memproses refresh token", err)
	}

	return helper.Success(c, "Token berhasil diperbarui", tokens)
}

// =====================================================================
// Handler: POST /api/auth/logout
// =====================================================================

func (s *AuthService) Logout(c *fiber.Ctx) error {
	var req model.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("Format body tidak valid")
	}

	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	ctx, batal := helper.RequestContext()
	defer batal()

	tokenHash := helper.SHA256Hex(req.RefreshToken)

	// Cabut refresh token. Jika tidak ditemukan, tetap anggap sukses
	// (idempotent — user tidak perlu tahu apakah token masih valid).
	if err := s.tokenRepo.CabutByHash(ctx, tokenHash); err != nil {
		if !errors.Is(err, model.ErrTidakDitemukan) {
			s.logger.Error("Gagal mencabut refresh token", "error", err)
			return helper.Internal("Gagal memproses logout", err)
		}
	}

	return helper.Success(c, "Logout berhasil", nil)
}

// =====================================================================
// Handler: GET /api/auth/me
// =====================================================================

func (s *AuthService) Me(c *fiber.Ctx) error {
	authUser := helper.AmbilAuthUser(c)
	if authUser == nil {
		return helper.Unauthorized("Autentikasi diperlukan")
	}

	// Ambil daftar permission berdasarkan role.
	permissions := s.setPermission.DaftarPermission(authUser.Role)

	return helper.Success(c, "Data user saat ini", model.MeResponse{
		User:        *authUser,
		Permissions: permissions,
	})
}

// =====================================================================
// Helper internal: buat pasangan access + refresh token
// =====================================================================

func (s *AuthService) buatTokenPair(ctx context.Context, user *model.User) (*model.TokenPair, error) {
	authUser := &model.AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	// Generate JWT access token.
		accessToken, expiresIn, err := s.jwtManager.BuatAccessToken(*authUser)
	if err != nil {
		return nil, err
	}

	// Generate random refresh token (32 bytes → 64 char hex).
	rawRefreshToken, err := helper.RandomToken()
	if err != nil {
   		return nil, err
	}
	tokenHash := helper.SHA256Hex(rawRefreshToken)
	expirasiPada := time.Now().Add(s.durasiRefreshToken)

	// Simpan hash ke database (bukan raw token).
	if err := s.tokenRepo.SimpanRefreshToken(ctx, user.ID, tokenHash, expirasiPada); err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}