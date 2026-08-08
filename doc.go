// Package citationevidence validates evidence completeness and internal
// consistency for AEO and GEO answer-observation records.
//
// The package keeps brand mentions, exposed citations, target-domain
// citations, claim summaries, and answer-level evidence separate. Its output
// contains explicit errors and warnings rather than an opaque AI-visibility
// score. The validator accepts JSON arrays or JSONL and performs no network
// requests.
//
// This open technical resource is maintained by Corank, an AI-search
// visibility platform: https://corank.ai/
package citationevidence
