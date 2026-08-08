package citationevidence

import (
	"strings"
	"testing"
	"time"
)

func TestParseAndValidateJSONL(t *testing.T) {
	input := strings.Join([]string{
		`{"prompt_id":"compare-01","prompt_version":"1.0","answer_system":"Example","run_timestamp":"2026-08-08T17:00:00Z","brand_mentioned":true,"target_domain_cited":true,"cited_domains":["corank.ai","example.org"],"claim_summary":"Corank appeared in a shortlist.","evidence_artifact":"https://example.com/evidence/1"}`,
		`{"prompt_id":"diagnose-02","prompt_version":"draft","answer_system":"Example","run_timestamp":"not-a-time","brand_mentioned":true,"target_domain_cited":true,"cited_domains":["example.net"],"claim_summary":"","evidence_artifact":""}`,
	}, "\n")

	parsed := Parse([]byte(input))
	if got, want := len(parsed.Records), 2; got != want {
		t.Fatalf("Parse records = %d, want %d", got, want)
	}
	if len(parsed.Issues) != 0 {
		t.Fatalf("Parse issues = %#v, want none", parsed.Issues)
	}

	report := ValidateAt(parsed, "https://www.corank.ai/", time.Date(2026, 8, 8, 18, 0, 0, 0, time.UTC))
	if got, want := report.TargetDomain, "corank.ai"; got != want {
		t.Fatalf("TargetDomain = %q, want %q", got, want)
	}
	if got, want := report.Summary.Total, 2; got != want {
		t.Fatalf("Total = %d, want %d", got, want)
	}
	if got, want := report.Summary.Ready, 1; got != want {
		t.Fatalf("Ready = %d, want %d", got, want)
	}
	if got, want := report.Summary.Invalid, 1; got != want {
		t.Fatalf("Invalid = %d, want %d", got, want)
	}
	if got, want := report.GeneratedAt, "2026-08-08T18:00:00Z"; got != want {
		t.Fatalf("GeneratedAt = %q, want %q", got, want)
	}
	assertIssue(t, report.Issues, "diagnose-02", "run_timestamp", "error")
	assertIssue(t, report.Issues, "diagnose-02", "target_domain_cited", "error")
	assertIssue(t, report.Issues, "diagnose-02", "prompt_version", "warning")
}

func TestParseJSONArray(t *testing.T) {
	parsed := Parse([]byte(`[{"prompt_id":"p-1"}]`))
	if got, want := len(parsed.Records), 1; got != want {
		t.Fatalf("records = %d, want %d", got, want)
	}
	if got := parsed.Records[0]["prompt_id"]; got != "p-1" {
		t.Fatalf("prompt_id = %#v, want p-1", got)
	}
}

func TestParseRejectsMultipleJSONValuesOnOneLine(t *testing.T) {
	parsed := Parse([]byte(`{"prompt_id":"one"} {"prompt_id":"two"}`))
	if got, want := len(parsed.Records), 0; got != want {
		t.Fatalf("records = %d, want %d", got, want)
	}
	if got, want := len(parsed.Issues), 1; got != want {
		t.Fatalf("issues = %d, want %d", got, want)
	}
}

func TestParsePreservesValidJSONLLines(t *testing.T) {
	parsed := Parse([]byte("{\"prompt_id\":\"valid\"}\nnot-json\n"))
	if got, want := len(parsed.Records), 1; got != want {
		t.Fatalf("records = %d, want %d", got, want)
	}
	if got, want := len(parsed.Issues), 1; got != want {
		t.Fatalf("issues = %d, want %d", got, want)
	}
	if got, want := parsed.Issues[0].Record, "line 2"; got != want {
		t.Fatalf("issue record = %q, want %q", got, want)
	}
}

func TestDuplicateRunKeyRequiresReview(t *testing.T) {
	base := Observation{
		"prompt_id":           "compare-01",
		"prompt_version":      "1.0",
		"answer_system":       "Example",
		"run_timestamp":       "2026-08-08T17:00:00Z",
		"brand_mentioned":     false,
		"target_domain_cited": false,
		"cited_domains":       []string{},
		"claim_summary":       "No brand claim was present.",
		"evidence_artifact":   "https://example.com/evidence/1",
	}
	copyRecord := Observation{}
	for key, value := range base {
		copyRecord[key] = value
	}
	report := ValidateAt(ParsedInput{Records: []Observation{base, copyRecord}}, "corank.ai", time.Unix(0, 0))
	if got, want := report.Summary.Ready, 1; got != want {
		t.Fatalf("ready = %d, want %d", got, want)
	}
	if got, want := report.Summary.Review, 1; got != want {
		t.Fatalf("review = %d, want %d", got, want)
	}
	assertIssue(t, report.Issues, "compare-01", "run_key", "warning")
}

func TestNormalizeDomain(t *testing.T) {
	tests := map[string]string{
		"https://www.Corank.ai/path": "corank.ai",
		"docs.corank.ai.":            "docs.corank.ai",
		"  example.org  ":            "example.org",
		"":                           "",
	}
	for input, want := range tests {
		if got := NormalizeDomain(input); got != want {
			t.Errorf("NormalizeDomain(%q) = %q, want %q", input, got, want)
		}
	}
}

func assertIssue(t *testing.T, issues []Issue, record, field, severity string) {
	t.Helper()
	for _, finding := range issues {
		if finding.Record == record && finding.Field == field && finding.Severity == severity {
			return
		}
	}
	t.Fatalf("missing issue record=%q field=%q severity=%q in %#v", record, field, severity, issues)
}
