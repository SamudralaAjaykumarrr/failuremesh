# Battling database performance

- **Candidate family:** postgresql connection pool exhaustion
- **Verification level:** L0 CANDIDATE. No reproduction, applicability, or execution verdict.
- **Organization/project/source:** incident.io
- **Public reference:** [Battling database performance](https://incident.io/blog/database-performance)
- **Source date:** 2023-04-20
- **Source type / quality:** First-party engineering incident analysis

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — observed symptoms:** API requests intermittently timed out; one trace waited nearly 20 seconds for a pool connection.
- **SOURCE-ESTABLISHED FACT — causal facts:** The team found many short unnecessary transactions around Slack modal submissions collectively held connections; removing them preceded four months without reported timeouts.
- **UNKNOWN:** Exact pool limits, traffic distribution, controlled counterfactual UNKNOWN.
- **INFERENCE — mechanism interpretation:** Long aggregate transaction hold time exhausted available pool capacity.
- **HYPOTHESIS — possible controls-under-test:** Transaction scope, pool sizing, admission, connection hold-time budgets.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** PostgreSQL-backed service with finite client connection pools and concurrent operations.
- **Causal sequence:** SOURCE-ESTABLISHED FACT: modal submissions open transactions → pool wait → request timeouts; removal associated with recovery.
- **Forbidden outcome (candidate):** Requests exceed declared pool-acquisition deadline.
- **Candidate invariant:** Pool acquisition remains within a declared deadline at bounded workload.
- **Candidate abstract experiment:** In an isolated PostgreSQL fixture, hold transactions under controlled concurrency and measure waits.
- **Required observations / proof needs:** Pool capacity/use/wait, transaction start/end, request deadline, database health.
- **Safety restrictions:** Level 1; bounded connections, no production database.
- **Provenance:** First-party engineering narrative with trace/metrics description.
- **Evidence limitations:** Report describes association and investigation; parameter thresholds are not supplied.
- **Portability question:** Does the contract express aggregate workload rather than a single fault?

The proposed experiment and invariant are **HYPOTHESIS**, not source observations. Missing facts remain UNKNOWN; no customer AC was evaluated.

## Repair-pass exact-family audit

- **Exact-family match:** YES, for client-side pool exhaustion.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** incident.io reports an API trace waiting nearly 20 seconds for a Go database/sql connection and attributes recurring waits to aggregate transaction holding; its database was PostgreSQL.
- **UNKNOWN — decisive criterion remaining:** Exact pool size, connection occupancy timeline, and a controlled counterfactual are UNKNOWN.
- **INFERENCE / supplementary source boundary:** The measurable invariant is scoped to a declared workload and deadline, not a universal latency guarantee.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary source class:** PUBLIC_INCIDENT.
- **Counts toward ten-public-incident milestone:** YES; the source reports an observed operational failure with the defining mechanism at the evidence level described above.
- **Mechanism-fit status:** YES, L0 candidate; downstream effects and missing traces remain UNKNOWN as stated above.
