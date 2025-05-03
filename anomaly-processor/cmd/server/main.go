package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

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
		WsServer:  websocketserver.NewWebsocketServer(Db, make(map[*websocket.Conn]bool)),
	}

	consumer := consumer.NewConsumer(app.QueueConn, app.Db, app.WsServer)
	go consumer.StartConsumer()
	fmt.Println("Consumer started")

	http.HandleFunc("/ws/live", app.WsServer.WsHandler)
	http.HandleFunc("/ws/anomalys", app.WsServer.WsHandlerAnomaly)

	// New HTTP API Endpoints
	http.HandleFunc("/api/anomalies/location", app.WsServer.AnomaliesByLocationHandler)   // GET /api/anomalies/location?lat=...&lon=...&radius=...
	http.HandleFunc("/api/anomalies/timerange", app.WsServer.AnomaliesByTimeRangeHandler) // GET /api/anomalies/timerange?start=...&end=...
	http.HandleFunc("/api/anomalies/density", app.WsServer.AnomalyDensityHandler)         // GET /api/anomalies/density?minLat=...&minLon=...&maxLat=...&maxLon=...

	fmt.Println("Websocket and API server running at :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
