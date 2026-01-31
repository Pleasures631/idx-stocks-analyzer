package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

var DB *sqlx.DB

func InitMySQL() {

	err := godotenv.Load(filepath.Join("../../.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PWD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FJakarta",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal("failed connect db:", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	DB = db
	log.Println("✅ MySQL connected (sqlx)")
}
