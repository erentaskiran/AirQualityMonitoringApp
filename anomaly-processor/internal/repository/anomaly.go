package repository

import (
	"api/internal/models"
	"database/sql"
	"encoding/json"
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
