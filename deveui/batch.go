package deveui

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/buntdb"
	"log"
	"net/http"
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
	isTaken bool
}

func CreateDevEUIs(batchSize int, discard bool) ([]string, error) {
	requests := make(chan string, maxConcurrentRequests)
	results := make(chan result, maxConcurrentRequests)
	for i := 0; i < maxConcurrentRequests; i++ {
		go handleServerCommunication(requests, results)
	}
	success, failure := monitorProgress(requests, results, batchSize, discard)
	log.Println("Success: ", success, "Failure: ", failure)
	var err error = nil
	if len(success) < batchSize {
		err = fmt.Errorf("aborted")
	}
	return success, err
}

func monitorProgress(requests chan string, results chan result, batchSize int, discard bool) (succeeded, failed []string) {
	defer close(requests)

	db, err := buntdb.Open("data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ids, err := checkForPreviousRun(db)
	if err != nil {
		panic(err)
	}

	bar := progressbar.Default(int64(batchSize), "generating")
	if len(ids) > 0 {
		if discard {
			if err = discardPreviousRun(db); err != nil {
				panic(err)
			}
		} else {
			succeeded = ids
			bar.Add(len(succeeded))
		}
	}
	activeRequestCount := 0
	checkResult := func(r result) {
		activeRequestCount--
		if r.success {
			succeeded = append(succeeded, r.devEUI)
			bar.Add(1)
		} else {
			failed = append(failed, r.devEUI)
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	abort := false
	wrapUp := func() {
		if !abort {
			bar.Describe("aborting")
			abort = true
		}
	}

	for {
		if len(succeeded) >= batchSize || (abort && activeRequestCount == 0) || len(failed) >= failureThreshold*batchSize {
			// Save and exit
			if err = saveCurrentRun(db, len(succeeded) >= batchSize, succeeded); err != nil {
				panic(err)
			}
			return
		}
		if abort || (activeRequestCount > 0 && batchSize > 0 &&
			(activeRequestCount == maxConcurrentRequests || activeRequestCount+len(succeeded) == batchSize)) {
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
		eui, open := <-requests
		if !open {
			return
		}
		res := result{devEUI: eui}
		resp, err := requestNewEUI(eui)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				res.success = true
			} else if resp.StatusCode == 422 {
				res.isTaken = true
			}
		}
		results <- res
	}
}
