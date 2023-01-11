package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/peterbourgon/diskv/v3"
)

type SensorOnboardingRequest struct {
	DevEUI string `json:"deveui"`
}

var d *diskv.Diskv

func sensorOnboarding(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	request := SensorOnboardingRequest{}
	if err := json.Unmarshal(r, &request); err != nil {
		panic(err)
	}
	if request.DevEUI == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Request for devEUI: ", request.DevEUI)
	if _, err := d.Read(request.DevEUI); err == nil {
		log.Println("devEUI already taken: ", request.DevEUI)
		w.WriteHeader(422)
		return
	}

	if err := d.Write(request.DevEUI, []byte(time.Now().String())); err != nil {
		panic(err)
	}
	log.Println("Allocated devEUI: ", request.DevEUI)
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Simplest transform function: put all the data files into the base dir.
	flatTransform := func(s string) []string { return []string{} }

	// Initialize a new diskv store, rooted at "my-data-dir", with a 1MB cache.
	d = diskv.New(diskv.Options{
		BasePath:     "store",
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024,
	})

	port := ":8090"
	http.HandleFunc("/sensor-onboarding-sample", sensorOnboarding)
	log.Println("Listening on port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}
}
