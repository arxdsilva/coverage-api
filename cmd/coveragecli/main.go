package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type packageCoverage struct {
	ImportPath      string  `json:"importPath"`
	CoveragePercent float64 `json:"coveragePercent"`
}

type ingestPayload struct {
	ProjectKey           string            `json:"projectKey"`
	ProjectName          string            `json:"projectName,omitempty"`
	ProjectGroup         *string           `json:"projectGroup,omitempty"`
	DefaultBranch        string            `json:"defaultBranch,omitempty"`
	Branch               string            `json:"branch"`
	CommitSHA            string            `json:"commitSha"`
	Author               string            `json:"author,omitempty"`
	TriggerType          string            `json:"triggerType"`
	RunTimestamp         string            `json:"runTimestamp"`
	TotalCoveragePercent float64           `json:"totalCoveragePercent"`
	Packages             []packageCoverage `json:"packages"`
}

type integrationPayload struct {
	ProjectKey    string         `json:"projectKey"`
	ProjectName   string         `json:"projectName,omitempty"`
	ProjectGroup  *string        `json:"projectGroup,omitempty"`
	DefaultBranch string         `json:"defaultBranch,omitempty"`
	Branch        string         `json:"branch"`
	CommitSHA     string         `json:"commitSha"`
	Author        string         `json:"author,omitempty"`
	TriggerType   string         `json:"triggerType"`
	RunTimestamp  string         `json:"runTimestamp"`
	Environment   *string        `json:"environment,omitempty"`
	GinkgoReport  map[string]any `json:"ginkgoReport"`
}

type integrationUploadResponse struct {
	Run struct {
		Status          string  `json:"status"`
		PassRatePercent float64 `json:"passRatePercent"`
	} `json:"run"`
	Comparison struct {
		DeltaPercent *float64 `json:"deltaPercent"`
	} `json:"comparison"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "integration-upload" {
		runIntegrationUpload(os.Args[2:])
		return
	}

	runCoverageUpload(os.Args[1:])
}

func runCoverageUpload(args []string) {
	fs := flag.NewFlagSet("coveragecli", flag.ExitOnError)
	coverprofile := fs.String("coverprofile", "coverage.out", "Path to go coverage profile")
	out := fs.String("out", "coverage-upload.json", "Path to output JSON payload file")
	projectKey := fs.String("project-key", "github.com/arxdsilva/opencoverage", "Project key")
	projectName := fs.String("project-name", "coverage-api", "Project display name")
	projectGroup := fs.String("project-group", "", "Project group (optional)")
	defaultBranch := fs.String("default-branch", "main", "Default branch")
	branch := fs.String("branch", envOrDefault("COVERAGE_BRANCH", "main"), "Current branch")
	commitSHA := fs.String("commit-sha", envOrDefault("COVERAGE_COMMIT_SHA", "local"), "Commit SHA")
	author := fs.String("author", envOrDefault("COVERAGE_AUTHOR", "local"), "Author")
	triggerType := fs.String("trigger-type", "manual", "Trigger type: push|pr|manual")
	upload := fs.Bool("upload", false, "Upload payload to API")
	apiURL := fs.String("api-url", envOrDefault("API_URL", "http://localhost:8080/v1/coverage-runs"), "Coverage API URL")
	apiKey := fs.String("api-key", os.Getenv("API_KEY"), "API key value")
	apiKeyHeader := fs.String("api-key-header", "X-API-Key", "API key header name")
	if err := fs.Parse(args); err != nil {
		exitErr("parse flags", err)
	}

	total, packages, err := parseCoverage(*coverprofile)
	if err != nil {
		exitErr("parse coverage", err)
	}
	if len(packages) == 0 {
		exitErr("parse coverage", fmt.Errorf("no package coverage entries found"))
	}

	var group *string
	if *projectGroup != "" {
		group = projectGroup
	}

	payload := ingestPayload{
		ProjectKey:           *projectKey,
		ProjectName:          *projectName,
		ProjectGroup:         group,
		DefaultBranch:        *defaultBranch,
		Branch:               *branch,
		CommitSHA:            *commitSHA,
		Author:               *author,
		TriggerType:          *triggerType,
		RunTimestamp:         time.Now().UTC().Format(time.RFC3339),
		TotalCoveragePercent: total,
		Packages:             packages,
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		exitErr("marshal payload", err)
	}

	if err := os.WriteFile(*out, body, 0o644); err != nil {
		exitErr("write payload file", err)
	}
	fmt.Printf("payload written: %s\n", *out)

	if !*upload {
		return
	}
	if strings.TrimSpace(*apiKey) == "" {
		exitErr("upload", fmt.Errorf("api key is required when -upload is set (use -api-key or API_KEY env var)"))
	}

	status, respBody, err := uploadPayload(*apiURL, *apiKeyHeader, *apiKey, body)
	if err != nil {
		exitErr("upload", err)
	}
	fmt.Printf("upload status: %d\n", status)
	fmt.Printf("upload response: %s\n", strings.TrimSpace(string(respBody)))
}

func runIntegrationUpload(args []string) {
	fs := flag.NewFlagSet("integration-upload", flag.ExitOnError)
	reportPath := fs.String("ginkgo-report", "", "Path to Ginkgo JSON report")
	apiURL := fs.String("api-url", envOrDefault("API_URL", "http://localhost:8080/v1/integration-test-runs"), "Integration test API URL")
	apiKey := fs.String("api-key", os.Getenv("API_KEY"), "API key value")
	apiKeyHeader := fs.String("api-key-header", "X-API-Key", "API key header name")
	projectKey := fs.String("project-key", envOrDefault("COVERAGE_PROJECT_KEY", "github.com/arxdsilva/opencoverage"), "Project key")
	projectName := fs.String("project-name", envOrDefault("COVERAGE_PROJECT_NAME", "coverage-api"), "Project display name")
	projectGroup := fs.String("project-group", "", "Project group (optional)")
	defaultBranch := fs.String("default-branch", envOrDefault("COVERAGE_DEFAULT_BRANCH", "main"), "Default branch")
	branch := fs.String("branch", envOrDefault("COVERAGE_BRANCH", "main"), "Current branch")
	commitSHA := fs.String("commit-sha", envOrDefault("COVERAGE_COMMIT_SHA", "local"), "Commit SHA")
	author := fs.String("author", envOrDefault("COVERAGE_AUTHOR", "local"), "Author")
	triggerType := fs.String("trigger-type", "manual", "Trigger type: push|pr|manual")
	environment := fs.String("environment", "", "Environment: test|stage|prod (optional)")
	runTimestamp := fs.String("run-timestamp", time.Now().UTC().Format(time.RFC3339), "Run timestamp (RFC3339)")
	if err := fs.Parse(args); err != nil {
		exitErr("parse flags", err)
	}

	if strings.TrimSpace(*reportPath) == "" {
		exitErr("validate input", fmt.Errorf("-ginkgo-report is required"))
	}
	if strings.TrimSpace(*apiKey) == "" {
		exitErr("validate input", fmt.Errorf("-api-key is required (or API_KEY env var)"))
	}
	if _, err := time.Parse(time.RFC3339, *runTimestamp); err != nil {
		exitErr("validate input", fmt.Errorf("run timestamp must be RFC3339: %w", err))
	}

	rawReport, err := os.ReadFile(*reportPath)
	if err != nil {
		exitErr("read ginkgo report", err)
	}

	var report map[string]any
	if err := json.Unmarshal(rawReport, &report); err != nil {
		exitErr("parse ginkgo report json", err)
	}

	var group *string
	if *projectGroup != "" {
		group = projectGroup
	}

	var env *string
	if *environment != "" {
		if *environment != "test" && *environment != "stage" && *environment != "prod" {
			exitErr("validate input", fmt.Errorf("-environment must be one of: test, stage, prod"))
		}
		env = environment
	}

	payload := integrationPayload{
		ProjectKey:    *projectKey,
		ProjectName:   *projectName,
		ProjectGroup:  group,
		DefaultBranch: *defaultBranch,
		Branch:        *branch,
		CommitSHA:     *commitSHA,
		Author:        *author,
		TriggerType:   *triggerType,
		RunTimestamp:  *runTimestamp,
		Environment:   env,
		GinkgoReport:  normalizeReport(report),
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		exitErr("marshal payload", err)
	}

	status, respBody, err := uploadPayload(*apiURL, *apiKeyHeader, *apiKey, body)
	if err != nil {
		exitErr("upload integration report", err)
	}

	fmt.Printf("upload status: %d\n", status)
	fmt.Printf("upload response: %s\n", strings.TrimSpace(string(respBody)))

	var parsed integrationUploadResponse
	if err := json.Unmarshal(respBody, &parsed); err == nil {
		delta := "-"
		if parsed.Comparison.DeltaPercent != nil {
			delta = fmt.Sprintf("%.2f", *parsed.Comparison.DeltaPercent)
		}
		fmt.Printf("summary: status=%s passRatePercent=%.2f deltaPercent=%s\n", parsed.Run.Status, parsed.Run.PassRatePercent, delta)
	}

	if status >= http.StatusBadRequest {
		exitErr("upload integration report", fmt.Errorf("server returned status %d", status))
	}
}

func normalizeReport(raw map[string]any) map[string]any {
	result := make(map[string]any)
	result["suiteDescription"] = firstString(raw, "suiteDescription", "SuiteDescription")
	result["suitePath"] = firstString(raw, "suitePath", "SuitePath")
	result["ginkgoVersion"] = firstString(raw, "ginkgoVersion", "GinkgoVersion")

	specReports := firstSlice(raw, "specReports", "SpecReports")
	normalizedSpecs := make([]map[string]any, 0, len(specReports))
	for _, item := range specReports {
		specMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		normalized := map[string]any{
			"leafNodeText":            firstString(specMap, "leafNodeText", "LeafNodeText"),
			"containerHierarchyTexts": firstSlice(specMap, "containerHierarchyTexts", "ContainerHierarchyTexts"),
			"state":                   firstString(specMap, "state", "State"),
			"runTime":                 firstFloat(specMap, "runTime", "RunTime"),
		}

		failureVal := firstMap(specMap, "failure", "Failure")
		if len(failureVal) > 0 {
			failure := map[string]any{
				"message": firstString(failureVal, "message", "Message"),
			}
			locationVal := firstMap(failureVal, "location", "Location")
			if len(locationVal) > 0 {
				failure["location"] = map[string]any{
					"fileName":   firstString(locationVal, "fileName", "FileName"),
					"lineNumber": int(firstFloat(locationVal, "lineNumber", "LineNumber")),
				}
			}
			normalized["failure"] = failure
		}

		normalizedSpecs = append(normalizedSpecs, normalized)
	}

	result["specReports"] = normalizedSpecs
	return result
}

func firstString(src map[string]any, keys ...string) string {
	for _, key := range keys {
		if raw, ok := src[key]; ok {
			if value, ok := raw.(string); ok {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func firstFloat(src map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if raw, ok := src[key]; ok {
			switch v := raw.(type) {
			case float64:
				return v
			case int:
				return float64(v)
			case json.Number:
				f, err := v.Float64()
				if err == nil {
					return f
				}
			}
		}
	}
	return 0
}

func firstSlice(src map[string]any, keys ...string) []any {
	for _, key := range keys {
		if raw, ok := src[key]; ok {
			if value, ok := raw.([]any); ok {
				return value
			}
		}
	}
	return nil
}

func firstMap(src map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if raw, ok := src[key]; ok {
			if value, ok := raw.(map[string]any); ok {
				return value
			}
		}
	}
	return nil
}

func parseCoverage(profilePath string) (float64, []packageCoverage, error) {
	cmd := exec.Command("go", "tool", "cover", "-func", profilePath)
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return 0, nil, fmt.Errorf("go tool cover failed: %s", string(ee.Stderr))
		}
		return 0, nil, err
	}

	lineRe := regexp.MustCompile(`^(.+):[0-9]+:\s+\S+\s+([0-9]+(?:\.[0-9]+)?)%$`)
	totalRe := regexp.MustCompile(`^total:\s+\(statements\)\s+([0-9]+(?:\.[0-9]+)?)%$`)

	type agg struct {
		sum   float64
		count int
	}
	byPackage := map[string]*agg{}
	var total float64
	foundTotal := false

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if m := totalRe.FindStringSubmatch(line); len(m) == 2 {
			t, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				return 0, nil, fmt.Errorf("parse total coverage: %w", err)
			}
			total = t
			foundTotal = true
			continue
		}

		m := lineRe.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		filePath := m[1]
		percent, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			return 0, nil, fmt.Errorf("parse package coverage: %w", err)
		}
		pkg := path.Dir(filePath)
		if byPackage[pkg] == nil {
			byPackage[pkg] = &agg{}
		}
		byPackage[pkg].sum += percent
		byPackage[pkg].count++
	}

	if !foundTotal {
		return 0, nil, fmt.Errorf("total coverage line not found in cover output")
	}

	pkgs := make([]packageCoverage, 0, len(byPackage))
	for pkg, a := range byPackage {
		if a.count == 0 {
			continue
		}
		pkgs = append(pkgs, packageCoverage{
			ImportPath:      pkg,
			CoveragePercent: round2(a.sum / float64(a.count)),
		})
	}
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].ImportPath < pkgs[j].ImportPath })

	return round2(total), pkgs, nil
}

func uploadPayload(url, apiKeyHeader, apiKey string, body []byte) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(apiKeyHeader, apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, respBody, nil
}

func envOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func exitErr(stage string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", stage, err)
	os.Exit(1)
}
