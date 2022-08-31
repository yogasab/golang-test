package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/yogasab/golang-test/helpers"
)

func main() {
	http.HandleFunc("/", parseTrackingHandler)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func parseTrackingHandler(w http.ResponseWriter, r *http.Request) {
	data := helpers.ResponseFormatter()
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	case "POST":
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "POST method requested"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Can't find method requested"}`))
	}
}
