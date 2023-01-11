package deveui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func requestNewEUI(eui string) error {
	req, err := json.Marshal(map[string]string{
		"deveui": eui,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post("http://localhost:8090/sensor-onboarding-sample", "application/json", bytes.NewReader(req))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

func generateEUI() string {
	eui, err := generateHexString(5)
	if err != nil {
		panic(err)
	}
	return eui
}

func spawnNewRequests(requests chan string, results chan result, batchSize int) (success, failure []string) {
	defer close(requests)
	activeRequestCount := 0
	checkResult := func(r result) {
		activeRequestCount--
		if r.success {
			success = append(success, r.devEUI)
		} else {
			failure = append(failure, r.devEUI)
		}
	}

	for {
		if len(success) >= batchSize || len(failure) >= failureThreshold*batchSize {
			// Exit
			return
		}
		if activeRequestCount > 0 && batchSize > 0 &&
			(activeRequestCount == maxConcurrentRequests || activeRequestCount+len(success) == batchSize) {
			// Wait for active requests to complete before spawning new requests
			checkResult(<-results)
			continue
		}

		// Multiplex b/w requests and results
		select {
		case r := <-results:
			checkResult(r)
		case requests <- generateEUI():
			activeRequestCount++
		}
	}
}

func handleRequests(requests chan string, results chan result) {
	for {
		eui := <-requests
		results <- result{
			devEUI:  eui,
			success: requestNewEUI(eui) == nil,
		}
	}
}

const maxConcurrentRequests = 5
const failureThreshold = 5

type result struct {
	devEUI  string
	success bool
}

func BatchRequest(batchSize int) ([]string, error) {
	requests := make(chan string, maxConcurrentRequests)
	results := make(chan result, maxConcurrentRequests)
	go handleRequests(requests, results)
	success, failure := spawnNewRequests(requests, results, batchSize)
	log.Println("Success: ", success, "Failure: ", failure)
	var err error = nil
	if len(success) < batchSize {
		err = fmt.Errorf("aborted when failures exceeded threshold")
	}
	return success, err
}
