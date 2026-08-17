package store

import (
	"context"
	"strings"
	"sync"

	"buoy-calibration/internal/model"
)

type Memory struct {
	mu       sync.RWMutex
	profiles map[string]model.Profile
	runs     map[string]model.Run
}

func (m *Memory) HasSensor(ctx context.Context, sensor string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	normalized := strings.ToLower(strings.TrimSpace(sensor))
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, profile := range m.profiles {
		if strings.ToLower(profile.SensorName) == normalized {
			return true, nil
		}
	}
	return false, nil
}

func NewMemory() *Memory {
	return &Memory{
		profiles: make(map[string]model.Profile),
		runs:     make(map[string]model.Run),
	}
}

func (m *Memory) CreateProfile(ctx context.Context, profile model.Profile) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profiles[profile.ID] = profile
	return nil
}

func (m *Memory) Profile(ctx context.Context, id string) (model.Profile, error) {
	if err := ctx.Err(); err != nil {
		return model.Profile{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	profile, ok := m.profiles[id]
	if !ok {
		return model.Profile{}, model.ErrNotFound
	}
	return profile, nil
}

func (m *Memory) CreateRun(ctx context.Context, run model.Run) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID] = cloneRun(run)
	return nil
}

func (m *Memory) Run(ctx context.Context, id string) (model.Run, error) {
	if err := ctx.Err(); err != nil {
		return model.Run{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	run, ok := m.runs[id]
	if !ok {
		return model.Run{}, model.ErrNotFound
	}
	return cloneRun(run), nil
}

func (m *Memory) UpdateRun(ctx context.Context, run model.Run) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.runs[run.ID]; !ok {
		return model.ErrNotFound
	}
	m.runs[run.ID] = cloneRun(run)
	return nil
}

func cloneRun(run model.Run) model.Run {
	run.Samples = append([]model.Sample(nil), run.Samples...)
	if run.Report != nil {
		report := *run.Report
		run.Report = &report
	}
	return run
}
