# PROPOSED BASELINE CHANGE

Status: **APPROVED on 2026-09-14** by the project owner: “Approve the bounded Phase 1 entry baseline change.” This record enacts that approval; it does not retroactively amend the historical [Gate A decision](../validation/gate-a/decision.md).

## Previous rule and approved rule

Previously, the ten-public-incident milestone was treated as a prerequisite for any Phase 1 implementation. Gate A's REVISE decision therefore recorded Phase 1 as NOT AUTHORIZED. The approved baseline distinguishes completion of research validation from explicit entry into a narrowly bounded engineering prototype. Ten qualifying public observed incidents, one per canonical family, are still required to claim the incident-validation milestone complete. A prototype may enter earlier only with recorded mechanism fit for its intended canonical scope, explicit unresolved evidence gaps, a narrow scope, local/reference-environment/customer-controlled execution, preserved privacy and safety controls, no implied production fault injection or broader evidence claims, and explicit project authorization. This approval supplies that authorization only for the existing first slice.

## Why and evidence

The distinction permits a controlled test of whether the conceptual PFC/AC and evidence model can work without pretending that missing public incidents have been found. The [Gate A checkpoint](../validation/gate-a/decision.md) records 10/10 exact-family L0 mechanism mappings, 8/10 qualifying public observed incidents, unresolved families 1 and 3, and three conceptual reference architecture shapes. The [fit review](../validation/gate-a/abstraction-fit.md) found no demonstrated model blocker to attempting the first slice. These are evidence for attempting a bounded prototype, not evidence that the missing incident milestone is complete or that FailureMesh has reproduced anything. Generic incident searching remains closed.

## Exact change and unchanged boundaries

The [canonical baseline](canonical-baseline-v1.0.md) gains a rule separating validation completion from bounded engineering entry; the [current disposition](../validation/gate-a/disposition.md) and [Phase 1 specification](../phases/phase-1-first-vertical-slice.md) apply this one approval. Gate A remains REVISE, mechanism fit remains 10/10 L0, qualifying incident coverage remains 8/10, and families 1 and 3 remain unresolved. The ten-incident milestone, mission, Portable Reliability Intelligence thesis, ten families, V0 targets and five fault primitives, PFC and AC definitions, three-state applicability and execution verdicts, verifier authority, provenance, UNKNOWN discipline, privacy, safety, and local/customer-controlled execution principles do not change. No source record or classification is changed.

## Authorization envelope

**BOUNDED PHASE 1 PROTOTYPE: AUTHORIZED.** Only the existing timeout-after-external-commit first vertical slice may be built: Go service, PostgreSQL, payment simulator, PFC, AC, deterministic applicability, causally synchronized post-commit response loss, retry and duplicate logical effect, EXPOSED counterexample, remediation, and rerun eligible for scoped PROVEN_RESILIENT only with complete evidence. Execution is limited to the synthetic level-1 reference environment under an approved safety plan. This governance change implements none of that software.

Broader Phase 1 or product expansion, customer execution, uncontrolled experiments, and production execution are **NOT AUTHORIZED** by this approval. Prototype authorization is not Gate A GO, a general “8/10 is enough” rule, public-incident validation for family 1 or 3, L1 reproduction, an earned execution verdict, customer applicability, portability, production readiness, market validation, commercial readiness, or broad product correctness. Future prototype results must be assessed against their actual evidence and scope; they do not automatically complete Gate A or historical reproduction of all families.

## Outstanding evidence and change control

Qualifying exact-family public observed incidents remain missing for families 1 and 3. Source chronology, independent reproduction, authoritative observer completeness, simulator fidelity, and transferability remain risks. Full incident-validation completion needs qualifying evidence for both missing families, a reviewed corpus/fit update, and a new explicit Gate A decision. Broader implementation, customer or production execution, or a different validation-entry rule requires separate authorization and, where it changes the canonical thesis or milestone rule, another **PROPOSED BASELINE CHANGE** with rationale and evidence. The historical checkpoint remains REVISE at 8/10.
