package repository

import (
	"api/internal/models"
	"database/sql"
	"fmt"
	"log"
)

type AirQualityRepository struct {
	Db *sql.DB
}

func NewAirQualityRepository(db *sql.DB) *AirQualityRepository {
	return &AirQualityRepository{
		Db: db,
	}
}
func (c *AirQualityRepository) SaveToDB(data models.AirQualityData) {
	_, err := c.Db.Exec(`INSERT INTO air_quality (latitude, longitude, parameter, value) VALUES ($1, $2, $3, $4)`,
		data.Latitude, data.Longitude, data.Parameter, data.Value)
	if err != nil {
		log.Printf("Failed to insert data: %s", err)
	}
}

func (c *AirQualityRepository) Get24HourDataForParameter(parameter string, latitude, longitude float64) ([]models.AirQualityData, error) {
	rows, err := c.Db.Query(`SELECT latitude, longitude, parameter, value FROM air_quality WHERE parameter = $1 AND latitude = $2 AND longitude = $3 AND timestamp >= NOW() - INTERVAL '24 hours'`, parameter, latitude, longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to query data: %w", err)
	}
	defer rows.Close()

	var results []models.AirQualityData
	for rows.Next() {
		var data models.AirQualityData
		if err := rows.Scan(&data.Latitude, &data.Longitude, &data.Parameter, &data.Value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, data)
	}

	return results, nil
}

func (c *AirQualityRepository) Get8HourDataForParameter(parameter string, latitude, longitude float64) ([]models.AirQualityData, error) {
	rows, err := c.Db.Query(`SELECT latitude, longitude, parameter, value FROM air_quality WHERE parameter = $1 AND latitude = $2 AND longitude = $3 AND timestamp >= NOW() - INTERVAL '8 hours'`, parameter, latitude, longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to query data: %w", err)
	}
	defer rows.Close()

	var results []models.AirQualityData
	for rows.Next() {
		var data models.AirQualityData
		if err := rows.Scan(&data.Latitude, &data.Longitude, &data.Parameter, &data.Value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, data)
	}

	return results, nil
}
