package api

import (
	"api/internal/repository"
	"api/pkg/utils"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Api struct {
	Db *sql.DB
}

func NewApi(Db *sql.DB) *Api {
	return &Api{
		Db: Db,
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Start-Time, X-End-Time")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (a *Api) StartApi() {
	http.HandleFunc("/api/anomalies/location", corsMiddleware(a.AnomaliesByLocationHandler))
	http.HandleFunc("/api/anomalies/timerange", corsMiddleware(a.AnomaliesByTimeRangeHandler))
	http.HandleFunc("/api/anomalies/density", corsMiddleware(a.AnomalyDensityHandler))

	log.Println("Starting API server on port 8081...")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("API server failed to start: %v", err)
	}
}

func (a *Api) AnomaliesByLocationHandler(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	radiusStr := r.URL.Query().Get("radius")

	if latStr == "" || lonStr == "" || radiusStr == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "Missing required query parameters: lat, lon, radius")
		return
	}

	lat, errLat := strconv.ParseFloat(latStr, 64)
	lon, errLon := strconv.ParseFloat(lonStr, 64)
	radius, errRadius := strconv.ParseFloat(radiusStr, 64)

	if errLat != nil || errLon != nil || errRadius != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid numeric value for lat, lon, or radius")
		return
	}

	repo := repository.NewAnomalyRepository(a.Db)
	anomalies, err := repo.GetAnomaliesByLocation(lat, lon, radius)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to retrieve anomalies by location")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, anomalies)
}

func (a *Api) AnomaliesByTimeRangeHandler(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.Header.Get("X-Start-Time")
	endTimeStr := r.Header.Get("X-End-Time")

	if startTimeStr == "" || endTimeStr == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "Missing required query parameters: X-Start-Time, X-End-Time")
		return
	}

	startTime, errStart := time.Parse(time.RFC3339, startTimeStr)
	endTime, errEnd := time.Parse(time.RFC3339, endTimeStr)

	if errStart != nil || errEnd != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid time format for start or end (use RFC3339: YYYY-MM-DDTHH:MM:SSZ)")
		return
	}

	if startTime.After(endTime) {
		utils.WriteJSONError(w, http.StatusBadRequest, "Start time cannot be after end time")
		return
	}

	repo := repository.NewAnomalyRepository(a.Db)
	anomalies, err := repo.GetAnomaliesByTimeRange(startTime, endTime)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to retrieve anomalies by time range")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, anomalies)
}

func (a *Api) AnomalyDensityHandler(w http.ResponseWriter, r *http.Request) {
	minLatStr := r.URL.Query().Get("minLat")
	minLonStr := r.URL.Query().Get("minLon")
	maxLatStr := r.URL.Query().Get("maxLat")
	maxLonStr := r.URL.Query().Get("maxLon")

	if minLatStr == "" || minLonStr == "" || maxLatStr == "" || maxLonStr == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "Missing required query parameters: minLat, minLon, maxLat, maxLon")
		return
	}

	minLat, errMinLat := strconv.ParseFloat(minLatStr, 64)
	minLon, errMinLon := strconv.ParseFloat(minLonStr, 64)
	maxLat, errMaxLat := strconv.ParseFloat(maxLatStr, 64)
	maxLon, errMaxLon := strconv.ParseFloat(maxLonStr, 64)

	if errMinLat != nil || errMinLon != nil || errMaxLat != nil || errMaxLon != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid numeric value for bounding box coordinates")
		return
	}

	if minLat >= maxLat || minLon >= maxLon {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid bounding box: min coordinates must be less than max coordinates")
		return
	}

	repo := repository.NewAnomalyRepository(a.Db)
	density, err := repo.GetAnomalyDensityByRegion(minLat, minLon, maxLat, maxLon)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to retrieve anomaly density")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, density)
}
