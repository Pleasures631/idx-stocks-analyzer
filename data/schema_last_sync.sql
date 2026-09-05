-- ============================================================
-- Schema Last Sync
-- Menyimpan tanggal terakhir sync untuk setiap jenis sync.
-- ============================================================

CREATE TABLE IF NOT EXISTS t_last_sync (
  id            INT UNSIGNED NOT NULL AUTO_INCREMENT,
  sync_type     VARCHAR(50)  NOT NULL, -- exodus_broker, trading_summary, idx_stocks, idx_broker
  last_sync_date DATE        NOT NULL, -- tanggal terakhir sync berhasil
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unique_sync_type (sync_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
