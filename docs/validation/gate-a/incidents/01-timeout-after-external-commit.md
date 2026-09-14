# WordPress media upload response loss creates a duplicate attachment

- **Candidate family:** timeout after external commit / response loss
- **Verification level:** L0 CANDIDATE. No FailureMesh reproduction, applicability, or execution verdict.
- **Organization/project/source:** WordPress Gutenberg issue #80741, community-reported project bug.
- **Public reference:** [Lost upload response makes the automatic retry create a duplicate attachment](https://github.com/WordPress/gutenberg/issues/80741).
- **Source date:** 2026-07-27.
- **Source type / quality:** Official project issue with a concrete fault-injection reproduction procedure, expected checks, and reported actual result. It is a reporter account in a WordPress Playground environment, not a first-party production postmortem or independently audited payment incident.

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — defining mechanism:** The reporter describes a media `POST /wp/v2/media` completing on the server and returning an attachment ID. Their browser-side wrapper captures that successful result, then hides the response by throwing a fetch error. The editor retries the same logical upload; the report says the retry creates a second attachment and original file. It gives checks for two distinct committed IDs/files while the editor inserts only the retry's attachment.
- **SOURCE-ESTABLISHED FACT — observed scope:** The issue reports a reproducible response-loss injection in a WordPress test environment and states the actual result is duplicate attachments. Its evidence is the reporter's procedure and account; no raw attachment ledger or independent rerun is published in the issue text.
- **UNKNOWN:** A naturally occurring production timeout with this precise order, payment-provider commits, customer charges, and independent execution-log completeness are UNKNOWN. The injected failure is a fetch error after a successful response was intercepted, not a measured network timeout.
- **INFERENCE:** WordPress media creation is a state-changing external API effect from the editor's perspective. The same causal window can apply to other external APIs, but portability is not established by this source.
- **HYPOTHESIS — controls-under-test:** Stable logical-upload identity, idempotent create/replay, or post-timeout reconciliation.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** Caller invokes a state-changing API, can lose success knowledge after server commit, and retries the logical operation. Effective idempotency is a control-under-test, not a prerequisite.
- **Causal sequence:** **SOURCE-ESTABLISHED REPORT:** server creates first attachment → success response is suppressed from browser → editor retries → server creates second attachment for one upload.
- **Forbidden outcome (candidate):** More than one authoritative durable effect for one logical operation.
- **Candidate invariant:** At most one committed effect per stable logical operation ID across post-commit response loss and retry.
- **Candidate abstract experiment:** In a synthetic external service, confirm the first effect committed, suppress its success response before caller receipt, let the caller retry, and compare authoritative effects.
- **Required observations / proof needs:** Logical operation ID, each request/attempt, authoritative commit IDs and order, response suppression, caller retry, effect ledger completeness.
- **Safety restrictions:** Level 1 only; synthetic effects, bounded calls, no live payments or customer uploads.
- **Provenance:** WordPress project issue #80741; reported reproduction is external to FailureMesh.
- **Evidence limitations:** This is an injected response-loss bug report about media attachments, not an observed payment outage. The source describes how to verify two effects but does not publish a complete independent trace.
- **Portability question:** How is one logical operation identified and reconciled across differing provider APIs and retry paths?

The proposed FailureMesh experiment and invariant are **HYPOTHESIS**. No customer AC was evaluated.

## Final focused exact-family audit

- **Exact-family match:** YES, at the defining external-commit/response-loss/retry/duplicate-effect mechanism level.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** The project issue's fault injection explicitly waits for first server success, hides that result from the browser, and reports an automatic retry that creates a second attachment.
- **UNKNOWN — downstream criterion:** Natural production frequency, payment behavior, and independent observer completeness remain UNKNOWN.
- **Previous source:** The [stripe/ai issue #402](https://github.com/stripe/ai/issues/402) is superseded as primary family-1 evidence; it remains supplementary for fresh idempotency keys on charge-tool retry but does not establish first commit before response loss. Stripe's [idempotency analysis](https://stripe.com/blog/idempotency) and [low-level error guidance](https://docs.stripe.com/error-low-level) explain mechanism semantics, not this incident.
- **Verification level remains:** L0 CANDIDATE; the external reporter's reproduction is not FailureMesh L1.

## Canonical public-incident audit

- **Primary source class:** SYNTHETIC_OR_INJECTED_REPRODUCTION.
- **Counts toward ten-public-incident milestone:** NO. The Playground injection is not an observed natural operational incident.
- **Mechanism-fit status:** YES, L0 candidate; this classification does not change the source-backed causal mapping.

## Final incident-search closeout

No qualifying exact-family public incident was established. The [OpenTofu S3 lockfile issue #4405](https://github.com/opentofu/opentofu/issues/4405) reports a real first conditional PUT that created a lock while the caller received a 412 and remained blocked for two days. It is **PROJECT_BUG_WITH_PRODUCTION_OBSERVATION** as an operational report, but it does **not** qualify for family 1: the author explicitly cannot determine whether an SDK retry occurred, and no duplicate committed lock/effect is shown. [Mollie API idempotency guidance](https://docs.mollie.com/reference/api-idempotency) is **MECHANISM_DOCUMENTATION**; its double-charge case is illustrative. The existing WordPress injection remains the stronger complete mechanism mapping, classified **SYNTHETIC_OR_INJECTED_REPRODUCTION**, and does not count as an incident. Searches across payments, cloud/object APIs, SDKs, and project issues found no primary source establishing all four required observed steps.
