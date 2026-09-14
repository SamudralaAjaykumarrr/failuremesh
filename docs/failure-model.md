# Failure model

A **failure mechanism** is a causal chain in which applicable architecture, a trigger, and potentially missing or insufficient protection can lead to a forbidden outcome. Applicability prerequisites establish that the causal experiment is meaningful and executable; missing or insufficient controls are vulnerability hypotheses or controls-under-test, not automatically prerequisites. An **incident** is evidence that may suggest a mechanism, not by itself a portable claim. A PFC abstracts the mechanism; an AC supplies system-specific facts; an experiment tests a declared invariant under controlled perturbation. See [PFC](architecture/pfc.md), [AC](architecture/architecture-contract.md), and [proof](architecture/proof.md).

The initial corpus target is the baseline's first ten families, in order:

1. timeout after external commit
2. duplicate queue delivery
3. stale worker after lease expiry
4. retry amplification / retry storm
5. PostgreSQL connection-pool exhaustion
6. poison-message retry loop
7. partial transaction / business-state divergence
8. dependency latency cascade
9. cache stampede
10. worker crash after durable write before acknowledgement

These are family names, not ten verified contracts. Each future PFC needs its own causal prerequisites, invariant, abstract experiment, safety restrictions, provenance, and verification level. Mechanisms may overlap; matching a family label alone never establishes applicability.

For the first family, a provider commits an effect, its response is lost, and the caller retries under uncertainty. The external effect, possible post-commit uncertainty, placeable fault, observable effects, and exercisable retry/recovery path establish applicability. Without effective idempotency or reconciliation, one logical operation can produce multiple external effects; that protection is tested by rerunning the same experiment after remediation. The forbidden outcome is a duplicate logical effect; a retry alone is not exposure. The [Phase 1 document](phases/phase-1-first-vertical-slice.md) defines the testable scope.
