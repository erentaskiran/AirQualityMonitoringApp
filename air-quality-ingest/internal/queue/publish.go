package queue

import (
	"api/internal/models"
	"encoding/json"

	"github.com/streadway/amqp"
)

type Queue struct {
	QueueConn *amqp.Connection
}

func NewQueue(QueueConn *amqp.Connection) *Queue {
	return &Queue{
		QueueConn: QueueConn,
	}
}

func (r *Queue) PublishToQueue(data models.AirQualityPayload) error {
	ch, err := r.QueueConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"mesurements",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
