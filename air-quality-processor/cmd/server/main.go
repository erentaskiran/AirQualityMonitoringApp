package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"api/internal/consumer"
	"api/pkg/db"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
)

type app struct {
	QueueConn *amqp.Connection
	Db        *sql.DB
	Redis     *redis.Client
}

func main() {
	_ = godotenv.Load()
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	Db := db.InitDB(dbURL)
	defer Db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("cannot connect to redis: %v", err)
	}

	app := &app{
		QueueConn: conn,
		Db:        Db,
		Redis:     rdb,
	}

	consumer := consumer.NewConsumer(app.QueueConn, app.Db, app.Redis)
	consumer.StartConsumer()
	fmt.Println("Consumer started")
}
