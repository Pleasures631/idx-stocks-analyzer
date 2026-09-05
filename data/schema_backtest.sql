-- ============================================================
-- Schema Backtest Strategy Engine V1 (Bandarmologi)
-- Referensi schema tabel yang sudah terpasang di database.
-- ============================================================

-- Tabel master untuk 1 sesi run backtest
CREATE TABLE IF NOT EXISTS t_backtest_run (
  id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  run_name             VARCHAR(100)    NOT NULL,
  start_date           DATE            NOT NULL,
  end_date             DATE            NOT NULL,
  tp_percent           DECIMAL(5,2)    NOT NULL,
  sl_percent           DECIMAL(5,2)    NOT NULL,
  max_holding_days     INT             NOT NULL DEFAULT 10,
  total_signals        INT             NOT NULL DEFAULT 0,
  total_trades         INT             NOT NULL DEFAULT 0,
  gross_profit         DOUBLE          NOT NULL DEFAULT 0,
  gross_loss           DOUBLE          NOT NULL DEFAULT 0,
  win_trades           INT             NOT NULL DEFAULT 0,
  loss_trades          INT             NOT NULL DEFAULT 0,
  expired_trades       INT             NOT NULL DEFAULT 0,
  win_rate             DECIMAL(5,2)    NOT NULL DEFAULT 0.00,
  profit_factor        DECIMAL(8,2)    NOT NULL DEFAULT 0.00,
  expectancy           DOUBLE          NOT NULL DEFAULT 0,
  avg_holding_days     DECIMAL(5,2)    NOT NULL DEFAULT 0.00,
  total_return_percent DECIMAL(8,2)    NOT NULL DEFAULT 0.00,
  max_drawdown         DOUBLE          NOT NULL DEFAULT 0,
  created_at           TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Tabel detail setiap trade dalam 1 run backtest
CREATE TABLE IF NOT EXISTS t_backtest_detail (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  backtest_run_id  BIGINT UNSIGNED NOT NULL,
  stock_code       VARCHAR(20)     NOT NULL,
  signal_date      DATE            NOT NULL,
  entry_date       DATE            NOT NULL,
  entry_price      DECIMAL(18,2)   NOT NULL,
  target_tp        DECIMAL(18,2)   NOT NULL,
  target_sl        DECIMAL(18,2)   NOT NULL,
  exit_date        DATE            NOT NULL,
  exit_price       DECIMAL(18,2)   NOT NULL,
  exit_reason      VARCHAR(20)     NOT NULL,
  holding_days     INT             NOT NULL DEFAULT 0,
  status           VARCHAR(20)     NOT NULL,
  return_percent   DECIMAL(8,2)    NOT NULL DEFAULT 0.00,
  created_at       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_backtest_detail_run (backtest_run_id),
  KEY idx_backtest_detail_stock (stock_code),
  CONSTRAINT fk_backtest_detail_run FOREIGN KEY (backtest_run_id) REFERENCES t_backtest_run (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
