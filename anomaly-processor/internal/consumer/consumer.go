package consumer

import (
	"database/sql"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

type Consumer struct {
	QueueConn *amqp.Connection
	Db        *sql.DB
	Clients   map[*websocket.Conn]bool
}

func NewConsumer(queueConn *amqp.Connection, db *sql.DB, Clients map[*websocket.Conn]bool) *Consumer {
	return &Consumer{
		QueueConn: queueConn,
		Db:        db,

		Clients: Clients,
	}
}

var clientsMu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (c *Consumer) StartConsumer() {
	ch, err := c.QueueConn.Channel()
	if err != nil {
		log.Fatal("Channel error:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"anomaly_alerts", // Kuyruk adı
		true,             // Durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Queue declare error:", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true, // Otomatik onay
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Consume error:", err)
	}

	for msg := range msgs {
		log.Printf("Received message: %s", msg.Body)
		log.Print("clientCount", len(c.Clients))
		c.BroadcastToClients(msg.Body)
	}
}

func (c *Consumer) BroadcastToClients(message []byte) {
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

func (c *Consumer) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Client'ı listeye ekle
	clientsMu.Lock()
	c.Clients[conn] = true
	clientsMu.Unlock()

	log.Println("Client connected")

	// Client bağlantısı kapanınca listeden çıkar
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
