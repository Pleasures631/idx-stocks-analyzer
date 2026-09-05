# SUMMARY.md - Project `indonesia-stocks-api`

## Executive Overview

**Tujuan API:**
`indonesia-stocks-api` adalah backend RESTful API berbasis Go yang dirancang untuk melakukan sinkronisasi, pemrosesan, dan analisis data pasar saham Indonesia (IDX / Bursa Efek Indonesia). API ini mengumpulkan data mentah *broker summary*, *trading summary*, serta daftar emiten/broker langsung dari server IDX, menyimpannya ke database MySQL, dan menyediakan algoritma kuantitatif (screener & backtest) untuk analisis strategi perdagangan saham seperti **Top Accumulation (Foreign Flow & Smart Money)**, **Silent Accumulation**, **Scalping/Swinger**, dan **Single Stock Technical Statistics (Volume Price Analysis)**. Proyek ini juga dilengkapi **Backtest Strategy Engine V1** untuk menguji strategi bandarmologi secara historis, serta **Telegram Bot** untuk menanyakan analisis broker flow langsung dari Telegram.

**Tech Stack Utama:**
* **Language:** Go (Golang 1.25+)
* **Web Framework:** Gin Gonic (`github.com/gin-gonic/gin`)
* **Database & Driver:** MySQL dengan SQL extension `sqlx` (`github.com/jmoiron/sqlx` & `github.com/go-sql-driver/mysql`)
* **Environment Management:** `godotenv` (`github.com/joho/godotenv`)
* **HTTP Client:** Native `net/http` dengan mekanisme *retry logic* & kustomisasi headers (User-Agent, Referer) untuk scraping IDX endpoint.

---

## Directory Structure

Tree folder proyek (maksimal Level 3) beserta penjelasan fungsi tiap komponen:

```text
indonesia-stocks-api/
├── cmd/
│   └── api/
│       └── main.go              # Entrypoint utama aplikasi (Inisialisasi DB, Bot & Server Gin)
├── data/                        # Skema/DDL database (migration)
│   ├── schema_exodus.sql        # Tabel t_exodus_broker_summary (Market Detector / Broker Summary)
│   └── schema_backtest.sql      # Tabel t_backtest_run & t_backtest_detail
├── internal/
│   ├── bot/                     # Telegram Bot (long-polling listener & command handler)
│   │   └── telegram.go
│   ├── config/                  # Konfigurasi aplikasi
│   │   └── config.go
│   ├── constants/               # Konstanta URL Endpoint IDX, Service Name, & Headers
│   │   └── endpoint.go
│   ├── database/                # Koneksi & konfigurasi Database MySQL (sqlx setup)
│   │   └── mysql.go
│   ├── handlers/                # HTTP Controllers / Request Handlers (Gin)
│   │   ├── backtest.go
│   │   ├── broker_summary.go
│   │   ├── broker_summary_exodus.go
│   │   ├── data_mapper.go
│   │   ├── exodus_broker_summary.go
│   │   ├── health.go
│   │   ├── sync_idx_data.go
│   │   └── trading_summary.go
│   ├── helpers/                 # Function utility (Date generator, number formatting, math round)
│   │   └── utility.go
│   ├── models/                  # Struct definitions untuk Request, Response, DB Tables, & Analytics
│   │   ├── backtest.go
│   │   ├── broker_summary.go
│   │   ├── exodus.go
│   │   ├── sync_idx_data.go
│   │   └── trading_summary.go
│   ├── repositories/            # Layer akses data ke MySQL (Raw/Complex SQL queries & Upserts)
│   │   ├── backtest.go
│   │   ├── broker.go
│   │   ├── exodus.go
│   │   └── stocks.go
│   ├── routes/                  # pendaftaran endpoint & HTTP routing
│   │   └── routes.go
│   └── services/                # Business logic, integrasi API eksternal, & simulasi backtest
│       ├── backtest_service.go
│       ├── exodus_flow_service.go
│       ├── exodus_services.go
│       └── idx_services.go
├── .gitignore                   # Abaikan file .env dan binary
├── AGENTS.MD                    # Agent directives & coding guidelines
├── SUMMARY.md                   # Dokumentasi proyek ini
├── go.mod                       # Dependensi Go module
└── server.exe                   # Executable file (windows build)
```

### Penjelasan Fungsi Folder Utama:
* **`cmd/api/`**: Tempat file `main.go` yang bertugas menjalankan fungsi `InitMySQL()`, menginisialisasi Telegram Bot (goroutine), mendaftarkan rute API, dan menyalakan server di port `:8080`.
* **`data/`**: Berisi file DDL/migration MySQL, yaitu `schema_exodus.sql` (tabel `t_exodus_broker_summary`) dan `schema_backtest.sql` (tabel `t_backtest_run` & `t_backtest_detail`).
* **`internal/bot/`**: Telegram Bot menggunakan `go-telegram-bot-api` (long-polling). Memproses perintah slash command dan teks bebas, misal `/analyze CUAN` untuk analisis broker flow.
* **`internal/constants/`**: Menyimpan base URL (`https://www.idx.co.id/primary`), nama modul/service IDX, dan Referrer headers untuk menghindari blokir request.
* **`internal/database/`**: Menangani koneksi ke MySQL menggunakan DSN dari `.env`, mengatur connection pool (`SetMaxOpenConns`, `SetMaxIdleConns`), dan mengekspor instance `database.DB`.
* **`internal/handlers/`**: Mengolah payload HTTP request, memanggil service atau layer repository, memetakan DTO (*data mapper*), dan mengembalikan response JSON.
* **`internal/models/`**: Definisi struktur data Go (`struct`) yang mencakup response JSON dari IDX, schema tabel MySQL (`t_trading_summary`, `t_exodus_broker_summary`, `m_list_stocks`, `m_list_broker`, `t_backtest_run`, `t_backtest_detail`), serta DTO hasil analisa (*TopAccumulation*, *TopSwinger*, *SilentAccumulation*, *BacktestResult*, *BacktestRun*).
* **`internal/repositories/`**: Berisi kueri SQL kompleks (SQL Window Functions, CTE, Lag, Moving Average) untuk menyimpan data (*upsert*) dan menjalankan skrining strategi teknikal & bandarmologi.
* **`internal/services/`**: Generic HTTP fetching client dengan support Generic (`[T any]`) dan *exponential backoff retry logic* untuk server IDX, serta business logic simulasi backtest (entry/exit/TP/SL) dan analisis broker flow.

---

## Data & Request Flow

Alur eksekusi data dari HTTP Request hingga ke Database / External API:

```text
[ Client / Postman ]                    [ Telegram Bot ]
       │                                       │
       ▼                                       ▼
[ routes/routes.go ]  ── (Mencocokkan HTTP Method & Endpoint)
       │
       ▼
[ handlers/* ]        ── (Validasi Param/JSON Query, Memanggil Service/Repo, Mapping Data)
       │
       ├─────────────────────────────────────────┐
       ▼                                         ▼
[ services/* ]                       [ repositories/* ]
 (Fetch HTTP ke IDX, simulasi)          (Menjalankan Query SQL via DB)
       │                                         │
       ▼                                         ▼
[ IDX Base Server ]                     [ MySQL Database ]
 (https://www.idx.co.id)                 (t_trading_summary, t_exodus_broker_summary,
                                          t_backtest_run, t_backtest_detail, m_list_*)
```

### Penjelasan Detail Alur:
1. **Routing:** Router Gin (`routes.RegisterRoutes`) menerima HTTP request dan mengarahkannya ke fungsi handler yang sesuai.
2. **Handlers & Data Mapping:**
   * Jika request bersifat **Sync / Fetch Data**: Handler memanggil `services.FetchIDX[T]()` untuk mengambil JSON mentah dari IDX, lalu melewatkannya ke transformer `data_mapper.go` untuk diubah menjadi struct DB. Data yang sudah rapi lalu dikirim ke layer Repository.
   * Jika request bersifat **Analytics / Screener**: Handler memanggil fungsi di `repositories/*` dengan menyertakan filter parameter (seperti `days`, `date`, `stock_code`).
   * Jika request bersifat **Backtest**: Handler menerima konfigurasi (TP%, SL%, holding days), lalu `services.RunBacktestV1()` menjalankan simulasi dan menyimpan hasilnya secara atomik ke `t_backtest_run` & `t_backtest_detail`.
   * **Telegram Bot**: Command `/analyze <TICKER>` memanggil `services.AnalyzeExodusBrokerFlowService()` — logic yang sama dengan endpoint `analyze-flow` — lalu memformat hasilnya menjadi pesan Markdown.
3. **Repository & MySQL:** Executed query SQL di `repositories/*` menggunakan fitur `sqlx`. Query ini melakukan agregasi berat (Moving Average 20/50, Volume Spike Ratio, Net Foreign Flow, Breakout Resistance, RSI 14, Top-3 Broker Net Buy) langsung di dalam engine database.
4. **Formatting & Response:** Handler menerima struct hasil kueri, memformat tampilan angka (seperti konversi angka besar ke string `1.5 B` / `2.3 T` via `helpers.FormatBigNumber`), lalu mengembalikannya sebagai JSON `HTTP 200 OK`.

---

## API Endpoints & Models Overview

Berikut adalah daftar endpoint utama beserta deskripsi dan handler yang menanganinya:

| HTTP Method | Endpoint | Handler | Deskripsi / Fungsi |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | `HealthCheck` | Pengecekan status server API. |
| `GET` | `/idx/brokersummary` | `FetchBrokerSummary` | Fetch data broker summary mentah langsung dari API IDX berdasarkan parameter `?date=YYYYMMDD`. |
| `GET` | `/idx/brokersummary/analyze` | `AnalyzeBrokerSummary` | Fetch data broker summary secara multi-day (maksimal range 7 hari). |
| `POST` | `/idx/syncbroker` | `SyncBrokerFromIDX` | Sinkronisasi & *upsert* daftar kode broker seluruh anggota bursa dari IDX ke tabel `m_list_broker`. |
| `POST` | `/idx/syncstocks` | `SyncStocksFromIDX` | Sinkronisasi & *upsert* daftar seluruh emiten saham aktif dari IDX ke tabel `m_list_stocks`. |
| `POST` | `/tradingsummary/insert` | `InsertTradingSummary` | Download data ringkasan perdagangan harian (*Trading Summary*) dari IDX berdasarkan tanggal/range dan simpan ke DB `t_trading_summary`. |
| `GET` | `/analyze/single-stocks` | `StatisticSingleStock` | Analisis mendalam 1 saham (`?stock_code=BBCA`), mengevaluasi aksi harga & volume (VPA / Volume Price Analysis: *SOS, SOW, Volume Trap, Markup*). |
| `GET` | `/analyze/top-accumulation` | `GetTopAccumulation` | Screener saham dengan akumulasi asing (Foreign Flow) & Close Strength tinggi dalam 7 hari terakhir. |
| `GET` | `/analyze/top-accumulation-eod` | `GetTopAccumulationEod` | Screener End-Of-Day (EOD) canggih 60 hari: deteksi Smart Money, Tren Bullish (MA20/MA50), RSI, Breakout Score, & rekomendasi Entry/SL. |
| `GET` | `/analyze/silent-accumulation` | `GetSilentAccumulation` | Screener deteksi akumulasi senyap (*Whale/Institutional accumulation*) pada saham sepi dengan partisipasi lokal/retail rendah (< 50%). |
| `GET` | `/analyze/top-scalping-daily` | `GetTopScalping` | Screener khusus calon saham swing/scalping berdasarkan lonjakan volume (*Vol Multiplier*) & Swing Score. |
| `GET` | `/backtest/top-accumulation-eod` | `RunBacktestEOD` | Fitur *Backtesting* harian EOD (`?date=YYYY-MM-DD`) untuk menguji Win Rate & Avg Profit sinyal di masa lalu. |
| `GET` | `/exodus/broker-summary` | `GetExodusBrokerSummary` | Ambil data broker summary Exodus dari DB (`?symbol`, `?from`, `?to`). |
| `POST` | `/exodus/broker-summary/fetch` | `FetchExodusBrokerSummary` | Tarik data broker summary Exodus (Market Detector) per saham & rentang tanggal, simpan ke `t_exodus_broker_summary`. |
| `POST` | `/exodus/broker-summary/bulkfetch` | `FetchExodusBrokerSummaryAll` | Bulk fetch broker summary Exodus untuk seluruh saham aktif. **Me-skip** `(stock_code, trade_date)` yang sudah ada di `t_exodus_broker_summary` supaya tidak hit Exodus ulang (hemat rate limit). |
| `GET` | `/exodus/broker-summary/analyze-flow` | `AnalyzeExodusBrokerFlow` | Analisis aliran dana broker (`?symbol`, `?from`, `?to`): fase Akumulasi/Distribusi, Net Foreign, Top-3 akumulasi & distribusi broker. |
| `POST` | `/api/v1/backtest/run` | `RunBacktestV1` | **Backtest Strategy Engine V1 (Bandarmologi)** — deteksi sinyal (broker accumulation + volume spike + bullish candle), simulasi entry/exit (TP/SL/max holding), lalu persist hasil ke `t_backtest_run` & `t_backtest_detail`. |

**Telegram Bot** (bukan endpoint HTTP, berjalan sebagai goroutine di dalam server):

| Command | Contoh | Deskripsi |
| :--- | :--- | :--- |
| `/analyze <TICKER>` | `/analyze CUAN` | Analisis broker flow 1 saham (default 7 hari terakhir). Alias: `/analyze_flow`, `/analyzeflow`, `/flow`, atau ketik teks `analyze CUAN`. |
| `/stock <TICKER>` | `/stock BBCA` | Analisis ringkas 1 saham. |
| `/topacc` | `/topacc` | Screener Top Accumulation Foreign. |
| `/start` | `/start` | Daftar perintah yang tersedia. |

---

## Backtest Strategy Engine V1 (Bandarmologi)

Backtest engine untuk menguji strategi *bandarmologi* (ikut jejak broker) secara historis. Sinyal terdeteksi dari kombinasi data broker summary Exodus (`t_exodus_broker_summary`) dan trading summary IDX (`t_trading_summary`).

**Endpoint:** `POST /api/v1/backtest/run`

**Contoh Request:**
```json
{
  "run_name": "Aggressive Bandarmologi Strategy V1",
  "start_date": "2026-01-01",
  "end_date": "2026-08-17",
  "tp_percent": 5,
  "sl_percent": 2,
  "max_holding_days": 10
}
```

### Signal Rules (Signal Date = T)
1. **Broker Accumulation** — Hitung Net Buy per broker (`SUM(BUY value) - SUM(SELL value)`), ambil **Top 3 broker** net buy. Valid jika `top3_net_buy / trading_summary.value >= 0.25`.
2. **Volume Spike** — `SMA5(volume)` dihitung dari **T-5 s/d T-1**. Valid jika `volume_T > SMA5 * 1.5`.
3. **Bullish Candle** — `close_price > open_price`.
4. **No Overlapping Position** — jika masih ada posisi aktif untuk `stock_code` yang sama, sinyal berikutnya di-skip.

### Entry & Exit Rules
* **Signal Date (T)** terpisah dari **Entry Date** (hari trading berikutnya setelah T) untuk menghindari *lookahead bias*.
* **Entry Price** = `open_price` pada entry date.
* **TP** = `entry_price * (1 + tp_percent/100)` → **WIN**.
* **SL** = `entry_price * (1 - sl_percent/100)` → **LOSS**.
* Jika **TP & SL tersentuh di hari yang sama** → **LOSS** (prioritas SL).
* Jika **max_holding_days** tercapai tanpa TP/SL → **EXPIRED**, exit di `close_price` terakhir.

### Metrics yang Dihitung
`total_trades`, `win_trades`, `loss_trades`, `expired_trades`, `win_rate`, `profit_factor` (gross profit / gross loss), `expectancy` (rata-rata return per trade), `avg_holding_days`, `total_return_percent`, `max_drawdown` (penurunan terbesar kurva ekuitas kumulatif).

### Persistensi (Database Transaction)
* `t_backtest_run` — 1 baris per run, berisi konfigurasi & metrik agregat.
* `t_backtest_detail` — detail per trade (`stock_code`, `signal_date`, `entry_date`, `entry_price`, `target_tp`, `target_sl`, `exit_date`, `exit_price`, `exit_reason`, `status` WIN/LOSS/EXPIRED, `return_percent`).

Transaction dibuat di Service Layer (`services.RunBacktestV1`): `Beginx` → insert run → insert detail (bulk) → `Commit`, dengan `Rollback` sebagai fallback. Repository menerima `*sqlx.Tx` dan tidak pernah melakukan commit/rollback sendiri.

### Arsitektur Implementasi
```text
Handler (POST /api/v1/backtest/run)
  → Service RunBacktestV1 (validasi, simulasi trade, hitung metrics, ownership transaction)
      → Repository GetBacktestSignals (deteksi sinyal via CTE)
      → Repository GetBacktestDailyBars (OHLC harian)
      → Repository InsertBacktestRun / InsertBacktestDetails (via *sqlx.Tx)
```

---

## Database Tables

| Tabel | Sumber | Deskripsi |
| :--- | :--- | :--- |
| `t_trading_summary` | IDX Trading Summary | Data harga/volume harian seluruh saham (open, high, low, close, volume, value, foreign flow). |
| `t_exodus_broker_summary` | Exodus Market Detector | Data aktivitas beli/jual per broker per saham per hari (`side`, `broker_type`, `value`, `lot`). |
| `t_backtest_run` | Backtest Engine | Master setiap sesi run backtest + metrik agregat. |
| `t_backtest_detail` | Backtest Engine | Detail setiap trade dalam sebuah run. |
| `t_interface_log` | Interface Log | Log setiap panggilan API eksternal (`function_name`, `request`, `response`, `http_status`, `success`, `created_at`). |
| `m_list_stocks` | IDX | Daftar emiten saham aktif. |
| `m_list_broker` | IDX | Daftar anggota bursa (broker). |

---

## Interface Log (t_interface_log)

Mencatat **setiap panggilan API eksternal** (IDX, Exodus, dsb) untuk audit & debugging.

**Kolom:**
- `function_name` — wakil dari tujuan hit, misal `FetchIDX`, `FetchExodusMarketDetector`.
- `request` — URL / request body yang dikirim.
- `response` — raw response body.
- `http_status` — status code HTTP.
- `success` — true jika berhasil, false jika ada error.
- `error_message` — pesan error (jika ada).
- `created_at` — otomatis terisi (tanggal & jam).

**Cara pakai (reusable & dinamis):** Modul baru yang melakukan HTTP call cukup memanggil satu fungsi service:
```go
services.LogInterfaceCall("FetchExodusMarketDetector", url, string(body), resp.StatusCode, err)
```
Logging bersifat *best-effort* — kegagalan insert tidak menghentikan alur utama. Sudah terpasang otomatis di `FetchIDX` (idx_services.go) dan `FetchExodusMarketDetector` (exodus_services.go).

---

## Local Setup & How to Run

### Prerequisites
* **Go** versi 1.25 atau lebih baru
* **MySQL Server** (versi 8.0+ direkomendasikan)
* Git

### Step-by-Step Setup

1. **Clone & Navigasi Project:**
   ```bash
   git clone <repository_url>
   cd indonesia-stocks-api
   ```

2. **Konfigurasi Environment (`.env`):**
   Buat file `.env` di root folder proyek dan sesuaikan kredensial MySQL Anda:
   ```env
   DB_USER=root
   DB_PWD=your_password
   DB_HOST=127.0.0.1
   DB_PORT=3306
   DB_NAME=indonesia_stocks_db

   # Opsional — Telegram Bot (jika ingin mengaktifkan /analyze)
   TELEGRAM_BOT_TOKEN=123456:ABC...

   # Opsional — Exodus Market Detector (fetch & sync broker summary)
   EXODUS_TOKEN=your_exodus_bearer_token
   ```

3. **Siapkan Database:**
   Jalankan file skema SQL yang tersedia di folder `data/` (sesuaikan dengan kebutuhan):
   ```bash
   mysql -u root -p indonesia_stocks_db < data/schema_exodus.sql
   mysql -u root -p indonesia_stocks_db < data/schema_backtest.sql
   ```

4. **Install Dependensi:**
   ```bash
   go mod download
   ```

5. **Jalankan Aplikasi:**
   ```bash
   go run cmd/api/main.go
   ```
   Jika berhasil, terminal akan menampilkan log koneksi database, status bot Telegram, dan server Gin aktif di port `:8080`:
   ```text
   ✅ MySQL connected (sqlx)
   🤖 Telegram Bot <bot_username> aktif!
   [GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware.
   [GIN-debug] Running on :8080
   ```

6. **Pengujian Endpoint Pertama Kali:**
   * Cek kesehatan server:
     ```bash
     curl http://localhost:8080/health
     ```
   * Sinkronisasi data saham dari IDX:
     ```bash
     curl -X POST http://localhost:8080/idx/syncstocks
     ```
   * Jalankan backtest:
     ```bash
     curl -X POST http://localhost:8080/api/v1/backtest/run -H "Content-Type: application/json" -d "{\"run_name\":\"Test\",\"start_date\":\"2026-01-01\",\"end_date\":\"2026-08-17\",\"tp_percent\":5,\"sl_percent\":2,\"max_holding_days\":10}"
     ```
