# FailureMesh — Canonical Company Baseline v1.0

Status: Authoritative baseline unless an explicit PROPOSED BASELINE CHANGE is reviewed and approved.

## Company

Working name: FailureMesh

Category: Portable Reliability Intelligence

Mission: No engineering team should have to learn a software failure that the industry has already learned.

Canonical definition: FailureMesh converts previously discovered software failures into portable, privacy-preserving, executable protection for systems that have not experienced those failures yet.

## Core loop

real/public/private incident
-> Portable Failure Contract (PFC)
-> Architecture Contract (AC)
-> Applicability Engine
-> APPLICABLE / NOT_APPLICABLE / UNKNOWN
-> customer-specific experiment compilation
-> safety planning
-> local/customer-controlled proof execution
-> EXPOSED / PROVEN_RESILIENT / UNKNOWN
-> reproducible counterexample or proof bundle

## PFC

A Portable Failure Contract is a machine-readable causal description of a software failure. It includes failure family, architectural prerequisites, missing/insufficient control, causal sequence, forbidden outcome, invariant, abstract experiment, proof requirements, safety restrictions, provenance, version, and verification level.

It is not merely an error fingerprint, stack trace, postmortem summary, chaos script, or LLM-generated warning.

## Architecture Contract

An Architecture Contract is a privacy-preserving description of the relevant structure and semantics of a customer's system: runtime/language, databases, transaction boundaries, queues and delivery semantics, workers, leases, concurrency, fencing, retries, backoff, jitter, timeouts, idempotency, reconciliation, external side effects, provider guarantees, caches, dependencies, and observability required for proof.

Core research problem:

Portable Failure Contract x Architecture Contract = Correct Applicability Decision

## Verdict semantics

Applicability: APPLICABLE / NOT_APPLICABLE / UNKNOWN.

Execution: EXPOSED / PROVEN_RESILIENT / UNKNOWN.

UNKNOWN is first-class. Never use IMMUNE as an absolute claim.

PROVEN_RESILIENT is always scoped to a specific PFC version, application/build, environment, experiment space, invariant, and collected evidence.

## AI policy

AI may read postmortems, extract candidate causal mechanisms, suggest prerequisites/missing controls, classify architecture facts, generate experiment/adapter candidates, generate hypotheses, and summarize evidence.

AI may not independently issue PROVEN_RESILIENT, invent proof, fabricate evidence, override safety policy, or convert uncertainty into false certainty.

Canonical rule: LLM = hypothesis generator. Verifier = authority.

## Privacy

Raw customer source code, logs, credentials, secrets, IPs, hostnames, identifiers, production rows, private postmortems, and raw traces remain local by default.

Only sufficiently generalized and explicitly approved structural information may enter shared intelligence.

## Product components

mesh-agent — local scanner, runner, redactor, safety enforcer, evidence collector.
mesh-compiler — transforms incidents/postmortems into candidate PFCs.
mesh-registry — versioned PFCs, provenance, verification history, architecture families, evidence, signatures.
mesh-match — evaluates PFC prerequisites against Architecture Contracts and explains the verdict.
mesh-runner — maps abstract failure operations into platform-specific execution.
mesh-proof — evaluates invariants and produces reproducible counterexamples or scoped proof bundles.
mesh-control — commercial control plane for registry intelligence, private PFCs, policy, RBAC, SSO, audit, execution history, analytics, governance, enterprise deployment.

## Initial wedge

Target backend/platform/SRE teams operating PostgreSQL, Kafka/SQS/RabbitMQ, background workers, distributed services, transactions, retries, leases, idempotency, external APIs, and concurrency.

## First ten PFC families

1. timeout after external commit
2. duplicate queue delivery
3. stale worker after lease expiry
4. retry amplification / retry storm
5. PostgreSQL connection-pool exhaustion
6. poison-message retry loop
7. partial transaction / business-state divergence
8. dependency latency cascade
9. cache stampede
10. worker crash after durable write before acknowledgement

## V0

Go, PostgreSQL, Docker, one CLI, three reference services, ten PFCs, five fault primitives, one deterministic proof format. No SaaS dashboard required initially.

First vertical slice: timeout-after-external-commit against Go + PostgreSQL + external payment simulator.

Canonical flow:
detect applicability
-> inject response loss after external commit
-> trigger retry
-> observe duplicate logical effect
-> EXPOSED
-> deterministic counterexample
-> add idempotency/reconciliation protection
-> rerun
-> scoped PROVEN_RESILIENT

## Open-source / commercial split

Likely open source: PFC specification, Architecture Contract specification, local runner, basic adapters, reference failure corpus, CLI.

Likely commercial: verified global PFC registry, cross-company reliability intelligence, advanced applicability engine, private PFC registries, execution history, evidence analytics, org policy, RBAC, SSO, governance, private/VPC deployment, enterprise support.

## Moat

Not UI, Go code, LLM prompts, or generic chaos injection.

Moat:
verified PFC corpus
+ architecture-to-failure mappings
+ cross-system execution evidence
+ adapter ecosystem
+ privacy-preserving contribution network
+ evidence about which remediations work across architecture families

## Verification levels

L0 CANDIDATE — extracted, never reproduced.
L1 REPRODUCED — reproduced against a reference implementation.
L2 VERIFIED — repeated deterministic reproduction with stable evidence.
L3 PORTABLE — verified across materially different implementations.
L4 FIELD VERIFIED — executed across multiple independent customer architecture families.

## Safety levels

0 static applicability only
1 synthetic reference architecture
2 developer machine / CI
3 ephemeral staging
4 customer staging
5 explicitly authorized controlled production

Production eventually requires explicit authorization, blast-radius constraints, abort conditions, health preconditions, time/rate budgets, audit records, and recovery/rollback controls where possible.

## Competitive boundary

FailureMesh is not merely incident management, chaos engineering, production failure reproduction, AI SRE, AI code review, or postmortem AI.

Preserve all four:
shared failure knowledge + architecture applicability + automatic adaptation + executable evidence.

Customer value chain:
KNOWLEDGE -> APPLICABILITY -> EXECUTION -> EVIDENCE

Standout question:
“What failures have systems like yours already learned the hard way, and which of those have you not proven yourself resilient against?”

## Validation sequence

PFC + AC specifications
-> 10 public incidents encoded
-> 3 reference architectures
-> 3 executable PFCs
-> 5 expert interviews
-> reproduce 1 real historical external incident
-> reproduce 3 materially different historical incidents
-> design partner
-> pilot
-> first payment

Kill or pivot if applicability remains noisy, false positives destroy trust, engineers refuse to install the runner, failure mechanisms do not transfer, experiments cannot be reproduced, or users see only demo value.

## Claims discipline

Never claim:
“FailureMesh proves your system is safe.”
“You cannot have this outage.”
“Our AI found every reliability issue.”

Make only scoped evidence claims tied to PFC, environment, build, invariant, experiment space, and collected evidence.

## Founder operating rules

Evidence over excitement.
Precision over alert volume.
UNKNOWN over hallucinated certainty.
Privacy by architecture.
No production fault injection by default.
Every important product claim needs a reproducible proof path.
Keep V0 narrow.
Change the canonical thesis only when evidence justifies it.

## Baseline change rule

Any change to this baseline must be explicitly labeled PROPOSED BASELINE CHANGE and state:
1. what changes,
2. why,
3. evidence,
4. difference from current baseline,
5. why the benefit justifies changing the thesis.

Until approved, this document remains authoritative.

## Long-term vision

FailureMesh should become collective memory for software architecture: portable, privacy-preserving, executable knowledge of how software systems fail.
