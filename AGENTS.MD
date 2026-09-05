# AGENTS.md

# Agent Directives & Coding Guidelines

## Context Overview

- **Project Name**: `indonesia-stocks-api`
- **Tech Stack**:
  - Go 1.25+
  - Gin Framework
  - MySQL (`sqlx`)
  - Native HTTP Client
- **Architecture**:
  - `cmd/`
  - `routes/`
  - `handlers/`
  - `services/`
  - `repositories/`
  - `database/`

Agent harus mengikuti struktur dan pola yang sudah digunakan pada project. Hindari membuat pola baru apabila sudah tersedia implementasi yang serupa.

---

# Architecture Rules

## Layer Responsibility

### Handler Layer

Tanggung jawab:

- Validasi request
- Parsing query parameter
- Parsing request body
- Memanggil service
- Mapping response HTTP

Tidak boleh:

- Menulis query SQL
- Mengakses database secara langsung
- Menjalankan business logic kompleks

### Service Layer

Tanggung jawab:

- Business logic
- Orchestration antar repository
- Transaction management
- Data transformation yang berkaitan dengan business flow

Tidak boleh:

- Menangani HTTP request/response
- Menulis response JSON

### Repository Layer

Tanggung jawab:

- Seluruh akses database
- Query SQL
- Mapping data database ke struct

Tidak boleh:

- Menjalankan business logic
- Mengakses Gin Context
- Menghasilkan response HTTP

---

# Database Rules

## Query Placement

- Semua query SQL harus berada di `internal/repositories/`.
- Jangan menulis query SQL di handler.
- Jangan menulis query SQL di route.

## SQL Guidelines

- Gunakan parameter binding.
- Hindari string concatenation untuk SQL query.
- Hindari SQL injection risk.
- Gunakan `NamedExec`, `NamedQuery`, `Get`, `Select`, atau helper `sqlx` lainnya jika sesuai.
- Hindari `SELECT *`.
- Selalu tuliskan nama kolom secara eksplisit.

Contoh:

Good:

```go
query := `
SELECT
    stock_code,
    trade_date,
    close_price
FROM t_trading_summary
WHERE stock_code = ?
`
```

Bad:

```go
query := fmt.Sprintf(
    "SELECT * FROM t_trading_summary WHERE stock_code='%s'",
    stockCode,
)
```

---

# Database Transaction Rules

## Transaction Ownership

- Transaction harus dibuat pada Service Layer.
- Repository menerima `*sqlx.Tx` jika diperlukan.
- Repository tidak boleh melakukan `Commit()`.
- Repository tidak boleh melakukan `Rollback()`.
- Service bertanggung jawab terhadap atomicity.

Contoh flow:

```text
Handler
  ↓
Service
  ↓
Begin Transaction
  ↓
Repository Calls
  ↓
Commit / Rollback
```

---

# Context Rules

## Context Propagation

- Handler harus meneruskan context ke service.
- Service harus meneruskan context ke repository.
- Repository menggunakan method berbasis context.

Contoh:

```go
ctx := c.Request.Context()
```

Gunakan:

```go
db.GetContext(...)
db.SelectContext(...)
tx.ExecContext(...)
```

Jika tersedia.

---

# Error Handling Rules

- Jangan mengabaikan error.
- Selalu cek hasil error dari:
  - SQL query
  - Transaction
  - HTTP request
  - JSON parsing
  - Time parsing
  - External API

Bad:

```go
result, _ := repo.GetStocks()
```

Good:

```go
result, err := repo.GetStocks()
if err != nil {
    return err
}
```

---

# Logging Rules

- Jangan gunakan `fmt.Println()` untuk logging production.
- Gunakan logger yang sudah digunakan pada project.
- Log error penting sebelum dikembalikan.

Bad:

```go
fmt.Println(err)
```

---

# Interface Log Rules (External API Call)

Setiap kali membuat panggilan API eksternal (ke IDX, Exodus, atau API pihak ketiga lainnya), **WAJIB** mencatat log request & response ke tabel `t_interface_log`.

- Gunakan service yang sudah tersedia:

  ```text
  internal/services/interface_log_service.go  ->  LogInterfaceCall(...)
  ```

- Contoh pemakaian:

  ```go
  // function_name wajib diisi untuk merepresentasikan tujuan hit
  services.LogInterfaceCall("FetchExodusMarketDetector", url, string(bodyBytes), resp.StatusCode, err)
  ```

- Aturan:
  - Selalu panggil `LogInterfaceCall` setelah mendapat response (sukses maupun error/retry).
  - Parameter `function_name` wajib diisi — wakil dari tujuan hit (mis. `FetchIDX`, `FetchExodusMarketDetector`).
  - Simpan raw request & raw response body (bukan hasil decode) agar bisa di-audit.
  - Logging bersifat *best-effort*: kegagalan insert log tidak boleh menghentikan alur utama, cukup di-log ke console.
  - Jangan menduplikasi: modul baru cukup memanggil `LogInterfaceCall`, jangan membuat fungsi logging baru.

---

# API Response Rules

Gunakan format response yang konsisten mengikuti pola endpoint yang sudah ada.

Jika belum terdapat standar khusus, gunakan:

Success:

```json
{
  "success": true,
  "data": {}
}
```

Error:

```json
{
  "success": false,
  "message": "error message"
}
```

---

# Helper Usage Rules

Gunakan helper yang sudah tersedia di:

```text
internal/helpers/
```

Contoh:

- Number formatting
- Date formatting
- Utility functions

Jangan menduplikasi helper yang sudah ada.

---

# Model Rules

- Gunakan model yang sudah tersedia jika sesuai.
- Jangan membuat struct duplikat dengan tujuan yang sama.
- Tambahkan model baru hanya jika memang diperlukan oleh fitur baru.

---

# Financial & Backtest Rules

Untuk seluruh fitur screening, analytics, dan backtest:

- Hindari lookahead bias.
- Gunakan data historis yang benar-benar tersedia pada tanggal simulasi.
- Pisahkan signal date dan execution date.
- Entry dilakukan pada trading day berikutnya.
- Jika TP dan SL tersentuh pada hari yang sama, prioritaskan SL.
- Jangan menggunakan data masa depan untuk menghasilkan signal masa lalu.

---

# File Protection Rules

Jangan mengubah atau menghapus file berikut kecuali diminta secara eksplisit:

```text
server.exe
tmp/
air_run.log
server.log
repomix-output.xml
```

---

# Environment Safety

- Jangan memasukkan credential database ke source code.
- Jangan membuat hardcoded username/password.
- Gunakan konfigurasi dari `.env`.
- Jangan mengubah file `.env`.

---

# Database Schema Safety

- Jangan menghapus tabel existing.
- Jangan mengubah schema existing tanpa instruksi eksplisit.
- Jika membutuhkan perubahan schema, buat migration terpisah.

---

# Code Generation Preference

Saat membuat fitur baru:

1. Buat Model jika diperlukan.
2. Buat Repository terlebih dahulu.
3. Buat Service.
4. Buat Handler.
5. Daftarkan Route.
6. Pastikan code dapat di-compile.

Untuk fitur besar:

- Pisahkan logic ke file yang sesuai.
- Hindari file dengan tanggung jawab terlalu banyak.
- Ikuti struktur project yang sudah ada.

---

# Final Rule

Sebelum membuat implementasi baru:

1. Cari apakah sudah ada helper yang serupa.
2. Cari apakah sudah ada repository yang serupa.
3. Cari apakah sudah ada service yang serupa.
4. Ikuti pola existing project terlebih dahulu.
5. Hindari membuat abstraction baru tanpa kebutuhan yang jelas.