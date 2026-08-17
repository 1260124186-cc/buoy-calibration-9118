package service

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"buoy-calibration/internal/calibration"
	"buoy-calibration/internal/model"
)

type Repository interface {
	CreateProfile(context.Context, model.Profile) error
	Profile(context.Context, string) (model.Profile, error)
	CreateRun(context.Context, model.Run) error
	Run(context.Context, string) (model.Run, error)
	UpdateRun(context.Context, model.Run) error
}

type Service struct {
	repo Repository
	seq  atomic.Uint64
	now  func() time.Time
}

func New(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) CreateProfile(ctx context.Context, sensor string, scale, bias float64) (model.Profile, error) {
	if strings.TrimSpace(sensor) == "" || scale == 0 {
		return model.Profile{}, fmt.Errorf("%w: sensor name and non-zero scale are required", model.ErrInvalidInput)
	}
	profile := model.Profile{
		ID:         s.nextID("profile"),
		SensorName: strings.TrimSpace(sensor),
		Scale:      scale,
		Bias:       bias,
	}
	if err := s.repo.CreateProfile(ctx, profile); err != nil {
		return model.Profile{}, err
	}
	return profile, nil
}

func (s *Service) StartRun(ctx context.Context, profileID string) (model.Run, error) {
	if _, err := s.repo.Profile(ctx, profileID); err != nil {
		return model.Run{}, err
	}
	run := model.Run{ID: s.nextID("run"), ProfileID: profileID, State: model.RunCollecting}
	if err := s.repo.CreateRun(ctx, run); err != nil {
		return model.Run{}, err
	}
	return run, nil
}

func (s *Service) AddSample(ctx context.Context, runID string, sample model.Sample) (model.Run, error) {
	if sample.ObservedAt.IsZero() {
		return model.Run{}, fmt.Errorf("%w: observed_at is required", model.ErrInvalidInput)
	}
	run, err := s.repo.Run(ctx, runID)
	if err != nil {
		return model.Run{}, err
	}
	if run.State != model.RunCollecting {
		return model.Run{}, fmt.Errorf("%w: run %s is already sealed", model.ErrInvalidState, run.ID)
	}
	// 把待写入样本纳入序列一起校验，拦截同一观测时刻的重复写入
	updated := append(run.Samples, sample)
	if err := calibration.ValidateSampleSequence(updated); err != nil {
		return model.Run{}, err
	}
	run.Samples = updated
	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return model.Run{}, err
	}
	return run, nil
}

func (s *Service) SealRun(ctx context.Context, runID string) (model.Report, error) {
	run, err := s.repo.Run(ctx, runID)
	if err != nil {
		return model.Report{}, err
	}
	if run.State != model.RunCollecting {
		return model.Report{}, fmt.Errorf("%w: run %s is not collecting", model.ErrInvalidState, run.ID)
	}
	profile, err := s.repo.Profile(ctx, run.ProfileID)
	if err != nil {
		return model.Report{}, err
	}
	report, err := calibration.BuildReport(run, profile, s.now())
	if err != nil {
		return model.Report{}, err
	}
	run.State = model.RunSealed
	run.Report = &report
	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return model.Report{}, err
	}
	return report, nil
}

func (s *Service) Report(ctx context.Context, runID string) (model.Report, error) {
	run, err := s.repo.Run(ctx, runID)
	if err != nil {
		return model.Report{}, err
	}
	if run.Report == nil {
		return model.Report{}, fmt.Errorf("%w: run %s has not been sealed", model.ErrInvalidState, run.ID)
	}
	return *run.Report, nil
}

func (s *Service) nextID(prefix string) string {
	return fmt.Sprintf("%s-%06d", prefix, s.seq.Add(1))
}
