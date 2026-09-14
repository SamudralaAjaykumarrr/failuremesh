# Engineering invariants

An **invariant** is a condition future phases must preserve. A **requirement** is an obligation for a particular deliverable. An **assumption** is an unverified input; a **hypothesis** is a proposition to test; **evidence** is a recorded observation; a **proof bundle** is a scoped, reproducible evidence package, not a mathematical proof. Unresolved facts produce `UNKNOWN`.

1. A PFC describes a causal failure mechanism, forbidden outcome, invariant, and proof needs; it is the portable unit. Incident text and scripts alone are not PFCs.
2. Applicability has exactly `APPLICABLE`, `NOT_APPLICABLE`, and `UNKNOWN`. Missing or conflicting architecture facts never become `NOT_APPLICABLE` or `APPLICABLE` by default.
3. Execution has exactly `EXPOSED`, scoped `PROVEN_RESILIENT`, and `UNKNOWN`. Never emit absolute `IMMUNE` or universal safety claims.
4. Every `PROVEN_RESILIENT` result binds PFC version, application/build identity, environment, invariant, experiment space, and collected evidence. A changed scope requires a new verdict.
5. The verifier, using versioned rules and recorded evidence, issues verdicts. LLM output is advisory and cannot create observations, relax proof requirements, override safety policy, or issue verdicts.
6. Evidence retains provenance, integrity, ordering, and enough execution context to audit or replay the claim. An absent observation is not evidence that an effect did not occur unless its observer and completeness are established.
7. Raw customer-sensitive inputs stay local by default. Shared structural contributions require generalization and explicit approval. Fail closed on ambiguous export classification.
8. Safety level is a ceiling enforced before and throughout execution. No production fault injection by default; level 5 needs explicit authorization and bounded controls.
9. Contracts, verifier rules, experiment definitions, adapters, and evidence formats are versioned. Historical verdicts remain interpretable against the versions used.
10. An experiment may claim resilience only for a declared finite experiment space with complete required observations and successful safety controls. Inconclusive or aborted runs are `UNKNOWN`.
11. Deterministic behavior must be reproducible from recorded inputs; random seeds, timing windows, and simulator behavior are recorded when relevant.
12. Architecture mapping and adaptation preserve the PFC's causal mechanism and invariant. An adapter that cannot establish that mapping must return `UNKNOWN`.
13. Applicability prerequisites establish whether the causal experiment is meaningful and executable, never whether exposure is predicted. Missing/insufficient controls are vulnerability hypotheses or controls-under-test. Control presence alone cannot yield `NOT_APPLICABLE`; only a PFC-defined, evidenced exclusion that defeats the experiment itself can do so.
14. Fault placement required by a PFC must be established by causal synchronization and evidence, not a wall-clock race. For timeout-after-external-commit, the provider commit and authoritative evidence precede response loss, which precedes caller success delivery.
