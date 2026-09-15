# Design-partner conversation offer

**Audience:** a small or mid-size SaaS engineering team with backend infrastructure. This is a prepared offer for discussion, not an existing design-partner relationship, pilot, or commercial result.

## One scenario, a bounded evaluation

Give FailureMesh **ONE known reliability mechanism or sanitized/public postmortem**. For this preparation stage, use public material or a fully generalized synthetic mechanism; do not contribute customer/private evidence or confidential incident information. Sanitization must remove sensitive content rather than merely hide names.

FailureMesh attempts:

- Causal mechanism extraction into a candidate Portable Failure Contract.
- Applicability analysis against a bounded architecture description with explicit unknowns.
- Safe local/non-production experiment design.
- Definition of authoritative evidence requirements.
- A scoped outcome tied to the mechanism, architecture, invariant, experiment, and evidence.

## Initial ask

**20–30 minute technical conversation + one bounded reliability scenario + permission to evaluate a non-production/synthetic representation.** Discuss which architecture facts can be represented safely, what would block adoption, and what evidence would earn trust.

Permission to discuss or evaluate a representation is not permission to run against customer infrastructure. Current work is preparation only; customer execution and production execution remain unauthorized. Any future experiment outside the already approved local reference scope needs separate governance authorization and a bounded safety plan. This offer does not authorize PFC #4 or Historical #003.

## Privacy and limits

No production credentials. No uncontrolled production fault injection. No requirement to centrally upload raw customer source code, logs, secrets, or data. Raw private material remains local by default; only sufficiently generalized, explicitly approved structural information may be shared.

There is no promise that FailureMesh will find a bug, no promise of production safety, and no claim of product-market fit. UNKNOWN and a well-supported applicability exclusion can be useful results.

## Desired result

A useful design-partner outcome is evidence that FailureMesh either:

- **A.** correctly determines the mechanism is not applicable (**NOT_APPLICABLE**),
- **B.** determines evidence is insufficient and returns **UNKNOWN**, or
- **C.** constructs a bounded test that exposes or disproves the scoped failure.

For C, an executed test would require authoritative evidence for EXPOSED or scoped PROVEN_RESILIENT; absence of an observed failure alone does not disprove it. These are desired outcomes, not completed evidence. See the [reviewer brief](reviewer-brief.md) for current proof limits and the [scorecard](review-scorecard.md) for criticism and adoption blockers.
