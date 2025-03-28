package api

import (
	"api/internal/models"
	"api/pkg/utils"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

type Router struct {
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/ingest", r.IngestHandler).Methods(http.MethodPost, http.MethodOptions)

	return router
}

func (r *Router) IngestHandler(w http.ResponseWriter, req *http.Request) {
	var payload models.AirQualityData
	err := utils.DecodeRequestBody(req, &payload)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := publishToQueue(payload); err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func publishToQueue(data models.AirQualityData) error {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"air_quality", // queue name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

var rabbitMQURL = os.Getenv("RABBITMQ_URL")
