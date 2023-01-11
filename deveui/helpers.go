package deveui

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func requestNewEUI(eui string) (*http.Response, error) {
	req, err := json.Marshal(map[string]string{
		"deveui": eui,
	})
	if err != nil {
		return nil, err
	}
	return http.Post("http://localhost:8090/sensor-onboarding-sample", "application/json", bytes.NewReader(req)) // TODO move url to env
}
