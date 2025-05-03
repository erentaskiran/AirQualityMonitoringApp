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

func (n *Notify) NotifyAnomaly(data models.AirQualityData, reason string) {
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

	alert := models.AnomalyData{
		Parameter:   data.Parameter,
		Value:       data.Value,
		Timestamp:   data.Timestamp,
		Latitude:    data.Latitude,
		Longitude:   data.Longitude,
		Description: reason,
	}

	body, _ := json.Marshal(alert)
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
