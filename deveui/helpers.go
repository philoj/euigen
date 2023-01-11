package deveui

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

const (
	allowedChars = "ABCDEF0123456789"
	devEUISize   = 5 // TODO change this or move to env
)

func generateHexString(length int) (string, error) {
	// TODO better generation
	// TODO batch-wise uniqueness
	max := big.NewInt(int64(len(allowedChars)))
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = allowedChars[n.Int64()]
	}
	return string(b), nil
}

func generateEUI() string {
	eui, err := generateHexString(devEUISize)
	if err != nil {
		panic(err)
	}
	return eui
}

func requestNewEUI(eui string) error {
	req, err := json.Marshal(map[string]string{
		"deveui": eui,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post("http://localhost:8090/sensor-onboarding-sample", "application/json", bytes.NewReader(req)) // TODO move url to env
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}
