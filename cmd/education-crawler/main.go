package main

import (
	"fmt"
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/repositories"
	"indonesia-stocks-api/internal/services"
)

func main() {
	database.InitMySQL()
	defer database.DB.Close()
	articles, err := services.CrawlEducationArticlesWithTimeout()
	if err != nil {
		panic(err)
	}
	if err := repositories.UpsertEducationArticles(articles); err != nil {
		panic(err)
	}
	fmt.Printf("education_articles_synced=%d\n", len(articles))
}
