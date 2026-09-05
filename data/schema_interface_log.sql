-- ============================================================
-- Schema Interface Log
-- Menyimpan log setiap panggilan API eksternal (IDX, Exodus, dll).
-- ============================================================

CREATE TABLE IF NOT EXISTS t_interface_log (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  function_name VARCHAR(100)    NOT NULL, -- wakil dari tujuan hit (FetchIDX, FetchExodusMarketDetector, ...)
  request       MEDIUMTEXT      NULL,     -- URL / request body yang dikirim
  response      MEDIUMTEXT      NULL,     -- raw response body
  http_status   INT             NOT NULL DEFAULT 0,
  success       TINYINT(1)      NOT NULL DEFAULT 0,
  error_message TEXT            NULL,
  created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_interface_log_fn (function_name),
  KEY idx_interface_log_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
