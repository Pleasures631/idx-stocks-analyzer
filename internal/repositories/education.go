package repositories

import (
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"
)

func UpsertEducationArticles(articles []models.EducationArticle) error {
	if len(articles) == 0 {
		return nil
	}
	_, err := database.DB.NamedExec(`
		INSERT INTO m_education_articles
			(level_no, level_title, title, slug, source_url, content_html, sort_order, fetched_at)
		VALUES (:level_no, :level_title, :title, :slug, :source_url, :content_html, :sort_order, NOW())
		ON DUPLICATE KEY UPDATE
			level_no = VALUES(level_no), level_title = VALUES(level_title), title = VALUES(title),
			content_html = VALUES(content_html), sort_order = VALUES(sort_order), fetched_at = NOW()
	`, articles)
	return err
}

func GetEducationArticles() ([]models.EducationArticle, error) {
	var articles []models.EducationArticle
	err := database.DB.Select(&articles, `
		SELECT id, level_no, level_title, title, slug, source_url, '' AS content_html, sort_order, fetched_at
		FROM m_education_articles ORDER BY level_no, sort_order, id
	`)
	return articles, err
}

func GetEducationArticle(slug string) (*models.EducationArticle, error) {
	var article models.EducationArticle
	err := database.DB.Get(&article, `
		SELECT id, level_no, level_title, title, slug, source_url, content_html, sort_order, fetched_at
		FROM m_education_articles WHERE slug = ?
	`, slug)
	if err != nil {
		return nil, err
	}
	return &article, nil
}
