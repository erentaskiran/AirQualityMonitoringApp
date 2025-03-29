package main

import (
	"database/sql"
	"fmt"
	"os"

	"api/internal/consumer"
	"api/pkg/db"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

type app struct {
	QueueConn *amqp.Connection
	Db        *sql.DB
}

func main() {
	_ = godotenv.Load()
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	dbURL := os.Getenv("DATABASE_URL")

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	Db := db.InitDB(dbURL)
	defer Db.Close()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := &app{
		QueueConn: conn,
		Db:        Db,
	}

	consumer := consumer.NewConsumer(app.QueueConn, app.Db)
	consumer.StartConsumer()
	fmt.Println("Consumer started")
}
