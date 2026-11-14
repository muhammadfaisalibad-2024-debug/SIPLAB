# SIPLAB — Sistem Informasi Peminjaman Laboratorium

Backend REST API untuk sistem pengelolaan dan peminjaman laboratorium, ruangan, dan peralatan laboratorium berbasis Go (Fiber), PostgreSQL, JWT, dan RBAC (Role-Based Access Control).

---

## 👥 Pembagian Tugas Kelompok

- **Anggota 1 (Git Master / Auth & RBAC)**:
  - Database Migration & Seeding RBAC (`migrations/001_auth_rbac.sql`)
  - User & Auth Models & DTO Validation (`app/model/`)
  - Repository Layer dengan Parameterized Queries (`app/repository/`)
  - Helper Keamanan: Password Hashing (Bcrypt), JWT Manager, Permission Set (`helper/`)
  - Service Layer & Pure Business Rules Unit Testing (`app/service/`)
  - Central Route & RBAC Guards (`route/route.go`)
  - Application Wiring & Graceful Shutdown (`main.go`)
- **Anggota 2**:
  - Konfigurasi Environment & Logger (`config/env.go`, `config/logger.go`, `config/app.go`)
  - Database Connection Pool (`database/postgres.go`)
  - Pengelolaan Ruangan & Alat Laboratorium (`RoomService`, `ToolService`)
- **Anggota 3**:
  - Global Error Handling & Standard JSON Response (`helper/errors.go`, `helper/response.go`)
  - Request Context, Cursor Pagination & Validator Helper (`helper/request.go`, `helper/cursor.go`, `helper/validator.go`)
  - Middleware: Auth JWT, Authorization RBAC, Rate Limiter (`middleware/`)
  - Pengelolaan Peminjaman & Verifikasi (`BookingService`)

📖 **Detail Kontrak Fungsi Antar Anggota, Keputusan Desain, & Matriks Hak Akses:**
Lihat panduan lengkap pada [PANDUAN_LAPORAN_DAN_KONTRAK_TIM.md](./PANDUAN_LAPORAN_DAN_KONTRAK_TIM.md).

---

## 🛡️ Matriks Hak Akses (RBAC) Ringkas

| Resource / Endpoint | Mahasiswa | Laboran | Admin | Penjaga |
| :--- | :---: | :---: | :---: | :--- |
| `POST /api/v1/auth/register` | ✅ (201) | ✅ (201) | ✅ (201) | Publik (Otomatis role `mahasiswa`) |
| `POST /api/v1/auth/login` | ✅ (200) | ✅ (200) | ✅ (200) | Publik + Rate Limiter |
| `GET /api/v1/auth/me` | ✅ (200) | ✅ (200) | ✅ (200) | Bearer JWT Token |
| `GET /api/v1/users` | ❌ (403) | ✅ (200) | ✅ (200) | `user:list` |
| `GET / PATCH /users/:id` | Milik Sendiri | Milik Sendiri / Any | Milik Sendiri / Any | Ownership / `user:read/update:any` |
| `PATCH /api/v1/users/:id/role` | ❌ (403) | ❌ (403) | ✅ (200) | `role:assign` (Anti-Self-Demotion) |
| `DELETE /api/v1/users/:id` | ❌ (403) | ❌ (403) | ✅ (204) | `user:delete` (Anti-Self-Delete) |
| `GET /rooms` & `GET /tools` | ✅ (200) | ✅ (200) | ✅ (200) | Authenticated |
| `POST / PUT /rooms` & `/tools` | ❌ (403) | ✅ (201/200) | ✅ (201/200) | `room/tool:create/update` |
| `POST /api/v1/bookings` | ✅ (201) | ❌ (403) | ✅ (201) | `booking:create` |
| `PATCH /bookings/:id/verify` | ❌ (403) | ✅ (200) | ✅ (200) | `booking:verify` |

---

## 🚀 Menjalankan Proyek

1. **Clone repositori:**
   ```bash
   git clone https://github.com/muhammadfaisalibad-2024-debug/SIPLAB.git
   cd SIPLAB
   ```

2. **Jalankan migrasi database:**
   ```bash
   psql -U postgres -d siplab -f migrations/001_auth_rbac.sql
   ```

3. **Jalankan unit tests:**
   ```bash
   go test -v ./app/service
   ```
