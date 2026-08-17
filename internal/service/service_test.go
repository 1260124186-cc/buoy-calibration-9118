package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"buoy-calibration/internal/model"
	"buoy-calibration/internal/service"
	"buoy-calibration/internal/store"
)

func TestSealRunBuildsCorrectionReport(t *testing.T) {
	svc := service.New(store.NewMemory())
	profile, err := svc.CreateProfile(context.Background(), "temperature", 1.5, -2)
	if err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), profile.ID)
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	for _, sample := range []model.Sample{
		{ObservedAt: time.Date(2026, 8, 1, 10, 2, 0, 0, time.UTC), RawValue: 10},
		{ObservedAt: time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC), RawValue: 8},
		{ObservedAt: time.Date(2026, 8, 1, 10, 1, 0, 0, time.UTC), RawValue: 12},
	} {
		if _, err := svc.AddSample(context.Background(), run.ID, sample); err != nil {
			t.Fatalf("AddSample() error = %v", err)
		}
	}
	report, err := svc.SealRun(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("SealRun() error = %v", err)
	}
	if report.CorrectedMean != 13 {
		t.Fatalf("CorrectedMean = %v, want 13", report.CorrectedMean)
	}
	if !report.FirstObserved.Equal(time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("FirstObserved = %s", report.FirstObserved)
	}
}

func TestRunCannotSealBeforeEnoughSamples(t *testing.T) {
	svc := service.New(store.NewMemory())
	profile, _ := svc.CreateProfile(context.Background(), "pressure", 1, 0)
	run, _ := svc.StartRun(context.Background(), profile.ID)
	_, err := svc.SealRun(context.Background(), run.ID)
	if !errors.Is(err, model.ErrTooFewSamples) {
		t.Fatalf("SealRun() error = %v, want ErrTooFewSamples", err)
	}
}
