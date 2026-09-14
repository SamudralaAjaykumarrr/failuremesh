# Vercel Workflow queue redelivery during live execution

- **Candidate family:** duplicate queue delivery
- **Verification level:** L0 CANDIDATE. No FailureMesh reproduction or verdict.
- **Organization/project/source:** vercel/workflow, issue #3811; reporter supplied production log excerpts.
- **Public reference:** [world-postgres queue delivery timeout issue](https://github.com/vercel/workflow/issues/3811)
- **Source date:** 2026-08-26.
- **Source type / quality:** Official project bug report with production observations and code-path analysis; reporter account is not independently audited.

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — observed symptoms:** The report shows Graphile Worker task failures near 300 seconds followed seconds later by recovery/redelivery logs; one step has two `step_started` events and the first execution continued after the second had begun.
- **SOURCE-ESTABLISHED FACT — causal facts:** The reporter traces queue delivery through an inline HTTP call using global `fetch`; its timeout marks the queue task failed and causes redelivery while the server-side handler can continue. The report states duplicate provider calls and a stale stream writer. This is a reporter claim supported by quoted log and database observations, not an independent forensic audit.
- **UNKNOWN:** Full provider-effect ledger, exact provider commit outcome, every delivery's durable broker state, and independent log completeness.
- **INFERENCE:** The generic transport-independent failure is repeat delivery while prior execution remains active; the report's actual transport is a Graphile/PostgreSQL-backed queue, so the queue boundary is source-supported.
- **HYPOTHESIS — controls-under-test:** Stable message-level idempotency, delivery-generation fencing, and transport timeout policy.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** Durable queue task, worker delivery, redelivery on failed task, executable business step, and stable task/step identity. Effective idempotency is a control-under-test, not a prerequisite.
- **Causal sequence:** **SOURCE-ESTABLISHED FACT (report):** inline handler remains live → delivery call times out → queue task is failed and redelivered → same step executes concurrently under a second delivery.
- **Forbidden outcome (candidate):** Duplicate committed business effect or stale-generation write from one logical task.
- **Candidate invariant:** At most one committed effect per logical task and no old-generation write after a replacement generation takes ownership.
- **Candidate abstract experiment:** In a synthetic queue, let a worker continue past a bounded delivery timeout, trigger redelivery, and observe both attempts and effects.
- **Required observations / proof needs:** Queue task ID, attempt/generation, failure and redelivery transitions, worker liveness, physical calls, authoritative committed effects, and observer completeness.
- **Safety restrictions:** Level 1, synthetic effects, bounded attempts and duration; abort on load or observer failure.
- **Provenance:** Project issue #3811, supplied production log excerpts and described code path; no raw private logs included.
- **Evidence limitations:** Reported duplicate calls are not proof of duplicate committed provider effects. The exact queue redelivery/overlap criterion **is supported**; exposure to a particular business invariant is UNKNOWN.
- **Portability question:** Which queue delivery timeout and acknowledgement semantics permit the same overlap on other brokers?

The invariant and experiment are **HYPOTHESIS**, not observed execution by FailureMesh. No customer AC was evaluated.

## Repair-pass exact-family audit

- **Exact-family match:** YES, for duplicate queue delivery and overlapping execution.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** The source reports the same Graphile queue task being failed and redelivered while the first handler remains live.
- **UNKNOWN — decisive criterion remaining:** Authoritative committed provider-effect duplication and full log completeness are UNKNOWN.
- **INFERENCE / supplementary source boundary:** The former TanStack Ship webhook report was removed from this candidate because a webhook is not evidence of this queue boundary.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary source class:** PROJECT_BUG_WITH_PRODUCTION_OBSERVATION.
- **Counts toward ten-public-incident milestone:** YES; the source reports an observed operational failure with the defining mechanism at the evidence level described above.
- **Mechanism-fit status:** YES, L0 candidate; downstream effects and missing traces remain UNKNOWN as stated above.
