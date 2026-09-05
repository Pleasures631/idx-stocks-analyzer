# Changelog

## [2026-09-05] Anomaly Detection Z-Score & Ticker Detail

- Analisis broker flow sekarang menampilkan deteksi **ANOMALI Z-Score** di endpoint analisis — setiap broker yang beraktivitas di luar normal diflag dengan nilai `z_score` dan status `is_market_maker`. Sebelumnya data ini hanya berupa data mentah tanpa pemrosesan anomali.
- Endpoint analisis broker flow kini mendukung filter tanggal (`from`/`to`) sehingga hasil akumulasi/distribusi bisa dipersempit ke rentang waktu yang diinginkan.
- Endpoint **detail saham** (`/stocks/:symbol`) — fitur baru untuk halaman ticker: mengembalikan chart harga harian (open/high/low/close/volume), volume agregat per broker, dan ringkasan buy/sell per broker dalam satu response lengkap.
- Opsi filter tanggal (`from`/`to`/`range`) pada endpoint detail saham sudah diverifikasi berfungsi memfilter data sesuai rentang yang diminta.
- Dokumentasi kontrak API (`openapi.yaml`) dilengkapi schema lengkap untuk response analisis broker flow (anomalies, akumulasi/distribusi broker) — WIP ini disimpan di file terpisah di luar repo.
