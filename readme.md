# CutiDesk — Sistem Manajemen Cuti Karyawan

REST API backend untuk pengelolaan pengajuan cuti karyawan, dibangun dengan arsitektur bersih (Clean Architecture) menggunakan **Go**, **Fiber v2**, dan **PostgreSQL**.

## Tech Stack

- **Bahasa**: Go (Golang)
- **Framework HTTP**: Fiber v2
- **Database**: PostgreSQL (driver: pgx v5 + pgxpool)
- **Autentikasi**: JWT HS256 (access token 15 menit + refresh token 7 hari)
- **Hashing Password**: bcrypt (cost 12)
- **Konfigurasi**: File `.env`

## Arsitektur

Project ini menggunakan pola **Clean Architecture** dengan lapisan:

```
model → repository → service → route
```

| Lapisan      | Tanggung Jawab                                      |
|--------------|------------------------------------------------------|
| `model`      | Definisi struct data dan request/response             |
| `repository` | Akses database (query SQL)                            |
| `service`    | Logika bisnis dan handler HTTP                        |
| `route`      | Pendaftaran endpoint dan middleware                   |
| `helper`     | Utilitas: hashing, JWT, validasi, response formatter  |
| `config`     | Konfigurasi environment dan permission RBAC           |
| `middleware` | Autentikasi JWT dan otorisasi RBAC                    |

## Fitur

- **Autentikasi & Otorisasi**
  - Registrasi user baru
  - Login dengan username & password
  - JWT access token (15 menit) + refresh token rotation (7 hari)
  - Logout dengan revokasi refresh token
  - Endpoint profil pengguna (`/auth/me`)

- **RBAC (Role-Based Access Control)**
  - 3 role: `admin`, `hr`, `employee`
  - 14 permission granular (user:create, cuti:approve, dll.)
  - Middleware otorisasi per-endpoint berdasarkan permission

- **Manajemen User** (Admin & HR)
  - CRUD user lengkap (buat, lihat, update lengkap, update parsial, hapus/nonaktifkan)
  - Ubah role user secara terpisah
  - Soft delete (nonaktifkan akun)

- **Manajemen Jenis Cuti** (Admin & HR)
  - CRUD jenis cuti (Cuti Tahunan, Cuti Sakit, Cuti Besar, dll.)
  - Update lengkap (PUT) dan update parsial (PATCH)

- **Pengajuan Cuti**
  - Employee mengajukan cuti
  - HR/Admin melihat seluruh pengajuan, Employee hanya miliknya
  - HR/Admin approve/reject pengajuan
  - Riwayat perubahan status pengajuan

- **Fitur API Lanjutan**
  - Cursor-based pagination (keyset pagination)
  - Filter dan pencarian pada daftar data
  - Content negotiation (JSON & CSV export)
  - Validasi input dengan pesan error per-field
  - Format response konsisten (success, message, data)
  - Request ID pada setiap response untuk tracing

## Konfigurasi

Buat file `.env` di root project:

```env
APP_PORT=3000
JWT_SECRET=rahasia_jwt_anda

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password_anda
DB_NAME=cutidesk
DB_SSLMODE=disable
```

## Menjalankan Server

```bash
go run main.go
```

Server berjalan di `http://localhost:3000`.

## Migrasi Database

Jalankan file migrasi secara berurutan di PostgreSQL:

```bash
psql -U postgres -d cutidesk -f migrations/001_create_user.sql
psql -U postgres -d cutidesk -f migrations/002_refresh_tokens.sql
psql -U postgres -d cutidesk -f migrations/003_create_jenis_cuti.sql
psql -U postgres -d cutidesk -f migrations/004_rbac.sql
psql -U postgres -d cutidesk -f migrations/005_create_pengajuan_cuti.sql
psql -U postgres -d cutidesk -f migrations/006_create_riwayat_status.sql
psql -U postgres -d cutidesk -f migrations/007_cursor_indexes.sql
psql -U postgres -d cutidesk -f migrations/008_seed_admin.sql
```

---

## Dokumentasi API

Base URL: `http://localhost:3000`

### Format Response

Semua response menggunakan format JSON konsisten:

**Sukses:**
```json
{
  "success": true,
  "message": "Pesan sukses",
  "data": { ... }
}
```

**Sukses dengan Pagination:**
```json
{
  "success": true,
  "message": "Pesan sukses",
  "data": [ ... ],
  "meta": {
    "limit": 10,
    "next_cursor": "eyJ...",
    "has_more": true
  }
}
```

**Error:**
```json
{
  "success": false,
  "code": "KODE_ERROR",
  "message": "Pesan error",
  "request_id": "uuid"
}
```

**Error Validasi:**
```json
{
  "success": false,
  "code": "VALIDATION_ERROR",
  "message": "validasi gagal",
  "fields": {
    "nama_field": "pesan error"
  },
  "request_id": "uuid"
}
```

### Kode Error

| Kode                | HTTP Status | Keterangan                    |
|---------------------|-------------|-------------------------------|
| `VALIDATION_ERROR`  | 400         | Input tidak valid              |
| `BAD_REQUEST`       | 400         | Request tidak sesuai           |
| `UNAUTHORIZED`      | 401         | Token tidak valid / tidak ada  |
| `FORBIDDEN`         | 403         | Tidak punya akses              |
| `NOT_FOUND`         | 404         | Resource tidak ditemukan        |
| `CONFLICT`          | 409         | Data duplikat                  |
| `INTERNAL_ERROR`    | 500         | Kesalahan server               |

---

### 1. Health Check

#### `GET /`

Cek status server.

**Response:**
```json
{
  "success": true,
  "message": "CutiDesk API berjalan"
}
```

---

### 2. Autentikasi

#### `POST /api/auth/register`

Registrasi user baru (role default: `employee`).

**Request Body:**
```json
{
  "username": "budi",
  "email": "budi@example.com",
  "password": "Budi@1234",
  "nama_lengkap": "Budi Santoso"
}
```

| Field          | Tipe   | Wajib | Validasi                      |
|----------------|--------|-------|-------------------------------|
| `username`     | string | Ya    | 3–50 karakter, unik           |
| `email`        | string | Ya    | Format email valid, unik       |
| `password`     | string | Ya    | Minimal 8 karakter             |
| `nama_lengkap` | string | Ya    | Tidak boleh kosong             |

**Response Sukses (201):**
```json
{
  "success": true,
  "message": "Registrasi berhasil",
  "data": {
    "id": 1,
    "username": "budi",
    "email": "budi@example.com",
    "nama_lengkap": "Budi Santoso",
    "role": "employee",
    "is_active": true,
    "created_at": "2026-10-12T00:00:00+07:00",
    "updated_at": "2026-10-12T00:00:00+07:00"
  }
}
```

---

#### `POST /api/auth/login`

Login dan mendapatkan token JWT.

**Request Body:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

| Field      | Tipe   | Wajib | Keterangan  |
|------------|--------|-------|-------------|
| `username` | string | Ya    | Username    |
| `password` | string | Ya    | Password    |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "user": {
      "id": 8,
      "username": "admin",
      "email": "admin@cutidesk.local",
      "nama_lengkap": "Administrator",
      "role": "admin",
      "is_active": true,
      "created_at": "2026-10-12T00:30:05+07:00",
      "updated_at": "2026-10-12T00:46:38+07:00"
    },
    "tokens": {
      "access_token": "eyJhbGci...",
      "refresh_token": "53cab49b...",
      "token_type": "",
      "expires_in": 900
    }
  }
}
```

---

#### `POST /api/auth/refresh`

Perbarui access token menggunakan refresh token (rotation).

**Request Body:**
```json
{
  "refresh_token": "53cab49be87477be92c63246..."
}
```

| Field           | Tipe   | Wajib | Keterangan                    |
|-----------------|--------|-------|-------------------------------|
| `refresh_token` | string | Ya    | Refresh token dari login      |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Token berhasil diperbarui",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "token_baru...",
    "token_type": "",
    "expires_in": 900
  }
}
```

---

#### `POST /api/auth/logout`

Logout dan revokasi refresh token. **Memerlukan autentikasi.**

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "refresh_token": "53cab49be87477be92c63246..."
}
```

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Logout berhasil"
}
```

---

#### `GET /api/auth/me`

Mendapatkan profil user yang sedang login. **Memerlukan autentikasi.**

**Header:**
```
Authorization: Bearer <access_token>
```

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Profil berhasil diambil",
  "data": {
    "id": 8,
    "username": "admin",
    "email": "admin@cutidesk.local",
    "nama_lengkap": "Administrator",
    "role": "admin",
    "is_active": true,
    "created_at": "2026-10-12T00:30:05+07:00",
    "updated_at": "2026-10-12T00:46:38+07:00"
  }
}
```

---

### 3. Manajemen User

Semua endpoint memerlukan autentikasi. Akses dibatasi berdasarkan permission RBAC.

#### `GET /api/users`

Daftar semua user. **Permission: `user:read`** (admin, hr).

**Header:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**

| Parameter | Tipe   | Keterangan                         |
|-----------|--------|------------------------------------|
| `limit`   | int    | Jumlah data per halaman (default 10) |
| `cursor`  | string | Cursor untuk halaman berikutnya     |
| `search`  | string | Pencarian di nama/username/email    |
| `role`    | string | Filter berdasarkan role             |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Daftar user berhasil diambil",
  "data": [
    {
      "id": 8,
      "username": "admin",
      "email": "admin@cutidesk.local",
      "nama_lengkap": "Administrator",
      "role": "admin",
      "is_active": true,
      "created_at": "2026-10-12T00:30:05+07:00",
      "updated_at": "2026-10-12T00:46:38+07:00"
    }
  ],
  "meta": {
    "limit": 10,
    "next_cursor": "",
    "has_more": false
  }
}
```

---

#### `GET /api/users/:id`

Detail user berdasarkan ID. **Permission: `user:read`** (admin, hr).

**Header:**
```
Authorization: Bearer <access_token>
```

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Detail user berhasil diambil",
  "data": {
    "id": 18,
    "username": "yusuf",
    "email": "yusuf@example.com",
    "nama_lengkap": "Yusuf Pratama",
    "role": "hr",
    "is_active": true,
    "created_at": "2026-10-12T01:50:43+07:00",
    "updated_at": "2026-10-12T01:50:43+07:00"
  }
}
```

---

#### `POST /api/users`

Buat user baru. **Permission: `user:create`** (admin, hr).

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "username": "dewi",
  "email": "dewi@example.com",
  "password": "Dewi@1234",
  "nama_lengkap": "Dewi Lestari",
  "role": "employee"
}
```

| Field          | Tipe   | Wajib | Validasi                                 |
|----------------|--------|-------|------------------------------------------|
| `username`     | string | Ya    | 3–50 karakter, unik                      |
| `email`        | string | Ya    | Format email valid, unik                  |
| `password`     | string | Ya    | Minimal 8 karakter                        |
| `nama_lengkap` | string | Ya    | Tidak boleh kosong                        |
| `role`         | string | Ya    | Salah satu: `admin`, `hr`, `employee`     |

**Response Sukses (201):**
```json
{
  "success": true,
  "message": "User berhasil dibuat",
  "data": {
    "id": 21,
    "username": "dewi",
    "email": "dewi@example.com",
    "nama_lengkap": "Dewi Lestari",
    "role": "employee",
    "is_active": true,
    "created_at": "2026-10-12T02:00:00+07:00",
    "updated_at": "2026-10-12T02:00:00+07:00"
  }
}
```

---

#### `PUT /api/users/:id`

Update data user. **Permission: `user:update`** (admin, hr).

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body (semua field wajib diisi):**
```json
{
  "username": "dewi",
  "email": "dewi_updated@example.com",
  "password": "DewiNew@1234",
  "nama_lengkap": "Dewi Lestari Utami",
  "role": "employee"
}
```

| Field          | Tipe   | Wajib | Keterangan                    |
|----------------|--------|-------|-------------------------------|
| `username`     | string | Ya    | Username (bisa sama/berubah)  |
| `email`        | string | Ya    | Email (bisa sama/berubah)     |
| `password`     | string | Ya    | Password baru                 |
| `nama_lengkap` | string | Ya    | Nama lengkap                  |
| `role`         | string | Ya    | Role user                     |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "User berhasil diperbarui",
  "data": {
    "id": 21,
    "username": "dewi",
    "email": "dewi_updated@example.com",
    "nama_lengkap": "Dewi Lestari Utami",
    "role": "employee",
    "is_active": true,
    "created_at": "2026-10-12T02:00:00+07:00",
    "updated_at": "2026-10-12T02:10:00+07:00"
  }
}
```

---

#### `PATCH /api/users/:id`

Update sebagian data user (partial update). **Permission: `user:update`** (admin, hr).

Hanya field yang dikirim yang akan diubah. Field yang tidak dikirim tetap seperti semula.

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body (kirim field yang ingin diubah saja):**
```json
{
  "nama_lengkap": "Dewi Lestari Utami",
  "email": "dewi_baru@example.com"
}
```

| Field          | Tipe   | Wajib | Keterangan                     |
|----------------|--------|-------|--------------------------------|
| `username`     | string | Tidak | Username baru                  |
| `email`        | string | Tidak | Email baru                     |
| `password`     | string | Tidak | Password baru                  |
| `nama_lengkap` | string | Tidak | Nama lengkap baru              |
| `role`         | string | Tidak | Role baru                      |
| `is_active`    | bool   | Tidak | Status aktif/nonaktif          |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "User berhasil diperbarui",
  "data": {
    "id": 21,
    "username": "dewi",
    "email": "dewi_baru@example.com",
    "nama_lengkap": "Dewi Lestari Utami",
    "role": "employee",
    "is_active": true,
    "created_at": "2026-10-12T02:00:00+07:00",
    "updated_at": "2026-10-12T02:20:00+07:00"
  }
}
```

---

#### `PATCH /api/users/:id/role`

Ubah role user secara spesifik. **Permission: `user:update`** (admin, hr).

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "role": "hr"
}
```

| Field  | Tipe   | Wajib | Validasi                             |
|--------|--------|-------|--------------------------------------|
| `role` | string | Ya    | Salah satu: `admin`, `hr`, `employee`|

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Role user berhasil diubah",
  "data": {
    "id": 21,
    "username": "dewi",
    "email": "dewi@example.com",
    "nama_lengkap": "Dewi Lestari",
    "role": "hr",
    "is_active": true,
    "created_at": "2026-10-12T02:00:00+07:00",
    "updated_at": "2026-10-12T02:25:00+07:00"
  }
}
```

---

### 4. Manajemen Jenis Cuti

#### `GET /api/jenis-cuti`

Daftar semua jenis cuti. **Permission: `jenis_cuti:read`** (admin, hr, employee).

**Header:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**

| Parameter | Tipe   | Keterangan                          |
|-----------|--------|-------------------------------------|
| `limit`   | int    | Jumlah data per halaman (default 10) |
| `cursor`  | string | Cursor untuk halaman berikutnya      |
| `search`  | string | Pencarian di nama jenis cuti         |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Daftar jenis cuti berhasil diambil",
  "data": [
    {
      "id": 1,
      "nama": "Cuti Ibadah Haji",
      "deskripsi": "Cuti untuk menunaikan ibadah haji",
      "kuota_hari": 40,
      "created_at": "2026-10-12T02:00:00+07:00",
      "updated_at": "2026-10-12T02:00:00+07:00"
    }
  ],
  "meta": {
    "limit": 10,
    "next_cursor": "",
    "has_more": false
  }
}
```

---

#### `GET /api/jenis-cuti/:id`

Detail jenis cuti. **Permission: `jenis_cuti:read`** (admin, hr, employee).

**Header:**
```
Authorization: Bearer <access_token>
```

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Detail jenis cuti berhasil diambil",
  "data": {
    "id": 1,
    "nama": "Cuti Ibadah Haji",
    "deskripsi": "Cuti untuk menunaikan ibadah haji",
    "kuota_hari": 40,
    "created_at": "2026-10-12T02:00:00+07:00",
    "updated_at": "2026-10-12T02:00:00+07:00"
  }
}
```

---

#### `POST /api/jenis-cuti`

Buat jenis cuti baru. **Permission: `jenis_cuti:create`** (admin, hr).

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "nama": "Cuti Tahunan",
  "deskripsi": "Cuti tahunan untuk karyawan tetap",
  "kuota_hari": 12
}
```

| Field        | Tipe   | Wajib | Validasi               |
|--------------|--------|-------|------------------------|
| `nama`       | string | Ya    | Tidak boleh kosong, unik |
| `deskripsi`  | string | Tidak | Deskripsi opsional      |
| `kuota_hari` | int    | Ya    | Minimal 1               |

**Response Sukses (201):**
```json
{
  "success": true,
  "message": "Jenis cuti berhasil dibuat",
  "data": {
    "id": 5,
    "nama": "Cuti Tahunan",
    "deskripsi": "Cuti tahunan untuk karyawan tetap",
    "kuota_hari": 12,
    "created_at": "2026-10-12T02:00:00+07:00",
    "updated_at": "2026-10-12T02:00:00+07:00"
  }
}
```

---

#### `PUT /api/jenis-cuti/:id`

Update jenis cuti. **Permission: `jenis_cuti:update`** (admin, hr).

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body (semua field wajib diisi):**
```json
{
  "nama": "Cuti Tahunan",
  "deskripsi": "Cuti tahunan karyawan tetap, maksimal 12 hari kerja",
  "kuota_hari": 12
}
```

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Jenis cuti berhasil diperbarui",
  "data": {
    "id": 5,
    "nama": "Cuti Tahunan",
    "deskripsi": "Cuti tahunan karyawan tetap, maksimal 12 hari kerja",
    "kuota_hari": 12,
    "created_at": "2026-10-12T02:00:00+07:00",
    "updated_at": "2026-10-12T02:15:00+07:00"
  }
}
```

---

#### `PATCH /api/jenis-cuti/:id`

Update sebagian data jenis cuti (partial update). **Permission: `jenis_cuti:update`** (admin, hr).

Hanya field yang dikirim yang akan diubah.

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body (kirim field yang ingin diubah saja):**
```json
{
  "kuota_hari": 15
}
```

| Field        | Tipe   | Wajib | Keterangan             |
|--------------|--------|-------|------------------------|
| `nama`       | string | Tidak | Nama jenis cuti baru   |
| `deskripsi`  | string | Tidak | Deskripsi baru          |
| `kuota_hari` | int    | Tidak | Kuota hari baru         |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Jenis cuti berhasil diperbarui",
  "data": {
    "id": 5,
    "nama": "Cuti Tahunan",
    "deskripsi": "Cuti tahunan untuk karyawan tetap",
    "kuota_hari": 15,
    "created_at": "2026-10-12T02:00:00+07:00",
    "updated_at": "2026-10-12T02:20:00+07:00"
  }
}
```

---

#### `DELETE /api/jenis-cuti/:id`

Hapus jenis cuti. **Permission: `jenis_cuti:delete`** (admin, hr).

**Header:**
```
Authorization: Bearer <access_token>
```

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Jenis cuti berhasil dihapus"
}
```

---

### 5. Pengajuan Cuti

#### `POST /api/pengajuan-cuti`

Ajukan cuti baru. **Permission: `cuti:create`** (employee).

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "jenis_cuti_id": 1,
  "tanggal_mulai": "2027-03-10",
  "tanggal_selesai": "2027-03-14",
  "alasan": "Persiapan pernikahan"
}
```

| Field             | Tipe   | Wajib | Validasi                                      |
|-------------------|--------|-------|-----------------------------------------------|
| `jenis_cuti_id`   | int    | Ya    | ID jenis cuti yang valid                       |
| `tanggal_mulai`   | string | Ya    | Format `YYYY-MM-DD`, tidak boleh di masa lalu  |
| `tanggal_selesai` | string | Ya    | Format `YYYY-MM-DD`, >= tanggal_mulai          |
| `alasan`          | string | Ya    | Tidak boleh kosong                             |

**Response Sukses (201):**
```json
{
  "success": true,
  "message": "Pengajuan cuti berhasil dibuat",
  "data": {
    "id": 1,
    "user_id": 20,
    "jenis_cuti_id": 1,
    "tanggal_mulai": "2027-03-10",
    "tanggal_selesai": "2027-03-14",
    "alasan": "Persiapan pernikahan",
    "status": "pending",
    "created_at": "2026-10-12T03:00:00+07:00",
    "updated_at": "2026-10-12T03:00:00+07:00"
  }
}
```

---

#### `GET /api/pengajuan-cuti`

Daftar pengajuan cuti. **Permission: `cuti:read`** (admin, hr, employee).

- Employee: hanya melihat pengajuan miliknya sendiri.
- HR/Admin: melihat semua pengajuan (dengan filter opsional).

**Header:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**

| Parameter       | Tipe   | Keterangan                                    |
|-----------------|--------|-----------------------------------------------|
| `limit`         | int    | Jumlah data per halaman (default 10)           |
| `cursor`        | string | Cursor untuk halaman berikutnya                |
| `search`        | string | Pencarian di nama pengaju / alasan             |
| `status`        | string | Filter: `pending`, `approved`, `rejected`      |
| `jenis_cuti_id` | int    | Filter berdasarkan jenis cuti                  |
| `user_id`       | int    | Filter berdasarkan user (khusus HR/Admin)      |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Daftar pengajuan cuti berhasil diambil",
  "data": [
    {
      "id": 1,
      "user_id": 20,
      "jenis_cuti_id": 1,
      "tanggal_mulai": "2027-03-10",
      "tanggal_selesai": "2027-03-14",
      "alasan": "Persiapan pernikahan",
      "status": "pending",
      "nama_pengaju": "Toni Setiawan Putra",
      "nama_jenis_cuti": "Cuti Ibadah Haji",
      "created_at": "2026-10-12T03:00:00+07:00",
      "updated_at": "2026-10-12T03:00:00+07:00"
    }
  ],
  "meta": {
    "limit": 10,
    "next_cursor": "",
    "has_more": false
  }
}
```

> **Content Negotiation:** Tambahkan header `Accept: text/csv` untuk mendapatkan response dalam format CSV.

---

#### `GET /api/pengajuan-cuti/:id`

Detail pengajuan cuti. **Permission: `cuti:read`** (admin, hr, employee — employee hanya miliknya).

**Header:**
```
Authorization: Bearer <access_token>
```

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Detail pengajuan cuti berhasil diambil",
  "data": {
    "id": 1,
    "user_id": 20,
    "jenis_cuti_id": 1,
    "tanggal_mulai": "2027-03-10",
    "tanggal_selesai": "2027-03-14",
    "alasan": "Persiapan pernikahan",
    "status": "pending",
    "nama_pengaju": "Toni Setiawan Putra",
    "nama_jenis_cuti": "Cuti Ibadah Haji",
    "created_at": "2026-10-12T03:00:00+07:00",
    "updated_at": "2026-10-12T03:00:00+07:00"
  }
}
```

---

#### `PATCH /api/pengajuan-cuti/:id/status`

Approve atau reject pengajuan cuti. **Permission: `cuti:approve`** (admin, hr).

Hanya pengajuan berstatus `pending` yang bisa diubah.

**Header:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "status": "approved",
  "catatan": "Disetujui, selamat menikah"
}
```

| Field     | Tipe   | Wajib | Validasi                              |
|-----------|--------|-------|---------------------------------------|
| `status`  | string | Ya    | `approved` atau `rejected`            |
| `catatan` | string | Tidak | Catatan/alasan (ada default otomatis) |

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Status pengajuan berhasil diubah",
  "data": {
    "id": 1,
    "user_id": 20,
    "jenis_cuti_id": 1,
    "tanggal_mulai": "2027-03-10",
    "tanggal_selesai": "2027-03-14",
    "alasan": "Persiapan pernikahan",
    "status": "approved",
    "created_at": "2026-10-12T03:00:00+07:00",
    "updated_at": "2026-10-12T03:30:00+07:00"
  }
}
```

---

#### `GET /api/pengajuan-cuti/:id/riwayat`

Riwayat perubahan status pengajuan cuti. **Permission: `riwayat:read`** (admin, hr, employee — employee hanya miliknya).

**Header:**
```
Authorization: Bearer <access_token>
```

**Response Sukses (200):**
```json
{
  "success": true,
  "message": "Riwayat status berhasil diambil",
  "data": [
    {
      "id": 1,
      "pengajuan_cuti_id": 1,
      "diubah_oleh": 20,
      "status_sebelum": null,
      "status_sesudah": "pending",
      "catatan": "Pengajuan cuti dibuat",
      "nama_pengubah": "Toni Setiawan Putra",
      "created_at": "2026-10-12T03:00:00+07:00"
    },
    {
      "id": 2,
      "pengajuan_cuti_id": 1,
      "diubah_oleh": 8,
      "status_sebelum": "pending",
      "status_sesudah": "approved",
      "catatan": "Disetujui, selamat menikah",
      "nama_pengubah": "Administrator",
      "created_at": "2026-10-12T03:30:00+07:00"
    }
  ]
}
```

---

### Ringkasan Endpoint

| #  | Method | Endpoint                            | Permission           | Role yang Boleh       |
|----|--------|-------------------------------------|----------------------|-----------------------|
| 1  | GET    | `/`                                 | —                    | Semua (public)        |
| 2  | POST   | `/api/auth/register`                | —                    | Semua (public)        |
| 3  | POST   | `/api/auth/login`                   | —                    | Semua (public)        |
| 4  | POST   | `/api/auth/refresh`                 | —                    | Semua (public)        |
| 5  | POST   | `/api/auth/logout`                  | Authenticated        | Semua (login)         |
| 6  | GET    | `/api/auth/me`                      | Authenticated        | Semua (login)         |
| 7  | GET    | `/api/users`                        | `user:read`          | admin, hr             |
| 8  | GET    | `/api/users/:id`                    | `user:read`          | admin, hr             |
| 9  | POST   | `/api/users`                        | `user:create`        | admin, hr             |
| 10 | PUT    | `/api/users/:id`                    | `user:update`        | admin, hr             |
| 11 | PATCH  | `/api/users/:id`                    | `user:update`        | admin, hr             |
| 12 | PATCH  | `/api/users/:id/role`               | `user:update`        | admin, hr             |
| 13 | GET    | `/api/jenis-cuti`                   | `jenis_cuti:read`    | admin, hr, employee   |
| 14 | GET    | `/api/jenis-cuti/:id`               | `jenis_cuti:read`    | admin, hr, employee   |
| 15 | POST   | `/api/jenis-cuti`                   | `jenis_cuti:create`  | admin, hr             |
| 16 | PUT    | `/api/jenis-cuti/:id`               | `jenis_cuti:update`  | admin, hr             |
| 17 | PATCH  | `/api/jenis-cuti/:id`               | `jenis_cuti:update`  | admin, hr             |
| 18 | DELETE | `/api/jenis-cuti/:id`               | `jenis_cuti:delete`  | admin, hr             |
| 19 | POST   | `/api/pengajuan-cuti`               | `cuti:create`        | employee              |
| 20 | GET    | `/api/pengajuan-cuti`               | `cuti:read`          | admin, hr, employee   |
| 21 | GET    | `/api/pengajuan-cuti/:id`           | `cuti:read`          | admin, hr, employee   |
| 22 | PATCH  | `/api/pengajuan-cuti/:id/status`    | `cuti:approve`       | admin, hr             |
| 23 | GET    | `/api/pengajuan-cuti/:id/riwayat`   | `riwayat:read`       | admin, hr, employee   |

---

### Autentikasi

Endpoint yang memerlukan autentikasi harus menyertakan header:

```
Authorization: Bearer <access_token>
```

Access token didapatkan dari endpoint login (`POST /api/auth/login`) dan berlaku selama **15 menit**. Gunakan refresh token untuk memperbarui access token melalui endpoint `POST /api/auth/refresh`. Refresh token berlaku selama **7 hari** dan menggunakan mekanisme **rotation** (setiap refresh menghasilkan refresh token baru, token lama di-revoke).

---

### RBAC (Role-Based Access Control)

Sistem menggunakan 3 role dengan permission berikut:

| Permission          | Admin | HR | Employee |
|---------------------|-------|----|----------|
| `user:create`       | ✅    | ✅ | ❌       |
| `user:read`         | ✅    | ✅ | ❌       |
| `user:update`       | ✅    | ✅ | ❌       |
| `user:delete`       | ✅    | ❌ | ❌       |
| `jenis_cuti:create` | ✅    | ✅ | ❌       |
| `jenis_cuti:read`   | ✅    | ✅ | ✅       |
| `jenis_cuti:update` | ✅    | ✅ | ❌       |
| `jenis_cuti:delete` | ✅    | ✅ | ❌       |
| `cuti:create`       | ✅    | ❌ | ✅       |
| `cuti:read`         | ✅    | ✅ | ✅       |
| `cuti:approve`      | ✅    | ✅ | ❌       |
| `riwayat:read`      | ✅    | ✅ | ✅       |
| `laporan:read`      | ✅    | ✅ | ❌       |
| `profil:update`     | ✅    | ✅ | ✅       |

---

### Struktur Database

```
users
├── id (BIGSERIAL PK)
├── username (VARCHAR 50, UNIQUE)
├── email (VARCHAR 100, UNIQUE)
├── password (VARCHAR 255, bcrypt hash)
├── nama_lengkap (VARCHAR 100)
├── role (VARCHAR 50)
├── role_id (BIGINT FK → roles)
├── is_active (BOOLEAN)
├── created_at (TIMESTAMPTZ)
└── updated_at (TIMESTAMPTZ)

roles
├── id (BIGSERIAL PK)
├── name (VARCHAR 50, UNIQUE)
├── description (VARCHAR 255)
└── created_at (TIMESTAMPTZ)

permissions
├── id (BIGSERIAL PK)
├── name (VARCHAR 100, UNIQUE)
├── description (VARCHAR 255)
└── created_at (TIMESTAMPTZ)

role_permissions
├── role_id (BIGINT FK → roles)
└── permission_id (BIGINT FK → permissions)

refresh_tokens
├── id (BIGSERIAL PK)
├── user_id (INTEGER FK → users)
├── token_hash (TEXT, UNIQUE)
├── expires_at (TIMESTAMPTZ)
├── revoked_at (TIMESTAMPTZ)
└── created_at (TIMESTAMPTZ)

jenis_cuti
├── id (SERIAL PK)
├── nama (VARCHAR 100, UNIQUE)
├── deskripsi (TEXT)
├── kuota_hari (INTEGER)
├── created_at (TIMESTAMPTZ)
└── updated_at (TIMESTAMPTZ)

pengajuan_cuti
├── id (SERIAL PK)
├── user_id (INTEGER FK → users)
├── jenis_cuti_id (INTEGER FK → jenis_cuti)
├── tanggal_mulai (DATE)
├── tanggal_selesai (DATE)
├── alasan (TEXT)
├── status (VARCHAR 20, default 'pending')
├── created_at (TIMESTAMPTZ)
└── updated_at (TIMESTAMPTZ)

riwayat_status
├── id (SERIAL PK)
├── pengajuan_cuti_id (INTEGER FK → pengajuan_cuti)
├── status_sebelum (VARCHAR 20, nullable)
├── status_sesudah (VARCHAR 20)
├── catatan (TEXT)
├── diubah_oleh (INTEGER FK → users)
└── created_at (TIMESTAMPTZ)
```