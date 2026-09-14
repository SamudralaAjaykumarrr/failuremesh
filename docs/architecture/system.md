# System architecture

The canonical flow is incident → PFC → AC → applicability → experiment compilation → safety planning → local/customer-controlled proof execution → execution verdict → counterexample or proof bundle. `UNKNOWN` is a valid terminal result at applicability or execution. [Baseline](../company/canonical-baseline-v1.0.md) governs all components.

| Boundary | Responsibility | Authority |
| --- | --- | --- |
| mesh-compiler | Extract candidate mechanisms and PFC fields from incidents/postmortems | Candidate only; human/verifier review required |
| mesh-registry | Versioned PFCs, provenance, verification history, architecture families, evidence references, signatures | Distribution and history, not execution proof |
| mesh-agent | Local scanner, runner, redactor, safety enforcer, evidence collector | Controls local data and actions |
| mesh-match | Explain prerequisite evaluation against an AC | Deterministic applicability rules |
| mesh-runner | Map abstract experiment operations to platform-specific actions | Executes only an approved safety plan |
| mesh-proof | Check invariant and evidence completeness; emit scoped verdict and bundle | Execution verdict authority |
| mesh-control | Possible commercial policy, governance, history, analytics, private deployment | Future scope, not Phase 0/1 implementation |

Data flows from local facts into an AC with explicit evidence and unknowns. A versioned PFC and AC yield a reasoned applicability decision. Only `APPLICABLE` permits compilation of a candidate experiment. That candidate binds adapter versions, fault parameters, build, environment, invariant, observation plan, and safety-planning inputs; compilation does not authorize execution. Safety planning then approves an execution plan bound to the candidate or denies execution with a recorded reason. Only an approved plan permits mutation or fault injection. A denial leaves applicability unchanged and, if execution was requested, leaves execution `UNKNOWN` (not established). The proof verifier consumes immutable run evidence and emits one of the three execution verdicts. Shared export is a distinct, approved step after local processing.

V0 constraints: Go, PostgreSQL, Docker, one CLI, three reference services, ten PFCs, five fault primitives, one deterministic proof format. These are target boundaries, not present implementations. Phase 1 narrows the implementation to one PFC and one reference flow.
