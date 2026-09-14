# Memory-file URI repeatedly re-enqueued

- **Candidate family:** poison message retry loop
- **Verification level:** L0 CANDIDATE. No reproduction, applicability, or execution verdict.
- **Organization/project/source:** OpenViking, issue #2734
- **Public reference:** [Memory-file URI repeatedly re-enqueued](https://github.com/volcengine/OpenViking/issues/2734)
- **Source date:** 2026-06-19
- **Source type / quality:** Project bug issue with code-path account

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — observed symptoms:** A memory-file URI repeatedly fails and re-enqueues, starving semantic queue work.
- **SOURCE-ESTABLISHED FACT — causal facts:** The issue identifies a non-directory URI raising in `_process_memory_directory` and entering transient retry handling.
- **UNKNOWN:** Production frequency, queue isolation, complete delivery log UNKNOWN.
- **INFERENCE — mechanism interpretation:** Deterministic invalid input is treated as transient, producing unbounded retries.
- **HYPOTHESIS — possible controls-under-test:** Permanent-error classification, dead-letter handling, retry budget.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** Queue/worker with redelivery and deterministic invalid payload class.
- **Causal sequence:** SOURCE-ESTABLISHED FACT (issue): dequeue → non-directory error → re-enqueue → repeat.
- **Forbidden outcome (candidate):** One poison item indefinitely consumes worker capacity or starves valid work.
- **Candidate invariant:** Every invalid item reaches terminal handling within a bounded attempt budget.
- **Candidate abstract experiment:** In a synthetic queue, enqueue one invalid and one valid item; observe bounded disposition and progress.
- **Required observations / proof needs:** Item identity, attempts, failure class, queue depth, valid-item latency, terminal state.
- **Safety restrictions:** Level 1; synthetic messages, bounded retries.
- **Provenance:** Project issue; no independent operational report.
- **Evidence limitations:** Broad outage impact is not established.
- **Portability question:** How to distinguish permanent invalidity from transient environment faults?

The proposed experiment and invariant are **HYPOTHESIS**, not source observations. Missing facts remain UNKNOWN; no customer AC was evaluated.

## Repair-pass exact-family audit

- **Exact-family match:** YES, for poison-message retry loop.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** OpenViking issue #2734 shows a file URI routed to a directory processor, exception, transient re-enqueue, and persistent queue starvation.
- **UNKNOWN — decisive criterion remaining:** Independent production fleet impact and full attempt log are UNKNOWN.
- **INFERENCE / supplementary source boundary:** The report is an official project issue; source date is 2026-06-19.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary source class:** PUBLIC_INCIDENT.
- **Counts toward ten-public-incident milestone:** YES. The issue reports an observed server traceback, repeated re-enqueue, and a roughly 31-minute hung client call; production deployment scope is not established.
- **Mechanism-fit status:** YES, L0 candidate; this classification does not change the source-backed causal mapping.
