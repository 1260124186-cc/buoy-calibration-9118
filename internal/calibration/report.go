package calibration

import (
	"fmt"
	"sort"
	"time"

	"buoy-calibration/internal/model"
)

func Correct(profile model.Profile, raw float64) float64 {
	return raw*profile.Scale + profile.Bias
}

func ValidateProfile(profile model.Profile) error {
	if profile.Scale <= 0 {
		return fmt.Errorf("%w: scale must be positive", model.ErrInvalidScale)
	}
	return nil
}

func BuildReport(run model.Run, profile model.Profile, generatedAt time.Time) (model.Report, error) {
	if err := ValidateProfile(profile); err != nil {
		return model.Report{}, err
	}
	if len(run.Samples) < 3 {
		return model.Report{}, model.ErrTooFewSamples
	}

	samples := append([]model.Sample(nil), run.Samples...)
	sort.Slice(samples, func(i, j int) bool {
		return samples[i].ObservedAt.Before(samples[j].ObservedAt)
	})

	first := Correct(profile, samples[0].RawValue)
	report := model.Report{
		RunID:          run.ID,
		ProfileID:      profile.ID,
		SampleCount:    len(samples),
		CorrectedMin:   first,
		CorrectedMax:   first,
		FirstObserved:  samples[0].ObservedAt,
		LastObserved:   samples[len(samples)-1].ObservedAt,
		GeneratedAtUTC: generatedAt.UTC(),
	}

	var total float64
	for _, sample := range samples {
		corrected := Correct(profile, sample.RawValue)
		if corrected < report.CorrectedMin {
			report.CorrectedMin = corrected
		}
		if corrected > report.CorrectedMax {
			report.CorrectedMax = corrected
		}
		total += corrected
	}
	report.CorrectedMean = total / float64(len(samples))
	return report, nil
}
