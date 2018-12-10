package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	on  string = "on"
	off string = "off"
)

var power string

func newBatteryLevel(level int) {
	if level >= 80 {
		power = off
	}
	if level <= 75 {
		power = on
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
	power = on
	bg()
	log.Fatal(http.ListenAndServe(getEnv("ADDR", ":8000"), router))
}

func bg() {
	var interval, err = time.ParseDuration(getEnv("INTERVAL", "10s"))
	if err != nil {
		log.Fatal(err)
	}
	time.AfterFunc(interval, bg)
	if power == on {
		execEnv("POWER_ON_CMD", "echo 0x1 > /sys/devices/platform/bcm2708_usb/buspower")
	}
	if power == off {
		execEnv("POWER_ON_CMD", "echo 0x0 > /sys/devices/platform/bcm2708_usb/buspower")
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func execEnv(key, fallback string) {
	var cmd = getEnv(key, fallback)
	var out, err = exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Printf("cmd='%v', err='%v'", cmd, err)
	}
	if len(out) > 0 {
		var outp = fmt.Sprintf("%s\n", out)
		log.Println(strings.Trim(outp, "\n "))
	}
}
