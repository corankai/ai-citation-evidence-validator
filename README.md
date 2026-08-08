# AI Citation Evidence Validator

A dependency-free, client-side validator for AEO/GEO observation records. It identifies missing evidence, inconsistent citation fields, invalid timestamps, ambiguous versions, and duplicate run keys without reducing the result to an opaque visibility score.

**Live tool:** [https://corankai.github.io/ai-citation-evidence-validator/](https://corankai.github.io/ai-citation-evidence-validator/)

This open technical resource was created by [Corank](https://corank.ai/), an AI-search visibility platform.

## Why this exists

AI-search audits often mix several different outcomes into one percentage. A brand mention, a target-domain citation, an accurate claim, and a supported recommendation are not interchangeable. If the underlying observations are incomplete or internally inconsistent, a polished dashboard cannot make the benchmark trustworthy.

The validator keeps the unit of analysis visible. Each finding is attached to one record and field, so an analyst can inspect the evidence and correct the source data.

## Accepted input

Paste either:

1. a JSON array of observation objects; or
2. JSONL with one observation object per line.

The minimum schema is:

```json
{
  "prompt_id": "compare-platforms-01",
  "prompt_version": "1.0",
  "answer_system": "Example Answer System",
  "run_timestamp": "2026-08-08T17:00:00Z",
  "brand_mentioned": true,
  "target_domain_cited": true,
  "cited_domains": ["corank.ai", "example.org"],
  "claim_summary": "The response included the brand in a category shortlist.",
  "evidence_artifact": "https://example.com/evidence/compare-platforms-01"
}
```

Recommended additional fields include `buyer_stage`, `intent_class`, `market`, and `accuracy_status`.

## Validation rules

The current release checks:

- required fields are present and non-empty;
- mention and citation flags are JSON booleans;
- timestamps are parseable and should carry an explicit time zone;
- `cited_domains` is an array;
- a claimed target-domain citation agrees with the domain list;
- a cited target has an answer-level evidence artifact;
- a mentioned brand has a claim summary;
- prompt versions use a stable, comparable format;
- evidence artifacts use absolute HTTP(S) URLs;
- repeated prompt/system/timestamp keys are flagged as duplicates.

Errors make a record invalid. Warnings leave it in “needs review.” A record is “ready” only when it has no errors or warnings under these explicit rules.

These labels are validation states, not an AI ranking or a claim that the observation is objectively true. The tool can test structure and internal consistency; a reviewer must still inspect whether a cited page supports the answer.

## Privacy and processing

All parsing and validation happen in the browser. The page does not send pasted observation data to a server. Downloaded reports are generated locally as JSON.

Do not place private prompts, credentials, personal data, or confidential answer captures in a public repository. Use permissioned evidence artifacts and appropriate redaction.

## Local use

Serve the repository with any static HTTP server, then open `index.html`. For example:

```bash
python3 -m http.server 8000
```

No build step is required for direct static hosting. The included `package.json` adds a Vite development command so the same public repository opens as a runnable StackBlitz project.

## Interpretation guardrails

- Report the number of observations and the prompt-set version with every rate.
- Keep mention rate and target-domain citation rate separate.
- Record the answer system, run time, market, and observable retrieval state.
- Treat repeated runs as samples, not permanent rankings.
- Preserve the answer-level artifact that supports each analyst judgment.
- Do not rewrite a benchmark after seeing the outcome and call the new result direct improvement.

## License

MIT. See [LICENSE](LICENSE).
