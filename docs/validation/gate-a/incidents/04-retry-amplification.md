# GitHub August 2026 authentication retry amplification

- **Candidate family:** retry amplification
- **Verification level:** L0 CANDIDATE. No reproduction, applicability, or execution verdict.
- **Organization/project/source:** GitHub
- **Public reference:** [GitHub August 2026 authentication retry amplification](https://github.blog/news-insights/company-news/github-availability-report-august-2026/)
- **Source date:** 2026-09-09 (August incident)
- **Source type / quality:** First-party availability report

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — observed symptoms:** Authentication latency and failures spread across services.
- **SOURCE-ESTABLISHED FACT — causal facts:** The report links load-balancer flow exhaustion, a sidecar concurrency ceiling, and a latent client retry bug that amplified authentication traffic; reducing retries aided recovery.
- **UNKNOWN:** Per-client retry policy, request multiplier, exact thresholds UNKNOWN.
- **INFERENCE — mechanism interpretation:** Retry pressure plausibly sustained overload; report describes amplification but not a transferable numeric threshold.
- **HYPOTHESIS — possible controls-under-test:** Retry limits, backoff, admission/load shedding.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** Shared dependency, retrying clients, bounded capacity.
- **Causal sequence:** SOURCE-ESTABLISHED FACT: peak load → sidecar/balancer saturation → retry amplification → broader errors.
- **Forbidden outcome (candidate):** Retry traffic drives dependency beyond a declared safe budget.
- **Candidate invariant:** Under bounded degradation, aggregate retry demand stays below a specified budget.
- **Candidate abstract experiment:** In a synthetic dependency, impose controlled partial failures and measure retry demand over a finite workload.
- **Required observations / proof needs:** Original and retry requests, concurrency, queue depth, capacity, error rate, completeness.
- **Safety restrictions:** Level 1; strict rate and duration caps.
- **Provenance:** First-party operational report.
- **Evidence limitations:** Several interacting causes; single-factor attribution is unsafe.
- **Portability question:** What workload and capacity normalization makes this portable?

The proposed experiment and invariant are **HYPOTHESIS**, not source observations. Missing facts remain UNKNOWN; no customer AC was evaluated.

## Repair-pass exact-family audit

- **Exact-family match:** YES, at the service-overload mechanism level.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** GitHub attributes amplified authentication traffic during the incident to a client retry bug and reports reduced retry pressure as part of recovery.
- **UNKNOWN — decisive criterion remaining:** Exact retry multiplier, per-client policy, and isolated contribution versus sidecar saturation are UNKNOWN.
- **INFERENCE / supplementary source boundary:** The candidate invariant requires a future declared workload/capacity budget; no numeric threshold is inferred from the report.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary source class:** PUBLIC_INCIDENT.
- **Counts toward ten-public-incident milestone:** YES; the source reports an observed operational failure with the defining mechanism at the evidence level described above.
- **Mechanism-fit status:** YES, L0 candidate; downstream effects and missing traces remain UNKNOWN as stated above.
