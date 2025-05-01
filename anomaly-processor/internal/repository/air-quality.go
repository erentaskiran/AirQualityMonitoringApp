package repository

import (
	"database/sql"
)

type AirQualityRepository struct {
	Db *sql.DB
}

func NewAirQualityRepository(db *sql.DB) *AirQualityRepository {
	return &AirQualityRepository{
		Db: db,
	}
}
