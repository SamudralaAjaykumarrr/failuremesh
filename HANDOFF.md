# Handoff

- Current maturity: **ARCHITECTURE FOUNDATION**.
- Current phase: **Phase 0**.
- Branch: `main` (unborn at Phase 0 creation).
- Completed work: canonical company baseline is present; Phase 0 engineering specifications, architecture decisions, and Phase 1 implementation obligations are documented.
- Current task: validate and review the Phase 0 foundation. No product implementation is authorized in Phase 0.
- Next task after Phase 0 acceptance: **Validation Gate A — Corpus & Reference Architecture Bootstrap**. Encode 10 public incidents as L0 candidate PFCs and specify 3 reference architecture shapes. Check that materially different real failure mechanisms fit the PFC abstraction without forcing them into it. This is pre-implementation validation, not reproduction or verification. Only after this gate passes, implement the [Phase 1 first vertical slice](docs/phases/phase-1-first-vertical-slice.md), timeout-after-external-commit. The remaining canonical validation sequence still governs later work.
- Authoritative docs: [canonical baseline](docs/company/canonical-baseline-v1.0.md) first; then [invariants](docs/invariants.md), [architecture](docs/architecture/system.md), [ADRs](docs/adr), and the current [phase specification](docs/phases/phase-1-first-vertical-slice.md).
- Known risks: architecture facts may be unavailable; instrumentation may fail to distinguish retry from duplicate logical effect; simulator fidelity may overstate transferability; negative execution evidence may be too weak for a resilience claim; adapters and safety controls remain unimplemented.
- Prohibited deviations: baseline thesis changes without the baseline change process; absolute immunity claims; LLM-issued verdicts; silent `UNKNOWN` coercion; raw sensitive data leaving the local boundary by default; uncontrolled production experiments; expanding Phase 1 beyond its single failure family; claiming future validation has occurred.
