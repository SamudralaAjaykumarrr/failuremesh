# FailureMesh

FailureMesh is a Portable Reliability Intelligence system. It converts previously discovered software failures into portable, privacy-preserving, executable protection for systems that have not experienced them yet. The [canonical company baseline](docs/company/canonical-baseline-v1.0.md) is authoritative; these engineering documents refine it without changing its thesis.

The loop is incident → Portable Failure Contract (PFC) → Architecture Contract (AC) → applicability (`APPLICABLE`, `NOT_APPLICABLE`, `UNKNOWN`) → experiment compilation → safety planning → customer-controlled execution → verifier verdict (`EXPOSED`, scoped `PROVEN_RESILIENT`, `UNKNOWN`) → counterexample or proof bundle. The LLM proposes hypotheses; the deterministic verifier evaluates evidence.

Phase 0 is complete; the repository is at **Pre-Phase-1 Validation Gate A (REVISE)**. It contains specifications and repository-quality checks, not a working product. Canonical incident coverage is 8/10, with families 1 and 3 unresolved; Phase 1 is not authorized. Start with [system architecture](docs/architecture/system.md), [invariants](docs/invariants.md), [failure model](docs/failure-model.md), and the [Gate A decision](docs/validation/gate-a/decision.md). [Roadmap](docs/roadmap.md) records the staged plan. No Go packages or product tests exist yet.

V0 is constrained to Go, PostgreSQL, Docker, one CLI, three reference services, ten PFCs, five fault primitives, and one deterministic proof format. Phase 1 implements only timeout-after-external-commit with a Go service, PostgreSQL, and an external payment simulator.

Local setup will be documented when implementation begins. The current `go.mod` establishes the module only; no dependencies are needed for Phase 0.
