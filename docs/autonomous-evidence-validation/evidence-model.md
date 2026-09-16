# Evidence model

Status: design requirements, not an implemented schema. All records carry `schema_version`, stable logical ID, immutable revision/content identity, sensitivity, and producing tool/policy identity. Unknown values are explicit, with a reason; an omitted field never means false. IDs identify objects; hashes identify exact representations. [Proof encoding](proof-bundle.md) governs digests.

## Source artifact

One logical `source_id` can have several immutable retrieval artifacts. A URL is a locator, not an identity or endorsement.

| Fields | Required meaning |
| --- | --- |
| `source_id`, `artifact_id`, `revision`, `supersedes` | Logical source, digest-bound capture, revision lineage; aliases do not create independent sources |
| `canonical_url`, `retrieved_url`, `publisher`, `publisher_identity_basis` | Claimed canonical location, actual location, publisher and basis for attribution; redirects recorded |
| `publication_timestamp`, `publication_timestamp_basis`, `retrieved_at` | Publication time may be unknown; retrieval time is collector-reported unless separately attested |
| `source_type`, `quality_class`, `quality_policy_version` | Incident report, issue, code, architecture document, etc.; admissibility is claim-specific |
| `content_digest`, `digest_algorithm`, `representation`, `byte_length` | Hash of exactly defined captured bytes, e.g. decoded HTTP entity; never present a normalized-text hash as an original-body hash |
| `retrieval_provenance` | Collector/build, request method, sanitized redirect chain, status, content type, upstream revision/ETag if present, acquisition errors; no credentials/cookies |
| `rights` | License identifier or UNKNOWN, rights basis/evidence, allowed local retention/export, excerpt scope, retention deadline, review status; public access is not redistribution permission |
| `locations` | Evidence IDs with artifact digest plus page/section/line/byte selector and selector representation; bounded excerpt digest where permitted |
| `mutability`, `availability`, `checked_at` | Mutable page or revision-pinned artifact; available/unavailable/deleted/retracted/unknown and observation time |
| `mirror_of`, `archive_of`, `revision_relation` | Claimed replica/origin relationship and evidence; disagreement creates a new artifact, never overwrites bytes |

Capture only permitted material needed for reproducibility: metadata, citation, digest, structured claims and, where allowed, bounded excerpts. Whole-document retention is opt-in under an explicit rights policy, not the default. A digest may be computed during permitted retrieval without keeping the body; declare `retained=false` and the consequent audit limitation. No fixed excerpt length is assumed legally sufficient. Unknown rights block source-text redistribution; metadata itself still receives sensitivity review. This is an engineering retention/export policy, not a determination of copyright entitlement.

`source_exists` means a retrieval observation established availability at a recorded time. `source_reports(P)` means a specific retained or recoverable location says P. `P` means a scoped fact supported under an admissibility rule. These are separate predicates and must never be substituted for each other. A first-party report can support a *reported incident mechanism* without certifying all details as independently observed truth.

For changed pages, create a new artifact and claim revisions; retain the old audit path where permitted. Link rot does not erase a valid pinned historical observation; it prevents fresh audit if supporting content cannot be recovered. A matching permitted archive can restore access, but archive time is not publication time and an archive does not add independent corroboration. Mirrors/secondary summaries share the origin cluster. Retractions and corrections append invalidation events and trigger dependency reevaluation. An unavailable, metadata-only source cannot newly substantiate a semantic claim from its hash alone.

## Atomic claim

| Fields | Required meaning |
| --- | --- |
| `claim_id`, `revision`, `predicate`, `value_type`, `value` | One assertion with a registered predicate and typed candidate value; boolean example `same_operation_retry=true` |
| `subject`, `scope` | System/component/edge/operation, build/configuration, environment, time interval, identity domain and quantifier |
| `epistemic_basis` | `reported`, `documented_design`, `code_observed`, `experiment_observed`, or `derived`; these are not interchangeable |
| `source_refs`, `evidence_refs` | Exact source revisions and location/observation IDs; an evidence ref resolves to bytes or an explicit access limitation |
| `extraction` | Method, extractor/version, input digests, transformation, prompt/model identity when used, public-only extraction context, proposer identity class |
| `advisory_confidence` | Optional heuristic score with method; never an admissibility threshold by itself |
| `support_status`, `contradiction_status`, `state` | Support present/absent/invalid; conflict none/open/resolved with resolution rule; derived effective state below |
| `freshness` | Observation time, valid interval, pinned revision, expiry/revalidation rule and fixed evaluation time |
| `sensitivity`, `boundary` | Public/synthetic only here; private, human-supplied private, secret/PII or unknown are quarantined; local/export audience explicit |
| `derivation` | Parent claim IDs/digests, rule ID/version, causal node/edge references; acyclic dependencies |
| `assessment` | Policy version, accepted and rejected supports, reasons, assessment time, supersession/invalidation references |

The effective state is deterministic over the full assessment and its evaluation time:

| State | Meaning | Matcher projection |
| --- | --- | --- |
| `SUPPORTED_TRUE` | Admissible current support for the positive proposition in exactly this scope | TRUE |
| `SUPPORTED_FALSE` | Admissible current support for its negation in exactly this scope | FALSE |
| `UNKNOWN` | Missing value, unresolved mapping or semantic interpretation | UNKNOWN |
| `CONFLICTING` | Unresolved materially contradictory evidence in the same scope | UNKNOWN |
| `STALE` | Otherwise sufficient support falls outside its version/time validity | UNKNOWN |
| `UNSUPPORTED` | Proposed value has no admissible support, invalid provenance, or only an AI assertion | UNKNOWN |

Retain all reasons even if several states apply. Assessment order is: reject malformed/unresolvable inputs as unsupported; identify material unresolved conflicts; check freshness of remaining support; validate scope and inference; then assign supported truth/falsehood or unknown/unsupported. A conflict remains visible even when stale or inadmissible; only material unresolved conflict blocks current support. The policy must record why a purported contradiction is out of scope, obsolete or non-evidentiary. Never resolve conflicts by majority vote, confidence score or convenient omission.

False requires positive evidence of a scoped negation, such as a complete bounded path with no retry, not absence of a retry reference in prose. An absence observation requires an observer completeness argument. Historical facts pinned to an old build do not expire simply because time passes; they do not support claims about a new build. Moving-target claims use an explicit freshness rule. No default universal TTL is invented.

Illustrative lineage, not a captured observation or new fixture: a synthetic run's complete caller trace could support `claim.retry@1`, predicate `same_operation_retry`, value `true`, subject `caller`, scope `(build B, run R, logical operation O)`, basis `experiment_observed`. Its evidence reference must resolve to the trace digest and the first-request/retry event selectors, with a pinned identity-comparison rule. An unrelated request with a different logical identity cannot satisfy it. Removing the only trace makes the candidate value UNSUPPORTED; an omitted identity field makes the interpretation UNKNOWN. A contradictory admissible trace makes it CONFLICTING. The example allocates no real claim ID and asserts no new evidence.

## Authority and source quality

Quality is a vector (attribution, directness, scope fit, version fit, integrity, completeness, independence), not a prestige ranking. Minimum admissibility: resolvable evidence, identified subject/scope/version, allowed source class for that predicate, valid extraction/derivation rule, freshness and no unresolved material contradiction.

| Category | May support | Cannot establish alone |
| --- | --- | --- |
| First-party incident report | What operator reports occurred, reported causal sequence and limitations | Exact unreported topology, independent truth certification, transferable execution proof |
| First-party issue/bug report | Reporter-observed symptoms/reproducer and explicitly attributed causal analysis | Maintainer confirmation or universal affected-version claims |
| First-party architecture docs | Documented design, versioned boundaries and guarantees | Deployed configuration or actual runtime adherence |
| Official cloud/reference architecture | Reference design facts in its documented scope | Any customer's implementation, measured failure or safety |
| Upstream code/test evidence | Reachable code path under pinned config; actual test observations if executed and captured | Production use of that path or completeness beyond tested scope |
| Reproducible local reference observation | Measured local mechanism, observer output under pinned experiment | Upstream incident truth or cross-system portability |
| Third-party technical analysis | Attributed hypotheses; corroborated facts through cited admissible primary evidence | Uncorroborated authoritative prerequisites |
| Community discussion | Discovery leads and reported statements | Authoritative mechanism or architecture predicates without qualifying underlying evidence |
| Secondary summary | Source discovery and origin links | Independent confirmation; duplicated stories count once |
| AI-generated claim | Candidate assertion, extraction or hypothesis | Any authoritative prerequisite, source truth or execution verdict |

An LLM can locate text but cannot certify that text entails a predicate. A deterministic validator checks typed evidence using versioned predicate-specific rules: for example an exact structured configuration key plus documented version semantics, a narrowly supported code-analysis rule with recorded coverage, or an observed event trace. It does not validate arbitrary prose by checking that a citation string exists. Unformalized prose interpretation remains an attributed candidate/report; prerequisite use requires a declared evidence-bound mapping and sufficient admissibility for the *reported/design* scope, or remains UNKNOWN. Distinguish an inspectable author interpretation from a mechanically checked entailment. Human approval alone does not promote either to runtime truth. Initial automation will have limited predicate coverage; measure that limit, do not hide it.

No universal “code beats docs” precedence: configured runtime observation may contradict a design without falsifying what the design document says. First split scopes/bases. Within the same scope, a documented explicit correction supersedes a prior assertion for future use; preserve both. A stale version does not defeat a pinned current observation. Two current admissible contradictory claims force CONFLICTING. A credible but not yet resolved primary-evidence challenge blocks certainty; an unsupported AI disagreement is logged as a hypothesis, not authoritative negative evidence. Policies define this distinction before evaluation and record every rejected challenge.

## PFC causal derivation

Each prerequisite has a lineage DAG:

```text
source artifact + evidence location → atomic claim
→ causal event (subject, action, state, ordering evidence)
→ causal relation (before/enables/requires, supporting claims, rule)
→ necessary prerequisite (predicate, quantifier, scope, derivation rule)
```

A temporal correlation alone cannot justify `causes` or `requires`. Every edge records whether it is reported, directly observed, derived under a named rule, or hypothetical. A prerequisite whose necessity remains hypothetical stays a candidate, and evaluations are explicitly conditional on that candidate PFC; it cannot earn a validated-transfer claim. Separate mechanism prerequisites, controls-under-test, adapter capabilities and exclusions. Observability/fault-placement prerequisites may come from the proof obligation rather than incident text: record that engineering derivation explicitly, link the incident causal relation and invariant, and never fabricate a source saying that a production fault hook existed.

Auditing “why this prerequisite?” must return the full path, rejected alternatives, unresolved assumptions and exact PFC version. Evidence invalidation marks dependent claim assessments and PFC versions challenged for current use; new definitive evaluations stop or return UNKNOWN with the blocking reason. Old evidence and verdicts remain historical records with appended status notices. A prerequisite/mechanism/invariant/proof/safety change creates a new PFC semantic version and requires rematching. Retraction does not erase a valid local experiment; it weakens the incident-transfer claim. Existing L0/L1/L2/L3/L4 definitions and evidence requirements remain unchanged.

## Independent Architecture Evidence Packet (AEP)

An AEP is an evidence-bearing envelope around the conceptual [Architecture Contract](../architecture/architecture-contract.md), not a replacement AC or runtime adapter. It is collected using a broad architecture inventory before target-PFC pairing wherever possible. Its factual packet and immutable AC projection have separate digests.

| Packet group | Required content |
| --- | --- |
| Identity | `architecture_id` (proposed `ac.<slug>`), `version`, packet schema, build/commit/configuration/environment, design-versus-deployed basis |
| Independence | Selection rationale, discovery source, collection window, collector exposure to PFCs/rules/results, freeze receipt, incident/project duplicate cluster |
| Provenance | Source artifacts, atomic claims, extraction method/tool/version, assessment policy, source and evidence-set digests |
| Boundaries | System, trust, transaction and persistence boundaries; components with IDs and typed communication edges |
| Semantics | Retry actor/path/count/backoff/identity; timeout/ack behavior; consistency, durability, idempotency and reconciliation; queue delivery/order/dedupe; lease/ownership/expiry/fencing |
| Proof capability | Observability coverage and completeness limits, causal synchronization capability, available local mapping; unavailable execution recorded independently |
| Controls | Known controls, coverage and evidence; no immunity inference |
| Uncertainty | Missing domains, explicit UNKNOWNs, conflicting facts, freshness/quality, interpretation limitations |
| Projection | AC schema, target-independent inventory, versioned mapping rules, every AC fact to claim references, rejected/unmapped fields |

Unspecified queue semantics or retry actors remain UNKNOWN. Do not infer facts from vendor names, familiar architecture patterns or absent prose. Preserve prose-versus-structured contradictions as candidate conflicts; prose cannot be silently ignored. If semantic extraction cannot resolve them, the affected claim remains UNKNOWN. A reference diagram is evidence of a reference design, not deployed behavior. Natural-source packets and locally constructed references are separate strata; synthetic mechanisms do not prove natural-architecture extraction quality.

At evaluation time choose a scoped operation and project only its facts, with the packet immutable. PFC-specific augmentation produces a new packet revision labeled `target_informed`; it cannot replace the independent holdout version. Architecture declarations, adapter executable capability and runtime evidence remain separate objects. Current PFC #2/#3 wrappers pin reference metadata and may refuse natural packets; that is a supported-scope limitation, not permission to rewrite the wrappers or claim a causal negative.

## Privacy, rights and source security

Phase 1 evidence is public/synthetic only. Private customer architecture, proprietary repositories, confidential postmortems, logs, PII and secrets are excluded even if accidentally exposed on a public URL. Quarantine uncertain classification locally; do not send it to an external model. Public source text, including instructions embedded in it, is untrusted data: never execute commands or follow instructions from a source. Future retrieval must reject private-network targets, credential-bearing URLs, unsafe redirects, oversized/decompression payloads and active content; collectors get no execution or governance credentials.

Retrieval, retention and export are separate permissions. Allowlisted public retrieval in a later authorized phase does not approve registry publication. Customer data remains local under [privacy policy](../privacy.md); any future generalized contribution needs artifact/version/destination approval. AI prompts, caches, telemetry, errors and proof references obey the same boundary. Shared bundles omit restricted text and sensitive digests that could disclose low-entropy secrets. A redacted export is a distinct derivative with its own digest and reduced replay capability; it never silently replaces the local original.
