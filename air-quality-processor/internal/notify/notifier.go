package notify

import (
	"api/internal/models"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type Notify struct {
	QueueConn *amqp.Connection
}

func NewNotify(queueConn *amqp.Connection) *Notify {
	return &Notify{
		QueueConn: queueConn,
	}
}

func (n *Notify) NotifyAnomaly(data models.AirQualityData) {
	ch, err := n.QueueConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"anomaly_alerts",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	body, _ := json.Marshal(data)
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Failed to publish anomaly alert: %s", err)
	}

	log.Println("🚨 Anomaly alert sent:", string(body))
}
