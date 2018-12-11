package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	on  string = "on"
	off string = "off"
)

var power = getEnv("INIT_POWER_STATE", on)
var powerMax = getIntEnv("POWER_MAX", 80)
var powerMin = getIntEnv("POWER_MIN", 75)

func newBatteryLevel(level int) {
	if level >= powerMax {
		power = off
	}
	if level <= powerMin {
		power = on
	}
	log.Printf("batt=%v,power=%v\n", level, power)
}

type request struct {
	Batt int `json:"batt,omitempty"`
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	var dump, _ = httputil.DumpRequest(r, true)
	log.Println(string(dump))

	var req request
	_ = json.NewDecoder(r.Body).Decode(&req)
	newBatteryLevel(req.Batt)
	json.NewEncoder(w).Encode(req)
}

func onHandler(w http.ResponseWriter, r *http.Request) {
	cmdHandler(w, r, on)
}

func offHandler(w http.ResponseWriter, r *http.Request) {
	cmdHandler(w, r, off)
}

func cmdHandler(w http.ResponseWriter, r *http.Request, p string) {
	if p != power {
		power = p
		var err = execCmd()
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
		}
	}
}

// https://www.codementor.io/codehakase/building-a-restful-api-with-golang-a6yivzqdo

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/update", updateHandler).Methods("POST")
	router.HandleFunc("/on", onHandler).Methods("GET", "POST")
	router.HandleFunc("/off", offHandler).Methods("GET", "POST")
	bg()
	log.Fatal(http.ListenAndServe(getEnv("LISTEN_ADDR", ":8000"), router))
}

func bg() {
	var interval, err = time.ParseDuration(getEnv("INTERVAL", "10s"))
	if err != nil {
		log.Fatal(err)
	}
	time.AfterFunc(interval, bg)
	execCmd()
}

func execCmd() error {
	if power == on {
		return execEnv("POWER_ON_CMD", "echo 0x1 > /sys/devices/platform/bcm2708_usb/buspower")
	}
	if power == off {
		return execEnv("POWER_OFF_CMD", "echo 0x0 > /sys/devices/platform/bcm2708_usb/buspower")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	var value = getEnv(key, strconv.Itoa(fallback))

	var i, err = strconv.ParseInt(value, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return int(i)
}

func execEnv(key, fallback string) error {
	var cmd = getEnv(key, fallback)
	var out, err = exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Printf("cmd='%v', err='%v'", cmd, err)
	}
	if len(out) > 0 {
		log.Println(strings.Trim(string(out), "\n "))
	}
	return err
}
