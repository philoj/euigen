package deveuigen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/buntdb"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	maxConcurrentRequests = 10 // TODO move to env
	failureThreshold      = 5  // TODO move to env
	//apiUrl                = "http://localhost:8090/sensor-onboarding-sample" // mock server is in examples dir
	apiUrl = "https://europe-west1-machinemax-dev-d524.cloudfunctions.net/sensor-onboarding-sample" // TODO move to env
)

type result struct {
	devEUI  string
	id      uint64
	success bool
}

func CreateDevEUIs(batchSize int, resume bool) ([]string, error) {
	requests := make(chan uint64, maxConcurrentRequests)
	results := make(chan result, maxConcurrentRequests)
	for i := 0; i < maxConcurrentRequests; i++ {
		go handleServerCommunication(requests, results)
	}
	success, _ := monitorProgress(requests, results, batchSize, resume)
	//log.Println("Success: ", success, "Failure: ", failure)
	log.Printf("DevEUIs:\n%s\n", strings.Join(success, "\n"))
	var err error = nil
	if len(success) < batchSize {
		err = fmt.Errorf("aborted")
	}
	return success, err
}

func monitorProgress(requests chan<- uint64, results <-chan result, batchSize int, resume bool) (succeeded, failed []string) {
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

	store := NewIdStore()
	if len(ids) > 0 {
		if resume {
			succeeded = ids
			store.resetStore(succeeded)
			fmt.Printf("Previous incomplete run found. Resuming...\n")
		} else if err = discardPreviousRun(db); err != nil {
			panic(err)
		}
	}
	bar := progressbar.Default(int64(batchSize), "Generating")
	bar.Add(len(succeeded))
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
			bar.Describe("Aborting")
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
			}
		}
		results <- res
	}
}

func requestNewEUI(eui string) (*http.Response, error) {
	req, err := json.Marshal(map[string]string{
		"deveuigen": eui,
	})
	if err != nil {
		return nil, err
	}
	return http.Post(apiUrl, "application/json", bytes.NewReader(req)) // TODO move url to env
}
