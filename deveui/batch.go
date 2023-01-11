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
	id      uint64
	success bool
	isTaken bool
}

func CreateDevEUIs(batchSize int, resume bool) ([]string, error) {
	requests := make(chan uint64, maxConcurrentRequests)
	results := make(chan result, maxConcurrentRequests)
	store := NewIdStore()
	for i := 0; i < maxConcurrentRequests; i++ {
		go handleServerCommunication(requests, results)
	}
	success, failure := monitorProgress(store, requests, results, batchSize, resume)
	log.Println("Success: ", success, "Failure: ", failure)
	var err error = nil
	if len(success) < batchSize {
		err = fmt.Errorf("aborted")
	}
	return success, err
}

func monitorProgress(store *IdStore, requests chan<- uint64, results <-chan result, batchSize int, resume bool) (succeeded, failed []string) {
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
		if resume {
			succeeded = ids
			bar.Add(len(succeeded))
			store.resetStore(succeeded)
		} else if err = discardPreviousRun(db); err != nil {
			panic(err)
		}
	}
	activeRequestCount := 0
	checkResult := func(r result) {
		store.updateStore(r)
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
		case requests <- store.generate():
			activeRequestCount++
		case <-sigs:
			wrapUp()
		}
	}
}

func handleServerCommunication(requests <-chan uint64, results chan<- result) {
	for {
		id, open := <-requests
		if !open {
			return
		}
		res := result{id: id, devEUI: fmt.Sprintf("%016X", id)}
		resp, err := requestNewEUI(res.devEUI)
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
