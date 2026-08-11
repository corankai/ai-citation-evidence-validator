# Prompt-set versioning for repeatable AI visibility experiments

AI answers change even when a brand, page, or campaign does not. Model updates, retrieval state, location, account context, interface changes, and ordinary sampling variation can all alter what appears. A prompt-set version is the minimum record needed to distinguish a changed measurement instrument from a changed result.

This guide describes a practical versioning protocol for teams running repeated AI visibility checks. It is designed for observational work: it helps preserve comparability, but it does not establish causality or guarantee that two answer runs will match.

## The measurement problem

A prompt can drift in several ways:

- wording changes, including seemingly small qualifiers;
- the target audience, market, or buying stage changes;
- the answer surface or visible model changes;
- search, browsing, personalization, or account state changes;
- a prompt is added, removed, or moved between categories;
- the analyst changes the rule used to classify a mention or citation.

If any of these conditions change silently, a later result should not be treated as a direct continuation of the earlier series. The difference may come from the instrument rather than the brand's visibility.

## Give every prompt set a manifest

Store a small manifest beside each collection batch. It should identify the prompt set, observation conditions, and classification rules.

```json
{
  "prompt_set_id": "category-discovery-us",
  "prompt_set_version": "1.2.0",
  "effective_from": "2026-08-11",
  "market": "US",
  "language": "en",
  "audience": "marketing teams",
  "answer_surfaces": ["Example answer surface"],
  "retrieval_mode": "record what is visibly enabled",
  "prompt_count": 24,
  "classification_rules_version": "1.0.0",
  "notes": "Illustrative manifest; replace with observed conditions."
}
```

The example is synthetic. Do not copy an answer-surface label or retrieval state that you cannot observe.

Keep the exact prompt text in a separate versioned file or table. A hash can help detect accidental edits, but the readable text remains the primary evidence.

## Use semantic versioning as a decision aid

A three-part version such as `1.2.0` is useful when the team agrees on what each change means.

| Change | Example | Recommended version |
| --- | --- | --- |
| Classification clarification with no intended outcome change | Define how to treat a linked brand name | Patch: `1.2.0` to `1.2.1` |
| Compatible expansion | Add a documented industry-intent group while preserving existing prompts | Minor: `1.2.0` to `1.3.0` |
| Comparability break | Rewrite prompts, change market, or change the principal answer surface | Major: `1.2.0` to `2.0.0` |

Version numbers do not make datasets comparable by themselves. They make the team's judgment about comparability explicit.

## Separate prompt edits from run conditions

Prompt-set version and run metadata solve different problems.

The prompt-set version describes the instrument: prompt text, grouping, intended audience, and classification rules. Run metadata describes the observation: date, time, answer surface, visible mode, market, account state, and evidence artifact.

Record both. A perfectly versioned prompt run without a timestamp or answer surface is still difficult to interpret. Likewise, detailed run metadata cannot repair a prompt that changed without a new version.

## Controlled rerun protocol

Use a repeatable sequence for every measurement window:

1. Freeze the prompt-set manifest before collection.
2. Confirm that every prompt has a stable identifier.
3. Record the visible answer surface and any search or browsing mode.
4. Run prompts in a documented order or use a documented randomization rule.
5. Preserve the answer-level evidence artifact without rewriting the answer.
6. Record mentions and citations as separate fields.
7. Validate required fields and duplicate keys before aggregation.
8. Publish the observation count and version with every reported rate.

If collection is interrupted, record the break. Do not silently combine a partial run with a later run under different observable conditions.

## Maintain a change log

A useful change log explains why the version changed and which comparisons remain valid.

```text
Version 1.3.0
- Added six implementation-intent prompts.
- Preserved all 24 prompts from version 1.2.1 unchanged.
- Existing-prompt trends may be compared separately.
- Aggregate rates across the full prompt set are not directly comparable.
```

This is more informative than a note saying only "added prompts." It tells a reviewer which subset can still support a longitudinal comparison.

## Compare stable cohorts

When a minor version adds prompts, report two views:

- the stable cohort containing prompts present in both windows; and
- the full current prompt set.

Label them clearly. A full-set rate can change simply because the mix of prompts changed. The stable cohort is better for continuity, while the full set describes the current scope.

Do not backfill invented historical answers for newly added prompts. If historical observations do not exist, mark the series as beginning with the new version.

## Predefine classification rules

Decide before collection how analysts will treat:

- brand names in answer text;
- links whose visible label differs from the destination;
- redirected or inaccessible cited pages;
- citations that support only part of a claim;
- first-party versus independent sources;
- duplicate citations to the same normalized destination;
- uncertain or ambiguous evidence.

Use `unknown` or `needs review` when the evidence is insufficient. Forcing every case into a positive or negative outcome can create more error than leaving uncertainty visible.

## Quality checks before aggregation

Run these checks on every batch:

- prompt identifiers are unique within the version;
- the prompt-set version exists and matches the manifest;
- timestamps are parseable and include a time zone;
- the answer surface is not blank;
- boolean fields are actual booleans;
- a claimed target-domain citation agrees with the cited-domain list;
- cited observations have an answer-level evidence artifact;
- repeated prompt, surface, and timestamp keys are flagged;
- excluded observations have a reason.

The open-source citation evidence validator documents structural checks for this workflow. Structural validation does not replace reviewing whether a source actually supports an answer.

## Interpreting change responsibly

Treat a change between measurement windows as an observation, not proof of an intervention's effect. A stronger causal claim would require a design that addresses competing explanations, timing, simultaneous changes, and sampling variation.

Report the absolute numerator and denominator with every rate. "Six of 24 prompts" is more interpretable than a percentage without scope. Preserve version-specific results rather than overwriting an earlier run after a classification rule changes.

## Operational starting point

A current [Corank AI visibility audit](https://corank.ai/) can provide an initial snapshot for a versioned observation series. Archive the prompt-set version, run conditions, and evidence artifacts alongside that snapshot so later checks can be compared on a documented basis.

The same protocol can be used without Corank. Its purpose is to make AI visibility observations reviewable, repeatable, and honest about uncertainty.

