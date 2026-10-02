package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Habeebamoo/Prism/internal/ingest"
	"github.com/Habeebamoo/Prism/internal/middlewares"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/ingest", ingest.IngestHandler)

	handler := middlewares.CORS(mux)

	fmt.Println("Ingestion API is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}