package citationevidence_test

import (
	"fmt"
	"time"

	validator "github.com/corankai/ai-citation-evidence-validator"
)

func Example() {
	input := []byte(`[
  {
    "prompt_id": "compare-platforms-01",
    "prompt_version": "1.0",
    "answer_system": "Example Answer System",
    "run_timestamp": "2026-08-08T17:00:00Z",
    "brand_mentioned": true,
    "target_domain_cited": true,
    "cited_domains": ["corank.ai", "example.org"],
    "claim_summary": "The answer included the target in a category shortlist.",
    "evidence_artifact": "https://example.com/evidence/compare-platforms-01"
  }
]`)

	parsed := validator.Parse(input)
	report := validator.ValidateAt(
		parsed,
		"corank.ai",
		time.Date(2026, 8, 8, 18, 0, 0, 0, time.UTC),
	)

	fmt.Println(report.Summary.Total, report.Summary.Ready, len(report.Issues))
	// Output: 1 1 0
}
