package consumer

import (
	"api/internal/repository"
	websocketserver "api/internal/websocket"
	"database/sql"
	"log"

	"github.com/streadway/amqp"
)

type Consumer struct {
	QueueConn *amqp.Connection
	Db        *sql.DB
	WsServer  *websocketserver.WebsocketServer
}

func NewConsumer(queueConn *amqp.Connection, db *sql.DB, WsServer *websocketserver.WebsocketServer) *Consumer {
	return &Consumer{
		QueueConn: queueConn,
		Db:        db,
		WsServer:  WsServer,
	}
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
	anomalyRepository := repository.NewAnomalyRepository(c.Db)

	for msg := range msgs {
		log.Printf("Received message: %s", msg.Body)
		anomalyRepository.SaveAnomalyToDB(msg.Body)
		c.WsServer.BroadcastToClients(msg.Body)
	}
}
