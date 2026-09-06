CREATE TABLE IF NOT EXISTS m_education_articles (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  level_no     TINYINT UNSIGNED NOT NULL,
  level_title  VARCHAR(180) NOT NULL,
  title        VARCHAR(255) NOT NULL,
  slug         VARCHAR(255) NOT NULL,
  source_url   VARCHAR(500) NOT NULL,
  content_html LONGTEXT NOT NULL,
  sort_order   SMALLINT UNSIGNED NOT NULL DEFAULT 0,
  fetched_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_education_source_url (source_url(191)),
  UNIQUE KEY uq_education_slug (slug),
  KEY idx_education_level_order (level_no, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
