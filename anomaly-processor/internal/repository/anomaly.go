package repository

import (
	"api/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type AnomalyRepository struct {
	Db *sql.DB
}

func NewAnomalyRepository(db *sql.DB) *AnomalyRepository {
	return &AnomalyRepository{Db: db}
}

func (r *AnomalyRepository) SaveAnomalyToDB(message []byte) {
	query := `INSERT INTO anomalies (parameter, value, time, location, description) VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6)`

	var anomaly models.Anomaly

	err := json.Unmarshal(message, &anomaly)
	if err != nil {
		log.Println("Error parsing anomaly message:", err)
		return
	}

	// Validate and set default time if empty
	if anomaly.Time == "" {
		anomaly.Time = time.Now().UTC().Format(time.RFC3339)
	}

	// Validate and set default value if empty
	if anomaly.Value == 0 {
		log.Println("Invalid or empty value in anomaly message")
		return
	}

	_, err = r.Db.Exec(query, anomaly.Parameter, anomaly.Value, anomaly.Time, anomaly.Longitude, anomaly.Latitude, anomaly.Description)
	if err != nil {
		log.Println("Error saving anomaly to DB:", err)
	}
}

func (r *AnomalyRepository) GetRecentAnomalies() ([]map[string]interface{}, error) {
	query := `SELECT parameter, value, time, 
			ST_X(location::geometry) AS longitude, 
			ST_Y(location::geometry) AS latitude, 
			description 
			FROM anomalies WHERE time >= NOW() - INTERVAL '2 hours'`

	rows, err := r.Db.Query(query)
	if err != nil {
		log.Println("Error querying anomalies:", err)
		return nil, err
	}
	defer rows.Close()

	var anomalies []map[string]interface{}

	for rows.Next() {
		var anomaly = make(map[string]interface{})
		var longitude, latitude float64
		var parameter, value, time, description interface{}
		if err := rows.Scan(&parameter, &value, &time, &longitude, &latitude, &description); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		anomaly["parameter"] = parameter
		anomaly["value"] = value
		anomaly["time"] = time
		anomaly["description"] = description
		anomaly["longitude"] = longitude
		anomaly["latitude"] = latitude
		anomalies = append(anomalies, anomaly)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		return nil, err
	}

	return anomalies, nil
}

func (r *AnomalyRepository) GetAnomaliesByLocation(latitude, longitude, radius float64) ([]models.Anomaly, error) {
	query := `
		SELECT parameter, value, time,
			   ST_X(location::geometry) AS longitude,
			   ST_Y(location::geometry) AS latitude,
			   description
		FROM anomalies
		WHERE ST_DWithin(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3 * 1000) -- radius in meters
		ORDER BY time DESC;` // Order by time, newest first

	rows, err := r.Db.Query(query, longitude, latitude, radius) // lon, lat, radius (km)
	if err != nil {
		log.Printf("Error querying anomalies by location: %v", err)
		return nil, err
	}
	defer rows.Close()

	var anomalies []models.Anomaly
	for rows.Next() {
		var anomaly models.Anomaly
		var anomalyTime time.Time // Use time.Time for scanning
		if err := rows.Scan(&anomaly.Parameter, &anomaly.Value, &anomalyTime, &anomaly.Longitude, &anomaly.Latitude, &anomaly.Description); err != nil {
			log.Printf("Error scanning anomaly row: %v", err)
			return nil, err
		}
		anomaly.Time = anomalyTime.Format(time.RFC3339) // Format back to string if needed by model
		anomalies = append(anomalies, anomaly)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating anomaly rows: %v", err)
		return nil, err
	}

	return anomalies, nil
}

func (r *AnomalyRepository) GetAnomaliesByTimeRange(startTime, endTime time.Time) ([]models.Anomaly, error) {
	query := `
		SELECT parameter, value, time,
			   ST_X(location::geometry) AS longitude,
			   ST_Y(location::geometry) AS latitude,
			   description
		FROM anomalies
		WHERE time >= $1 AND time <= $2
		ORDER BY time DESC;`

	rows, err := r.Db.Query(query, startTime, endTime)
	if err != nil {
		log.Printf("Error querying anomalies by time range: %v", err)
		return nil, err
	}
	defer rows.Close()

	var anomalies []models.Anomaly
	for rows.Next() {
		var anomaly models.Anomaly
		var anomalyTime time.Time
		if err := rows.Scan(&anomaly.Parameter, &anomaly.Value, &anomalyTime, &anomaly.Longitude, &anomaly.Latitude, &anomaly.Description); err != nil {
			log.Printf("Error scanning anomaly row: %v", err)
			return nil, err
		}
		anomaly.Time = anomalyTime.Format(time.RFC3339)
		anomalies = append(anomalies, anomaly)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating anomaly rows: %v", err)
		return nil, err
	}

	return anomalies, nil
}

// GetAnomalyDensityByRegion calculates anomaly counts within a bounding box.
// Returns a map where keys represent grid cells (e.g., "lat_lon") and values are counts.
// This is a simplified approach; more sophisticated grid/clustering could be used.
func (r *AnomalyRepository) GetAnomalyDensityByRegion(minLat, minLon, maxLat, maxLon float64) (map[string]int, error) {
	query := `
		SELECT COUNT(*) AS count,
			   -- Define a simple grid cell identifier (e.g., rounding coordinates)
			   -- Adjust precision for desired grid size
			   ROUND(ST_Y(location::geometry)::numeric, 2) AS grid_lat,
			   ROUND(ST_X(location::geometry)::numeric, 2) AS grid_lon
		FROM anomalies
		WHERE ST_Contains(
			ST_MakeEnvelope($1, $2, $3, $4, 4326), 
			location::geometry
		) -- Check if location is within the envelope
		GROUP BY grid_lat, grid_lon;`

	rows, err := r.Db.Query(query, minLon, minLat, maxLon, maxLat) // Note: ST_MakeEnvelope uses minLon, minLat, maxLon, maxLat
	if err != nil {
		log.Printf("Error querying anomaly density: %v", err)
		return nil, err
	}
	defer rows.Close()

	density := make(map[string]int)
	for rows.Next() {
		var count int
		var gridLat, gridLon float64
		if err := rows.Scan(&count, &gridLat, &gridLon); err != nil {
			log.Printf("Error scanning density row: %v", err)
			return nil, err
		}
		// Create a simple key for the map
		gridKey := fmt.Sprintf("%.2f_%.2f", gridLat, gridLon)
		density[gridKey] = count
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating density rows: %v", err)
		return nil, err
	}

	return density, nil
}
