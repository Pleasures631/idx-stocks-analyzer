-- Migration: password hashes, rotating sessions, and password reset tokens.
ALTER TABLE m_users ADD COLUMN password_hash VARCHAR(255) NULL AFTER email;

CREATE TABLE IF NOT EXISTS m_user_sessions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  access_token_hash CHAR(64) NOT NULL,
  refresh_token_hash CHAR(64) NOT NULL,
  access_expires_at DATETIME NOT NULL,
  refresh_expires_at DATETIME NOT NULL,
  revoked_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_m_user_sessions_access (access_token_hash),
  UNIQUE KEY uq_m_user_sessions_refresh (refresh_token_hash),
  KEY ix_m_user_sessions_user (user_id),
  CONSTRAINT fk_m_user_sessions_user FOREIGN KEY (user_id) REFERENCES m_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS m_password_reset_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  token_hash CHAR(64) NOT NULL,
  expires_at DATETIME NOT NULL,
  used_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_m_password_reset_tokens_hash (token_hash),
  KEY ix_m_password_reset_tokens_user (user_id),
  CONSTRAINT fk_m_password_reset_tokens_user FOREIGN KEY (user_id) REFERENCES m_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
