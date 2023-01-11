package deveui

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	maxConcurrentRequests = 5 // TODO move to env
	failureThreshold      = 5 // TODO move to env
)

type result struct {
	devEUI  string
	success bool
}

func CreateDevEUIs(batchSize int) ([]string, error) {
	requests := make(chan string, maxConcurrentRequests)
	results := make(chan result, maxConcurrentRequests)
	for i := 0; i < maxConcurrentRequests; i++ {
		go handleServerCommunication(requests, results)
	}
	success, failure := monitorProgress(requests, results, batchSize)
	log.Println("Success: ", success, "Failure: ", failure)
	var err error = nil
	if len(success) < batchSize {
		err = fmt.Errorf("aborted")
	}
	return success, err
}

func monitorProgress(requests chan string, results chan result, batchSize int) (success, failure []string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
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

	abort := false
	wrapUp := func() {
		if !abort {
			log.Println("Aborting", activeRequestCount, batchSize-len(success))
			abort = true
		}
	}

	for {
		if len(success) >= batchSize || (abort && activeRequestCount == 0) || len(failure) >= failureThreshold*batchSize {
			// Exit
			return
		}
		if abort || (activeRequestCount > 0 && batchSize > 0 &&
			(activeRequestCount == maxConcurrentRequests || activeRequestCount+len(success) == batchSize)) {
			// Wait for active requests to complete without spawning new requests
			select {
			case r := <-results:
				checkResult(r)
			case <-sigs:
				wrapUp()
			}
			continue
		}

		// Multiplex b/w requests and results
		select {
		case r := <-results:
			checkResult(r)
		case requests <- generateEUI():
			activeRequestCount++
		case <-sigs:
			wrapUp()
		}
	}
}

func handleServerCommunication(requests chan string, results chan result) {
	for {
		eui := <-requests
		log.Println("Requesting", eui)
		results <- result{
			devEUI:  eui,
			success: requestNewEUI(eui) == nil,
		}
	}
}
