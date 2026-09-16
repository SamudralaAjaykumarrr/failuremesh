# Content-addressed proof bundle

Status: proposed evidence envelope, not a replacement for current proof formats or verifier behavior. A proof is an auditable empirical argument within a declared scope. Content identity is not authenticity, truth, chronology, authorization or universal safety.

## Layout and contents

Use JSON artifacts and Markdown replay instructions to fit existing repository conventions. Logical layout for a future bundle:

```text
proof/
  manifest.json
  sources.json
  claims.json
  causal-model.json
  pfc.json
  architecture.json
  applicability.json
  mutations.json
  experiment.json
  execution-evidence.json
  verdict.json
  provenance.json
  replay.md
  artifacts/sha256/<digest>
```

| Artifact | Required content |
| --- | --- |
| `manifest.json` | Format/profile version, file roles/paths/sizes/digests/media types, external permitted artifact references, required/optional status, scope, registration digest, policy/tool identities |
| `sources.json` | Exact captures, origin/revision clusters, location IDs, retention/rights/availability and provenance limitations |
| `claims.json`, `causal-model.json` | Atomic claims and assessments; source→event→relation→prerequisite DAG with rule versions and hypotheses |
| `pfc.json` | Exact versioned PFC bytes or explicit envelope referencing unmodified legacy bytes; prerequisites, invariant and proof requirements |
| `architecture.json` | Frozen packet, independent-collection record, AC projection and full fact-to-claim mapping |
| `applicability.json` | Evaluator/build/policy version, evaluation time, PFC/AC digests, complete prerequisite/exclusion explanation tree and scoped verdict |
| `mutations.json` | Registered operators and all attempts, exact parent/child patches and digests, expectations, raw results and permanent counterexample links |
| `experiment.json` | Candidate manifest, capability mapping, safety plan and authorization reference, budgets/space or explicit `not_requested`, `denied`, `unavailable` status |
| `execution-evidence.json` | Stimuli, causal ordering, identities, complete authoritative observations, observer provenance/completeness, run status, errors and omissions |
| `verdict.json` | Applicability and execution in separate fields; execution is EXPOSED/scoped PROVEN_RESILIENT/UNKNOWN with reason and evidence references; grading and historical comparison use separate namespaces |
| `provenance.json` | Registration and run receipts, parent bundle roots, toolchain/build/source revisions, custody, signatures/attestation limits, invalidation/update references |
| `replay.md` | Offline audit and deterministic decision replay, separately authorized execution recipe, prerequisites, reset/cleanup, comparator and expected limits |

For a static-only evaluation, keep explicit experiment/evidence status records, no fabricated empty “successful run.” Execution verdict is UNKNOWN with `not_requested` or other reason. Missing optional artifacts are distinguished from missing required evidence; empty experiment space cannot prove resilience. An index of counterfactual children may reference their own bundles to avoid duplication.

## Byte identity and validation

Choose exact-byte SHA-256 addressing for the initial design, with an algorithm tag for future migration. Each artifact is hashed over its stored bytes; semantically equivalent but differently formatted JSON is a different artifact. This avoids pretending a new cross-language canonicalizer exists. Writers later use a pinned encoding profile (UTF-8, LF, deterministic key order, finite numbers, no duplicate keys); readers verify bytes first and reject duplicate keys, unsupported schema/algorithm, dangling references and ambiguous parse forms. A new serialization changes the root, even if a versioned semantic comparator finds no difference.

The root ID is `sha256:<digest of exact manifest.json bytes>`. The manifest lists every payload digest, but not its own digest. Attestations and a root identifier file, if used, are detached and reference the root, preventing self-reference. Registration and output manifests are separate: the output references the registration; the registration never references future output. Proof payloads use internal role/claim IDs, not their own root hash. New invalidation notices reference old roots without modifying them.

Validate path uniqueness, relative paths, sizes, digests, schema, referential integrity, DAG acyclicity, PFC/AC/scope consistency, policy versions and evidence completeness before recomputing decisions. Reject traversal, symlinks escaping the bundle, oversized archives and executable content. Untrusted replay instructions are documentation; inspection never automatically runs shell commands. Resource limits apply to parsing as well as execution.

An artifact not redistributable can be represented by citation, permitted excerpt, digest and access requirement. Mark replay capability accordingly. A missing body hash does not mean the source is false; it limits audit. A bundle whose decisive source cannot be retrieved or whose extraction cannot be reproduced must not claim complete end-to-end reproducibility. Hash-only evidence cannot establish what text said to a new verifier.

## Three replay modes

| Mode | Meaning and required result |
| --- | --- |
| Integrity/audit | Check bytes, lineage, scope, source access and custody; report unavailable evidence and unsupported claims without execution |
| Decision replay | Recompute claim assessments, applicability, mutation relations and verifier logic from frozen permitted evidence/policies/evaluation time; same inputs yield the same semantic result |
| Experiment reproduction | Separately authorize a fresh isolated environment, recreate pinned build/config/initial state, execute finite experiment and collect new authoritative evidence; produce a new linked bundle |

Decision replay can show consistency of recorded observations without independently proving those observations occurred. Experiment reproduction needs authoritative observation again; replaying supplied JSON is not a fresh execution proof. New run times/IDs/DB-generated identifiers may differ: register a comparator that permits only declared incidental differences while preserving logical identity correspondence, causal ordering, parameters, invariant and complete effect counts. Do not strip inconvenient differences or claim byte-identical executions unless achieved.

Replay instructions require dependency/toolchain digests, build and environment versions, configuration, observer/query versions, reset/isolation, finite parameter set, seeds/logical scheduling, expected causal placement, budgets/stop conditions, safety authorization requirements, output collection and failure handling. Network refetch is optional and explicitly separates current-source reevaluation from frozen replay. No paid service or cloud account is required by the format; unsupported backends are reported unavailable.

## Authoritative observation and existing compatibility

The [current PFC #1 types/evaluator](../../internal/phase1/contracts.go) use `phase1.Digest` over Go `json.Marshal`, and its [verifier](../../internal/phase1/run.go) consumes evidence collected by the local runner from PostgreSQL. The CLI passes that captured evidence to `phase1.Verify`; this is not a generalized attested-source verifier. [PFC #2](../../internal/pfc2/run.go), [PFC #3](../../internal/pfc3/run.go) and the [Iceberg](../../internal/historical/iceberg16282/model.go)/[GitHub](../../internal/historical/githubmay4/model.go) historical paths additionally re-read authoritative DB state through `VerifyAgainstDB`. Historical MATCH/PASS is not an execution verdict.

Preserve these original artifact bytes, digests, version semantics and limitations when later enveloping them. Do not relabel the existing digest as a raw-file digest or imply legacy proofs have full source-to-prerequisite lineage. A legacy import declares missing lineage, missing independent attestation and replay limits; it cannot satisfy new acceptance requirements by filling them with invented metadata. No migration or verifier alteration occurs in this phase.

Future collectors must bind observations to run/environment/build/observer identity, authoritative store and query, logical/physical operation identities, ordered fault placement, initial-state evidence and completeness checks. Backend exit status, LLM narrative and a collector signature alone cannot prove the invariant. A verifier that cannot establish observer authority or completeness returns UNKNOWN. Detached signatures improve custody evidence only under a declared key/trust policy; they do not certify source truth.

## Immutability and invalidation

Content addressing detects changed bytes relative to a known root; it does not prevent deletion or replacement of the root a user was shown. Preserve roots in append-only records with declared custody and optional independent anchoring as in the [protocol](evaluation-protocol.md). No physical write-once storage is claimed to exist.

Bundle revisions, source retractions, policy changes, disputes and corrected verdicts append new objects linked to old roots. Consumers check the current invalidation index before reusing a verdict; offline audit reports the index's last-known checkpoint and cannot claim current validity. Withdrawal of restricted material removes access according to retention policy while preserving a non-sensitive tombstone where permitted; it reduces replay capability rather than silently recreating forbidden text. Exported redacted derivatives identify their parent only when safe, state omissions, and never inherit stronger proof claims than retained evidence supports.
