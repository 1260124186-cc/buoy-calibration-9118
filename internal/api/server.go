package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"buoy-calibration/internal/model"
	"buoy-calibration/internal/service"
)

type Server struct {
	service *service.Service
}

func NewServer(svc *service.Service) *Server {
	return &Server{service: svc}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /profiles", s.createProfile)
	mux.HandleFunc("POST /runs", s.startRun)
	mux.HandleFunc("POST /runs/{id}/samples", s.addSample)
	mux.HandleFunc("POST /runs/{id}/seal", s.sealRun)
	mux.HandleFunc("GET /runs/{id}/report", s.getReport)
	return mux
}

type profileRequest struct {
	SensorName string  `json:"sensor_name"`
	Scale      float64 `json:"scale"`
	Bias       float64 `json:"bias"`
}

type runRequest struct {
	ProfileID string `json:"profile_id"`
}

type sampleRequest struct {
	ObservedAt time.Time `json:"observed_at"`
	RawValue   float64   `json:"raw_value"`
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) createProfile(w http.ResponseWriter, r *http.Request) {
	var input profileRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	profile, err := s.service.CreateProfile(r.Context(), input.SensorName, input.Scale, input.Bias)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, profile)
}

func (s *Server) startRun(w http.ResponseWriter, r *http.Request) {
	var input runRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	run, err := s.service.StartRun(r.Context(), input.ProfileID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

func (s *Server) addSample(w http.ResponseWriter, r *http.Request) {
	var input sampleRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	run, err := s.service.AddSample(r.Context(), r.PathValue("id"), model.Sample{
		ObservedAt: input.ObservedAt,
		RawValue:   input.RawValue,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) sealRun(w http.ResponseWriter, r *http.Request) {
	report, err := s.service.SealRun(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) getReport(w http.ResponseWriter, r *http.Request) {
	report, err := s.service.Report(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrInvalidState), errors.Is(err, model.ErrTooFewSamples):
		status = http.StatusConflict
	case errors.Is(err, model.ErrDuplicateObservation):
		status = http.StatusConflict
	case errors.Is(err, contextCanceled()):
		status = http.StatusRequestTimeout
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func contextCanceled() error {
	return errors.New(strings.TrimSpace("context canceled"))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
