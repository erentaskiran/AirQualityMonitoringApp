package models

type Anomaly struct {
	Parameter   string  `json:"parameter"`
	Value       float64 `json:"value"`
	Time        string  `json:"time"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Description string  `json:"description"`
}
