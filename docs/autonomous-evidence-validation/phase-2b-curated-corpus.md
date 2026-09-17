# Phase 2B — Curated corpus tooling design

Status: **design/specification only; ready for independent architecture review, not implementation approval**. The owner authorized this bounded design and existing repository checks. No production code, crawler, scraping, real corpus collection, extraction, matching, new service or execution capability is delivered. The [canonical baseline](../company/canonical-baseline-v1.0.md), [invariants](../invariants.md), [system architecture](../architecture/system.md), [privacy](../privacy.md), [safety](../safety.md) and [ADRs](../adr) remain authoritative.

FailureMesh transfers causal failure knowledge across systems. This layer would preserve the provenance and limitations of manually selected incidents so that later work can challenge whether mechanisms transfer. Corpus membership is neither causal evidence nor a qualifying Gate A incident. The design refines the [foundation](phase-1-foundation.md) and [evaluation protocol](evaluation-protocol.md) without changing them.

## Design package and boundary

- [Record model](corpus-record-model.md): intake, origin/episode/project dependence, family hypotheses and rights.
- [Selection protocol](corpus-selection-protocol.md): policy-first curation, rejected candidates, split constraints, exposure and registration.
- This document: future local package/CLI, deterministic manifest, acceptance and stop criteria.

Future flow: manually supplied candidate → intake revision → source references → origin clustering → tentative family → rights/audit limitations → decision → split → frozen manifest → separately authorized later extraction. No arrow promotes an interpretation into an observation. The register remains small and local under mesh-registry's existing responsibility; no service, database, search product, LLM training dataset or leaderboard is proposed.

## Proposed local tooling contract

A future `internal/corpus` package and `failuremesh corpus` subcommands are design names only. They accept explicit local input files and a trusted local store root. No network, source retrieval, shell execution, embedded scripts, model calls, telemetry or automatic matcher invocation. Source instructions are inert data. CLI and library share the same validation path. Pin schema, encoding, policy, tool/build and complete inputs; never use ambient time, filesystem iteration order or randomness in deterministic results. Policy-check time is supplied explicitly and reported; it is a caller assertion, not attestation.

| Conceptual operation | Inputs → outputs | Failure boundary |
| --- | --- | --- |
| `candidate add` | Safe intake envelope, policy, expected ledger head, operation ID → revision 1 digest and attempt/event receipt | Duplicate logical ID, malformed classification or disallowed retention refused; safe failed-attempt reason retained |
| `candidate revise` | Candidate/head digest, proposed complete revision and reasons → next revision and event | Stale parent, silent replacement, lineage fork or exposure reset refused |
| `candidate exclude` | Candidate/head and coded rationale → EXCLUDED revision | Cannot delete candidate history or remove frozen membership |
| `cluster` | Explicit relation proposals, basis, pinned candidate set → cluster revision, allocation components and count report | Cycles in directed ancestry, dangling endpoints, unexplained identity split or leakage block freeze; no confidence-based independence |
| `assign-family` | Candidate revision, vocabulary, hypothesis/rationale/alternatives → revised candidate | No natural semantic assessment, PFC creation, AI-confidence promotion or EVIDENCE_SUPPORTED emission |
| `assign-split` | Complete explicit table, policy, cluster/exposure heads → assignment digest and deficit report | Component split, development promotion or stale exposure refused |
| `exposure append` | Actor/context/access event, affected IDs and basis → immutable event | Cannot replace prior YES or clear inherited exposure |
| `freeze` | Closed inventory, policy, assignments, ledger head, supplied check/freeze times → manifest, local receipt and summary | Invalid closure/permissions/transitions fail without a receipt; count deficits remain explicit |
| `verify` | Manifest digest, store, declared historical/current mode, ledger head and policy time → deterministic report | Distinguish integrity failure, current-use block and disclosed content-audit limitation; never issue applicability/execution verdicts |
| `inspect` | Pinned object/manifest and view scope → metadata/history/limitations | No implicit source-body expansion or text export; unknown schema refused |
| `retire` | Manifest/component, current head and reason → retirement/invalidation event | Never erase historical freeze or restore freshness |

Exit contract: 0 = requested structural operation completed (may contain explicit SHORTFALL/CONDITIONAL_AUDIT); 2 = malformed input/integrity/unsupported version; 3 = policy or lifecycle refusal; 4 = I/O or concurrency failure. Stable reason codes and affected IDs accompany every failure. No partial success reported as freeze. Output must state independently `integrity`, `selection_target_status`, `current_use_status`, `audit_capability`, `freshness` and `chronology_tier`; no single PASS implying truth. Failed writes must never log prohibited input bytes. If the failure receipt itself cannot be stored, report that durability failure and refuse subsequent freeze until reconciliation.

## Relationship to Phase 2A

Reuse [internal/evidence](../../internal/evidence) digest representation and its versioned source/rights semantics; reference existing exact evidence-record/artifact identities. Do not modify its schema, synthetic predicate, assessor or APIs, and do not introduce fake atomic claims for source-only intake. Natural-family hypotheses never call the synthetic retry assessment to claim support. Existing source records are optional; absent captures are explicit citations with limitations.

Corpus objects contain no source bodies/excerpts. Existing permitted source material remains in its evidence store, with one declared store identity and current read/retention checks. An intake reference is not permission to copy it. Register verification checks corpus bytes and reference bindings; full evidence validation additionally invokes the existing evidence validator with required permitted transient parent bytes. Missing parents produce an explicit audit refusal rather than a fake successful source validation. Phase 2A is the independently approved evidence substrate landed via PR #16; Phase 2B does not expand its authority or eliminate its documented replay, parent-byte or store limitations.

## Storage, encoding and immutable boundaries

Proposed local root contains `objects/sha256/<hex>.json` for intake, relation/cluster, policy, assignment, exposure and manifest objects; `events/<sequence>.json` for hash-linked receipts; `bindings/` for immutable logical-ID/revision-to-digest bindings; and a rebuildable non-authoritative index. Evidence storage is referenced separately, not duplicated. Permission-restricted local root and regular files only; reject symlinks, path traversal and supplied paths escaping the root. IDs never become unchecked paths. Administrators remain inside the trust boundary.

Proposed `corpus-go-json-v1` follows Phase 2A's exact Go JSON profile: UTF-8, no trailing newline, explicit fields, finite integers, strict schema, unknown/duplicate fields and noncanonical re-encodings refused. Implementation must pin exact struct field order and golden bytes before acceptance. Set-like arrays sort by stable ID then revision/digest, reject duplicates; ordered histories preserve sequence. No locale sorting or URL normalization in identity hashes. Limit each object to 4 MiB; oversized input fails, never truncates. Hash the exact serialized object using tagged SHA-256; store its digest outside its hashed bytes. No self-hash field. These are local versioned bytes, not a claim of universal JSON canonicalization.

A cooperating-process lock protects expected-head checks, bindings and receipt publication. Objects and bindings use exclusive atomic publication: identical repeats are idempotent; conflicting bytes fail. The freeze receipt is published only after closure verification and durable object publication. Interrupted writes yield explicit unregistered/orphan status; restart cannot turn them into a completed freeze without revalidation. The CLI cannot stop a filesystem administrator rewriting the entire store. Integrity is checked relative to supplied retained identities and heads, not claimed write-once custody.

## Frozen manifest

The immutable manifest uses `corpus-manifest-v1`, the encoding above, and a unique corpus ID plus increasing corpus version. Parent manifest is null only for the first version. It binds:

| Binding | Minimum exact content |
| --- | --- |
| Scope and policy | Selection-round ID, selection-policy ID/version/digest, family-vocabulary digest, intended use and all policy deviations |
| Closed inventory | Every submitted candidate's ID/revision/digest through cutoff, including excluded/deferred/duplicate/weak/inaccessible/rights-blocked entries; safe failed-attempt IDs and event references |
| Sources | Citation entries and exact captured-source record/artifact references, content identities, unavailable references explicitly typed; no invented digest for uncaptured bytes |
| Dependence | Cluster/episode/project and publisher-group snapshot digest, full relation graph, independence limitations, count representative per origin, allocation units |
| Decisions and families | Inclusion/exclusion reasons and primary/alternative family hypotheses with states, bound to intake revisions; materialized summaries must equal referenced records |
| Splits | Assignment table digest, member IDs/components, historical split and current freshness, inherited synthetic parents, deficits by family/split |
| Rights and audit | Per-reference rights snapshot/check time, admissibility NOT_ASSESSED or referenced separate assessment, retention limits, audit tiers and missing parent dependencies |
| Exposure | Ledger pre-freeze sequence/head, actor/context matrix and inherited effective state; prior LLM training UNKNOWN disclosure |
| Determinism | Selection method/version; seed or explicit NOT_USED; stable ordered inputs, schema/encoding/hash identities |
| Registration | Supplied freeze timestamp, explicit self-reported semantics, tool version/build digest, source commit and dirty-tree digest where applicable; no receipt self-reference |
| Revision and counts | Parent-manifest digest, correction reasons, selected/natural/synthetic/origin/project/publisher/family/split totals and exact shortfalls |

Verify recomputes counts, components, rights/exposure summaries and all digests from the bound closure, refusing caller-supplied false summaries. Every intake included at inventory cutoff must appear exactly once as a current revision, with all revision ancestry resolvable. Later discoveries do not retroactively enter that cutoff. The receipt references this manifest and the prior ledger head. Inspection must display parent/child versions and later invalidation/exposure notices separately from original bytes. Hashes establish byte identity, never chronology, source truth, rights, completeness outside the store or independence.

## Human and LLM authority

Later authorized LLM assistance may propose leads, metadata summaries, duplicates, family hypotheses and inclusion concerns. Every proposal records model/context provenance and exposure; none certifies causal truth, source admissibility or independence. Confidence cannot decide selection authority or replace evidence. Humans record policy decisions and evidence-bound reasoning, not truth by approval. Neither may secretly optimize selection for matcher outcomes, promote weak sources, delete failures or rewrite frozen splits. Deterministic corpus checks enforce declared structural constraints only; verifier logic and recorded admissible evidence retain their existing authority for later scoped verdicts.

## Implementation acceptance — future tests, not current results

Use authored synthetic metadata/bytes only during initial tooling acceptance. Each row requires a retained fixture, expected reason/status, actual result and immutable-history comparison. All negative cases must fail the intended boundary without changing committed objects; unknown/limited states must remain inspectable. No real 25-incident collection is required to test this layer.

| Adversarial fixture | Required measurable outcome |
| --- | --- |
| One incident submitted through three URLs | Three candidate/locator histories, exactly one natural representative and one origin |
| Mirror declared independent; copied secondary article | Known provenance collapses count; unresolved relation defers count/joins leakage component; no URL-based independence |
| Issue plus postmortem plus follow-up in one episode | Exactly one natural origin; independent episode override requires recorded contrary evidence/new revision |
| Secondary article references two unrelated incidents | Preserve two justified origins; article citation edges do not spuriously merge them |
| Duplicate/project/derivative split leakage, including late cluster merge | Freeze refused; old affected manifests receive current-use contamination notice; no silent reassignment |
| Repeated failures with unknown episode boundaries | At most one count or explicit deferral; never independent merely from timestamps |
| Holdout revised after reveal, renamed ID, new model or new round | Old bytes unchanged; inherited EXPOSED; zero fresh-holdout count for descendants |
| AI confidence 1.0 or human approval of wrong family | Remains CANDIDATE/DISPUTED; EVIDENCE_SUPPORTED request refused; no claim/matcher projection |
| Rights unknown; public body but export unknown | Safe metadata-only record or minimal rejected attempt; no body retention/export; audit limitation visible |
| Complementary excerpts/forged parent ranges/parent loss in referenced evidence | Existing Phase 2A failures preserved; corpus cannot bypass them or call missing-parent audit successful |
| Source deleted after registration | Original manifest identity retained; append loss notice; current audit downgraded/refused, not semantic support from digest |
| Family shortfall and easy-family substitution | Explicit per-family deficits; no target-complete status; new policy round cannot rewrite original selection |
| One publisher disguised through aliases dominates | Recomputed normalized concentration; cap refusal or registered pre-result DIVERSITY_SHORTFALL exception |
| Candidate chosen after matcher output or undeclared shared context revealed | Exposure event invalidates prospective eligibility; development only; prior freeze stays visible |
| Synthetic descendant marked public/natural | Origin ancestry forces synthetic count zero for natural quota and parent split inheritance |
| Tampered manifest/intake digest; duplicate keys; unknown schema | Integrity error before any successful receipt; stable rejection reason |
| Dangling source/intake/parent refs | Integrity refusal; uncaptured citation remains distinguishable from missing promised object |
| Revision silently replaces frozen history or removes excluded candidate | Binding/closure refusal; historical bytes and receipt unchanged |
| Exposure NO asserted after YES; stale ledger head supplied | YES retained; stale writes refused; historical read cannot claim current freshness |
| Concurrent writers, crash before receipt, retry | At most one committed next head; no partial frozen success; exact retry idempotent and orphan report explicit |
| Reordered sets, identical inputs/times/build | Byte-identical objects/manifests/report across clean local stores; golden encoding checks pass |
| Private marker, hostile instructions, sensitive metadata | Safe rejection/quarantine, no execution/model/network/export, no sensitive payload in diagnostics |
| False count summary, hidden PENDING/rejection, omitted cutoff candidate | Freeze refused; recomputed inventory reconciles every submitted attempt |
| Corpus selected/frozen successfully | Gate A/baseline files and matcher/verifier/execution/PFC behavior unchanged; no new authorization |

Implementation exit requires all cases above, compatible Phase 2A regressions, full repository quality gates, and a synthetic end-to-end intake → duplicate → exclude → split → freeze → inspect → expose → revise demonstration with recorded digests. Semantic evidence truth and independent chronology remain unearned. Review must inspect the complete patch and authority boundaries; passing tests alone is not corpus validation.

## Stop and thesis-pressure criteria

At each predeclared budget/cutoff, publish the full deficit/concentration/audit report. Never manufacture cases to reach 25. Stop the affected prospective round immediately for matcher-aware selection, uncontrolled holdout access, hidden attempts, unresolvable split leakage, integrity failure or prohibited retention. Preserve failures and quarantine current use; no favorable rerun replaces them.

If multiple proposed families still lack adequately evidenced independent origins after their registered effort budgets, origin clustering collapses apparent diversity, most scientifically useful cases cannot be audited under permitted retention, or family assignment requires hidden incident-specific knowledge, stop expansion and report a research feasibility problem. Report exact counts and candidate IDs, not unsupported generalizations. If fresh holdouts cannot be obtained under the exposure constraints, publish that shortfall and stop blind-transfer claims; development research may proceed only within separate authorization and explicit informed status.

These observations challenge the proposed validation approach and potentially the transfer thesis; they do not by themselves prove the company thesis false. Record competing explanations (source scarcity, rights/access, model insufficiency, selection bias), the evidence needed to distinguish them and bounded next research proposals. Any thesis change must be labeled **PROPOSED BASELINE CHANGE** with what changes, why, evidence, difference from baseline and benefit justifying change. Do not edit the baseline or silently replace difficult families.

## Self-review and remaining implementation decisions

Within the declared local trust boundary, the design addresses the requested inflation/leakage paths through conservative clustering, component-wide exposure, immutable inventory closure and no source-text export. Residual limitations are explicit: a local administrator can rewrite unanchored history; a curator can conceal external access; source attribution/independence are evidence-bound judgments; public-model pretraining is unknown; corpus checks cannot certify prose truth. None is reported as solved by a hash or a second agent.

Before implementation authorization, pin concrete Go field layouts/golden encoding fixtures, safe failure-event persistence and crash recovery details, exact size limits for a whole store/closure, and the trusted-root filesystem support. Before actual curation, supply finite search budgets, source-specific rights bases, curator roles/access restrictions and explicit selection-round policy. Independent receipt custody and source-text release history remain later work; neither blocks an honestly local metadata register. Phase 2A B1 independent re-review remains separate. No unresolved decision permits relaxing a specified boundary.

## Local design verification — 2026-09-17

Passed repository Markdown-link checker; Gate A structural/checkpoint and canonical-baseline byte integrity validator; all 25 Python tooling tests; `gofmt -l .` (empty); `go mod verify`; `go vet ./...`; fresh `go test -count=1 ./...`; and fresh `go test -race -count=1 ./...`. Both Go suites used the existing isolated synthetic PostgreSQL fixture and passed all six library packages; the CLI has no test files. Tracked diff and every untracked document passed whitespace checks. These checks validate repository compatibility, not the unimplemented corpus tooling, source truth or independent architecture approval. Self-review covered all ten requested attacks, including immutable exposure inheritance, publisher aliases, retained exclusions and unchanged Gate A. No staging, commit, push, PR or merge occurred.

## Governance checkpoint

Gate A remains **REVISE**. Mechanism fit remains **10/10**. Qualifying public incidents remain **8/10**. Families **1 and 3 remain unresolved**. PFC #4 remains **unauthorized**. Historical #003 remains **unauthorized**. Customer execution remains **unauthorized**. Production execution remains **unauthorized**. Broader Phase 1 remains unauthorized. Selected corpus candidates do not automatically change any Gate A count. Existing three-state applicability, scoped execution verdicts, safety gates and privacy boundary are unchanged.
