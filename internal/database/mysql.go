package database

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitMySQL() {
	dbUser := "root"
	dbPass := "Alz081897997!"
	dbHost := "localhost"
	dbPort := "3306"
	dbName := "selca_stocks_idx"

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
