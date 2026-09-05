-- ============================================================
-- Schema Market Maker
-- Mapping broker yang berperan sebagai market maker untuk saham tertentu.
-- -- ============================================================

CREATE TABLE IF NOT EXISTS m_market_maker (
  id          INT UNSIGNED NOT NULL AUTO_INCREMENT,
  stock_code  VARCHAR(20) NOT NULL,
  broker_code VARCHAR(20) NOT NULL,
  created_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_mm (stock_code, broker_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Contoh seed (isi manual sesuai data kalian):
-- INSERT INTO m_market_maker (stock_code, broker_code) VALUES ('CUAN', 'DX');
