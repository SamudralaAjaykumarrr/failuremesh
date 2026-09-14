# FailureMesh

FailureMesh is a Portable Reliability Intelligence system. It converts previously discovered software failures into portable, privacy-preserving, executable protection for systems that have not experienced them yet. The [canonical company baseline](docs/company/canonical-baseline-v1.0.md) is authoritative; these engineering documents refine it without changing its thesis.

The loop is incident → Portable Failure Contract (PFC) → Architecture Contract (AC) → applicability (`APPLICABLE`, `NOT_APPLICABLE`, `UNKNOWN`) → experiment compilation → safety planning → customer-controlled execution → verifier verdict (`EXPOSED`, scoped `PROVEN_RESILIENT`, `UNKNOWN`) → counterexample or proof bundle. The LLM proposes hypotheses; the deterministic verifier evaluates evidence.

Phase 0 is complete. **Gate A remains REVISE** with 10/10 exact-family L0 mechanism mappings and 8/10 qualifying public observed incidents; families 1 and 3 remain unresolved, and the ten-incident validation milestone is incomplete. The [approved bounded-entry change](docs/company/baseline-change-001-bounded-phase1-entry.md) authorizes only the first Phase 1 prototype in a synthetic reference environment. Broader Phase 1, customer execution, and production execution are not authorized. The repository still contains specifications and repository-quality checks, not a working product or earned FailureMesh verdict. Start with [system architecture](docs/architecture/system.md), [invariants](docs/invariants.md), [failure model](docs/failure-model.md), and the historical [Gate A decision](docs/validation/gate-a/decision.md). [Roadmap](docs/roadmap.md) records the staged plan. No Go packages or product tests exist yet.

V0 is constrained to Go, PostgreSQL, Docker, one CLI, three reference services, ten PFCs, five fault primitives, and one deterministic proof format. Phase 1 implements only timeout-after-external-commit with a Go service, PostgreSQL, and an external payment simulator.

Local setup will be documented when implementation begins. The current `go.mod` establishes the module only; no dependencies are needed for Phase 0.
