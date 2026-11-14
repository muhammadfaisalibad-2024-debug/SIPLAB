# Dokumentasi Tugas Kelompok SIPLAB — Bagian Git Master (Auth, RBAC, User, Route, Main)

Dokumen ini memuat:
1. **Kontrak Fungsi yang Dibutuhkan dari Anggota 2 & Anggota 3**
2. **Urutan Micro Commits Git**
3. **Alasan Keputusan Desain (Bahan Tanya Jawab Dosen)**
4. **Tabel Matriks Hak Akses (Admin, Laboran, Mahasiswa)**
5. **Daftar Kode Error Endpoint Auth & Users**

---

## 1. Kontrak yang Saya Butuhkan dari Anggota Lain

Agar seluruh kode pada repositori `siplab` dapat dikompilasi tanpa *error*, pastikan **Anggota 2** (`config/`, `database/`) dan **Anggota 3** (`helper/`, `middleware/`) menyediakan *signature* fungsi berikut:

### Dari Anggota 2 (`config/` dan `database/`)
```go
// package config (config/env.go)
func LoadEnv()
func GetEnv(key, fallback string) string
func GetEnvInt(key string, fallback int) int

// package config (config/logger.go)
func NewLogger() *slog.Logger

// package config (config/app.go)
func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App

// package database (database/postgres.go)
func NewPool(ctx context.Context) (*pgxpool.Pool, error)
```

### Dari Anggota 3 (`helper/` dan `middleware/`)
```go
// package helper (helper/errors.go)
type AppError struct {
    Status  int
    Code    string
    Message string
    Fields  map[string]string
    // cause error (unexported, hanya untuk log server)
}
func (e *AppError) Error() string
func (e *AppError) Unwrap() error
func BadRequest(message string) *AppError         // 400 BAD_REQUEST
func Unauthorized(message string) *AppError       // 401 UNAUTHORIZED
func Forbidden(message string) *AppError          // 403 FORBIDDEN
func NotFound(message string) *AppError           // 404 NOT_FOUND
func Conflict(message string) *AppError           // 409 CONFLICT
func Validation(fields map[string]string) *AppError // 422 VALIDATION_ERROR
func ServiceUnavailable(message string) *AppError // 503 SERVICE_UNAVAILABLE
func Internal(cause error) *AppError              // 500 INTERNAL_ERROR

// package helper (helper/response.go)
func Success(c *fiber.Ctx, status int, message string, data any) error
func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error
func Created(c *fiber.Ctx, message string, data any, location string) error
func NoContent(c *fiber.Ctx) error

// package helper (helper/request.go & helper/cursor.go & helper/validator.go)
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc)
func ParamID(c *fiber.Ctx) (int, bool)
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool)
func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error)
func EncodeCursor(createdAt time.Time, id int) string
func DecodeCursor(encoded string) (model.Cursor, error)
func ValidateStruct(s any) map[string]string

// package middleware (middleware/middleware.go, middleware/auth.go, middleware/authz.go)
func RequireJSON(c *fiber.Ctx) error
func LoginRateLimiter() fiber.Handler
func RequireAuth(jwt *helper.JWTManager) fiber.Handler
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler
```

---

## 2. Urutan Micro Commits Git (Sesuai Langkah Pengerjaan)

Jalankan perintah berikut secara bertahap agar riwayat Git terlihat rapi dan modular:

1. **Langkah 1 (Migration SQL):**
   `git add migrations/001_auth_rbac.sql`
   `git commit -m "feat(auth): add sql migration for users, refresh_tokens, and rbac schema with seed data"`
2. **Langkah 2 (Model & DTO Validasi):**
   `git add app/model/user.go app/model/auth.go`
   `git commit -m "feat(auth): add user and auth models with declarative validation tags"`
3. **Langkah 3 (Repository Layer):**
   `git add app/repository/errors.go app/repository/user_repository.go app/repository/token_repository.go app/repository/role_repository.go`
   `git commit -m "feat(auth): implement user, token, and role repositories with parameterized queries"`
4. **Langkah 4 (Helper Khusus Auth & RBAC):**
   `git add helper/password.go helper/jwt.go helper/authz.go`
   `git commit -m "feat(auth): add bcrypt hashing, jwt manager, and fail-closed permission set"`
5. **Langkah 5 (Service Layer & Unit Tests):**
   `git add app/service/auth_rules.go app/service/auth_rules_test.go app/service/auth_service.go app/service/user_service.go`
   `git commit -m "feat(auth): implement pure business rules, unit tests, auth service, and user service"`
6. **Langkah 6 (Routing & RBAC Guards):**
   `git add route/route.go`
   `git commit -m "feat(auth): register auth and user routes with permission middleware and team slots"`
7. **Langkah 7 (Main Application Wiring):**
   `git add main.go`
   `git commit -m "feat(auth): wire application dependencies and graceful shutdown in main.go"`

---

## 3. Alasan Keputusan Desain (Untuk Penjelasan ke Dosen)

1. **Pencegahan Mass Assignment pada Register (`RegisterRequest` tanpa field `Role`):**
   `RegisterRequest` sengaja tidak memiliki field `Role`, dan fungsi murni `BuildRegisteredUser` mengunci `Role: "mahasiswa"`. Jika memakai struct `User` secara langsung pada `c.BodyParser`, pengguna jahat cukup mengirim `{"role": "admin"}` saat registrasi untuk mengambil alih sistem.
2. **Batas `max=72` pada Tag Validasi Password:**
   Algoritma `bcrypt` hanya memproses 72 byte pertama dari input dan mengabaikan sisanya. Membatasi `max=72` pada tag validator mencegah ilusi keamanan di mana dua password berbeda di karakter ke-73 dianggap sama.
3. **Penggunaan `LEFT JOIN` + `COALESCE` pada `RoleRepository.LoadPermissions`:**
   Bila suatu saat ada role yang tidak memiliki permission eksplisit di `role_permissions` (hanya mengandalkan ownership), `INNER JOIN` akan menghilangkan role tersebut dari hasil query sehingga `IsKnownRole()` mengembalikan `false`. `LEFT JOIN` memastikan seluruh baris pada tabel `roles` tetap termuat ke dalam `PermissionSet`.
4. **Prinsip *Fail Closed* pada `PermissionSet.Can()`:**
   Setiap kondisi ragu (`p == nil`, *role* tidak ditemukan di map, atau *permission* salah ketik) selalu mengembalikan `false`. Jika salah ketik nama permission di `route.go`, akses langsung tertutup (`403 Forbidden`) dan segera ketahuan saat pengujian, bukan terbuka diam-diam (*fail open*).
5. **Pemisahan Pemeriksaan Akses: Middleware vs Service:**
   Keputusan yang tidak bergantung pada isi baris database (seperti `GET /users`, `PATCH /users/:id/role`, `DELETE /users/:id`) diletakkan di `route.go` memakai `middleware.RequirePermission` agar `route.go` terbaca sebagai peta hak akses. Sebaliknya, keputusan yang butuh tahu "data ini milik siapa" (`GET /users/:id`, `PATCH /users/:id`) diperiksa di `UserService` menggunakan `CanAccessUser` sebelum menyentuh database (mencegah *timing attack* enumerasi ID).
6. **Aturan Bisnis Anti-Self-Demotion & Anti-Self-Delete:**
   Admin dilarang mengubah role dirinya sendiri (`422 VALIDATION_ERROR` di `ValidateAssignRole`) dan dilarang menghapus akunnya sendiri (`403 FORBIDDEN` di `Delete`) agar sistem SIPLAB tidak pernah terkunci tanpa admin aktif.
7. **Keyset/Cursor Pagination dengan `(created_at, id) < ($1, $2)`:**
   Menggantikan `OFFSET` yang melambat secara linear dan rentan menduplikasi/melewatkan baris saat ada *insert* baru. Pasangan `(created_at DESC, id DESC)` menjamin urutan yang 100% deterministik karena `id` (Primary Key) menjadi pemecah seri saat dua baris memiliki `created_at` pada mikrodetik yang sama.

---

## 4. Tabel Matriks Hak Akses SIPLAB

| Endpoint / Aksi | Tanpa Token | Mahasiswa | Laboran | Admin | Penjaga (Middleware / Service Rule) |
| :--- | :---: | :---: | :---: | :---: | :--- |
| `POST /api/v1/auth/register` | 201 | 201 | 201 | 201 | Publik (Otomatis role `mahasiswa`) |
| `POST /api/v1/auth/login` | 200 | 200 | 200 | 200 | Publik + `LoginRateLimiter` |
| `POST /api/v1/auth/refresh` | 200 | 200 | 200 | 200 | Publik (Validasi hash refresh token) |
| `POST /api/v1/auth/logout` | 200 | 200 | 200 | 200 | Publik (Revoke hash refresh token) |
| `GET /api/v1/auth/me` | 401 | 200 | 200 | 200 | `RequireAuth` |
| `GET /api/v1/users` | 401 | 403 | 200 | 200 | `RequirePermission("user:list")` |
| `GET /api/v1/users/:id` (Diri Sendiri) | 401 | 200 | 200 | 200 | Service: `CanAccessUser` (Ownership) |
| `GET /api/v1/users/:id` (Orang Lain) | 401 | 403 | 200 | 200 | Service: `CanAccessUser("user:read:any")` |
| `PATCH /api/v1/users/:id` (Diri Sendiri) | 401 | 200 | 200 | 200 | Service: `CanAccessUser` (Ownership) |
| `PATCH /api/v1/users/:id` (Orang Lain) | 401 | 403 | 403 | 200 | Service: `CanAccessUser("user:update:any")` |
| `PATCH /api/v1/users/:id/role` (Orang Lain) | 401 | 403 | 403 | 200 | `RequirePermission("role:assign")` |
| `PATCH /api/v1/users/:id/role` (Diri Sendiri) | 401 | 403 | 403 | 422 | Middleware lolos, ditolak `ValidateAssignRole` |
| `DELETE /api/v1/users/:id` (Orang Lain) | 401 | 403 | 403 | 204 | `RequirePermission("user:delete")` |
| `DELETE /api/v1/users/:id` (Diri Sendiri) | 401 | 403 | 403 | 403 | Ditolak di Service (`current.UserID == id`) |
| `GET /api/v1/rooms` & `GET /api/v1/tools` | 401 | 200 | 200 | 200 | `RequireAuth` |
| `POST / PUT /rooms` & `POST / PUT /tools` | 401 | 403 | 201 / 200 | 201 / 200 | `RequirePermission("room:create/update")` |
| `DELETE /api/v1/rooms/:id` & `/tools/:id` | 401 | 403 | 403 | 204 | `RequirePermission("room:delete / tool:delete")` |
| `POST /api/v1/bookings` | 401 | 201 | 403 | 201 | `RequirePermission("booking:create")` |
| `PATCH /api/v1/bookings/:id/verify` | 401 | 403 | 200 | 200 | `RequirePermission("booking:verify")` |

---

## 5. Daftar Kode Error Endpoint Auth & Users

| Endpoint | HTTP Status | Error Code (`code`) | Situasi Pemicu |
| :--- | :---: | :--- | :--- |
| Semua Endpoint `POST`/`PATCH` | 415 | `UNSUPPORTED_MEDIA_TYPE` | Header `Content-Type` bukan `application/json` |
| Semua Endpoint `POST`/`PATCH` | 400 | `BAD_REQUEST` | Body JSON rusak / tidak dapat di-parse |
| `POST /api/v1/auth/register` | 422 | `VALIDATION_ERROR` | Format email salah, username tidak valid, atau password lemah |
| `POST /api/v1/auth/register` | 409 | `CONFLICT` | Username atau email sudah terdaftar di database (`SQLSTATE 23505`) |
| `POST /api/v1/auth/login` | 422 | `VALIDATION_ERROR` | Username atau password kosong |
| `POST /api/v1/auth/login` | 401 | `UNAUTHORIZED` | Username tidak ditemukan atau password salah |
| `POST /api/v1/auth/login` | 403 | `FORBIDDEN` | Kredensial benar tetapi `is_active = false` |
| `POST /api/v1/auth/login` | 429 | `TOO_MANY_REQUESTS` | Melebihi batas percobaan login (`LoginRateLimiter`) |
| `POST /api/v1/auth/refresh` | 422 | `VALIDATION_ERROR` | Field `refresh_token` kosong |
| `POST /api/v1/auth/refresh` | 401 | `UNAUTHORIZED` | Refresh token tidak dikenal, sudah kadaluarsa, atau sudah di-revoke |
| `POST /api/v1/auth/refresh` | 403 | `FORBIDDEN` | Akun pemilik token telah dinonaktifkan |
| `POST /api/v1/auth/logout` | 401 | `UNAUTHORIZED` | Refresh token tidak ditemukan atau sudah dicabut sebelumnya |
| `GET /api/v1/auth/me` | 401 | `UNAUTHORIZED` | Header `Authorization: Bearer <token>` tidak ada, rusak, atau expired |
| `GET /api/v1/users` | 400 | `BAD_REQUEST` | Query string `cursor` rusak (bukan base64 valid) atau `is_active` bukan boolean |
| `GET /api/v1/users` | 403 | `FORBIDDEN` | Role pemanggil tidak memiliki permission `user:list` |
| `GET / PATCH /users/:id` | 400 | `BAD_REQUEST` | Parameter `:id` bukan angka bulat positif |
| `GET / PATCH /users/:id` | 403 | `FORBIDDEN` | Mengakses/mengubah ID milik orang lain tanpa permission `:any` |
| `GET / PATCH /users/:id` | 404 | `NOT_FOUND` | ID user tidak ditemukan di database |
| `PATCH /api/v1/users/:id` | 400 | `BAD_REQUEST` | Body PATCH kosong (`{}`) tanpa satu pun field yang diubah |
| `PATCH /api/v1/users/:id/role` | 403 | `FORBIDDEN` | Role pemanggil tidak memiliki permission `role:assign` |
| `PATCH /api/v1/users/:id/role` | 422 | `VALIDATION_ERROR` | Role tidak dikenal atau admin mencoba mengubah role dirinya sendiri |
| `DELETE /api/v1/users/:id` | 403 | `FORBIDDEN` | Tidak punya `user:delete` atau admin mencoba menghapus akunnya sendiri |
| Semua Endpoint | 500 | `INTERNAL_ERROR` | Kegagalan tak terduga di sisi database/server (pesan asli hanya dicatat di log) |
