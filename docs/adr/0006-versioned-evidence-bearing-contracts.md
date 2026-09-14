# ADR 0006: Versioned, evidence-bearing contracts

Status: Accepted for Phase 0. Date: 2026-09-13.

Decision: PFCs, ACs, experiment manifests, adapter behavior, verifier rules, and proof bundles carry versions and provenance. AC facts include evidence or explicit unknowns; PFC verification levels require recorded reproduction evidence. Historical artifacts remain interpretable under the versions that produced them.

Reason: Silent contract changes make verdicts non-reproducible and can turn old evidence into a new claim. Consequence: substantive PFC revisions create new versions, changed AC scope/facts trigger re-match, and claims bind exact artifact identities. See [PFC](../architecture/pfc.md), [AC](../architecture/architecture-contract.md), and [registry](../architecture/registry.md).
