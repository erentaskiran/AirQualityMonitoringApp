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
	_, err := c.Db.Exec(`
		INSERT INTO measurements (location, parameter, value)
		VALUES (ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
		        $3,
		        $4)
	`, data.Longitude, data.Latitude, data.Parameter, data.Value) // dikkat: lon, lat sırası!
	if err != nil {
		log.Printf("Failed to insert data: %v", err)
	}
}

func (c *AirQualityRepository) Get24HourDataForParameter(parameter string, latitude, longitude float64) ([]models.AirQualityData, error) {
	query := `
		SELECT
			ST_Y(location::geometry) AS latitude,
			ST_X(location::geometry) AS longitude,
			parameter,
			value
		FROM measurements
		WHERE parameter = $1
		  AND location = ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography
		  AND time >= NOW() - INTERVAL '24 hours'
	`
	rows, err := c.Db.Query(query, parameter, longitude, latitude) // lon, lat!!
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
	query := `
		SELECT
			ST_Y(location::geometry) AS latitude,
			ST_X(location::geometry) AS longitude,
			parameter,
			value
		FROM measurements
		WHERE parameter = $1
		  AND location = ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography
		  AND time >= NOW() - INTERVAL '8 hours'
	`
	rows, err := c.Db.Query(query, parameter, longitude, latitude)
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
