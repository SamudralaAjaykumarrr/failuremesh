# Applicability semantics

Applicability asks whether the PFC's causal mechanism can operate in the AC's scoped architecture, given the available facts. It is a deterministic, explainable predicate evaluation, not an execution verdict. Evaluate PFC and AC versions together and record each prerequisite, its AC fact/evidence, and any unresolved or conflicting input.

| Verdict | Exact meaning |
| --- | --- |
| `APPLICABLE` | Every required prerequisite is supported by sufficiently current evidence for the scoped operation, no known exclusion defeats the mechanism, and the abstract experiment has a causal mapping to that architecture. This does not assert that the forbidden outcome occurs. |
| `NOT_APPLICABLE` | At least one necessary prerequisite is demonstrably false, or a specified exclusion is demonstrably true, for the scoped operation. The decisive fact and evidence must be recorded. It does not mean the whole system is immune to the family. |
| `UNKNOWN` | Neither of the above is justified: a required fact is missing, stale, conflicting, insufficiently evidenced, or the mapping/exclusion cannot be resolved. No experiment is authorized from this result. |

Only facts necessary for a meaningful, executable causal experiment are applicability prerequisites. Vulnerability hypotheses and controls-under-test, including idempotency/reconciliation effectiveness, determine the observed outcome and are not generic applicability predicates. A PFC may define a mechanism-defeating exclusion only where evidence shows the experiment itself is impossible or meaningless. A protected architecture can therefore be `APPLICABLE` and later earn scoped `PROVEN_RESILIENT` through execution; control presence alone proves neither verdict.

Evaluation order is evidence-aware: a proved false necessary condition is enough for `NOT_APPLICABLE` even if other facts are unknown; otherwise all prerequisites must be proved true for `APPLICABLE`; all remaining cases are `UNKNOWN`. Disjunctions and conditional prerequisites must be evaluated using the same three-valued logic, with an explanation tree. An AI classification is a candidate AC fact until corroborated by an allowed source. Any change to PFC version, AC scope, or decisive facts invalidates reuse of the old decision.
