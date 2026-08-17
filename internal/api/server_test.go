package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestCanceledCreateRequestReturnsRequestTimeout(t *testing.T) {
	server := api.NewServer(service.New(store.NewMemory())).Routes()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewBufferString(
		`{"sensor_name":"salinity","scale":1,"bias":0}`,
	)).WithContext(ctx)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusRequestTimeout {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusRequestTimeout)
	}
}
