package anomaly

import (
	"api/internal/models"
	"api/internal/repository"
	"database/sql"
	"fmt"
	"math"

	"github.com/redis/go-redis/v9"
)

type AnomalyDetector struct {
	Redis *redis.Client
	Db    *sql.DB
}

func NewAnomalyDetector(redis *redis.Client, Db *sql.DB) *AnomalyDetector {
	return &AnomalyDetector{
		Redis: redis,
		Db:    Db,
	}
}

func (a *AnomalyDetector) IsAnomalous(data models.AirQualityData) bool {
	if a.CheckTresholds(data) {
		fmt.Println("⚠️ Anomaly Detected (Threshold):", data)
		a.triggerAnomalyActions(data)
		return true
	}

	// Z-score-based detection (Statistical method)
	if a.isZScoreAnomalous(data) {
		fmt.Println("⚠️ Anomaly Detected (Z-score):", data)
		a.triggerAnomalyActions(data)
		return true
	}

	if a.isTimeSeriesAnomalous(data) {
		fmt.Println("⚠️ Anomaly Detected (Time Series):", data)
		a.triggerAnomalyActions(data)
		return true
	}

	if a.isGeospatialAnomalous(data) {
		fmt.Println("⚠️ Anomaly Detected (Geospatial):", data)
		a.triggerAnomalyActions(data)
		return true
	}

	return false
}

func (a *AnomalyDetector) CheckTresholds(data models.AirQualityData) bool {
	thresholds := map[string]float64{
		"PM2.5": 15.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"PM10":  45.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"NO2":   25.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"SO2":   40.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"O3":    100.0, // WHO 2021 8 saatlik ortalama sınır değeri
	}
	airQualityRepository := repository.NewAirQualityRepository(a.Db)

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

func (a *AnomalyDetector) isZScoreAnomalous(data models.AirQualityData) bool {
	mean := 50.0
	stdDev := 10.0
	zScore := (data.Value - mean) / stdDev
	return math.Abs(zScore) > 3
}

func (a *AnomalyDetector) isTimeSeriesAnomalous(data models.AirQualityData) bool {
	return false
}

func (a *AnomalyDetector) isGeospatialAnomalous(data models.AirQualityData) bool {
	return false
}

func (a *AnomalyDetector) triggerAnomalyActions(data models.AirQualityData) {
	a.markOnMap(data)

	a.sendAlert(data)
}

func (a *AnomalyDetector) markOnMap(data models.AirQualityData) {
	fmt.Println("📍 Marking anomaly on map:", data)
}

func (a *AnomalyDetector) sendAlert(data models.AirQualityData) {
	fmt.Println("🚨 Sending alert to warning panel:", data)
}
