package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound             = errors.New("record not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidState         = errors.New("invalid run state")
	ErrTooFewSamples        = errors.New("too few samples")
	ErrDuplicateObservation = errors.New("duplicate observation")
)

type DuplicateObservationError struct {
	ObservedAt time.Time
}

func (e *DuplicateObservationError) Error() string {
	return "duplicate observation at " + e.ObservedAt.UTC().Format(time.RFC3339Nano)
}

func (e *DuplicateObservationError) Is(target error) bool {
	return target == ErrDuplicateObservation
}

type Profile struct {
	ID         string  `json:"id"`
	SensorName string  `json:"sensor_name"`
	Scale      float64 `json:"scale"`
	Bias       float64 `json:"bias"`
}

type Sample struct {
	ObservedAt time.Time `json:"observed_at"`
	RawValue   float64   `json:"raw_value"`
}

type RunState string

const (
	RunCollecting RunState = "collecting"
	RunSealed     RunState = "sealed"
)

type Run struct {
	ID        string   `json:"id"`
	ProfileID string   `json:"profile_id"`
	State     RunState `json:"state"`
	Samples   []Sample `json:"samples"`
	Report    *Report  `json:"report,omitempty"`
}

type Report struct {
	RunID          string    `json:"run_id"`
	ProfileID      string    `json:"profile_id"`
	SampleCount    int       `json:"sample_count"`
	CorrectedMin   float64   `json:"corrected_min"`
	CorrectedMax   float64   `json:"corrected_max"`
	CorrectedMean  float64   `json:"corrected_mean"`
	FirstObserved  time.Time `json:"first_observed"`
	LastObserved   time.Time `json:"last_observed"`
	GeneratedAtUTC time.Time `json:"generated_at_utc"`
}
