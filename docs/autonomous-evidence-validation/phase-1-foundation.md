# Phase 1 architecture foundation

Status: **design specification for independent architecture review**. Requirements here govern subsequent proposals; no implementation is delivered or authorized. Existing [baseline](../company/canonical-baseline-v1.0.md), [system architecture](../architecture/system.md), [safety](../safety.md) and [privacy](../privacy.md) remain authoritative.

## Decisions and responsibilities

The system transfers causal failure knowledge across systems and tests that transfer against evidence. It does not become an incident search engine, postmortem summarizer, AI reliability chatbot, scenario recommender, generic chaos library or generic fault injector. Its differentiator is the combination of cross-system knowledge transfer, causal applicability, evidence provenance, adversarial self-challenge and bounded executable proof. Search and execution are replaceable supporting capabilities.

| Decision | Reason and consequence |
| --- | --- |
| Evidence-bearing envelope before matching | Matcher correctness and architecture truth are separate; inspect admissibility, scope and derivation before projecting AC facts |
| Atomic claims with non-binary support states | Missing, stale, conflicting and unsupported inputs remain visible and project to UNKNOWN |
| Independent target-blind packet inventory | Reduces fixtures constructed around known rules; independence and exposure are recorded, not assumed |
| Immutable PFC/packet versions and append-only assessments | Corrections never rewrite historical evidence or labels |
| Registered case-level evaluation and counterfactuals | Negative results and oracle disputes remain inspectable; metrics cannot replace cases |
| Content-addressed evidence envelope | Supports lineage and replay without claiming hashes establish truth or chronology |
| Replaceable execution adapters, independent proof authority | FailureMesh owns mechanism, rationale, evidence requirements and proof; backend owns bounded actuation/collection capabilities |
| Policy-bounded autonomous progress | Deterministic checks can progress without expert approval; authorization, source truth and commercial validation cannot be inferred |

Retain existing component names rather than introduce a new service architecture:

- **mesh-compiler:** discovers/extracts candidates, emits causal DAG/PFC proposals with evidence; no authoritative verdict or execution permission.
- **mesh-agent:** future public/local evidence capture and packet construction, classification, authoritative observers and local safety enforcement. Arbitrary source content remains data.
- **mesh-registry:** source/claim/PFC/packet versions, origin clusters, frozen manifests, invalidations and counterexamples. Storage does not certify truth.
- **mesh-match:** deterministic claim-policy projection and applicability trace, with separately versioned evidence validator and matcher. Unsupported mappings abstain.
- **mesh-runner:** candidate experiment compilation and replaceable backend capability mapping; execution only within an approved plan.
- **mesh-proof:** completeness, integrity, invariant checks, scoped execution verdict and proof envelope. A match or successful backend job is not proof.
- **mesh-control:** no implementation here; future governance/commercial functions retain separate authorization.

A future local CLI/library is sufficient for the first implementation proposals. No crawler, distributed service, cloud registry, customer agent or new PFC is necessary to review this design.

## Current implementation boundary

The three [contracts](../../contracts) are narrowly pinned L0 candidate PFCs with string prerequisites and synthetic provenance. The shared [evaluator](../../internal/phase1/contracts.go) validates supplied boolean facts, nonempty evidence/source strings, quality, stale/conflict flags and some AC metadata; it does not verify semantic entailment from source artifacts. PFC #2/#3 wrappers also constrain reference identity, controls and metadata. Their scope refusals must not be mistaken for natural architecture causal negatives. The [proof design](proof-bundle.md) records differing current verifier paths accurately.

This package adds no schemas, parser, mutation executable, evaluator preprocessor or automatic fact promotion. New evidence validation will require separately authorized schema/tooling work, explicit predicate coverage and compatibility tests. Legacy proofs remain replayable only under their original trust model. Existing architecture docs and all six ADRs stay unchanged.

## Autonomous technical validation flow

1. Within an allowed source policy, detect a candidate and record attribution, retrieval, content identity, rights and uncertainty.
2. Extract atomic candidate claims. Validate typed evidence with pinned predicate-specific rules; keep unsupported prose interpretations explicit.
3. Build a causal DAG and candidate PFC with every prerequisite justified; unresolved necessity remains hypothetical, not a validated transferable contract.
4. Independently freeze architecture packets, then register pairings, evidence sets, expected procedure and all relevant versions.
5. Deterministically evaluate supported scoped facts into APPLICABLE / NOT_APPLICABLE / UNKNOWN and retain the full explanation tree.
6. Challenge the result using registered counterfactuals. Preserve failures, disputes and regression candidates.
7. If APPLICABLE, optionally compile a bounded experiment. Verify capability and separately approved safety authority before any execution.
8. Collect authoritative observations and evaluate complete evidence into EXPOSED / scoped PROVEN_RESILIENT / UNKNOWN.
9. Emit a content-addressed bundle, replay it at its declared capability tier, append invalidations/counterexamples and expose inspectable results.

These stages form the future continuous pipeline: public sources → candidate detection → source capture → extraction → deterministic evidence validation → candidate PFC → independent architecture corpus → applicability matrix → counterfactual tests → authorized bounded reproductions → proof bundles → regression corpus. Queues should be idempotent by artifact/policy digest, bounded by explicit acquisition/compute budgets, and resumable without discarding failed attempts. Reprocessing under a new rule creates a new assessment. No crawler or bulk ingestion is built now.

## Deterministic versus LLM authority

LLMs may discover sources, extract candidate facts, suggest causal links/PFC mappings/architecture facts, generate adversarial hypotheses and summarize evidence. Their output is marked candidate and its model/prompt/tool context recorded. AI-only claims never satisfy authoritative prerequisites. Confidence is advisory. A human approving an LLM answer does not turn it into an observation.

Versioned deterministic rules validate structural integrity, typed evidence, policy admissibility, freshness, scope, conflict handling, matching, safety eligibility, completeness and invariants. Semantics that cannot be mechanically established or bound to admissible scoped evidence remain UNKNOWN. Neither LLM nor evaluator certifies internet truth. No LLM upgrades UNKNOWN, reconciles contradictions silently, edits grading after results, owns a verdict, declares a system safe, or authorizes execution.

## Bounded execution compilation — abstract design only

APPLICABLE permits a candidate compilation, never execution by itself. Require exact PFC/packet/AC/evaluator identities, current admissibility and no invalidation, scoped application/build/environment, executable causal mapping, finite parameter space, invariant, observer completeness and backend capability evidence. If mapping/fault placement cannot be demonstrated, record execution unavailable/UNKNOWN without injection; preserve the applicability result and its scope.

The candidate manifest binds logical identities, abstract actions/causal barriers, backend/adapter versions, fault budget (maximum operations, concurrency, duration, rate and affected resources), environment allowlist, stop conditions, health preconditions, initial-state/reset/cleanup and recovery plan. No unbounded or implicit default budgets. A safety plan binds the candidate digest, authorizing policy/owner record, maximum level and validity interval. Effective authority is the intersection of all ceilings, not the most permissive one. Check immediately before and during execution.

Current authorized experiments remain only the separately approved local synthetic level-1 slices/reproductions. Future local/container/CI execution requires a named pre-authorization envelope; level 2 is not granted by mentioning CI. Ephemeral/customer staging and production are not authorized. Production is forbidden by default, and this phase creates no production path.

Stop on budget exhaustion, unexpected target identity, loss of authoritative observer, violated health threshold, unplaceable causal fault, escape from resource boundary, expired/revoked authority or failed cleanup preconditions. Abort safely, record partial evidence, and do not infer resilience. UNKNOWN is the normal incomplete/aborted outcome; the existing execution specification permits EXPOSED only if independently sufficient trustworthy authorized-run evidence already establishes the violation. Cleanup failure is explicit and prevents an automatic next run.

| Adapter capability | Contract with FailureMesh |
| --- | --- |
| Discover/describe | Return versioned capabilities and supported target scope; no authority to execute |
| Compile/map | Map abstract actions and causal synchronization points; declare unsupported features |
| Plan/validate | Verify target isolation, permissions, finite budgets, preconditions and abort/recovery capability |
| Execute/stop | Enforce approved manifest exactly; idempotent run identity, observable stop acknowledgment |
| Observe/export locally | Correlated authoritative evidence with observer identity, ordered events, completeness and errors |
| Reset/replay | Recreate isolated initial state and declared experiment space without touching unrelated resources |

Potential adapters include built-in local runner, local containers, Kubernetes-native execution, Gremlin-like, AWS FIS-like, Azure Chaos Studio-like and future adapters. These are abstract target classes, not verified capability claims or integration commitments. An adapter lacking required causal placement or authoritative observations is unsuitable for that PFC. FailureMesh owns WHAT mechanism applies, WHY, WHAT evidence supports it and WHAT constitutes proof; an execution vendor's success flag never replaces the verifier. No services are purchased or integrations implemented here.

## Autonomy envelope

Later explicitly authorized tooling may automatically retrieve allowed public sources, propose claims/models, evaluate deterministic rules, run counterfactuals, generate bundles and add regression cases. It may execute only experiments already covered by an explicit local/sandbox pre-authorization, with a valid per-candidate safety plan. Technical UNKNOWNs and failures can be terminal outputs without asking a human to choose the “right” fact. External criticism can arrive asynchronously and create appended challenges.

It must not access private customer systems, run in production, buy services, bypass gates, promote weak evidence, change accepted policy or governance, remove outstanding expert milestones, claim customers will pay, declare commercial validation, or claim general safety. This design task authorizes documentation and requested quality checks only. No later roadmap stage is automatically authorized. Governance changes follow the baseline process; technical records do not depend on experts agreeing with them.

## Acceptance for subsequent implementation

Architecture review is a design milestone, not proof that these capabilities work. Subsequent implementation acceptance requires recorded artifacts and adversarial checks for every obligation below, under separate authorization.

| Obligation | Measurable acceptance evidence |
| --- | --- |
| Complete lineage | 100% of decisive facts and PFC prerequisites resolve through versioned derivations to permitted evidence or explicit engineering proof obligations; zero dangling refs or unlabeled hypotheses |
| No AI-only authority | Every AI-only/missing/unsupported required fact challenge projects to UNKNOWN; no confidence-based promotion |
| Deterministic states | All six claim states, conflict/scope/freshness failures, decisive FALSE with unrelated UNKNOWN, and control/exclusion cases exercise documented semantics |
| Honest extraction | Every authoritative fact names its entailment rule or evidence-bound interpretation; unsupported prose cannot pass on citation presence alone |
| Natural independence | Corpus inventory targets approximately 25 incidents/five families and 20–30 packets; at least 20 natural packets and 80% frozen before target assignment, with all deviations disclosed |
| Registration | Every evaluation binds frozen inputs, expected procedure and evaluator; any post-run edit creates a visible revision; chronology tier explicit |
| Inspectable errors | Every registered case/attempt retained; taxonomy, metrics and exclusions resolve to case IDs; no forced FP/FN or manufactured passing labels |
| Adversarial coverage | Every listed mutation instantiated where valid, with reason for exceptions; demonstrate a deliberately faulty test evaluator is caught and retained as a labeled synthetic regression, without claiming a natural FP |
| Proof replay | At least one complete permitted bundle replays in two clean environments under the declared comparator; tampered/missing decisive artifacts rejected; restricted-source replay limits explicit |
| Execution separation | Static/denied/unavailable cases cannot yield positive execution proof; any later execution demonstrates stop/budget/observer/isolation enforcement and authoritative evidence |
| Privacy/security | Injected private markers, secrets and source instructions are quarantined; export/prompt/telemetry paths tested; no restricted text redistributed by default |
| Compatibility/governance | Existing authorized regressions pass; baseline/Gate A/engine semantics unchanged unless separately approved; no new execution authority inferred |

These are proposed future acceptance requirements, not results earned by these documents. The curated corpus is a target with shortfalls visible; it is not a license to lower evidence quality to hit counts.

## Failure and kill criteria

Pre-register numerical budgets before the first validation reveal. The following are initial decision thresholds for that future protocol, not current measurements, statistically established population limits, or new governance gates. Small samples require reporting counts and case evidence; an inconclusive denominator is not success. A stop suspends broader technical claims and expansion, while preserving counterexamples and permitting separately authorized diagnosis. It does not autonomously pivot the company or edit the baseline.

| Risk | Evidence that requires stop/revision or weakens the thesis |
| --- | --- |
| False-positive burden | Any unjustified execution-enabling certainty blocks release; two resolved natural-case false positives in a validation round, or FP rate above a pre-registered 10% investigation threshold, require abstraction/evidence review |
| Missing necessary context | At least three independently sourced transfer cases need hidden incident-specific context that cannot be represented without reconstructing the original system; investigate whether PFC portability is achievable |
| Excessive UNKNOWN | More than half of eligible natural pairings remain UNKNOWN after one budgeted evidence-repair round; report whether extraction, source quality or mechanism mapping caused it |
| Failure to transfer | Three adequately evidenced materially similar independent systems defeat the claimed mapping for one family; repeated failure across two families after a registered revision seriously weakens the transfer thesis |
| Weak provenance | Any positive result relies on AI-only/irrecoverable decisive evidence; systematic inability to retain legally permitted audit material blocks trustworthy corpus expansion |
| Unreliable architecture extraction | Over 10% of decisive fact assessments are overturned in independent evidence checks, or recurrent scope/identity errors cause certainty; stop automated promotion |
| Unreproducible experiments | Two clean authorized repeats fail to place/observe the registered mechanism; no positive reproduction claim until explained. Persistent failure across supported families weakens executable-proof value |
| Unusable proof cost | Median audit/replay exceeds the pre-registered operating budget; provisional local target is 30 minutes operator time per existing supported case, excluding first-time dependency setup; record compute/storage separately |
| Ordinary chaos testing suffices | In a registered comparison, causal lineage and transfer add no discriminating applicability or proof information beyond a generic fault scenario across independently sourced cases; reconsider differentiation |
| Public cases contradict semantics | Recurrent resolved natural counterexamples persist after one registered correction round; freeze affected rules and retain failures rather than adjusting labels |

Thresholds trigger investigation, not automatic truth judgments. Persistent failures after a budgeted revision, especially across families, justify a **PROPOSED BASELINE CHANGE** with evidence under the canonical process or a recommendation to stop the thesis. Commercial runner adoption, design-partner usefulness, pilots and willingness to pay remain distinct later kill tests; technical replay cannot establish them.

## Review recommendation and remaining questions

The design is ready to challenge when all linked requirements are internally consistent, legacy limits are explicit, repository checks pass and the working tree is documentation-only. Exact implementation choices still requiring a later design decision are listed in the [roadmap](roadmap.md); none is hidden as an implemented capability or a human-approval dependency for individual technical verdicts.
