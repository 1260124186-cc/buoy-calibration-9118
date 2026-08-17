package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"buoy-calibration/internal/api"
	"buoy-calibration/internal/service"
	"buoy-calibration/internal/store"
)

func TestCreateProfileRejectsZeroScale(t *testing.T) {
	server := api.NewServer(service.New(store.NewMemory())).Routes()
	request := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBufferString(
		`{"sensor_name":"salinity","scale":0,"bias":1}`,
	))
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected validation message")
	}
}

func TestSampleEndpointRejectsDuplicateObservationTime(t *testing.T) {
	server := api.NewServer(service.New(store.NewMemory())).Routes()
	profileID := createProfile(t, server)
	runID := startRun(t, server, profileID)
	at := time.Date(2026, 8, 17, 7, 0, 0, 0, time.UTC)

	for index, value := range []float64{8.1, 8.2} {
		request := httptest.NewRequest(http.MethodPost, "/runs/"+runID+"/samples", bytes.NewBufferString(
			`{"observed_at":"`+at.Format(time.RFC3339)+`","raw_value":`+formatFloat(value)+`}`,
		))
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if index == 0 && response.Code != http.StatusOK {
			t.Fatalf("first sample status = %d, want %d", response.Code, http.StatusOK)
		}
		if index == 1 && response.Code != http.StatusConflict {
			t.Fatalf("second sample status = %d, want %d", response.Code, http.StatusConflict)
		}
	}
}

func createProfile(t *testing.T, server http.Handler) string {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBufferString(
		`{"sensor_name":"conductivity","scale":1,"bias":0}`,
	))
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	return body.ID
}

func startRun(t *testing.T, server http.Handler, profileID string) string {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/runs", bytes.NewBufferString(
		`{"profile_id":"`+profileID+`"}`,
	))
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode run: %v", err)
	}
	return body.ID
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
