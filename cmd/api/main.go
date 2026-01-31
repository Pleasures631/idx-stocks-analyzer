package main

import (
	"indonesia-stocks-api/internal/routes"

	"indonesia-stocks-api/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitMySQL()

	r := gin.Default()
	routes.RegisterRoutes(r)
	r.Run(":8080")
}
