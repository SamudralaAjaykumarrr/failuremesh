# GitHub May 2026 shared-database timeout cascade

- **Candidate family:** dependency latency cascade
- **Verification level:** L0 CANDIDATE. No FailureMesh reproduction or verdict.
- **Organization/project/source:** GitHub availability report, May 4 incident.
- **Public reference:** [GitHub availability report: May 2026](https://github.blog/news-insights/company-news/github-availability-report-may-2026/)
- **Source date:** 2026-06-11 (May 4 incident).
- **Source type / quality:** First-party availability report with trigger, shared dependency, symptoms, and mitigation.

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — observed symptoms:** GitHub reports elevated latency and errors across pull requests, issues, actions, webhooks, Git operations, and dependent services including Pages and Copilot.
- **SOURCE-ESTABLISHED FACT — causal facts:** A schema migration and peak traffic saturated connection capacity on a heavily used primary database. GitHub explicitly reports query contention and cascading timeouts across services depending on that database; pausing the migration preceded dependent-service recovery.
- **UNKNOWN:** Per-hop traces, exact deadlines, retry policies, whether each affected request traversed the same path, and independently complete latency measurements.
- **INFERENCE:** A shared slow dependency propagated deadline misses to multiple callers; GitHub's report supports the cascade at service level, not a fully reconstructed request-by-request graph.
- **HYPOTHESIS — controls-under-test:** Migration throttling, admission control, timeout budgets, circuit breaking, pool headroom, and dependency isolation.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** Shared critical-path database dependency, concurrent callers, finite connection capacity and deadlines. Presence or effectiveness of controls remains under test.
- **Causal sequence:** **SOURCE-ESTABLISHED FACT:** migration plus traffic → database connection saturation/query contention → cascading timeouts in dependent services → recovery after migration pause.
- **Forbidden outcome (candidate):** A bounded dependency slowdown causes dependent paths to exceed their declared latency/error budget.
- **Candidate invariant:** In a declared finite workload and dependency-delay range, caller deadlines and error budgets remain within specified limits.
- **Candidate abstract experiment:** In an isolated dependency graph, induce bounded query delay/contention and measure resulting caller latency and timeout propagation.
- **Required observations / proof needs:** Per-edge requests and deadlines, dependency latency/capacity, caller latency/errors, workload, control actions, and complete trace/metric coverage.
- **Safety restrictions:** Level 1; bounded synthetic load and delay, no live shared database or migration.
- **Provenance:** GitHub first-party monthly report; no raw traces published.
- **Evidence limitations:** The exact dependency-timeout cascade criterion **is supported at service level**; individual hop order and numerical bounds remain UNKNOWN. This overlaps pool exhaustion but asks a different invariant about propagation to dependent services.
- **Portability question:** Which critical-path edges and deadline budgets define equivalent cascades outside GitHub?

The proposed invariant and experiment are **HYPOTHESIS**. No customer AC was evaluated.

## Repair-pass exact-family audit

- **Exact-family match:** YES, at service-level cascade.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** GitHub May 4 report explicitly links shared database saturation/query contention to cascading timeouts across dependent services.
- **UNKNOWN — decisive criterion remaining:** Exact per-request hop/deadline/retry chain and observer completeness are UNKNOWN.
- **INFERENCE / supplementary source boundary:** The previous October 2025 provider-outage report was removed because it did not establish latency propagation.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary source class:** PUBLIC_INCIDENT.
- **Counts toward ten-public-incident milestone:** YES; the source reports an observed operational failure with the defining mechanism at the evidence level described above.
- **Mechanism-fit status:** YES, L0 candidate; downstream effects and missing traces remain UNKNOWN as stated above.
