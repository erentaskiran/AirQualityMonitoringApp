package models

import "time"

type AirQualityData struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Parameter string    `json:"parameter"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}
type AnomalyData struct {
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Parameter   string    `json:"parameter"`
	Value       float64   `json:"value"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}
