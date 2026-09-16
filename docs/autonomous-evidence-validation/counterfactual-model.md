# Deterministic counterfactual challenges

Status: design only. A mutation changes a synthetic descendant artifact or evidence view, never the historical source or natural packet. No execution mutation runs without separate authorization. [Applicability semantics](../architecture/applicability.md) and [invariant 13](../invariants.md) remain authoritative.

## Mutation record and oracle

Each mutation records ID, operator/version, parent registration and packet/PFC digests, target claim/edge and scope, typed precondition, seed, exact patch, derived artifact digests, evidence changes, semantic expectation or relation, justification independent of matcher output, actual result and grade. Deterministic generation orders eligible targets by ID; identical parent/operator/seed produces identical descendants. A mutation is synthetic even if its parent is natural. Never manufacture a new authoritative natural fact by toggling a boolean.

Use two modes: (1) a coherent synthetic architecture with supplied bounded support for a changed fact, testing matcher semantics; (2) an evidence perturbation, testing admissibility and uncertainty. A bare value flip unsupported by evidence belongs to mode 2 and yields UNSUPPORTED, not evidenced FALSE. Proposed LLM mutations pass the same typed validator as any other candidate. Invalid preconditions produce `not instantiated` with a reason; they do not count as passing tests.

For the table, assume a valid parent APPLICABLE case with a conjunctive necessary prerequisite set, all other prerequisites supported TRUE, no exclusion, and unchanged scope/mapping unless stated. Recompute the whole explanation tree. A supported false necessary condition elsewhere can still make NOT_APPLICABLE despite an unknown changed fact. For disjunctions/conditionals use the registered three-valued expression, not this simplified conjunction. If the expected label cannot be logically derived, register a relation or disputed expectation before running.

| Mutation | Required semantic expectation |
| --- | --- |
| TRUE prerequisite → evidenced FALSE | NOT_APPLICABLE for a coherent scoped necessary-condition negation; unsupported value flip instead yields UNKNOWN |
| TRUE prerequisite → UNKNOWN | UNKNOWN under table assumptions; never a default negative |
| Evidence removed | UNKNOWN if last admissible support removed; unchanged if independent sufficient support remains |
| Evidence stale | UNKNOWN if all admissible support fails the frozen freshness policy; pinned historical facts are not aged arbitrarily |
| Evidence downgraded | UNKNOWN if no longer meets predicate admissibility; otherwise unchanged with updated evidence rationale |
| Contradictory evidence introduced | CONFLICTING fact → UNKNOWN when contradiction is material/admissible/unresolved; record rejection of out-of-scope or unsupported contradiction |
| Source provenance weakened | UNKNOWN when attribution/digest/scope cannot satisfy the rule; a new URL for identical verified bytes alone need not change truth |
| Architecture prose conflicts with structured fact | If conflict is established, block the affected fact pending resolution; semantic detection failure is an extraction counterexample, not license to ignore prose |
| Protective control introduced, mechanism retained | Remains APPLICABLE if mapping and proof capability remain; no inferred PROVEN_RESILIENT, control tested separately |
| Mechanism-breaking exclusion introduced | NOT_APPLICABLE only for a PFC-defined exclusion with coherent scoped evidence; generic “has protection” is insufficient |
| Causal ordering altered | Demonstrated impossibility of a required order → NOT_APPLICABLE; merely unestablished order → UNKNOWN; reordered execution evidence cannot prove the original mechanism |
| Scope changed | Rebind all supports; without support in new scope UNKNOWN; explicit false in new scope can yield NOT_APPLICABLE. Old verdict cannot transfer |
| Identity/retry semantics changed | Different logical operation or no retry → NOT_APPLICABLE if this negates a necessary predicate; ambiguous identity → UNKNOWN; more retries with unchanged mechanism need not alter applicability |
| Cross-boundary ambiguity introduced | UNKNOWN when transaction, ownership, effect or observer boundary is required and unresolved |
| Irrelevant similarity injected | Verdict and decisive causal trace unchanged for non-causal vendor/name/word similarity; if it changes scope it is not an irrelevant mutation |

Evidence additions cannot automatically strengthen certainty: relevant contradictory additions may reduce it. Removing evidence cannot create a newly supported fact. A protected architecture remaining APPLICABLE is a required challenge, not a false positive. Failures to recognize unsupported assertions or a contradictory prose field must remain visible even if the legacy matcher is behaving as currently implemented.

## Challenge coverage

Register each operator against at least one eligible supported scope and a negative precondition case. Cover both independent support retained and last-support removed; decisive FALSE with unrelated UNKNOWN; conditional/disjunctive expressions when later supported; and control versus exclusion distinctions. Cross-operator compositions include weakened provenance plus conflicting scope and identity change plus protective control. Cap composition depth and count in the registered budget so an operator cannot flood metrics with trivial variants.

Use metamorphic relations when labels are not known: irrelevant similarity preserves the verdict, equivalent fact ordering preserves explanation meaning, invalidating the only support forbids an unsupported determinate result. Counterfactual success is separate from natural-case transfer success. No target mutation pass rate authorizes broader execution or labels a corpus accurate.

## Permanent counterexamples

Use proposed stable IDs `FM-COUNTEREXAMPLE-000001` onward; allocate once from an append-only registry in later tooling. Distributed collection first uses content digests, then assigns a unique display ID; never recycle withdrawn IDs. No IDs are allocated by this document.

A counterexample includes original and minimized case digests, expected semantic argument, observed raw result, violated rule, source/claim lineage, evaluator version, mutation chain, replay, sensitivity, split/origin cluster, severity and dispute history. Preserve unsuccessful minimization and the original evidence. Minimization must retain the independent rationale and failure; it cannot simplify away the contested context.

Classify failures as extraction, source admissibility, PFC abstraction, projection, matcher, oracle/grading, execution mapping or verifier failure. A surprising result is initially a candidate counterexample; disputed grading is not erased to make a test pass. Once reproduced and resolved under a versioned rule, retain the case permanently in regression. Fixes append resolution and new expected behavior/version; old outputs and original grades remain inspectable. Automated regression additions are permitted only under a later authorized tooling envelope; changing production semantics, accepted policy, or governance is a separate action.
