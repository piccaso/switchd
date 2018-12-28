package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
)

var (
	mqttBroker   = getEnv("MQTT_BROKER", "tcp://mqtt:1883")
	mqttClientID = getEnv("MQTT_CLIENT_ID", "switchd")
	mqttTopic    = getEnv("MQTT_TOPIC", "owntracks/+/+")
	mqttDebug    = getEnv("MQTT_DEBUG", "False")
)

func connectionLost(client mqtt.Client, err error) {
	log.Fatal(err)
}

func subscribe() {
	if len(mqttBroker) < 3 {
		return
	}
	var debug, _ = strconv.ParseBool(mqttDebug)
	if debug {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
		mqtt.WARN = log.New(os.Stdout, "", 0)
	}
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "", 0)

	var opts = mqtt.NewClientOptions().AddBroker(mqttBroker).SetClientID(mqttClientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.OnConnectionLost = connectionLost
	var client = mqtt.NewClient(opts)
	var token = client.Connect()
	token.Wait()
	var err = token.Error()
	if err != nil {
		log.Fatal(err)
	}
	token = client.Subscribe(mqttTopic, 0, mqttMessageHandler)
	token.Wait()
	err = token.Error()
	if err != nil {
		log.Fatal(err)
	}
}

func mqttMessageHandler(client mqtt.Client, msg mqtt.Message) {
	var payload = msg.Payload()
	log.Printf("mqtt payload=%s, topic=%v\n", payload, msg.Topic())
	var r request
	json.Unmarshal(payload, &r)
	newBatteryLevel(r.Batt, r.Tst)
}

const (
	on  string = "on"
	off string = "off"
)

var power = getEnv("INIT_POWER_STATE", on)
var powerMax = getIntEnv("POWER_MAX", 80)
var powerMin = getIntEnv("POWER_MIN", 78)
var lastTimestamp = timeNow()
var lastBatteryLevel int
var lock sync.Mutex

func newBatteryLevel(level int, timestamp uint64) {
	lock.Lock()
	defer lock.Unlock()
	if timestamp < lastTimestamp {
		log.Printf("timestamp=%v is to small", timestamp)
		return
	}
	if level >= powerMax {
		power = off
	}
	if level <= powerMin {
		power = on
	}
	lastTimestamp = timestamp
	lastBatteryLevel = level
	log.Printf("batt=%v,power=%v,timestamp=%v\n", level, power, timestamp)
}

type request struct {
	Batt int    `json:"batt,omitempty"`
	Tst  uint64 `json:"tst,omitempty"`
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	var dump, _ = httputil.DumpRequest(r, true)
	log.Println(string(dump))

	var req request
	_ = json.NewDecoder(r.Body).Decode(&req)
	newBatteryLevel(req.Batt, req.Tst)
	json.NewEncoder(w).Encode(req)
}

func onHandler(w http.ResponseWriter, r *http.Request) {
	cmdHandler(w, r, on)
}

func offHandler(w http.ResponseWriter, r *http.Request) {
	cmdHandler(w, r, off)
}

func cmdHandler(w http.ResponseWriter, r *http.Request, p string) {
	lock.Lock()
	defer lock.Unlock()
	if p != power {
		power = p
		var err = execCmd()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
	}
}

func timeNow() uint64 {
	return uint64(time.Now().Unix())
}

func getDataAge() uint64 {
	now := timeNow()
	lock.Lock()
	defer lock.Unlock()
	age := now - lastTimestamp
	return age
}

func dbgHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	k := vars["k"]
	response := ""
	status := http.StatusOK

	switch k {
	case "timestamp":
		response = fmt.Sprintf("%v", lastTimestamp)
	case "power":
		if power == on {
			response = "1"
		}
		if power == off {
			response = "0"
		}
	case "level":
		age := getDataAge()
		if age < 3600 {
			response = fmt.Sprintf("%v", lastBatteryLevel)
		}
	case "age":
		age := getDataAge()
		response = fmt.Sprintf("%v", age)
	default:
		status = http.StatusNotFound
	}

	w.WriteHeader(status)
	w.Write([]byte(response))
}

// https://www.codementor.io/codehakase/building-a-restful-api-with-golang-a6yivzqdo

func main() {
	go subscribe()
	router := mux.NewRouter()
	router.HandleFunc("/update", updateHandler).Methods("POST")
	router.HandleFunc("/on", onHandler).Methods("GET", "POST")
	router.HandleFunc("/off", offHandler).Methods("GET", "POST")
	router.HandleFunc("/dbg/{k}", dbgHandler).Methods("GET")
	muninDir := getEnv("MUNIN_DIR", "/var/cache/munin/www/")
	fi, err := os.Stat(muninDir)
	if err == nil && fi.IsDir() {
		log.Printf("/munin -> %v\n", muninDir)
		muninFs := http.FileServer(http.Dir(muninDir))
		router.PathPrefix("/munin/").Handler(http.StripPrefix("/munin", muninFs))
	} else {
		if err == nil {
			err = errors.New("not a directory")
		}
		log.Printf("Error stating munin directory '%v': %v\n", muninDir, err)
	}

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
	age := getDataAge()
	lock.Lock()
	defer lock.Unlock()
	if age > 3*60*60 {
		power = on
	}
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
