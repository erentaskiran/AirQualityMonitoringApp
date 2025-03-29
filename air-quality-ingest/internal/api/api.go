package api

import (
	"api/internal/models"
	"api/internal/queue"
	"api/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

type Router struct {
	QueueConn *amqp.Connection
}

func NewRouter(QueueConn *amqp.Connection) *Router {
	return &Router{
		QueueConn: QueueConn,
	}
}

func (r *Router) NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/ingest", r.IngestHandler).Methods(http.MethodPost, http.MethodOptions)

	return router
}

func (r *Router) IngestHandler(w http.ResponseWriter, req *http.Request) {
	var payload models.AirQualityPayload
	err := utils.DecodeRequestBody(req, &payload)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	latitude, err := strconv.ParseFloat(payload.Latitude, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid latitude value"+err.Error())
		return
	}

	longitude, err := strconv.ParseFloat(payload.Longitude, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid longitude value")
		return
	}
	value, err := strconv.ParseFloat(payload.Value, 64)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid value")
		return
	}

	data := models.AirQualityData{
		Latitude:  latitude,
		Longitude: longitude,
		Parameter: payload.Parameter,
		Value:     value,
	}

	queue := queue.NewQueue(r.QueueConn)

	if err := queue.PublishToQueue(data); err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
