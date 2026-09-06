CREATE TABLE IF NOT EXISTS t_stockbit_ihsg_chart (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  symbol VARCHAR(20) NOT NULL,
  trade_date DATE NOT NULL,
  `interval` VARCHAR(40) NOT NULL DEFAULT 'INTERVAL_CHART_MINUTELY',
  observed_at DATETIME NOT NULL,
  xlabel VARCHAR(20) NOT NULL DEFAULT '',
  value DECIMAL(18,4) NOT NULL,
  percentage DECIMAL(12,4) NOT NULL DEFAULT 0.0000,
  change_value DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_stockbit_ihsg_chart (symbol, `interval`, observed_at),
  KEY idx_stockbit_ihsg_chart_date (symbol, trade_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
