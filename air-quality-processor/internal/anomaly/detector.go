package anomaly

import (
	"api/internal/models"
	"fmt"
	"math"
)

func IsAnomalous(data models.AirQualityData) bool {
	thresholds := map[string]float64{
		"PM2.5": 15.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"PM10":  45.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"NO2":   25.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"SO2":   40.0,  // WHO 2021 24 saatlik ortalama sınır değeri
		"O3":    100.0, // WHO 2021 8 saatlik ortalama sınır değeri
	}

	// WHO Threshold-based detection
	if val, exists := thresholds[data.Parameter]; exists {
		if data.Value > val {
			fmt.Println("⚠️ Anomaly Detected (Threshold):", data)
			triggerAnomalyActions(data)
			return true
		}
	}

	// Z-score-based detection (Statistical method)
	if isZScoreAnomalous(data) {
		fmt.Println("⚠️ Anomaly Detected (Z-score):", data)
		triggerAnomalyActions(data)
		return true
	}

	if isTimeSeriesAnomalous(data) {
		fmt.Println("⚠️ Anomaly Detected (Time Series):", data)
		triggerAnomalyActions(data)
		return true
	}

	if isGeospatialAnomalous(data) {
		fmt.Println("⚠️ Anomaly Detected (Geospatial):", data)
		triggerAnomalyActions(data)
		return true
	}

	return false
}

func isZScoreAnomalous(data models.AirQualityData) bool {
	mean := 50.0
	stdDev := 10.0
	zScore := (data.Value - mean) / stdDev
	return math.Abs(zScore) > 3
}

func isTimeSeriesAnomalous(data models.AirQualityData) bool {
	return false
}

func isGeospatialAnomalous(data models.AirQualityData) bool {
	return false
}

func triggerAnomalyActions(data models.AirQualityData) {
	markOnMap(data)

	sendAlert(data)
}

func markOnMap(data models.AirQualityData) {
	fmt.Println("📍 Marking anomaly on map:", data)
}

func sendAlert(data models.AirQualityData) {
	fmt.Println("🚨 Sending alert to warning panel:", data)
}
