package models

type AirQualityData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Parameter string  `json:"parameter"`
	Value     float64 `json:"value"`
}
