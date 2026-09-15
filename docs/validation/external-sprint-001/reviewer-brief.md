# FailureMesh: adversarial reviewer brief

For senior SRE, platform, backend, and distributed-systems engineers; approximately 5–10 minutes. **Criticism is the desired outcome. Try to invalidate FailureMesh.**

## Mission and definition

No engineering team should have to learn a software failure that the industry has already learned.

FailureMesh converts previously discovered software failures into portable, privacy-preserving, executable protection for systems that have not experienced those failures yet.

This is the mission and proposed product model. Current implementation is a bounded synthetic prototype; general transferability and usefulness remain hypotheses.

## Core model and authority

```text
incident
-> Portable Failure Contract
-> Architecture Contract
-> applicability
-> bounded experiment
-> evidence
-> scoped verdict
```

A **Portable Failure Contract (PFC)** preserves causal prerequisites, controls under test, sequence, forbidden outcome, invariant, abstract experiment, proof requirements, provenance, version, and safety restrictions. An **Architecture Contract (AC)** describes relevant evidenced system structure and semantics: transaction boundaries, retries, delivery, leases, dependencies, controls, and observability, with scope, provenance, and explicit unknowns.

| Decision | Vocabulary | Meaning |
| --- | --- | --- |
| Applicability | APPLICABLE / NOT_APPLICABLE / UNKNOWN | The mechanism can be meaningfully mapped; an evidenced exclusion defeats it; or evidence is insufficient/conflicting. |
| Execution | EXPOSED / PROVEN_RESILIENT / UNKNOWN | The forbidden outcome is proved; the invariant is proved within the declared experiment; or proof is insufficient. |
| Historical comparison | MATCH / PARTIAL / MISMATCH / UNKNOWN | Agreement between reported mechanism and bounded local representation, with explicit limits. |

APPLICABLE does not predict exposure. A protective control alone does not earn NOT_APPLICABLE; controls are tested. Missing facts remain UNKNOWN. Compilation requires applicability, and execution separately requires an approved bounded safety plan. Historical **PASS** is a milestone recommendation for that bounded comparison, not an execution verdict or Gate A result.

**PROVEN_RESILIENT is scoped, never absolute immunity:** it binds PFC version, application/build, environment, finite experiment space, invariant, and complete collected evidence. A new scope requires a new verdict. A digest identifies evidence bytes; it is not independent attestation.

**AI boundary:** AI may generate hypotheses or mappings. The deterministic verifier plus recorded authoritative evidence control final proof. AI cannot invent observations, relax proof rules, coerce UNKNOWN, or approve execution.

**Privacy boundary:** local-first. Raw source code, private logs, secrets, rows, and private postmortems need not leave the customer's environment. Shared structural information must be sufficiently generalized and explicitly approved. This review accepts public/synthetic material only; do not submit private evidence.

## Currently executable proof

Three separately approved local Go/PostgreSQL reference slices exist:

- [PFC #1: timeout after external commit](../../phases/phase-1-local-run.md): independent provider commit, response loss, retry, duplicate effect; durable idempotency prevents the duplicate in the remediated serial experiment.
- [PFC #2: duplicate queue delivery](../../phases/phase-1-pfc2-local-run.md): two deliveries of one message; consumer idempotency binds the second delivery to the first committed effect.
- [PFC #3: stale worker after lease expiry](../../phases/phase-1-pfc3-local-run.md): worker A resumes after B acquires a newer token; transactional fencing rejects A's stale effect in the remediated experiment.

Each vulnerable run earns EXPOSED and its remediated run earns scoped PROVEN_RESILIENT only with complete required evidence. These are serial/logically scheduled fixtures, not real queue, payment-provider, or distributed lease integrations, concurrent-race proofs, or customer verdicts.

## Historical Reproduction #001: Apache Iceberg #16282

The public reporter describes durable file registration before control-offset advancement, coordinator shutdown/recovery, replay of the same DATA_WRITTEN event, and registration of the same file in a second snapshot. The bounded Go/PostgreSQL model records one event/file, two consumptions, unchanged durable offset through recovery, and two distinct snapshot registrations. Authoritative database verification supports APPLICABLE and local MATCH / PASS.

Reported production observations and reporter causal analysis remain distinct from local observations. REST versus JDBC catalog is unresolved; exact crash/offset chronology is not independently established. The fixture executes no upstream Iceberg, Kafka, GCS, or catalog, and does not reproduce the reported timing, row-count symptom, or an upstream fix.

Read: [source record](../historical-reproductions/001-iceberg-16282/source-record.md) · [architecture](../historical-reproductions/001-iceberg-16282/architecture.md) · [reproduction](../historical-reproductions/001-iceberg-16282/reproduction.md) · [comparison](../historical-reproductions/001-iceberg-16282/comparison.md) · [result](../historical-reproductions/001-iceberg-16282/result.json).

## Historical Reproduction #002: GitHub May 4 cascade

GitHub's first-party account attributes shared database connection saturation to overlapping migration and production load, with contention, cross-service timeouts, and recovery after migration pause. The local model uses a finite capacity ledger, shared dependency edges, waiting requests, logical deadlines, and release/recovery rows. PostgreSQL verification supports APPLICABLE at **SERVICE-LEVEL CASCADE**, with local MATCH / PASS.

This mechanism is materially different from #001's recovery replay. Reported prerequisites are separated from direct local fixture capabilities. Capacity numbers, ticks, requests, and graph are synthetic choices, not GitHub telemetry. Exact pool sizes, request paths, deadlines, retry policy, and full topology remain UNKNOWN. No real pool exhaustion, GitHub services, physical latency, error rates, or upstream execution are reproduced.

Read: [source record](../historical-reproductions/002-github-may4-cascade/source-record.md) · [architecture](../historical-reproductions/002-github-may4-cascade/architecture.md) · [reproduction](../historical-reproductions/002-github-may4-cascade/reproduction.md) · [comparison](../historical-reproductions/002-github-may4-cascade/comparison.md) · [result](../historical-reproductions/002-github-may4-cascade/result.json).

## Reviewer challenge

1. Is a causal mechanism represented incorrectly?
2. Can applicability become APPLICABLE without enough evidence?
3. Can MATCH/PASS or PROVEN_RESILIENT be earned without authoritative proof?
4. Are UNKNOWNs being hidden or converted into assumptions?
5. Does the PFC abstraction lose information required for transfer?
6. Does the AC leak too much system information?
7. Could the proof produce dangerous false confidence?
8. What real architecture would break the current abstraction?
9. What would prevent you from running this at work?
10. What would you need to see before trusting it?

Use the [reproduction guide](reproduction-guide.md) and [scorecard](review-scorecard.md). Provide a scoped counterexample, verifier concern, or failed attempt where possible. Finding a flaw is a successful review outcome.

Gate A remains REVISE, mechanism fit 10/10, qualifying public incidents 8/10, families 1 and 3 unresolved, and expert validation incomplete. PFC #4, Historical #003, customer execution, and production execution remain unauthorized. This package establishes no design partner, pilot, payment, portability, or product readiness.
