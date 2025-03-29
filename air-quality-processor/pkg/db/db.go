package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(dbURL string) *sql.DB {
	var err error
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to TimescaleDB: %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS air_quality (
		id SERIAL PRIMARY KEY,
		latitude DOUBLE PRECISION,
		longitude DOUBLE PRECISION,
		parameter TEXT,
		value DOUBLE PRECISION,
		timestamp TIMESTAMPTZ DEFAULT now()
	);`)
	if err != nil {
		log.Fatalf("Failed to create table: %s", err)
	}

	return db
}
