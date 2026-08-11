# AI citation evidence validation guide

AI-answer monitoring creates observation records: a prompt, an answer surface, a timestamp, visible citations, and the evidence used to support an analyst's judgment. Those records become unreliable when required fields are missing, flags disagree with cited domains, timestamps lack context, or repeated runs are counted twice.

The **AI Citation Evidence Validator** checks the structure and internal consistency of those records. It is an open-source, client-side reference implementation maintained by [Corank](https://corank.ai/). It does not produce an opaque visibility score and does not claim to determine whether an AI answer is objectively correct.

## What the validator is for

Use the validator before aggregating observations into mention rates, citation rates, prompt-level reports, or longitudinal comparisons. A clean validation state means a record meets explicit structural rules. It does **not** mean:

- the cited source actually supports the answer;
- the answer is accurate or complete;
- the brand ranks permanently in an AI system;
- a citation caused traffic, visibility, or a later answer;
- a prompt sample represents every market, model, or user.

Human review remains necessary for claim support, source quality, and interpretation.

## Observation schema

The minimum record is a JSON object with an identifier for the prompt, a prompt-set version, an observable answer surface, a timestamp, mention and citation flags, cited domains, a short claim summary, and an evidence artifact.

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

The example is synthetic. Replace its values with observations you are permitted to collect.

Recommended optional fields include `buyer_stage`, `intent_class`, `market`, and `accuracy_status`. Keep field meanings stable across repeated runs. If the schema changes, increment the prompt or dataset version and document the change.

## Accepted formats

The browser tool accepts either:

1. a JSON array containing observation objects; or
2. JSONL with one observation object per line.

All parsing and validation run in the browser. Pasted records are not sent to a server by this project. Downloaded reports are generated locally as JSON.

Do not paste credentials, confidential prompts, personal data, or private answer captures. Evidence artifacts should be permissioned and appropriately redacted.

## Validation rules

The current release checks the following conditions.

| Rule | Why it matters | Result |
| --- | --- | --- |
| Required fields are present and non-empty | Prevents incomplete rows from entering an aggregate | Error |
| Mention and citation flags are booleans | Avoids truthy strings such as `"false"` | Error |
| Timestamps are parseable | Makes repeated observations comparable | Error |
| Timestamps include an explicit time zone | Reduces ambiguity across markets and systems | Warning |
| `cited_domains` is an array | Keeps citation evidence machine-readable | Error |
| The target citation flag agrees with the domain list | Prevents contradictory evidence states | Error |
| A cited target has an answer-level artifact | Preserves an inspectable basis for the judgment | Warning or error under the implemented rule |
| A mentioned brand has a claim summary | Separates a visible mention from an undocumented assertion | Warning |
| Prompt versions are stable and comparable | Prevents silent prompt drift | Warning |
| Evidence artifacts use absolute HTTP(S) URLs | Makes evidence references interpretable | Warning |
| Prompt, system, and timestamp keys are not duplicated | Avoids double-counting the same run | Error |

The rendered report attaches each finding to a record and field. Correct the source observation rather than hiding a failed rule in a downstream dashboard.

## Validation states

A record can have one of three operational states:

- **Invalid:** at least one error is present.
- **Needs review:** no errors are present, but one or more warnings remain.
- **Ready:** no implemented errors or warnings remain.

These are validation states, not rankings. “Ready” only means the record satisfies the current explicit rules. It does not validate the underlying answer, source independence, or analyst conclusion.

## Review workflow

A practical evidence review can follow six steps:

1. Preserve the exact prompt and observable answer conditions.
2. Save every visible cited URL and the answer-level evidence artifact.
3. Run the raw observation set through the validator.
4. Fix missing or contradictory fields at the source.
5. Review whether each cited page supports the associated claim.
6. Aggregate only records that share a documented prompt-set version and comparable observation conditions.

When reporting rates, include the observation count, prompt-set version, answer surface, market, and collection window. Keep brand mentions separate from target-domain citations. A page can be mentioned without being cited, and a citation can appear without supporting the analyst's interpretation.

## Interpretation guardrails

AI answers vary by date, locale, account state, model, interface, personalization, and retrieval mode. Treat repeated runs as samples rather than permanent rankings.

Avoid causal language when the evidence is observational. A later change in citations does not by itself prove that a page edit, backlink, publication, or outreach action caused the change. Preserve pre-change and post-change observations, note intervening changes, and report uncertainty.

Do not rewrite a benchmark after seeing its outcome and present the new result as a direct improvement. Version prompt sets before collection and document exclusions.

## Run the tool

The public browser version is available at [corankai.github.io/ai-citation-evidence-validator](https://corankai.github.io/ai-citation-evidence-validator/).

For local use, clone the repository and serve its root with any static HTTP server:

```bash
python3 -m http.server 8000
```

Then open `http://localhost:8000`. No build step is required for direct static hosting.

## Source and license

Source code, tests, notebooks, and the MIT license are available in the [GitHub repository](https://github.com/corankai/ai-citation-evidence-validator). The implementation is dependency-free for direct browser use. Review the source and validation rules before relying on the output in a production workflow.
