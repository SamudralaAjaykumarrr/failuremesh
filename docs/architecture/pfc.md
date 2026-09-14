# Portable Failure Contract v1 (conceptual)

A PFC is the versioned, machine-readable portable unit of causal failure knowledge. This is a conceptual contract, not a Go type or finalized serialization schema. It must carry:

| Field group | Required meaning |
| --- | --- |
| Identity | Stable PFC ID, semantic version, family, lifecycle state, verification level |
| Provenance | Source class/reference, extraction/review history, evidence references and confidentiality classification |
| Mechanism | Applicability prerequisites, vulnerability hypotheses and controls-under-test (including missing or insufficient control), causal sequence, trigger, forbidden outcome |
| Test | Invariant, abstract experiment operations and parameter domain, required observations, proof requirements |
| Constraints | Safety restrictions, supported architecture assumptions, known exclusions and uncertainty |

Applicability prerequisites must be evaluable predicates over AC facts and their evidence, not prose-only labels. They make the causal experiment meaningful and executable; they do not predict exposure. Missing or insufficient controls are recorded separately as vulnerability hypotheses or controls-under-test. A present control does not make the PFC `NOT_APPLICABLE` unless a PFC-defined, evidenced mechanism-defeating exclusion makes the causal experiment impossible or meaningless. For timeout-after-external-commit, prerequisites include an external side effect, possible caller uncertainty after provider commit, a placeable post-commit/pre-success fault, observable logical operation and provider effects, and an exercisable defined retry/recovery path. Idempotency/reconciliation is normally a control-under-test, so remediation does not prevent the same experiment. The causal sequence distinguishes provider commit from caller knowledge of that commit; a timeout by itself cannot prove commit. The invariant states a measurable property, such as at most one provider effect per logical operation. The abstract experiment names operations and temporal ordering without hard-coding a platform adapter. Proof requirements specify what must be observed to conclude exposure or scoped resilience, including observer completeness. Safety restrictions are enforced before execution.

Every revision that changes prerequisites, mechanism, invariant, experiment, proof requirements, or safety restrictions creates a new version. Existing verdicts remain bound to their original PFC version. Verification level is evidence-backed and cannot be promoted by LLM output: L0 CANDIDATE (extracted, never reproduced); L1 REPRODUCED (reference reproduction); L2 VERIFIED (repeated deterministic reproduction with stable evidence); L3 PORTABLE (materially different implementations); L4 FIELD VERIFIED (multiple independent customer architecture families). See [registry](registry.md).

A PFC is not an error fingerprint, stack trace, postmortem summary, chaos script, or LLM warning. Phase 1 must document a single v1 contract for timeout-after-external-commit before implementing any types.
