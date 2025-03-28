package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"api/internal/api"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

type app struct {
	QueueConn *amqp.Connection
}

func main() {
	_ = godotenv.Load()
	rabbitMQURL := os.Getenv("RABBITMQ_URL")

	time.Sleep(10 * time.Second)

	port := ":8080"

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := &app{
		QueueConn: conn,
	}

	defer conn.Close()

	router := api.NewRouter(app.QueueConn)

	r := router.NewRouter()

	fmt.Println("Server is running on port", port)

	err = http.ListenAndServe(port, r)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
