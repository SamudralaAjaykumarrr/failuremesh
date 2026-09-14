# ADR 0003: Verifier, not LLM, is authority

Status: Accepted for Phase 0. Date: 2026-09-13.

Decision: AI may extract candidates, suggest mappings, experiments, and summaries. Deterministic rules evaluate applicability, safety eligibility, evidence completeness, invariants, and execution verdicts. AI cannot issue `PROVEN_RESILIENT`, fabricate evidence, override safety policy, or turn uncertainty into certainty.

Reason: Verdicts need reproducible, auditable premises. Consequence: every verdict records rule/version and evidence; AI suggestions stay explicitly unverified until checked. See [proof](../architecture/proof.md).
