package models

import "time"

type EducationArticle struct {
	ID          uint64    `db:"id" json:"id"`
	Level       int       `db:"level_no" json:"level"`
	LevelTitle  string    `db:"level_title" json:"level_title"`
	Title       string    `db:"title" json:"title"`
	Slug        string    `db:"slug" json:"slug"`
	SourceURL   string    `db:"source_url" json:"source_url"`
	ContentHTML string    `db:"content_html" json:"content_html"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	FetchedAt   time.Time `db:"fetched_at" json:"fetched_at"`
}
