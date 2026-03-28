package domain

import "fmt"

const (
	DefaultThresholdPercent = 80.0
	DefaultBranch           = "main"
)

type Direction string

const (
	DirectionUp    Direction = "up"
	DirectionDown  Direction = "down"
	DirectionEqual Direction = "equal"
	DirectionNew   Direction = "new"
)

type ThresholdStatus string

const (
	ThresholdPassed ThresholdStatus = "passed"
	ThresholdFailed ThresholdStatus = "failed"
)

func ValidateCoveragePercent(v float64) error {
	if v < 0 || v > 100 {
		return ErrInvalidCoverage
	}
	return nil
}

func CompareCoverage(current float64, previous *float64) (*float64, Direction) {
	if previous == nil {
		return nil, DirectionNew
	}
	delta := current - *previous
	if delta > 0 {
		return &delta, DirectionUp
	}
	if delta < 0 {
		return &delta, DirectionDown
	}
	return &delta, DirectionEqual
}

func EvaluateThreshold(current float64, threshold float64) ThresholdStatus {
	if current >= threshold {
		return ThresholdPassed
	}
	return ThresholdFailed
}

func ValidateTriggerType(trigger string) error {
	switch trigger {
	case "push", "pr", "manual":
		return nil
	default:
		return fmt.Errorf("invalid triggerType: %s", trigger)
	}
}
