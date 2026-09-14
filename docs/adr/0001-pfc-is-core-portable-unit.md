# ADR 0001: PFC is the core portable unit

Status: Accepted for Phase 0. Date: 2026-09-13.

Decision: A versioned PFC carries the causal mechanism, prerequisites, missing control, forbidden outcome, invariant, abstract experiment, proof requirements, safety restrictions, provenance, and verification level. Incident documents and platform scripts are inputs or adapters, not the portable unit. See [PFC v1](../architecture/pfc.md).

Reason: Portability requires causal and testable meaning independent of one incident's wording or one platform's injection technique. Consequence: a candidate extracted from text remains L0 until reproduced; Phase 1 must instantiate the concept before adding breadth.
