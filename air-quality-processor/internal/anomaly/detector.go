package anomaly

import (
	"api/internal/models"
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"math"
	"time"
)

type Detector struct {
	db *sql.DB
}

func NewAnomalyDetector(db *sql.DB) *Detector {
	return &Detector{
		db: db,
	}
}

func (d *Detector) IsAnomalous(data models.AirQualityData) (string, bool) {
	ctx := context.Background()

	var cutoff int64
	if data.Parameter == "O3" {
		cutoff = data.Timestamp.Add(-8 * time.Hour).UnixMilli()
	} else {
		cutoff = data.Timestamp.Add(-24 * time.Hour).UnixMilli()
	}

	sum, count, err := d.fetchDataFromDB(ctx, data.Parameter, cutoff)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return "", false
	}

	if d.CheckTreshold(data.Parameter, data.Value) {
		reason := "Threshold"
		fmt.Printf("⚠️ Anomaly Detected (%s): %v\n", reason, data)
		d.triggerAnomalyActions(data, reason)
		return reason, true
	}

	if count == 0 {
		return "", false
	}

	average := float64(sum) / float64(count)

	if data.Value > average*1.5 {
		reason := "Percentage Increase"
		fmt.Printf("⚠️ Anomaly Detected (%s): %v\n", reason, data)
		d.triggerAnomalyActions(data, reason)
		return reason, true
	}

	if d.isZScoreAnomalous(data, average, float64(count)) {
		reason := "Z-score"
		fmt.Printf("⚠️ Anomaly Detected (%s): %v\n", reason, data)
		d.triggerAnomalyActions(data, reason)
		return reason, true
	}

	if d.isTimeSeriesAnomalous(data) {
		reason := "Time Series"
		fmt.Printf("⚠️ Anomaly Detected (%s): %v\n", reason, data)
		d.triggerAnomalyActions(data, reason)
		return reason, true
	}

	if d.isGeospatialAnomalous(data) {
		reason := "Geospatial"
		fmt.Printf("⚠️ Anomaly Detected (%s): %v\n", reason, data)
		d.triggerAnomalyActions(data, reason)
		return reason, true
	}

	return "", false
}

func (d *Detector) fetchDataFromDB(ctx context.Context, parameter string, cutoff int64) (int64, int64, error) {
	query := `
	SELECT COALESCE(SUM(value), 0) AS sum, COALESCE(SUM(value), 0) AS count 
	FROM measurements 
	WHERE parameter = $1 AND time >= to_timestamp($2)
	`
	row := d.db.QueryRowContext(ctx, query, parameter, cutoff)

	var sum int64
	var count int64
	if err := row.Scan(&sum, &count); err != nil {
		return 0, 0, err
	}

	return sum, count, nil
}

func (a *Detector) CheckTreshold(parameter string, value float64) bool {
	thresholds := map[string]float64{
		"PM2.5": 15.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"PM10":  45.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"NO2":   25.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"SO2":   40.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"O3":    100.0, // WHO 2021 8 saatlik ortalama sınır değeri
	}

	return value > thresholds[parameter]
}

func (a *Detector) isZScoreAnomalous(data models.AirQualityData, mean, count float64) bool {
	stdDev := math.Sqrt(mean)
	zScore := (data.Value - mean) / stdDev
	return math.Abs(zScore) > 3
}

func (a *Detector) isTimeSeriesAnomalous(data models.AirQualityData) bool {
	return false
}

func (a *Detector) isGeospatialAnomalous(data models.AirQualityData) bool {
	return false
}

func (a *Detector) triggerAnomalyActions(data models.AirQualityData, reason string) {
	a.markOnMap(data, reason)
	a.sendAlert(data, reason)
}

func (a *Detector) markOnMap(data models.AirQualityData, reason string) {
	fmt.Printf("📍 Marking anomaly on map (%s): %v\n", reason, data)
}

func (a *Detector) sendAlert(data models.AirQualityData, reason string) {
	fmt.Printf("🚨 Sending alert to warning panel (%s): %v\n", reason, data)
}
