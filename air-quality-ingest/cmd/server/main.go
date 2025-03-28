package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"api/internal/api"
	"github.com/joho/godotenv"
)

type app struct {
	db *sql.DB
}

func main() {
	_ = godotenv.Load()

	time.Sleep(10 * time.Second)

	port := ":8080"

	router := api.NewRouter()

	r := router.NewRouter()

	fmt.Println("Server is running on port", port)

	err := http.ListenAndServe(port, r)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
