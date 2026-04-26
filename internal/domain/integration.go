package domain

import "time"

type IntegrationRunStatus string

const (
	IntegrationRunStatusPassed IntegrationRunStatus = "passed"
	IntegrationRunStatusFailed IntegrationRunStatus = "failed"
)

type IntegrationSpecState string

const (
	IntegrationSpecStatePassed  IntegrationSpecState = "passed"
	IntegrationSpecStateFailed  IntegrationSpecState = "failed"
	IntegrationSpecStateSkipped IntegrationSpecState = "skipped"
	IntegrationSpecStatePending IntegrationSpecState = "pending"
	IntegrationSpecStateFlaky   IntegrationSpecState = "flaky"
)

type IntegrationTestRun struct {
	ID               string
	ProjectID        string
	Branch           string
	CommitSHA        string
	Author           string
	TriggerType      string
	RunTimestamp     time.Time
	GinkgoVersion    string
	SuiteDescription string
	SuitePath        string
	TotalSpecs       int
	PassedSpecs      int
	FailedSpecs      int
	SkippedSpecs     int
	FlakedSpecs      int
	PendingSpecs     int
	Interrupted      bool
	TimedOut         bool
	DurationMS       int64
	Status           IntegrationRunStatus
	CreatedAt        time.Time
}

type IntegrationSpecResult struct {
	ID                  string
	IntegrationRunID    string
	SpecPath            string
	LeafNodeText        string
	State               IntegrationSpecState
	DurationMS          int64
	FailureMessage      *string
	FailureLocationFile *string
	FailureLocationLine *int
}

func EvaluateIntegrationRunStatus(failedSpecs int, interrupted bool, timedOut bool) IntegrationRunStatus {
	if failedSpecs == 0 && !interrupted && !timedOut {
		return IntegrationRunStatusPassed
	}
	return IntegrationRunStatusFailed
}
