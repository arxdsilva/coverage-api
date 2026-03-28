package domain

import "time"

type CoverageRun struct {
	ID                   string
	ProjectID            string
	Branch               string
	CommitSHA            string
	Author               string
	TriggerType          string
	RunTimestamp         time.Time
	TotalCoveragePercent float64
	CreatedAt            time.Time
}

type PackageCoverage struct {
	ID                string
	RunID             string
	PackageImportPath string
	CoveragePercent   float64
}
