package main

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	"api/internal/api"
	"api/internal/consumer"
	websocketserver "api/internal/websocket"
	"api/pkg/db"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

type app struct {
	QueueConn *amqp.Connection
	Db        *sql.DB
	Clients   map[*websocket.Conn]bool
	WsServer  *websocketserver.WebsocketServer
	Api       *api.Api
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

	clients := make(map[*websocket.Conn]bool)
	app := &app{
		QueueConn: conn,
		Db:        Db,
		Clients:   clients,
		WsServer:  websocketserver.NewWebsocketServer(Db, clients),
		Api:       api.NewApi(Db),
	}

	consumer := consumer.NewConsumer(app.QueueConn, app.Db, app.WsServer)
	go consumer.StartConsumer()
	fmt.Println("Consumer started")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		app.WsServer.StartWebsocketServer()
	}()
	fmt.Println("WebSocket server starting on port 8080")

	go func() {
		defer wg.Done()
		app.Api.StartApi()
	}()
	fmt.Println("API server starting on port 8081")

	wg.Wait()
}
