package anomaly

import (
	"api/internal/models"
	"api/internal/repository"
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

	zKey := fmt.Sprintf(zKeyTmpl, data.Parameter)
	rollUp := fmt.Sprintf(rollTmpl, data.Parameter)

	now := data.Timestamp.UnixMilli()
	cutoff := data.Timestamp.Add(-window).UnixMilli()

	res, err := d.script.Run(ctx, d.rdb,
		[]string{zKey, rollUp},
		now, data.Value, cutoff).Result()
	if err != nil {
		// fall back or log…
		return false
	}

	arr := res.([]interface{})
	sum := arr[0].(int64)
	count := arr[1].(int64)
	avg := float64(sum) / float64(count)
	return data.Value > 0.5*avg

	/*
		if d.CheckTreshold(data) {
			fmt.Println("⚠️ Anomaly Detected (Threshold):", data)
			d.triggerAnomalyActions(data)
			return true
		}

		// Z-score-based detection (Statistical method)
		if d.isZScoreAnomalous(data) {
			fmt.Println("⚠️ Anomaly Detected (Z-score):", data)
			d.triggerAnomalyActions(data)
			return true
		}

		if d.isTimeSeriesAnomalous(data) {
			fmt.Println("⚠️ Anomaly Detected (Time Series):", data)
			d.triggerAnomalyActions(data)
			return true
		}

		if d.isGeospatialAnomalous(data) {
			fmt.Println("⚠️ Anomaly Detected (Geospatial):", data)
			d.triggerAnomalyActions(data)
			return true
		}

		return false
	*/
}

func (a *Detector) CheckTreshold(data models.AirQualityData) bool {
	thresholds := map[string]float64{
		"PM2.5": 15.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"PM10":  45.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"NO2":   25.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"SO2":   40.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"O3":    100.0, // WHO 2021 8 saatlik ortalama sınır değeri
	}
	airQualityRepository := repository.NewAirQualityRepository(a.db)

	var dataList []models.AirQualityData
	var err error

	if data.Parameter == "O3" {
		dataList, err = airQualityRepository.Get8HourDataForParameter(data.Parameter, data.Latitude, data.Longitude)
	} else {
		dataList, err = airQualityRepository.Get24HourDataForParameter(data.Parameter, data.Latitude, data.Longitude)
	}
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return false
	}

	sum := 0.0
	for _, d := range dataList {
		sum += d.Value
	}
	average := sum + data.Value/float64(len(dataList)+1)

	return average > thresholds[data.Parameter]
}

func (a *Detector) isZScoreAnomalous(data models.AirQualityData) bool {
	mean := 50.0
	stdDev := 10.0
	zScore := (data.Value - mean) / stdDev
	return math.Abs(zScore) > 3
}

func (a *Detector) isTimeSeriesAnomalous(data models.AirQualityData) bool {
	return false
}

func (a *Detector) isGeospatialAnomalous(data models.AirQualityData) bool {
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
