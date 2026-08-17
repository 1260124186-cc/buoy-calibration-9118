package store_test

import (
	"context"
	"testing"
	"time"

	"buoy-calibration/internal/model"
	"buoy-calibration/internal/store"
)

func TestReturnedRunDoesNotShareSampleStorage(t *testing.T) {
	repo := store.NewMemory()
	run := model.Run{
		ID:        "run-1",
		ProfileID: "profile-1",
		State:     model.RunCollecting,
		Samples:   make([]model.Sample, 1, 4),
	}
	run.Samples[0] = model.Sample{ObservedAt: time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC), RawValue: 7}
	if err := repo.CreateRun(context.Background(), run); err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}
	snapshot, _ := repo.Run(context.Background(), run.ID)
	updated, _ := repo.Run(context.Background(), run.ID)
	updated.Samples[0].RawValue = 99
	if err := repo.UpdateRun(context.Background(), updated); err != nil {
		t.Fatalf("UpdateRun() error = %v", err)
	}
	if snapshot.Samples[0].RawValue != 7 {
		t.Fatalf("snapshot raw value = %v, want 7", snapshot.Samples[0].RawValue)
	}
}
