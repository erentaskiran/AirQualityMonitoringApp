package anomaly

import (
	"api/internal/models"
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed window.lua
var luaScript string

type Detector struct {
	db     *sql.DB
	rdb    *redis.Client
	script *redis.Script
}

const (
	window   = 24 * time.Hour
	zKeyTmpl = "sensor:%s:z"  // per‑sensor zset
	rollTmpl = "sensor:%s:hp" // per‑sensor hash (sum,count)
)

func NewAnomalyDetector(rdb *redis.Client, db *sql.DB) *Detector {
	return &Detector{
		db:     db,
		rdb:    rdb,
		script: redis.NewScript(luaScript),
	}
}

func (d *Detector) IsAnomalous(data models.AirQualityData) bool {
	ctx := context.Background()

	var cutoff int64
	if data.Parameter == "O3" {
		cutoff = data.Timestamp.Add(-window / 3).UnixMilli()
	} else {
		cutoff = data.Timestamp.Add(-window).UnixMilli()
	}

	sum, count, err := d.fetchDataFromCacheOrDB(ctx, data.Parameter, cutoff)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return false
	}

	average := float64(sum) / float64(count)

	// WHO threshold check
	if d.CheckTreshold(data.Parameter, data.Value) {
		fmt.Println("⚠️ Anomaly Detected (Threshold):", data)
		d.triggerAnomalyActions(data)
		return true
	}

	// Percentage increase check
	if data.Value > average*1.5 {
		fmt.Println("⚠️ Anomaly Detected (Percentage Increase):", data)
		d.triggerAnomalyActions(data)
		return true
	}

	// Z-score-based detection
	if d.isZScoreAnomalous(data, average, float64(count)) {
		fmt.Println("⚠️ Anomaly Detected (Z-score):", data)
		d.triggerAnomalyActions(data)
		return true
	}

	// Time series analysis
	if d.isTimeSeriesAnomalous(data) {
		fmt.Println("⚠️ Anomaly Detected (Time Series):", data)
		d.triggerAnomalyActions(data)
		return true
	}

	// Geospatial anomaly detection
	if d.isGeospatialAnomalous(data) {
		fmt.Println("⚠️ Anomaly Detected (Geospatial):", data)
		d.triggerAnomalyActions(data)
		return true
	}

	return false
}

func (d *Detector) fetchDataFromCacheOrDB(ctx context.Context, parameter string, cutoff int64) (int64, int64, error) {
	zKey := fmt.Sprintf(zKeyTmpl, parameter)
	rollUp := fmt.Sprintf(rollTmpl, parameter)

	// Try to fetch data from Redis
	res, err := d.script.Run(ctx, d.rdb, []string{zKey, rollUp}, cutoff).Result()
	if err == nil {
		arr, ok := res.([]interface{})
		if ok && len(arr) == 2 {
			sum, ok1 := arr[0].(int64)
			count, ok2 := arr[1].(int64)
			if ok1 && ok2 {
				return sum, count, nil
			}
		}
	}

	// If Redis fetch fails, fallback to DB
	fmt.Println("Data not found in Redis, fetching from DB...")
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

	// Write the fetched data back to Redis
	_, err = d.rdb.HSet(ctx, rollUp, "sum", sum, "count", count).Result()
	if err != nil {
		fmt.Println("Failed to write data to Redis:", err)
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
	// Standard deviation approximation using Redis data
	stdDev := math.Sqrt(mean) // Simplified for demonstration
	zScore := (data.Value - mean) / stdDev
	return math.Abs(zScore) > 3
}

func (a *Detector) isTimeSeriesAnomalous(data models.AirQualityData) bool {
	// Placeholder for time series analysis logic
	// Implement ARIMA, Holt-Winters, or other methods here
	return false
}

func (a *Detector) isGeospatialAnomalous(data models.AirQualityData) bool {
	// Placeholder for geospatial anomaly detection
	// Compare with nearby sensors within a 25km radius
	return false
}

func (a *Detector) triggerAnomalyActions(data models.AirQualityData) {
	a.markOnMap(data)

	a.sendAlert(data)
}

func (a *Detector) markOnMap(data models.AirQualityData) {
	fmt.Println("📍 Marking anomaly on map:", data)
}

func (a *Detector) sendAlert(data models.AirQualityData) {
	fmt.Println("🚨 Sending alert to warning panel:", data)
}
