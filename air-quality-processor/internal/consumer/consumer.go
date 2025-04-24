package consumer

import (
	"api/internal/anomaly"
	"api/internal/models"
	"api/internal/notify"
	"api/internal/repository"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
)

type Consumer struct {
	QueueConn *amqp.Connection
	Db        *sql.DB
	Redis     *redis.Client
}

func NewConsumer(queueConn *amqp.Connection, db *sql.DB, Redis *redis.Client) *Consumer {
	return &Consumer{
		QueueConn: queueConn,
		Db:        db,
		Redis:     Redis,
	}
}

func (c *Consumer) StartConsumer() {
	ch, err := c.QueueConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	notify := notify.NewNotify(c.QueueConn)
	airQualityRepository := repository.NewAirQualityRepository(c.Db)
	AnomalyDetector := anomaly.NewAnomalyDetector(c.Redis, c.Db)

	q, err := ch.QueueDeclare(
		"mesurements",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var data models.AirQualityData
			if err := json.Unmarshal(d.Body, &data); err != nil {
				log.Printf("Error decoding message: %s", err)
				continue
			}

			fmt.Printf("Received a message: %+v\n", data)

			if AnomalyDetector.IsAnomalous(data) {
				fmt.Println("⚠️ Anomaly detected!", data)
				notify.NotifyAnomaly(data)
			}

			airQualityRepository.SaveToDB(data)
		}
	}()

	log.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
