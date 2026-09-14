# ADR 0005: Scoped PROVEN_RESILIENT verdict

Status: Accepted for Phase 0. Date: 2026-09-13.

Decision: `PROVEN_RESILIENT` is an empirical verdict only for a PFC version, application/build identity, environment, invariant, finite experiment space, and collected evidence. The verifier must confirm complete runs and observations. Never use absolute `IMMUNE`.

Reason: Passing a bounded test cannot establish universal safety. Consequence: changed build, environment, PFC, invariant, experiment space, or insufficient evidence requires reevaluation or `UNKNOWN`. See [proof semantics](../architecture/proof.md).
