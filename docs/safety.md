# Execution safety

Safety level describes the maximum environment and action allowed, exactly as in the [baseline](company/canonical-baseline-v1.0.md):

| Level | Scope |
| --- | --- |
| 0 | static applicability only |
| 1 | synthetic reference architecture |
| 2 | developer machine / CI |
| 3 | ephemeral staging |
| 4 | customer staging |
| 5 | explicitly authorized controlled production |

After an `APPLICABLE` verdict, compilation produces a candidate experiment manifest; it does not authorize execution. Safety planning evaluates that candidate against the PFC's safety restrictions, environment identity, requested safety level and fault primitive, policy, operator authorization, and applicable blast-radius, time, rate, health-precondition, and abort controls. The effective level is no higher than every applicable ceiling. Planning either approves an execution plan bound to the compiled candidate or denies execution with an explicit reason. No mutation or fault injection may occur without the approved plan. Denial leaves the applicability verdict unchanged; if execution was requested, its result remains `UNKNOWN` (not established), with the reason recorded.

Production is never the default. Level 5 eventually requires explicit authorization, blast-radius constraints, abort conditions, health preconditions, time and rate budgets, audit records, and recovery/rollback controls where possible. The runner must monitor abort conditions during execution and stop safely when triggered. A stopped experiment is not negative proof. Phase 1 runs at level 1 only; it has no production path.
