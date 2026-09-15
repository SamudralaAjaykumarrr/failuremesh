# FailureMesh Match Evaluation #001

External reviewer feedback, as summarized in the task requesting this package, identified inspectable case-level matching quality and grading provenance as more useful than headline benchmark numbers. This artifact responds with three concrete synthetic cases an engineer can challenge individually. No reviewer identity, private feedback transcript, completed expert review, or reviewer endorsement is asserted.

**Question to audit:** “Show me exactly why this matched, what evidence existed, what the human expected, what FailureMesh returned, and where I can disagree with the grade.”

## Read and reproduce

1. Read [grading method, registration hashes and executable replay](grading-method.md).
2. Inspect [001 positive match](case-001-positive-match.md), [002 near-match challenge](case-002-near-match-challenge.md), and [003 insufficient evidence](case-003-insufficient-evidence.md). Each separates source facts, architecture facts, assumptions, expected judgment, raw system output, grading, ambiguity and uncertainty.
3. Read [results and quality checks](results.md). Reproduce and dispute individual grades; no aggregate benchmark is claimed.

All three use the existing PFC #1 at version 1.0.0 to isolate a material causal prerequisite: same-operation retry. Case 001 uses the implemented synthetic reference. Case 002 is a bounded no-retry design. Case 003 deliberately withholds retry evidence. Pre-registered expected judgments were written before matcher invocation by the coding agent for human review; they are neither authoritative truth nor independent human labels.

## Selection context

The existing matching materials were inspected before selection:

| Existing scope | Evidence and matching distinction | Why used or not used here |
| --- | --- | --- |
| [PFC #1](../../phases/phase-1-local-run.md) | Independent provider commit, post-commit uncertainty/barrier, same-operation retry, observable ledger; idempotency is a control | Selected: existing shared evaluator accepts static variant ACs without engine changes |
| [PFC #2](../../phases/phase-1-pfc2-local-run.md) | Duplicate reachability, identity and consumer effects; wrapper pins reference metadata and requires control evidence | Not selected: changing architecture can trigger reference-identity UNKNOWN in addition to causal decisions |
| [PFC #3](../../phases/phase-1-pfc3-local-run.md) | Lease expiry, reacquisition, stale worker and observable tokens/effects; fencing is a control | Not selected: similarly bounded reference wrapper; no generalized lease claim |
| [Historical #001 architecture](../historical-reproductions/001-iceberg-16282/architecture.md) and [comparison](../historical-reproductions/001-iceberg-16282/comparison.md) | Reported recovery/replay prerequisites; catalog conflict is explicitly non-causal; local PostgreSQL proof is separate | Not selected: incident-specific mapping, not a fourth generalized PFC |
| [Historical #002 architecture](../historical-reproductions/002-github-may4-cascade/architecture.md) and [comparison](../historical-reproductions/002-github-may4-cascade/comparison.md) | Reported shared-capacity cascade versus direct local observation capabilities; quantitative details remain unknown | Not selected: service-level incident model, not a new historical reproduction |

The historical source records, recorded results, reproduction guidance and evaluator implementations inform this distinction; their evidence was not rerun or relabeled as matching-case evidence. No additional web claims or public-incident qualification is introduced.

## Scope and governance

This is an external-validation artifact under [External Sprint #001](../external-sprint-001/README.md), not a new PFC phase, Historical #003, customer execution or production execution. Matching is static only. Repository quality tests separately exercise the already authorized local synthetic fixtures.

[Canonical baseline](../../company/canonical-baseline-v1.0.md), [invariants](../../invariants.md), and [HANDOFF](../../../HANDOFF.md) retain authority. Gate A remains **REVISE**, mechanism fit **10/10**, qualifying public incidents **8/10**; families **1 and 3 remain unresolved**. PFC #4, Historical #003, broader Phase 1, customer execution and production execution remain **unauthorized**. No engine/verifier behavior, safety authorization, source classification or proof semantics is changed.
