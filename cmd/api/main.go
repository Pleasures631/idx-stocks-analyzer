package main

import (
	"log"

	"indonesia-stocks-api/internal/bot"
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load file .env untuk mengambil TELEGRAM_BOT_TOKEN
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ File .env tidak ditemukan, menggunakan environment OS")
	}

	// 2. Inisialisasi Database
	database.InitMySQL()

	// 3. Jalankan Telegram Bot di Background (Goroutine)
	// database.DB diambil dari instance DB yang diexport dari package database
	tgBot, err := bot.NewTelegramBot(database.DB)
	if err != nil {
		log.Printf("⚠️ Telegram Bot gagal diinisialisasi: %v", err)
	} else {
		go tgBot.StartBot()
	}

	// 4. Inisialisasi & Jalankan Gin Server
	r := gin.Default()
	routes.RegisterRoutes(r)

	log.Println("🚀 Server running on :8080")
	r.Run(":8080")
}
