package database

import (
	"database/sql"
	"log"
	"os"

	"github.com/Habeebamoo/Prism/internal/configs"

	_ "github.com/lib/pq"
)

func Init(cfg *configs.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.PostgresUrl)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return db
}