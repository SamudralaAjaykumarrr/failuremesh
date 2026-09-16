# Autonomous evidence validation — Phase 1 foundation

Status: **architecture specification for review; no implementation or new execution authorization**. This package is authoritative for the requested design scope, subordinate to the [canonical baseline](../company/canonical-baseline-v1.0.md), [invariants](../invariants.md), and existing approved scope. “Phase 1” here names the evidence-validation design initiative, not permission to implement broader prototype Phase 1.

FailureMesh turns real software failures into portable, evidence-backed failure models that other systems can automatically evaluate and safely test before experiencing the same failure themselves. The thesis and product loop are preserved. The design removes external approval as a prerequisite for producing technical evidence within an explicitly authorized envelope. Experts remain useful critics; customers and design partners remain necessary for later commercial validation. Neither votes facts into truth.

## Read the design

| Document | Authority within this package |
| --- | --- |
| [Foundation](phase-1-foundation.md) | Decisions, component boundaries, execution, acceptance and failure criteria |
| [Evidence model](evidence-model.md) | Sources, atomic claims, admissibility, causal lineage, independent architecture packets, privacy |
| [Evaluation protocol](evaluation-protocol.md) | Freeze/registration, grading, corpus selection, leakage and metrics |
| [Counterfactual model](counterfactual-model.md) | Deterministic mutations, conditional expectations, permanent counterexamples |
| [Proof bundle](proof-bundle.md) | Content identity, evidence trust, replay and immutable history |
| [Roadmap](roadmap.md) | Separately authorized future stages and unresolved decisions |

```text
public failure source → source artifact → atomic evidence claims
→ causal failure model / versioned PFC
                         × independent architecture evidence packet
→ deterministic applicability → adversarial counterfactuals
→ optional separately authorized bounded execution
→ authoritative execution evidence → scoped verdict
→ content-addressed replayable proof → permanent regression corpus
```

Every arrow records input identities, policy/tool versions, output identity and limitations. Source availability, what a source reports, and the truth of that report are separate claims. Determinism establishes reproducibility of a decision from premises; it does not establish that arbitrary prose or declarations are true.

## Existing evidence and design motivation

[Match Evaluation #001](../validation/match-evaluation-001/README.md) provides three inspectable designed cases, not natural architecture truth or a population benchmark. Its [grading record](../validation/match-evaluation-001/grading-method.md) explicitly discloses agent-prepared labels, knowledge of matcher rules, self-reported chronology, and declaration-trusting evaluation. Its [results](../validation/match-evaluation-001/results.md) contain no observed FP/FN. The present task also reports independent criticism of easy cases and outstanding natural samples; that task context is motivation, not a newly recorded expert-review completion.

This design adds independent packet collection, evidence admissibility before matching, a sealed evaluation procedure, inspectable error cases, and challenges that can fail the matcher. No accuracy improvement, public-corpus acquisition, attestation service, generalized extraction, or new runtime capability is claimed to exist.

## Governance checkpoint

Gate A remains **REVISE**. Mechanism fit remains **10/10**. Qualifying public incidents remain **8/10**. Families **1 and 3 remain unresolved**. PFC #4 remains **unauthorized**. Historical #003 remains **unauthorized**. Customer execution remains **unauthorized**. Production execution remains **unauthorized**. Broader Phase 1 implementation remains unauthorized. The approximately 25-incident research corpus proposed here is separate from Gate A's ten canonical-family incident requirement; it neither reopens the closed generic search nor changes its counts.

[External Sprint #001](../validation/external-sprint-001/README.md) and the five-review target remain outstanding. The [approved parallel validation sequence](../company/baseline-change-004-nonblocking-validation-sequence.md) and [Historical #002 authorization](../company/baseline-change-005-historical-reproduction-002.md) remain unchanged. No baseline amendment or new ADR is necessary to specify this evidence layer: it refines [ADRs 0001–0006](../adr) without replacing their decisions. Any later proposal to remove governance milestones or broaden execution must follow the baseline change process.
