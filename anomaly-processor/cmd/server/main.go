package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"api/internal/consumer"
	"api/pkg/db"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

type app struct {
	QueueConn *amqp.Connection
	Db        *sql.DB
	Clients   map[*websocket.Conn]bool
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

	app := &app{
		QueueConn: conn,
		Db:        Db,
		Clients:   make(map[*websocket.Conn]bool),
	}

	consumer := consumer.NewConsumer(app.QueueConn, app.Db, app.Clients)
	go consumer.StartConsumer()
	fmt.Println("Consumer started")
	http.HandleFunc("/ws", consumer.WsHandler)

	fmt.Println("Websocket server running at :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
