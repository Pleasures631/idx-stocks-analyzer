package handlers

import (
	"net/http"

	"indonesia-stocks-api/internal/repositories"

	"github.com/gin-gonic/gin"
)

func GetEducationArticles(c *gin.Context) {
	articles, err := repositories.GetEducationArticles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed fetch education articles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "total": len(articles), "data": articles})
}

func GetEducationArticle(c *gin.Context) {
	article, err := repositories.GetEducationArticle(c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "education article not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": article})
}
