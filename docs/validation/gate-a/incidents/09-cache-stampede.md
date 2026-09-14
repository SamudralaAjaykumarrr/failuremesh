# Facebook configuration cache feedback outage

- **Candidate family:** cache stampede
- **Verification level:** L0 CANDIDATE. No reproduction, applicability, or execution verdict.
- **Organization/project/source:** Facebook Engineering
- **Public reference:** [Facebook configuration cache feedback outage](https://engineering.fb.com/2010/09/23/uncategorized/more-details-on-today-s-outage/)
- **Source date:** 2010-09-23
- **Source type / quality:** First-party postmortem

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — observed symptoms:** Facebook was down or unreachable for many users for about 2.5 hours.
- **SOURCE-ESTABLISHED FACT — causal facts:** An invalid persistent configuration value made many clients attempt repair via database reads; database errors caused cache-key deletion, feeding further queries and overload.
- **UNKNOWN:** Exact cache-key cardinality and authoritative per-key refill log UNKNOWN.
- **INFERENCE — mechanism interpretation:** This is a cache/repair thundering herd with positive feedback, not a simple TTL-expiry stampede.
- **HYPOTHESIS — possible controls-under-test:** Request coalescing, bounded repair, error-aware cache invalidation.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** Shared cached configuration, concurrent clients, database refill/repair path.
- **Causal sequence:** SOURCE-ESTABLISHED FACT: invalid value → parallel repairs → DB overload → cache deletion on errors → persistent loop.
- **Forbidden outcome (candidate):** Cache recovery traffic overwhelms origin and prevents service recovery.
- **Candidate invariant:** Refill/repair work per key and origin demand stay within declared limits under a bounded fault.
- **Candidate abstract experiment:** In a synthetic cache, introduce invalid backing value and concurrent readers; measure repair/refill amplification.
- **Required observations / proof needs:** Key state, misses/repairs, database demand/capacity, errors, recovery, observer completeness.
- **Safety restrictions:** Level 1; synthetic configuration, strict load caps.
- **Provenance:** First-party technical postmortem.
- **Evidence limitations:** Mechanism mixes repair feedback with stampede; simple cache-miss model would distort it.
- **Portability question:** Can a PFC include feedback state transitions and origin capacity?

The proposed experiment and invariant are **HYPOTHESIS**, not source observations. Missing facts remain UNKNOWN; no customer AC was evaluated.

## Repair-pass exact-family audit

- **Exact-family match:** YES, for cache/repair stampede feedback.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** Facebook reports many clients simultaneously attempted configuration repair, overwhelming the database; errors deleted cache keys and fed more queries.
- **UNKNOWN — decisive criterion remaining:** Per-key cardinality and complete origin-demand measurement are UNKNOWN.
- **INFERENCE / supplementary source boundary:** This is not asserted to be a simple TTL-expiry stampede; the candidate captures the reported repair feedback.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary source class:** PUBLIC_INCIDENT.
- **Counts toward ten-public-incident milestone:** YES; the source reports an observed operational failure with the defining mechanism at the evidence level described above.
- **Mechanism-fit status:** YES, L0 candidate; downstream effects and missing traces remain UNKNOWN as stated above.
