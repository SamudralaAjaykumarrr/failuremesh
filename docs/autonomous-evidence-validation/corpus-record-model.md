# Corpus record model — proposed v1

Status: **Phase 2B design only**. Read with the [tooling specification](phase-2b-curated-corpus.md) and [selection protocol](corpus-selection-protocol.md). No records, captures, claims or corpus are created here. This model extends the [evidence model](evidence-model.md) with curation metadata; it does not change `evidence-v1` or its authority.

## Common envelope and unknowns

Every object has `schema_version` (proposed `corpus-v1` for register objects; `corpus-manifest-v1` for manifests), `object_kind`, logical ID, positive integer `revision`, `parent_revision_digest` (null only at revision 1), producing actor/tool/version, policy digest, reported UTC time and sensitivity. Revisions are consecutive and refer to the same logical ID. Corrections require reason codes and rationale. No reuse of an ID/revision for different bytes, branching revision heads, cycles or dangling parents. Ordered event history, rather than timestamps, determines local transition order.

An uncertain field uses `{state: UNKNOWN, value: null, reason: ...}`; a known field uses `{state: KNOWN, value: ..., reason: ...}` with basis references. Empty string, omitted field and false are not synonyms for unknown. Dates support a date, timestamp or interval with precision, timezone/unknown timezone and source basis; never invent midnight or incident time from publication time. Internal event timestamps use canonical UTC RFC3339Nano. IDs are opaque locally assigned strings, unrelated to titles, URLs or matcher results.

References in persisted object payloads point only to already published objects. A transition event points to its output revision; that revision may cite earlier basis events but never its own receipt. Split/exposure fields in intake are snapshots at that revision, not mutable current state. Later assignment and exposure events are authoritative for the manifest, which binds both the intake snapshot and the later events and checks their ordered transitions. This avoids cyclic hashes and contradictory copied summaries.

## Intake revision

All following groups are required; unknown values and empty lists need reasons where they denote missing information.

| Field/group | Required content and constraint |
| --- | --- |
| `candidate_id`, envelope | Stable candidate identity; revision, parent digest, actor, tool and reason history as above |
| `discovery` | Manual submission method, discoverer identity/class, reported discovery time, lead/citation, search scope or supplied lead, selection-round ID, exposure-ledger references; discovery is not retrieval |
| `source_references` | Canonical locator(s), publisher attribution basis, source type and role (origin/report/secondary/architecture); locator status and citation-only limitation; no inferred content digest |
| `captured_sources` | When already captured: evidence record digest, source ID/revision, artifact ID, content digest, exact representation and location references. Record digest and raw-content digest are distinct |
| `publisher`, `project` | Normalized publisher group, aliases/ownership basis, project and `project_lineage_id`, independent attribution unknowns |
| `incident_time`, `publication_date` | Separate typed values, precision and basis; unknown permitted |
| `system_context` | Technology, stack/version, system/domain, vendor and architecture-context references or explicit unknowns; names do not prove semantics |
| `family_assignments`, `mechanism_hypothesis` | Versioned family vocabulary, primary accounting family or unassigned, alternative families, proposed ordering/mechanism, scope, status and rationale; no prerequisite projection |
| `classification` | `public`, `synthetic`, `private`, `secret`, `unknown`; plus `natural_incident`, `synthetic_descendant`, `synthetic_original`, `unknown` origin class. Public visibility never makes a synthetic story natural |
| `rights_retention` | Per-reference rights snapshot/basis and current policy-check reference, metadata/digest/text permissions, retention deadline, audit/export limitations; never a single “public = allowed” flag |
| `evidence_availability` | Per source: `available`, `unavailable`, `deleted`, `retracted`, `unknown`; check time/basis, captured/uncaptured, parent-byte dependency; separate register-integrity and content-audit status |
| `selection_assessments` | Causal clarity, architecture context, bounded reproduction feasibility and directness: `adequate`, `limited`, `unavailable`, `unknown`, each with rationale, references and assessor; no combined prestige score |
| `clustering` | Origin cluster, project lineage, causal episode if known, duplicate relations, independence assessment and unresolved links; never infer independence from missing edges |
| `exposure_refs` | All applicable actor/context/event IDs and ledger head; model pretraining knowledge `UNKNOWN` explicitly |
| `decision` | `PENDING`, `INCLUDED`, `DEFERRED`, `EXCLUDED`; nonempty reason codes/rationale, policy revision, decision actor, supporting and rejected alternatives, round ID |
| `split_assignment` | `UNASSIGNED`, `DEVELOPMENT`, `VALIDATION`, `HOLDOUT`, or `NOT_SELECTED`, assignment event and allocation-unit ID; freshness tracked separately |
| `freeze_status` | Revision-local `DRAFT` or `SELECTED`; frozen membership is derived from registration events pointing to this revision, avoiding a self-referential manifest hash |
| `revision_lineage` | Parent revision plus correction/withdrawal references; synthetic parent candidate/revision and transformation identity where applicable |

An uncaptured citation is valid intake metadata, not a dangling artifact reference. A claimed captured artifact must resolve exactly or validation fails. Private/unknown-sensitive submissions retain only a safe rejection envelope locally, without copying their payload or sending it to a model. If even citation/metadata retention is not permitted, retain only an opaque attempt ID, reason and permitted administrative event. Thus selection failure remains visible without keeping forbidden content.

## Origins, episodes and dependence

`origin_cluster_id` groups every representation of one underlying incident and its derivatives. `causal_episode_id` identifies the bounded initiating event and continuation/recovery interval when defensible. `project_lineage_id` groups shared project/code/architecture lineage, including renamed projects and relevant forks; separate outages can have separate origins but remain project-dependent. Publisher group is an additional concentration dimension, not a substitute for either lineage.

A relation object binds endpoint candidate revisions or source references, relation type, `ESTABLISHED`, `POSSIBLE` or `REJECTED` status, basis, actor, policy and rationale. Types: `SAME_INCIDENT`, `MIRROR_OF`, `COPIED_SUMMARY_OF`, `FOLLOWUP_OF`, `ISSUE_REPORT_OF`, `QUOTES_ORIGIN`, `REVISION_OF`, `SYNTHETIC_DESCENDANT_OF`, `SAME_CAUSAL_EPISODE`, `SHARED_PROJECT_LINEAGE`. Directional derivations must be acyclic; equivalence groups use connected components. Rejected proposals remain in history.

Established incident/episode/derivative relations collapse natural counts to at most one representative per origin. Three URLs are three locators, never three independent incidents. A downstream article can reference several origins without merging those distinct incidents: model its citations separately, not as an equivalence edge connecting everything it mentions. One report covering several incidents needs separately justified episode boundaries and provenance locations; otherwise count at most one. Repeated failures during one episode stay together unless independently documented initiating events justify a revision separating them.

Possible overlap blocks an independence claim. Candidates with unresolved origin/episode identity are deferred from the selected natural count, and possible links conservatively join allocation units for leakage protection. An unresolved project identity similarly prevents fresh validation/holdout placement. Absence of discovered overlap only supports a documented, bounded independence assessment, never proof of independence.

Allocation units are connected components of established origin/episode/derivative/project relations plus unresolved possible leakage links. They are deliberately broader than incident count units. Synthetic descendants inherit all parent dependencies; multi-parent descendants connect all parents. A connection across existing splits blocks a new freeze and appends a contamination notice to affected old manifests. It cannot be “fixed” by deleting the relation. Splitting an erroneous cluster requires explicit contrary evidence and new revisions; accumulated exposure never resets.

## Family hypotheses

Vocabulary v1 has five hypotheses: durable-effect replay/identity; ownership/lease/fencing; shared-capacity/deadline propagation; retry feedback amplification; cache/coordination contention. These are research strata, not canonical-family renumbering or authority for new PFCs.

| State | Meaning and transition requirement |
| --- | --- |
| `UNASSIGNED` | No defensible association yet; unknown reason required |
| `CANDIDATE` | Attributed human/tool/LLM hypothesis, including scope, rationale and alternative explanations |
| `EVIDENCE_SUPPORTED` | Reserved for a later, separately authorized evidence-bound assessment with supported scope, evidence and mapping-rule identity; Phase 2B cannot emit this state |
| `DISPUTED` | Material competing association or challenge recorded; retain alternatives and why unresolved |

Phase 2B assigns `CANDIDATE` or `DISPUTED` from metadata/manual rationale only. It performs no extraction or semantic validation. Existing later assessments may be referenced historically, but their authority must be checked by their own policy and cannot be synthesized here. The current Phase 2A synthetic retry rule does not support natural-family classification. High confidence, human approval or a citation cannot promote a hypothesis. One primary tentative family is used for accounting before results; alternatives are visible and never double-count one incident. Disputed/unassigned cases may be registered as challenge cases but do not fill an established family-coverage claim. Even candidate-family totals must be labeled tentative.

## Rights and current audit capability

Reuse Phase 2A `Rights`: `license`, `basis`, `review`, `metadata`, `digest`, `excerpt`, `whole_body`, `export`, `excerpt_scope`, `retain_until`; permission values remain `allow`, `deny`, `unknown`. Retrieval permission and its basis are separate discovery metadata, not an extension to the implemented type. Public access grants none of these permissions automatically. This is conservative engineering policy, not a copyright ruling.

Intake stores metadata, citations, permitted digests and references by default. No source text in free-form rationale as a retention workaround. Permitted bounded excerpts stay in separately governed evidence storage; whole-document retention requires explicit permission. No arbitrary excerpt length is presumed lawful. Unknown text/export rights block those uses while a metadata-permitted candidate may remain included with `REGISTER_ONLY` or `CONDITIONAL_AUDIT` limitations. Such membership cannot count as full replay capability.

Audit tiers are `REGISTER_ONLY` (inspect selection/provenance metadata), `CONDITIONAL_AUDIT` (requires permitted reacquisition or transient parents), and `LOCAL_AUDIT_AVAILABLE` (currently permitted pinned evidence available). None means semantic truth or full evaluation replay. Admissibility for later causal predicates remains `NOT_ASSESSED`, not “admissible because selected.” Inspect reports record historical tier and current tier separately.

Phase 2A requires true parent bytes to bind excerpt ranges/length and cumulative coverage; missing parents refuse that validation. Corpus verification must surface this limitation, never trust saved “verified” flags. Register-only verification may succeed while content audit is unavailable. Metadata/digest-only evidence needs no body to check its record integrity, but its hash says nothing about source meaning. Do not fabricate a target claim to fit the Phase 2A `Record` API (which requires a target); uncaptured citations remain corpus-native references. Already valid evidence records can be referenced unchanged.

Corpus tooling has no source-text export operation. Later export requires separately scoped approval for exact bytes/destination, current rights/classification checks and complete composition review. Independent stores/releases cannot be used to bypass cumulative excerpt limits; Phase 2A currently has only one-store persistence and complete-record export coverage, not a global ledger. Source expiry/revocation/deletion appends an availability/withdrawal event and blocks affected current use. Required deletion takes precedence over historical byte preservation: retain permitted tombstone/digest only, report the lost audit path, never claim complete replay. No deletion scheduler or rights adjudicator is implied.
