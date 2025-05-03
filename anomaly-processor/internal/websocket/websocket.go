package websocketserver

import (
	"api/internal/repository"
	"api/pkg/utils"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketServer struct {
	Db      *sql.DB
	Clients map[*websocket.Conn]bool
}

func NewWebsocketServer(Db *sql.DB, Clients map[*websocket.Conn]bool) *WebsocketServer {
	return &WebsocketServer{
		Db:      Db,
		Clients: Clients,
	}
}

var clientsMu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (c *WebsocketServer) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	c.Clients[conn] = true
	clientsMu.Unlock()

	log.Println("Client connected")

	for {
		if _, _, err := conn.NextReader(); err != nil {
			clientsMu.Lock()
			delete(c.Clients, conn)
			clientsMu.Unlock()
			log.Println("Client disconnected")
			break
		}
	}
}

func (c *WebsocketServer) WsHandlerAnomaly(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}
	defer conn.Close()

	repo := repository.NewAnomalyRepository(c.Db)
	anomalies, err := repo.GetRecentAnomalies()
	if err != nil {
		log.Println("Error fetching recent anomalies:", err)
		return
	}

	if err := conn.WriteJSON(anomalies); err != nil {
		log.Println("Error writing JSON to WebSocket:", err)
	}
}

func (c *WebsocketServer) BroadcastToClients(message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range c.Clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Write error, removing client:", err)
			client.Close()
			delete(c.Clients, client)
		}
	}
}

// HTTP handler for getting anomalies by location
func (c *WebsocketServer) AnomaliesByLocationHandler(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	radiusStr := r.URL.Query().Get("radius") // Radius in km

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

	repo := repository.NewAnomalyRepository(c.Db)
	anomalies, err := repo.GetAnomaliesByLocation(lat, lon, radius)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to retrieve anomalies by location")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, anomalies)
}

// HTTP handler for getting anomalies by time range
func (c *WebsocketServer) AnomaliesByTimeRangeHandler(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.URL.Query().Get("start") // Expected format: RFC3339 or similar parseable by time.Parse
	endTimeStr := r.URL.Query().Get("end")

	if startTimeStr == "" || endTimeStr == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "Missing required query parameters: start, end")
		return
	}

	// Use a flexible parsing approach if needed, RFC3339 is standard
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

	repo := repository.NewAnomalyRepository(c.Db)
	anomalies, err := repo.GetAnomaliesByTimeRange(startTime, endTime)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to retrieve anomalies by time range")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, anomalies)
}

// HTTP handler for getting anomaly density by region (bounding box)
func (c *WebsocketServer) AnomalyDensityHandler(w http.ResponseWriter, r *http.Request) {
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

	repo := repository.NewAnomalyRepository(c.Db)
	density, err := repo.GetAnomalyDensityByRegion(minLat, minLon, maxLat, maxLon)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to retrieve anomaly density")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, density)
}
