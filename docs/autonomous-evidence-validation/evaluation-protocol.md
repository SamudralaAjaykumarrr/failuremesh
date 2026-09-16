# Evaluation protocol and curated corpus

Status: future protocol requirements. No cases, labels, corpus acquisitions or independent attestations are created by this design. [Match Evaluation #001](../validation/match-evaluation-001/grading-method.md) remains a disclosed development example; its chronology is not upgraded by these requirements.

## Two distinct questions

1. Does matching implement the declared causal rules correctly on the supplied evidence states?
2. Do the evidence packet and PFC accurately represent the independently documented mechanism and architecture, within their declared scopes?

Rule truth tables answer the first. Source inspection, predicate-specific evidence checks, independently collected architecture packets, and bounded observation can challenge the second. Agreement between a matcher and a test oracle built from the same code does not establish either source truth or transferability. Keep separate scorecards for evidence extraction/admissibility, causal model adequacy, matcher semantics, execution capability and proof verification.

## Freeze, run, reveal, append

1. **Select before matching.** Register sampling criteria, incident/project duplicate clusters, candidate inclusion/exclusion reasons and development/validation/holdout allocation. Inventory architecture sources independently of target PFCs. Record all collector exposure; “blind” is not inferred from separate files or agents.
2. **Freeze the evidence.** Pin PFC version and bytes, architecture packet and AC projection versions, source revision set, accepted/rejected evidence set, scope, evaluation time/freshness policy, and unresolved facts. Freeze pairing rules and deterministic selection seed before observing results. No new source can silently repair a case after reveal.
3. **Freeze the procedure.** Pin evaluator source commit plus dirty-tree content digest if necessary, binary/build/toolchain, schemas, source-quality policy, extraction/projection rules, expected-label procedure, grading taxonomy, exclusion rules, mutation plan, replay comparator and resource budget. Store an expected judgment and evidence rationale when defensible; otherwise freeze `grading disputed` or an allowed semantic relation, not an invented label. A reference rule interpreter must be independent of production matcher code and use recorded evidence, not its output.
4. **Register.** Produce a content-addressed registration manifest covering all input/procedure/expected-label artifacts. A run requires its registration digest and receipt. Keep registration separate from outputs so registration cannot contain the eventual result. A future CI workflow records receipt before starting evaluation and emits immutable-addressed artifacts. Registration can be fully automated within the authorized scope; no external human signature is mandatory.
5. **Evaluate.** Run only frozen inputs and versions in an isolated process without access to sealed expected labels. Record all attempts, including failures/timeouts, resource usage and raw output. Do not choose the best rerun. Deterministic mutations use the registered seed and preconditions; invalid mutations are reported, not silently discarded.
6. **Reveal and grade.** Join observed outputs with committed expectations only after output digest recording. Produce per-case evidence trace, agreement/disagreement, limitations and challenge links. Expected labels are evidence-bound judgments, not truth created by a seal. An unresolved defensible dispute stays disputed.
7. **Append corrections.** Record a new case/assessment revision with parent digest, rationale, newly acquired evidence and exposure status. Preserve original labels, results, attempts and metrics. A corrected holdout is now exposed/development material; it cannot count as a fresh blind test.

Registration requires the manifest and referenced bytes to be available for later inspection. If expected labels must be hidden from the runner, keep them in a separate access-controlled artifact store; commit their digest with a randomly generated nonce to resist guessing three possible labels. Reveal bytes and nonce afterward. This is a future access-control requirement, not a claim that the current repository isolates agents or hides files.

## Provenance strength

| Term | Required evidence | Limitation |
| --- | --- | --- |
| Merely self-reported | Author timestamp, local log or unsigned local manifest | Author can rewrite both chronology and content |
| Pre-registered, locally enforced | Frozen digest plus append-only run linkage and process-enforced write-before-run order | Auditably ordered within that process; same-host administrator can rewrite history |
| Reproducible | All necessary permitted inputs, rules, toolchain and replay procedure recoverable; comparator agrees | Does not establish chronology, source truth or independence |
| Independently attestable registration | External/independently administered receipt binding registration digest, identity and time before a run receipt referring to it | Trust in service, identity and custody; does not prove no earlier undisclosed experiment |
| Blind evaluation | Evidence of withheld target/result/label information under the stated roles and access policy | Cannot erase prior model training or a curator's prior knowledge |

Git commit IDs identify content/history, not trustworthy wall-clock time. Hash chains expose edits only relative to a previously retained/anchored head; a locally rewritable chain is not immutable chronology. CI artifact retention may expire or be administratively mutable; record provider, receipt, retention and custody limitations. A signature authenticates its signer and bytes, not their truth or signing time. No timestamp authority, independent custody or signing infrastructure is implemented here. Later registration tooling must declare its achieved tier rather than claim cryptographic chronology from a digest.

Use append records containing sequence, previous-record digest, event type, artifact digest, reported time and optional receipt/signature reference. Duplicate runs, invalidations and withdrawals are first-class records. Public registration contains only approved metadata/digests, never restricted source text.

## Initial corpus: small and difficult

Target **approximately 25 curated public incidents across roughly five materially different causal families**, and **20–30 independently sourced architecture packets**. This is a later research corpus target, not current evidence or authorization to collect it. It is separate from Gate A and does not require selecting the five easiest existing PFC matches.

Candidate strata are durable-effect replay/identity, ownership and fencing, shared-capacity/deadline propagation, retry feedback amplification, and cache/coordination contention. These are selection hypotheses, not five new executable PFCs or claims that enough qualifying incidents exist. Prefer about five independent incident clusters per selected family. If a family lacks adequate evidence, record the deficit; do not fill it with synthetic stories or silently replace the family after observing output.

Select for first-party observations, causal ordering clarity, explicit context and limitations, recoverable provenance, permitted audit material, materially different implementations, and feasible bounded reproduction. Include version-dependent, ambiguous, cross-boundary, protected, and apparent near-match cases. Record weak-source candidates in the exclusion/uncertainty register rather than erasing selection failures. Seek multiple stacks/vendors and independent publishers; cap any one publisher at roughly one third of the incident corpus unless the pre-registration documents a shortfall. Report the actual concentration, not a diversity claim from names alone.

Acquire architecture packets from official reference designs, OSS architecture/design docs and repositories, vendor technical docs, and reproducible local references. Inventory the same domains without seeing the target match. Target at least 20 natural-source packets; any local synthetic references are separately counted and cannot fill that target. Freeze at least 80% of the natural packets before specific target-PFC assignment; disclose rule-aware collection even when assignment is hidden. Packet source must be independent of the failure incident being transferred for an independence claim; same project/release or copied topology is not independent just because the URL differs.

Approximately 25 incidents could be split by origin/project cluster into 15 development, 5 validation and 5 holdout cases, about 3/1/1 per family where supported. For 25 packets use roughly 15/5/5 by architecture project lineage; with 20–30 preserve similar proportions. Keep all versions, incident mirrors, derivatives and synthetic mutations in the same split. These small holdouts support falsification and audit, not precise population performance estimates. Pre-register departures from target counts before reveal.

Pair by a registered stratified matrix including plausible transfers, mechanism-breaking near matches, and missing-evidence cases. Preserve all attempted cells and exclusions. Do not count all incident×packet combinations as independent samples or compare unsupported PFC implementations as if they were supported. Existing three executable PFCs and two incident-specific models remain their own bounded strata. Future generalized family mappings need separate implementation authorization.

## Contamination and leakage controls

| Risk | Control and disclosure |
| --- | --- |
| Architecture selected because it obviously matches | Separate broad packet inventory/freeze from pairing; record selection history and rejected candidates |
| Labels written after output | Seal expectation/procedure before run; runner cannot read labels; append corrections |
| Known training/development case reused | Stable origin clusters and split ledger; no promotion from development to holdout |
| Source summary leaks label | Label-free architecture packet view; raw source remains inspectable; disclose unavoidable incident/outcome cues |
| Agent knows desired result | Record model/context/tool access and rule exposure; no blind claim for informed fixtures; use withheld natural packets later |
| Duplicate evidence | Cluster by originating incident, causal episode, code lineage and derivative citations, not URL count |
| Model pretraining contamination | Cannot exclude knowledge of public material; disclose as unknown and prioritize temporal/project separation and independent deterministic checks |
| Mutation leakage | Descendants inherit parent split; metamorphic challenges are regression evidence, not new independent incidents |
| Repeated holdout tuning | One registered reveal per holdout round; retire exposed cases to regression; replace through a newly registered sampling round |

An external reviewer is optional for progress, but cannot be simulated by a second agent sharing the same task context. Mechanical separation and public evidence make challenges possible; neither establishes independence by assertion.

## Case-level grading and errors

Each case records original expectation, its evidence basis, actual verdict, every prerequisite trace, source/architecture limits, grading version, execution availability and dispute history. Grade only within a validated scope; a tool invocation failure is not a matcher UNKNOWN output and cannot count as a correct abstention.

| Classification | Condition |
| --- | --- |
| Correct positive | Expected APPLICABLE and observed APPLICABLE with admissible rationale |
| Correct negative | Expected NOT_APPLICABLE and observed NOT_APPLICABLE with decisive false prerequisite/exclusion |
| Correct UNKNOWN | Expected UNKNOWN and observed UNKNOWN for the justified unresolved inputs |
| False positive | Expected NOT_APPLICABLE, observed APPLICABLE |
| False negative | Expected APPLICABLE, observed NOT_APPLICABLE |
| Unjustified certainty | Expected UNKNOWN, observed either determinate verdict |
| Unjustified UNKNOWN | Evidence supports a determinate expectation, but evaluator abstains without a supported limitation |
| Grading disputed | Competing defensible interpretations or inadequate independent expectation; preserve both |
| Evidence invalid | Digest, provenance, entailment, scope or admissibility failure; original result retained and current use blocked |
| Source conflict | Material source disagreement unresolved |
| Architecture ambiguity | Required system/identity/boundary semantics unresolved |
| Execution unavailable | No authorized capable backend/observer; orthogonal to applicability correctness |

Source conflict and ambiguity are cause tags, not substitute verdicts. Disputed/invalid cases are excluded from resolved-label accuracy denominators with visible counts and reasons, never silently dropped. Distinguish justified applicability UNKNOWN due to an unestablished required causal mapping from needless abstention on a fact the evaluator claims to support. Mere absence of an installed/authorized backend is execution unavailable and does not change an otherwise supported applicability judgment.

Metrics are secondary indexes into cases. Publish numerator/denominator case IDs, split, family, cluster dependence, abstentions and exclusions with every metric. For resolved expectations, report the complete 3×3 applicability confusion table (including expected UNKNOWN), then:

- FP rate = observed APPLICABLE among expected NOT_APPLICABLE / all expected NOT_APPLICABLE; FN rate analogously uses observed NOT_APPLICABLE among expected APPLICABLE. Report UNKNOWN counts separately so abstention cannot hide errors.
- Precision = correct positives / all observed APPLICABLE with resolved expectations; certainty against expected UNKNOWN is an error in this denominator. Recall = correct positives / all expected APPLICABLE, including abstentions as missed positives.
- UNKNOWN rate = observed UNKNOWN / successfully evaluated cases; invocation failure rate uses all attempted cases. Coverage = determinate verdicts / successfully evaluated cases, paired with certainty error rate.
- Disagreement rate = differing expected/actual labels / cases with frozen resolved expectations. Publish disputed and evidence-invalid fractions over all registered cases separately.
- Evidence sufficiency = supported decisive predicate slots / required predicate slots, plus unknown/conflict reason counts. Advisory confidence can be assessed against later resolved evidence in bins, but no probability calibration claim follows from deterministic verdicts or tiny samples.

Zero denominators are `not estimable`, not zero error. No one headline accuracy number. Inspect natural-source errors separately from constructed counterfactuals. Match correctness is not exposure prediction, architecture truth, execution success or commercial validation.
