# FailureMesh

FailureMesh is a Portable Reliability Intelligence system. It converts previously discovered software failures into portable, privacy-preserving, executable protection for systems that have not experienced them yet. The [canonical company baseline](docs/company/canonical-baseline-v1.0.md) is authoritative; these engineering documents refine it without changing its thesis.

The loop is incident → Portable Failure Contract (PFC) → Architecture Contract (AC) → applicability (`APPLICABLE`, `NOT_APPLICABLE`, `UNKNOWN`) → experiment compilation → safety planning → customer-controlled execution → verifier verdict (`EXPOSED`, scoped `PROVEN_RESILIENT`, `UNKNOWN`) → counterexample or proof bundle. The LLM proposes hypotheses; the deterministic verifier evaluates evidence.

Phase 0 is complete. **Gate A remains REVISE** with 10/10 exact-family L0 mechanism mappings and 8/10 qualifying public observed incidents; families 1 and 3 remain unresolved, and the ten-incident validation milestone is incomplete. The [first bounded-entry change](docs/company/baseline-change-001-bounded-phase1-entry.md) authorizes the timeout-after-external-commit prototype; a [separate approved change](docs/company/baseline-change-002-bounded-pfc2-entry.md) authorizes only the duplicate-queue-delivery synthetic slice. Broader Phase 1, customer execution, and production execution are not authorized. These bounded reference flows have produced local synthetic proofs; they do not supply qualifying public incidents. Start with [system architecture](docs/architecture/system.md), [invariants](docs/invariants.md), [failure model](docs/failure-model.md), and the historical [Gate A decision](docs/validation/gate-a/decision.md). [Roadmap](docs/roadmap.md) records the staged plan.

V0 is constrained to Go, PostgreSQL, Docker, one CLI, three reference services, ten PFCs, five fault primitives, and one deterministic proof format. The bounded Phase 1 reference uses a Go caller, PostgreSQL, and an in-process payment simulator with an independently committed provider effect.

The first runnable reference flow and exact commands are documented in [Phase 1 local run](docs/phases/phase-1-local-run.md). The separately approved synthetic duplicate-delivery slice is documented in [PFC #2 local run](docs/phases/phase-1-pfc2-local-run.md).

The separately approved [PFC #3 local run](docs/phases/phase-1-pfc3-local-run.md) demonstrates deterministic synthetic stale-worker fencing against PostgreSQL. It remains an L0 candidate and does not close the family-3 Gate A incident gap.
