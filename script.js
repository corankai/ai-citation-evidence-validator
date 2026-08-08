const requiredFields = [
  "prompt_id",
  "prompt_version",
  "answer_system",
  "run_timestamp",
  "brand_mentioned",
  "target_domain_cited",
  "cited_domains",
  "claim_summary",
  "evidence_artifact",
];

const sampleRecords = [
  {
    prompt_id: "compare-platforms-01",
    prompt_version: "1.0",
    buyer_stage: "vendor-comparison",
    intent_class: "comparative",
    answer_system: "Example Answer System",
    run_timestamp: "2026-08-08T17:00:00Z",
    market: "en-US",
    brand_mentioned: true,
    target_domain_cited: true,
    cited_domains: ["corank.ai", "example.org"],
    claim_summary: "The response included Corank in a category shortlist.",
    accuracy_status: "needs-review",
    evidence_artifact: "https://example.com/evidence/compare-platforms-01",
  },
  {
    prompt_id: "diagnose-citations-02",
    prompt_version: "1.0",
    buyer_stage: "method-exploration",
    intent_class: "diagnostic",
    answer_system: "Example Answer System",
    run_timestamp: "not-a-timestamp",
    market: "en-US",
    brand_mentioned: true,
    target_domain_cited: true,
    cited_domains: ["example.net"],
    claim_summary: "",
    accuracy_status: "unknown",
    evidence_artifact: "",
  },
];

const targetDomainInput = document.querySelector("#targetDomain");
const observationsInput = document.querySelector("#observations");
const resultsSection = document.querySelector("#results");
const issueRows = document.querySelector("#issueRows");
const downloadButton = document.querySelector("#downloadButton");
let currentReport = null;

document.querySelector("#sampleButton").addEventListener("click", () => {
  observationsInput.value = sampleRecords.map((record) => JSON.stringify(record)).join("\n");
  observationsInput.focus();
});

document.querySelector("#clearButton").addEventListener("click", () => {
  observationsInput.value = "";
  resultsSection.hidden = true;
  currentReport = null;
  downloadButton.disabled = true;
  observationsInput.focus();
});

document.querySelector("#analyzeButton").addEventListener("click", () => {
  const targetDomain = normalizeDomain(targetDomainInput.value);
  const parsed = parseRecords(observationsInput.value);
  currentReport = buildReport(parsed, targetDomain);
  renderReport(currentReport);
});

downloadButton.addEventListener("click", () => {
  if (!currentReport) return;
  const blob = new Blob([JSON.stringify(currentReport, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = "ai-citation-evidence-validation.json";
  anchor.click();
  URL.revokeObjectURL(url);
});

function parseRecords(rawText) {
  const raw = rawText.trim();
  if (!raw) {
    return { records: [], parseIssues: [{ record: "input", field: "input", severity: "error", message: "No observation records were supplied." }] };
  }

  if (raw.startsWith("[")) {
    try {
      const value = JSON.parse(raw);
      if (!Array.isArray(value)) {
        return { records: [], parseIssues: [{ record: "input", field: "input", severity: "error", message: "The top-level JSON value must be an array." }] };
      }
      return { records: value, parseIssues: [] };
    } catch (error) {
      return { records: [], parseIssues: [{ record: "input", field: "input", severity: "error", message: `Invalid JSON array: ${error.message}` }] };
    }
  }

  const records = [];
  const parseIssues = [];
  raw.split(/\r?\n/).forEach((line, index) => {
    if (!line.trim()) return;
    try {
      records.push(JSON.parse(line));
    } catch (error) {
      parseIssues.push({
        record: `line ${index + 1}`,
        field: "input",
        severity: "error",
        message: `Invalid JSONL record: ${error.message}`,
      });
    }
  });
  return { records, parseIssues };
}

function buildReport(parsed, targetDomain) {
  const issues = [...parsed.parseIssues];
  const recordResults = [];
  const duplicateKeys = new Map();

  parsed.records.forEach((record, index) => {
    const label = record && typeof record === "object" && !Array.isArray(record) && record.prompt_id
      ? String(record.prompt_id)
      : `record ${index + 1}`;
    const recordIssues = validateRecord(record, label, targetDomain);

    if (record && typeof record === "object" && !Array.isArray(record)) {
      const key = [record.prompt_id, record.answer_system, record.run_timestamp].map((value) => String(value ?? "").trim()).join("|");
      if (key !== "||") {
        if (duplicateKeys.has(key)) {
          recordIssues.push({ record: label, field: "run_key", severity: "warning", message: `Duplicate prompt/system/timestamp key; first seen in record ${duplicateKeys.get(key) + 1}.` });
        } else {
          duplicateKeys.set(key, index);
        }
      }
    }

    issues.push(...recordIssues);
    const hasError = recordIssues.some((issue) => issue.severity === "error");
    const hasWarning = recordIssues.some((issue) => issue.severity === "warning");
    recordResults.push({
      index,
      label,
      status: hasError ? "invalid" : hasWarning ? "review" : "ready",
      error_count: recordIssues.filter((issue) => issue.severity === "error").length,
      warning_count: recordIssues.filter((issue) => issue.severity === "warning").length,
    });
  });

  const summary = {
    total: parsed.records.length,
    ready: recordResults.filter((record) => record.status === "ready").length,
    review: recordResults.filter((record) => record.status === "review").length,
    invalid: recordResults.filter((record) => record.status === "invalid").length,
    parse_errors: parsed.parseIssues.length,
  };

  return {
    generated_at: new Date().toISOString(),
    target_domain: targetDomain || null,
    methodology: "Explicit field and consistency checks; no proprietary visibility score.",
    summary,
    record_results: recordResults,
    issues,
  };
}

function validateRecord(record, label, targetDomain) {
  const issues = [];
  if (!record || typeof record !== "object" || Array.isArray(record)) {
    return [{ record: label, field: "record", severity: "error", message: "Each observation must be a JSON object." }];
  }

  requiredFields.forEach((field) => {
    if (!(field in record) || record[field] === null || (typeof record[field] === "string" && !record[field].trim())) {
      issues.push({ record: label, field, severity: "error", message: "Required field is missing or empty." });
    }
  });

  ["brand_mentioned", "target_domain_cited"].forEach((field) => {
    if (field in record && typeof record[field] !== "boolean") {
      issues.push({ record: label, field, severity: "error", message: "Value must be a JSON boolean, not a string or number." });
    }
  });

  if (record.run_timestamp && Number.isNaN(Date.parse(record.run_timestamp))) {
    issues.push({ record: label, field: "run_timestamp", severity: "error", message: "Timestamp is not parseable; use an ISO 8601 value with an explicit time zone." });
  }

  if ("cited_domains" in record && !Array.isArray(record.cited_domains)) {
    issues.push({ record: label, field: "cited_domains", severity: "error", message: "Value must be an array of domains or URLs." });
  }

  const citedDomains = Array.isArray(record.cited_domains)
    ? record.cited_domains.map(normalizeDomain).filter(Boolean)
    : [];
  const hasTarget = targetDomain ? citedDomains.some((domain) => domain === targetDomain || domain.endsWith(`.${targetDomain}`)) : false;

  if (record.target_domain_cited === true && targetDomain && !hasTarget) {
    issues.push({ record: label, field: "target_domain_cited", severity: "error", message: `Marked true, but cited_domains does not contain ${targetDomain}.` });
  }

  if (record.target_domain_cited === false && targetDomain && hasTarget) {
    issues.push({ record: label, field: "target_domain_cited", severity: "warning", message: `Marked false, but cited_domains contains ${targetDomain}.` });
  }

  if (record.brand_mentioned === true && (!record.claim_summary || !String(record.claim_summary).trim())) {
    issues.push({ record: label, field: "claim_summary", severity: "warning", message: "The brand is marked as mentioned, but no claim summary explains the context." });
  }

  if (record.target_domain_cited === true && (!record.evidence_artifact || !String(record.evidence_artifact).trim())) {
    issues.push({ record: label, field: "evidence_artifact", severity: "error", message: "A target citation needs an answer-level evidence artifact." });
  }

  if (record.brand_mentioned === false && record.target_domain_cited === true) {
    issues.push({ record: label, field: "brand_mentioned", severity: "warning", message: "The target is cited while the brand is marked absent; verify whether the citation supports another claim." });
  }

  if (record.prompt_version && !/^\d+(\.\d+){0,2}([+-][a-z0-9.-]+)?$/i.test(String(record.prompt_version))) {
    issues.push({ record: label, field: "prompt_version", severity: "warning", message: "Use a stable, comparable version such as 1.0 or 2.1.0." });
  }

  if (record.evidence_artifact && !isAbsoluteHttpUrl(String(record.evidence_artifact))) {
    issues.push({ record: label, field: "evidence_artifact", severity: "warning", message: "Evidence artifact is not an absolute HTTP(S) URL; confirm that reviewers can resolve it." });
  }

  return issues;
}

function normalizeDomain(value) {
  const raw = String(value ?? "").trim().toLowerCase();
  if (!raw) return "";
  try {
    const url = new URL(raw.includes("://") ? raw : `https://${raw}`);
    return url.hostname.replace(/^www\./, "").replace(/\.$/, "");
  } catch {
    return raw.replace(/^www\./, "").replace(/\/$/, "");
  }
}

function isAbsoluteHttpUrl(value) {
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

function renderReport(report) {
  resultsSection.hidden = false;
  document.querySelector("#totalMetric").textContent = report.summary.total;
  document.querySelector("#readyMetric").textContent = report.summary.ready;
  document.querySelector("#reviewMetric").textContent = report.summary.review;
  document.querySelector("#invalidMetric").textContent = report.summary.invalid;

  const summaryText = report.summary.total
    ? `${report.summary.ready} ready, ${report.summary.review} needing review, and ${report.summary.invalid} invalid. Parse errors: ${report.summary.parse_errors}.`
    : `No valid records were parsed. Parse errors: ${report.summary.parse_errors}.`;
  document.querySelector("#resultSummary").textContent = summaryText;

  issueRows.replaceChildren();
  if (!report.issues.length && report.summary.total) {
    const row = document.createElement("tr");
    row.append(
      cell("All records"),
      statusCell("ready"),
      cell("—"),
      cell("No validation issues found for the supplied schema and consistency rules."),
    );
    issueRows.append(row);
  } else {
    report.issues.forEach((issue) => {
      const row = document.createElement("tr");
      row.append(cell(issue.record), statusCell(issue.severity), cell(issue.field), cell(issue.message));
      issueRows.append(row);
    });
  }

  downloadButton.disabled = false;
  resultsSection.scrollIntoView({ behavior: "smooth", block: "start" });
}

function cell(value) {
  const element = document.createElement("td");
  element.textContent = value;
  return element;
}

function statusCell(status) {
  const element = document.createElement("td");
  const badge = document.createElement("span");
  badge.className = `status ${status}`;
  badge.textContent = status === "error" ? "Error" : status === "warning" ? "Warning" : "Ready";
  element.append(badge);
  return element;
}
