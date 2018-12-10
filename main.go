package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var power string

func newBatteryLevel(level int) {
	if level >= 80 {
		power = "off"
	}
	if level <= 75 {
		power = "on"
	}
	log.Printf("batt=%v,power=%v\n", level, power)
}

type request struct {
	Batt int `json:"batt,omitempty,string"`
}

func update(w http.ResponseWriter, r *http.Request) {
	var req request
	_ = json.NewDecoder(r.Body).Decode(&req)
	newBatteryLevel(req.Batt)
	json.NewEncoder(w).Encode(req)
}

// https://www.codementor.io/codehakase/building-a-restful-api-with-golang-a6yivzqdo

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/update", update).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", router))
}
