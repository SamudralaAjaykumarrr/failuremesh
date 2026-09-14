# ADR 0002: Three-state applicability

Status: Accepted for Phase 0. Date: 2026-09-13.

Decision: Applicability returns `APPLICABLE`, `NOT_APPLICABLE`, or `UNKNOWN` with a reasoned prerequisite/evidence trace. A proved false necessary condition yields `NOT_APPLICABLE`; all proved true conditions plus valid mapping yield `APPLICABLE`; unresolved cases yield `UNKNOWN`. See [semantics](../architecture/applicability.md).

Necessary conditions concern whether the causal experiment can be meaningfully executed. A missing or insufficient control is a vulnerability hypothesis/control-under-test, not automatically a necessary condition. A protected architecture remains `APPLICABLE` when the experiment can still exercise its causal path; only an explicit, evidenced mechanism-defeating exclusion can change that.

Reason: Binary matching would turn absent or conflicting architecture facts into unjustified certainty. Consequence: `UNKNOWN` is a useful work item for fact collection and cannot authorize an experiment.
