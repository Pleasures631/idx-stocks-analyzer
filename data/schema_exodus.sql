-- Table khusus data tarikan Exodus (Market Detector / Broker Summary)
-- Jangan dicampur dengan tabel lama.
CREATE TABLE IF NOT EXISTS t_exodus_broker_summary (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  stock_code VARCHAR(20) NOT NULL,
  trade_date DATE NOT NULL,
  broker_code VARCHAR(20) NOT NULL,
  side VARCHAR(10) NOT NULL, -- 'BUY' / 'SELL'
  broker_type VARCHAR(30) NOT NULL DEFAULT '', -- Pemerintah / Lokal / Asing
  lot DOUBLE NOT NULL DEFAULT 0,
  volume DOUBLE NOT NULL DEFAULT 0,
  value DOUBLE NOT NULL DEFAULT 0,
  turnover DOUBLE NOT NULL DEFAULT 0,
  avg_price DOUBLE NOT NULL DEFAULT 0,
  frequency BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_exodus_broker (stock_code, trade_date, broker_code, side, broker_type),
  KEY idx_exodus_stock_date (stock_code, trade_date),
  KEY idx_exodus_trade_date (trade_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabel derived: net value per broker per saham per hari (hasil agregasi t_exodus_broker_summary)
CREATE TABLE IF NOT EXISTS t_exodus_broker_net (
  stock_code VARCHAR(20) NOT NULL,
  trade_date DATE NOT NULL,
  broker_code VARCHAR(20) NOT NULL,
  broker_type VARCHAR(30) NOT NULL DEFAULT '',
  net_value DOUBLE NOT NULL DEFAULT 0,
  PRIMARY KEY (stock_code, trade_date, broker_code),
  KEY idx_net_trade_date (trade_date, stock_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabel derived: agregasi top-3 broker & foreign net per saham per hari (untuk backtest)
CREATE TABLE IF NOT EXISTS t_exodus_broker_agg (
  stock_code VARCHAR(20) NOT NULL,
  trade_date DATE NOT NULL,
  top3_net_buy DOUBLE NOT NULL DEFAULT 0,
  broker_count INT NOT NULL DEFAULT 0,
  top1_asing INT NOT NULL DEFAULT 0,
  foreign_net_buy DOUBLE NOT NULL DEFAULT 0,
  PRIMARY KEY (stock_code, trade_date),
  KEY idx_agg_trade_date (trade_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;