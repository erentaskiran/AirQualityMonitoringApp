package websocketserver

import (
	"api/internal/repository"
	"database/sql"
	"log"
	"net/http"
	"sync"

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

// Setup WebSocket routes and start server
func (c *WebsocketServer) StartWebsocketServer() {
	http.HandleFunc("/ws/live", c.WsHandler)

	log.Println("Starting WebSocket server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("WebSocket server failed to start: %v", err)
	}
}

// WebSocket handler for live connections
func (c *WebsocketServer) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}
	defer conn.Close()

	if _, ok := c.Clients[conn]; !ok {
		repo := repository.NewAnomalyRepository(c.Db)
		anomalies, err := repo.GetRecentAnomalies()
		if err != nil {
			log.Println("Error fetching recent anomalies:", err)
			return
		}

		if err := conn.WriteJSON(anomalies); err != nil {
			log.Println("Error writing JSON to WebSocket:", err)
			return
		}
		clientsMu.Lock()
		c.Clients[conn] = true
		clientsMu.Unlock()
		log.Println("Sent recent anomalies to client")
	}

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
