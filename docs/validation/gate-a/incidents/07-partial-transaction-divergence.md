# Dodo Payments committed payment with missing merchant webhook

- **Candidate family:** partial transaction / business-state divergence
- **Verification level:** L0 CANDIDATE. No FailureMesh reproduction, applicability, or execution verdict.
- **Organization/project/source:** Dodo Payments engineering, Ayush Agarwal (Co-founder & CPTO).
- **Public reference:** [Building Webhooks That Never Fail: Our Journey to 99.99%+ Delivery Reliability](https://dodopayments.com/engineering/building-webhooks-never-fail).
- **Source date:** 2026-01-21.
- **Source type / quality:** First-party engineering retrospective describing observed production webhook loss and the old application's in-process delivery architecture. It is not a dated postmortem for one enumerated payment; no raw event ledger or independently audited loss calculation is published. A [vendor case study quoting Dodo's co-founder](https://restate.dev/use-cases/dodo-payments) corroborates the account but is secondary and has promotional interest.

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — observed behavior:** Dodo reports it was losing about 0.3% of webhooks before rebuilding delivery. Its engineering article describes successful payment and committed database state while the webhook notification disappeared before merchant receipt. The article says old event creation occurred in application code and in-process webhook delivery lacked durability; deployments, process termination, memory loss, and network failures could lose in-flight work. The later design moves event creation into the payment transaction with PostgreSQL triggers and uses durable downstream stages.
- **SOURCE-ESTABLISHED FACT — source classification:** This is category **B: an engineering retrospective describing observed production behavior**. It reports an aggregate loss rate and causal architecture, and uses an illustrative payment narrative; it does not provide a timestamped trace for a specific transaction. The numerical article also has an internally inconsistent chart label, so the rate is reported as the author's approximate claim, not independently validated precision.
- **UNKNOWN:** IDs and timing for individual lost notifications, exact attribution of each lost webhook to process failure versus merchant endpoint failure, whether a durable domain-event row was missing in a particular case, and whether each merchant failed to provision access are UNKNOWN. A failed merchant webhook does not by itself prove permanent business-state divergence if later reconciliation occurred.
- **INFERENCE:** A committed payment not reflected at the merchant because its required notification was lost is a cross-resource business-state divergence. The first committed state and missing paired notification are the source-supported mechanism; long-term permanence is not established.
- **HYPOTHESIS — controls-under-test:** Atomic event capture with the business write, durable relay/retry, delivery auditing, and reconciliation.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** Authoritative payment state in PostgreSQL, a required merchant-facing event/notification in a separate delivery path, and a failure window between commit and durable delivery. Control effectiveness is under test.
- **Causal sequence:** **SOURCE-ESTABLISHED RETROSPECTIVE:** payment succeeds and database state commits → in-process webhook/event path can lose the required notification → merchant does not receive that payment event; Dodo reports real aggregate webhook loss in this architecture.
- **Forbidden outcome (candidate):** A committed authoritative business transition has no corresponding delivered or recoverable required notification within a declared bound.
- **Candidate invariant:** Every in-scope committed payment transition has a durable event and a traceable terminal delivery or reconciliation outcome within a finite declared window.
- **Candidate abstract experiment:** In a synthetic payment/event fixture, commit business state and interrupt the application between that commit and durable event handoff; verify event presence and downstream convergence.
- **Required observations / proof needs:** Payment commit ledger, event creation and persistence, delivery attempts and receipts, merchant-side observation or reconciliation, timing window, and observer completeness.
- **Safety restrictions:** Level 1; synthetic payments and merchant endpoint, no real funds or production webhooks.
- **Provenance:** Dodo first-party retrospective; secondary Restate case study is corroborative only.
- **Evidence limitations:** Aggregate reported loss and architectural explanation establish the mechanism; no per-payment forensic trace or independent audit supports stronger impact claims.
- **Portability question:** Which event and delivery boundary defines equivalent business-state divergence across other paired systems?

The proposed FailureMesh experiment and invariant are **HYPOTHESIS**. No customer AC was evaluated.

## Final focused exact-family audit

- **Exact-family match:** YES, at the defining committed-business-state/missing-paired-notification mechanism level.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** Dodo's first-party retrospective reports actual webhook loss and explicitly describes payment success plus database commit with a missing notification under its former in-process delivery design.
- **UNKNOWN — downstream criterion:** Individual merchant access loss, permanent divergence, and complete per-event chronology remain UNKNOWN.
- **Previous source:** The [GitHub October 2018 database-partition report](https://github.blog/news-insights/company-news/oct21-post-incident-analysis/) is superseded as primary family-7 evidence; it remains a replication-divergence near-match, not a paired business-write incident. [AWS transactional-outbox guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html) remains supplementary mechanism documentation, not observed incident evidence.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary source class:** OBSERVED_OPERATIONAL_RETROSPECTIVE.
- **Counts toward ten-public-incident milestone:** YES; the source reports an observed operational failure with the defining mechanism at the evidence level described above.
- **Mechanism-fit status:** YES, L0 candidate; downstream effects and missing traces remain UNKNOWN as stated above.
