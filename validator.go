package citationevidence

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	requiredFields = []string{
		"prompt_id",
		"prompt_version",
		"answer_system",
		"run_timestamp",
		"brand_mentioned",
		"target_domain_cited",
		"cited_domains",
		"claim_summary",
		"evidence_artifact",
	}
	promptVersionPattern = regexp.MustCompile(`^\d+(\.\d+){0,2}([+-][a-zA-Z0-9.-]+)?$`)
)

// Observation is one answer-level event decoded from JSON or JSONL.
type Observation map[string]any

// Issue is one explicit validation finding attached to a record and field.
type Issue struct {
	Record   string `json:"record"`
	Field    string `json:"field"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// ParsedInput contains decoded observations and any line- or document-level
// parse errors. Parse errors do not prevent valid JSONL lines from being
// returned.
type ParsedInput struct {
	Records []Observation `json:"records"`
	Issues  []Issue       `json:"issues"`
}

// RecordResult summarizes the validation state of one observation.
type RecordResult struct {
	Index        int    `json:"index"`
	Label        string `json:"label"`
	Status       string `json:"status"`
	ErrorCount   int    `json:"error_count"`
	WarningCount int    `json:"warning_count"`
}

// Summary reports record counts without combining them into a score.
type Summary struct {
	Total       int `json:"total"`
	Ready       int `json:"ready"`
	Review      int `json:"review"`
	Invalid     int `json:"invalid"`
	ParseErrors int `json:"parse_errors"`
}

// Report is the complete machine-readable validation result.
type Report struct {
	GeneratedAt   string         `json:"generated_at"`
	TargetDomain  string         `json:"target_domain,omitempty"`
	Methodology   string         `json:"methodology"`
	Summary       Summary        `json:"summary"`
	RecordResults []RecordResult `json:"record_results"`
	Issues        []Issue        `json:"issues"`
}

// Parse accepts either a top-level JSON array of objects or newline-delimited
// JSON. Empty lines in JSONL input are ignored.
func Parse(data []byte) ParsedInput {
	raw := bytes.TrimSpace(data)
	if len(raw) == 0 {
		return ParsedInput{Issues: []Issue{issue("input", "input", "error", "No observation records were supplied.")}}
	}

	if raw[0] == '[' {
		var records []Observation
		if err := decodeJSON(raw, &records); err != nil {
			return ParsedInput{Issues: []Issue{issue("input", "input", "error", "Invalid JSON array: "+err.Error())}}
		}
		return ParsedInput{Records: records}
	}

	parsed := ParsedInput{}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var record Observation
		if err := decodeJSON(line, &record); err != nil {
			parsed.Issues = append(parsed.Issues, issue(
				fmt.Sprintf("line %d", lineNumber),
				"input",
				"error",
				"Invalid JSONL record: "+err.Error(),
			))
			continue
		}
		parsed.Records = append(parsed.Records, record)
	}
	if err := scanner.Err(); err != nil {
		parsed.Issues = append(parsed.Issues, issue("input", "input", "error", "Unable to scan JSONL input: "+err.Error()))
	}
	return parsed
}

// Validate checks parsed observations against explicit field, type, timestamp,
// URL, domain-consistency, and duplicate-run rules.
func Validate(parsed ParsedInput, targetDomain string) Report {
	return ValidateAt(parsed, targetDomain, time.Now().UTC())
}

// ValidateAt is equivalent to Validate but accepts a report timestamp. It is
// useful for reproducible tests and deterministic processing pipelines.
func ValidateAt(parsed ParsedInput, targetDomain string, generatedAt time.Time) Report {
	target := NormalizeDomain(targetDomain)
	report := Report{
		GeneratedAt:  generatedAt.UTC().Format(time.RFC3339),
		TargetDomain: target,
		Methodology:  "Explicit field and consistency checks; no proprietary visibility score.",
		Issues:       append([]Issue(nil), parsed.Issues...),
	}

	duplicateKeys := make(map[string]int)
	for index, record := range parsed.Records {
		label := recordLabel(record, index)
		recordIssues := validateRecord(record, label, target)

		key := duplicateKey(record)
		if key != "||" {
			if firstIndex, exists := duplicateKeys[key]; exists {
				recordIssues = append(recordIssues, issue(
					label,
					"run_key",
					"warning",
					fmt.Sprintf("Duplicate prompt/system/timestamp key; first seen in record %d.", firstIndex+1),
				))
			} else {
				duplicateKeys[key] = index
			}
		}

		result := RecordResult{Index: index, Label: label, Status: "ready"}
		for _, finding := range recordIssues {
			switch finding.Severity {
			case "error":
				result.ErrorCount++
			case "warning":
				result.WarningCount++
			}
		}
		if result.ErrorCount > 0 {
			result.Status = "invalid"
		} else if result.WarningCount > 0 {
			result.Status = "review"
		}

		report.RecordResults = append(report.RecordResults, result)
		report.Issues = append(report.Issues, recordIssues...)
	}

	report.Summary.Total = len(parsed.Records)
	report.Summary.ParseErrors = countSeverity(parsed.Issues, "error")
	for _, result := range report.RecordResults {
		switch result.Status {
		case "ready":
			report.Summary.Ready++
		case "review":
			report.Summary.Review++
		case "invalid":
			report.Summary.Invalid++
		}
	}
	return report
}

// NormalizeDomain converts a domain or URL to a lowercase hostname without a
// leading www label or trailing dot.
func NormalizeDomain(value string) string {
	raw := strings.TrimSpace(strings.ToLower(value))
	if raw == "" {
		return ""
	}
	toParse := raw
	if !strings.Contains(toParse, "://") {
		toParse = "https://" + toParse
	}
	parsed, err := url.Parse(toParse)
	if err != nil || parsed.Hostname() == "" {
		return strings.TrimSuffix(strings.TrimPrefix(raw, "www."), "/")
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	return strings.TrimPrefix(host, "www.")
}

func validateRecord(record Observation, label, targetDomain string) []Issue {
	var issues []Issue
	for _, field := range requiredFields {
		value, exists := record[field]
		if !exists || value == nil || isBlankString(value) {
			issues = append(issues, issue(label, field, "error", "Required field is missing or empty."))
		}
	}

	brandMentioned, brandIsBool := booleanValue(record["brand_mentioned"])
	if value, exists := record["brand_mentioned"]; exists && value != nil && !brandIsBool {
		issues = append(issues, issue(label, "brand_mentioned", "error", "Value must be a JSON boolean, not a string or number."))
	}
	targetCited, targetIsBool := booleanValue(record["target_domain_cited"])
	if value, exists := record["target_domain_cited"]; exists && value != nil && !targetIsBool {
		issues = append(issues, issue(label, "target_domain_cited", "error", "Value must be a JSON boolean, not a string or number."))
	}

	if timestamp := stringValue(record["run_timestamp"]); timestamp != "" {
		if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
			issues = append(issues, issue(label, "run_timestamp", "error", "Timestamp is not RFC 3339; use an ISO 8601 value with an explicit time zone."))
		}
	}

	domains, domainsValid := domainValues(record["cited_domains"])
	if value, exists := record["cited_domains"]; exists && value != nil && !domainsValid {
		issues = append(issues, issue(label, "cited_domains", "error", "Value must be an array of domains or URLs."))
	}
	hasTarget := containsDomain(domains, targetDomain)
	if targetIsBool && targetCited && targetDomain != "" && !hasTarget {
		issues = append(issues, issue(label, "target_domain_cited", "error", fmt.Sprintf("Marked true, but cited_domains does not contain %s.", targetDomain)))
	}
	if targetIsBool && !targetCited && targetDomain != "" && hasTarget {
		issues = append(issues, issue(label, "target_domain_cited", "warning", fmt.Sprintf("Marked false, but cited_domains contains %s.", targetDomain)))
	}

	if brandIsBool && brandMentioned && stringValue(record["claim_summary"]) == "" {
		issues = append(issues, issue(label, "claim_summary", "warning", "The brand is marked as mentioned, but no claim summary explains the context."))
	}
	if targetIsBool && targetCited && stringValue(record["evidence_artifact"]) == "" {
		issues = append(issues, issue(label, "evidence_artifact", "error", "A target citation needs an answer-level evidence artifact."))
	}
	if brandIsBool && targetIsBool && !brandMentioned && targetCited {
		issues = append(issues, issue(label, "brand_mentioned", "warning", "The target is cited while the brand is marked absent; verify whether the citation supports another claim."))
	}

	if version := stringValue(record["prompt_version"]); version != "" && !promptVersionPattern.MatchString(version) {
		issues = append(issues, issue(label, "prompt_version", "warning", "Use a stable, comparable version such as 1.0 or 2.1.0."))
	}
	if artifact := stringValue(record["evidence_artifact"]); artifact != "" && !isAbsoluteHTTPURL(artifact) {
		issues = append(issues, issue(label, "evidence_artifact", "warning", "Evidence artifact is not an absolute HTTP(S) URL; confirm that reviewers can resolve it."))
	}
	return issues
}

func decodeJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values found")
		}
		return fmt.Errorf("trailing data: %w", err)
	}
	return nil
}

func recordLabel(record Observation, index int) string {
	if promptID := stringValue(record["prompt_id"]); promptID != "" {
		return promptID
	}
	return fmt.Sprintf("record %d", index+1)
}

func duplicateKey(record Observation) string {
	return strings.Join([]string{
		stringValue(record["prompt_id"]),
		stringValue(record["answer_system"]),
		stringValue(record["run_timestamp"]),
	}, "|")
}

func booleanValue(value any) (bool, bool) {
	result, ok := value.(bool)
	return result, ok
}

func domainValues(value any) ([]string, bool) {
	switch values := value.(type) {
	case []any:
		domains := make([]string, 0, len(values))
		for _, value := range values {
			text, ok := value.(string)
			if !ok {
				return nil, false
			}
			if normalized := NormalizeDomain(text); normalized != "" {
				domains = append(domains, normalized)
			}
		}
		return domains, true
	case []string:
		domains := make([]string, 0, len(values))
		for _, value := range values {
			if normalized := NormalizeDomain(value); normalized != "" {
				domains = append(domains, normalized)
			}
		}
		return domains, true
	default:
		return nil, false
	}
}

func containsDomain(domains []string, target string) bool {
	if target == "" {
		return false
	}
	for _, domain := range domains {
		if domain == target || strings.HasSuffix(domain, "."+target) {
			return true
		}
	}
	return false
}

func stringValue(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func isBlankString(value any) bool {
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) == ""
}

func isAbsoluteHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func countSeverity(issues []Issue, severity string) int {
	count := 0
	for _, finding := range issues {
		if finding.Severity == severity {
			count++
		}
	}
	return count
}

func issue(record, field, severity, message string) Issue {
	return Issue{Record: record, Field: field, Severity: severity, Message: message}
}
