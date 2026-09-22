# Roadmap and unresolved architectural questions

Status: planning only. Each stage below requires its own explicit authorization and scoped specification before implementation. Completion of a previous stage, passing metrics, this document, or absence of human objections authorizes none of the next stage. Baseline-changing proposals use the canonical change process. PFC #4, Historical #003, customer execution and production execution remain unauthorized.

## Staged progression

| Stage | Reviewable deliverable and exit evidence | Authorization boundary |
| --- | --- | --- |
| Foundation (this task) | Evidence, evaluation, mutations, proof, acceptance and failure design; documentation checks | Architecture only; independent review is recommended, not claimed complete |
| Curated corpus tooling | [Phase 2B local register](phase-2b-implementation.md) implements manual metadata intake and frozen selection records with synthetic tests; independent implementation review pending | Local tooling authorized; real corpus collection still requires separate authorization; no crawler/bulk ingestion or automatic Gate A promotion |
| Architecture packet builder | [Design specification](architecture-packet-builder.md) now defines broad inventory, six-state facts, AC sidecar projection, frozen packet/exposure manifests and replay; independent architecture review pending, no packets built | Design only; separate implementation and collection authorization required; public/synthetic inputs only, no retrieval, customer scan or execution capability |
| Deterministic claim validator | Versioned predicate/evidence rules, six-state assessments, conflict/freshness/entailment tests | No generic prose truth oracle; unsupported predicates abstain |
| Blind match arena | Registered manifests/receipts, withheld expectations, case-level outputs and dispute history | Static evaluation; any evaluator changes separately scoped |
| Counterfactual engine | Reproducible typed mutations, semantic relations, permanent counterexample registry | Synthetic descendants; no new natural-case counts or execution permission |
| Proof bundle generator | Exact-byte manifests, legacy envelopes, integrity and decision replay in two clean environments | No retrospective attestation or source redistribution by default |
| Bounded execution compiler | Capability mapping, finite manifests, safety plan and observer fidelity tests for named local scope | New adapters/experiments require named authorization; no cloud/customer/production grant |
| Continuous public-evidence pipeline | Allowlisted incremental discovery/capture, bounded queues, retractions and idempotent reassessment | Separate permission for retrieval automation; no paid service or private access |
| Larger corpus | New registered sampling round, independent holdouts, case-auditable transfer results | Expand only after falsification/UNKNOWN/cost review; preserve Gate A distinction |
| Product/API layer | Inspectable scoped decisions, evidence availability and local privacy boundaries | No implication of market fit or customer rollout |
| Design partner | Explicitly recorded interest and separately scoped agreement, commercial learning | Technical corpus does not imply partnership or customer execution |
| Pilot | Named customer authorization, approved environment, privacy and safety plan | No automatic production access |
| Payment | Actual commercial transaction and evidence of value | No inferred willingness to pay from technical success |

The order expresses dependencies; it does not replace the canonical company validation sequence. External expert criticism continues in parallel under existing approvals, and outstanding review targets remain open. Future technical tooling may issue evidence-bound results without waiting for those reviews. Stronger customer/product claims still require the baseline's reconciliation and commercial evidence.

## Exact unresolved questions

These require decisions in the corresponding later authorization, not arbitrary human labels for each case:

1. **Predicate coverage and entailment:** Which initial predicates admit reliable structured/code/observation rules, and what exact evidence-bound prose mappings can be accepted for reported/design scope? Until specified, unsupported semantic extraction remains UNKNOWN. Resolve with adversarial fixtures and scoped extraction benchmarks, not a confidence threshold.
2. **Packet wire schema and legacy projection:** The [packet-builder design](architecture-packet-builder.md) chooses a versioned AEP sidecar and separately hashed evidence-bearing conceptual AC projection; it leaves legacy AC schema 1 unchanged. It specifies required domains, six-state projection, closure, exposure, encoding profile and replay boundaries. Before separately authorized implementation, pin exact field layouts/golden bytes and supported predicate rules. A later compatibility adapter must preserve scope/evidence or refuse; no migration or legacy evaluator change is authorized.
3. **Registration custody:** Which independent receipt/CI storage, retention period and signing trust root are available without paid services? Default is honestly labeled local registration. No independent chronology claim until custody is established.
4. **Rights-aware retention:** Which source classes permit retained excerpts or snapshots, with what per-source retention/export policy? Default is metadata/citations/digests and only explicitly permitted excerpts; restricted decisive evidence reduces replay coverage. No legal entitlement is inferred here.
5. **Freshness and contradiction profiles:** The packet specification requires immutable build/config scope or explicit moving-state validity, fixed assessment time, visible semantic scope-overlap/containment conflicts and append-only invalidation. Its B1/B2 corrections define shared-region failure and competing destination-contributor resolution. Initial predicate-specific windows, accepted evidence and correction rules must still be pinned before implementation; there is no universal TTL or generic code-over-docs precedence. Missing current support stays UNKNOWN.
6. **Independent corpus feasibility:** Can approximately five materially different families supply 25 sufficiently grounded incidents and 20–30 natural architecture packets within a registered curation budget? Which projects/publishers must share a contamination cluster? No named sources are acquired or prequalified by this plan.
7. **Backend evidence fidelity:** Which first local adapter can prove causal placement and complete authoritative observation beyond the already bounded fixtures? Cloud/Kubernetes adapter capabilities are unknown; no integration is selected or endorsed here.
8. **Operational cost and failure budgets:** What curation time, per-case compute/storage and audit budgets are sustainable? Freeze the provisional failure thresholds and final budgets before reveal; do not tune them after seeing results. Small-sample uncertainty remains explicit.

## Architecture Packet Builder design checkpoint

The [packet specification](architecture-packet-builder.md) is documentation only. Independent architecture review returned REVISE for B1 semantic scope overlap/containment and B2 competing AC mappings; focused corrections and normative acceptance vectors now await independent re-review. It retains adversarial cases A01–A20, stop conditions, claim boundaries and local validation history. No production code, corpus, packet artifact, retrieval or execution capability was added. Gate A remains REVISE, mechanism fit 10/10, qualifying incidents 8/10, and families 1 and 3 unresolved. PFC #4, Historical #003, customer execution and production execution remain unauthorized. Existing Phase 2A/2B review status and company-validation work are unaffected.

## Foundation verification record

Checks on 2026-09-16 passed: repository Markdown links, Gate A/baseline integrity, 25 Python tooling tests, `gofmt -l .` (no output), `go mod verify`, `go vet ./...`, fresh `go test -count=1 ./...`, fresh `go test -race -count=1 ./...`, tracked diff whitespace and explicit untracked-file whitespace inspection. Go used the existing temporary build cache; both test suites used the existing isolated loopback synthetic PostgreSQL fixture and passed all five library packages. The CLI has no test files. No new experiment or production/customer target was introduced.

Complete changed/untracked-file review confirms seven new Markdown files in this package and one informational HANDOFF line, with no engine/verifier, contract, baseline, Gate A or authorization-file changes, and no secret/private evidence introduced. No files were staged or committed. These checks establish documentation/tooling integrity and existing prototype regression behavior, not source truth, independent review, the future corpus, or implementation acceptance. The [foundation acceptance table](phase-1-foundation.md) is the checklist for later implemented evidence.
