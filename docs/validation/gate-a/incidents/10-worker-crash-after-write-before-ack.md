# RabbitMQ consumer result stored before acknowledgement loss

- **Candidate family:** worker crash after durable write before acknowledgement
- **Verification level:** L0 CANDIDATE. No FailureMesh reproduction, applicability, or execution verdict.
- **Organization/project/source:** RabbitMQ rabbitmq-server discussion #16909, posted by an external contributor and answered by a project maintainer.
- **Public reference:** [Formal proof of ACK temporal collision with TLA+ and chaos cross-validation](https://github.com/rabbitmq/rabbitmq-server/discussions/16909).
- **Source date:** 2026-07-09.
- **Source type / quality:** Public technical failure analysis and reported reproduction in the RabbitMQ project discussion. It is not presented as a customer production incident or a RabbitMQ implementation bug. The author reports model-checking and a Docker/Toxiproxy pilot; FailureMesh has not independently audited the artifacts or rerun the experiment. The maintainer confirms the general redelivery/idempotence semantics, not the pilot's every observation.

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — defining mechanism:** The contributor's scenario shows a consumer receiving and processing a message, storing a result in a database, losing the acknowledgement before the broker receives it, then receiving a redelivery and processing/storing the same logical work again. The described crash is after the first store and before ACK.
- **SOURCE-ESTABLISHED FACT — reported experiment:** The contributor reports a RabbitMQ/PostgreSQL/Toxiproxy pilot that waits for the first store, drops the AMQP connection, waits for redelivery, and counts two database rows with a non-idempotent consumer versus one with a skip guard. The physical pilot injects connection loss; the crash is in the scenario/model, not shown as a process-kill step in the published pilot summary.
- **UNKNOWN:** Independently verified durable-commit and broker-ACK traces, every consumer process state, behavior across other implementations, and customer production impact are UNKNOWN. A complete external audit of the linked code and runs is not part of this Gate A pass.
- **INFERENCE:** The reported connection loss and the described crash share the decisive broker-ACK gap after a database store. They are not identical physical faults.
- **HYPOTHESIS — controls-under-test:** Atomic idempotent consumer guard, stable task identity, and reconciliation.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** At-least-once broker delivery, consumer-controlled ACK, durable result store outside broker ACK transaction, and redelivery of unacknowledged work. Idempotence is a control-under-test.
- **Causal sequence:** **SOURCE-ESTABLISHED TECHNICAL ANALYSIS:** delivery → processing → database result stored → crash or connection loss before broker ACK → unacknowledged message redelivered → same work processed again. The discussion reports two durable rows for the connection-loss pilot.
- **Forbidden outcome (candidate):** More than one committed result for one logical message when the declared business invariant requires one.
- **Candidate invariant:** At most one durable result per stable logical task ID across ACK-loss and redelivery.
- **Candidate abstract experiment:** In an isolated queue/store fixture, synchronize on a confirmed first durable write, interrupt the consumer or AMQP connection before broker ACK, restore delivery, and compare authoritative results for the same task ID.
- **Required observations / proof needs:** Task identity, delivery generations, authoritative committed records, store-before-fault ordering, broker ACK frontier, consumer/connection interruption, redelivery, and observer completeness.
- **Safety restrictions:** Level 1; synthetic messages, bounded redelivery and database writes, no production broker.
- **Provenance:** Contributor account and reported pilot in official project discussion; maintainer response supports generic redelivery semantics only.
- **Evidence limitations:** The reported pilot tests connection loss, not a literal process crash; independent artifacts/trace audit and customer incident evidence are absent.
- **Portability question:** Which broker ACK and store-commit semantics make the same gap observable on another implementation?

The proposed FailureMesh experiment and invariant are **HYPOTHESIS**. The source's reproduction is external and does not promote this candidate to L1. No customer AC was evaluated.

## Targeted exact-family audit

- **Exact-family match:** YES, at the defining durable-write-before-ACK-loss/redelivery mechanism level.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** The discussion states the crash-before-ACK sequence and reports a connection-loss pilot with redelivery and two stored rows. The maintainer confirms redelivery is expected and consumers may need idempotence.
- **UNKNOWN — downstream criterion:** Customer production impact and independent completeness of the reported experiment remain UNKNOWN. The physical pilot's exact crash step is not established.
- **Previous source:** The [Honcho issue #982](https://github.com/plastic-labs/honcho/issues/982) is superseded as primary evidence; it remains a supplementary HTTP client replay near-match, not broker-ACK evidence.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary mechanism source class:** TECHNICAL_FAILURE_ANALYSIS. The RabbitMQ formal scenario and chaos pilot are not a customer operational incident.
- **Additional canonical incident source class:** PROJECT_BUG_WITH_PRODUCTION_OBSERVATION.
- **Counts toward ten-public-incident milestone:** YES, through the Apache Iceberg production report below, not the RabbitMQ discussion.
- **Mechanism-fit status:** YES, L0 candidate; this classification does not change the source-backed causal mapping.

## Final incident-evidence pass: Apache Iceberg Kafka Connect

- **Organization/project:** Apache Iceberg Kafka Connect, reported by a production operator in the official project issue.
- **Source:** [Duplicate file registration across snapshots during coordinator recovery, issue #16282](https://github.com/apache/iceberg/issues/16282).
- **Date:** Issue opened 2026-05-11; one documented incident occurred 2026-05-07.
- **Source type/classification:** PROJECT_BUG_WITH_PRODUCTION_OBSERVATION; reporter-supplied logs and metadata observations from two claimed production incidents.
- **SOURCE-ESTABLISHED FACT:** The reporter says the coordinator committed file X into Iceberg snapshot S1. The supplied log shows a completed commit at 13:14:42 after a worker `Connection timed out` and `Committer lost leader partition` at 13:14:22. The reporter says the control-topic offset for `DATA_WRITTEN` had not advanced before task crash/coordinator shutdown. A replacement coordinator re-consumed that event and committed the same physical file in S2 at 13:36:44. The issue supplies a metadata query showing the same file path and record count in both snapshots, with 4,873 extra rows reported for the May 7 incident.
- **Defining causal sequence:** durable table commit → worker timeout/coordinator shutdown before durable control-topic offset advancement → replacement coordinator re-consumes the event → same file is committed a second time. Kafka offset advancement is the acknowledgement frontier for this flow.
- **UNKNOWN:** The exact process termination point, complete raw broker and catalog traces, independent reproduction, whether the two claimed incidents share every step, and transfer to RabbitMQ-style per-message ACK remain UNKNOWN. The source's step-by-step account is reporter analysis corroborated in part by supplied log and metadata excerpts, not an independent forensic audit.
- **Canonical incident decision:** YES. The source explicitly reports observed production failures and gives dated commit/recovery/duplicate-effect evidence for the defining write-before-ack gap. It is a different architecture from the RabbitMQ mechanism source and does not make the RabbitMQ pilot an incident.
- **Verification level:** L0 CANDIDATE. No FailureMesh reproduction, applicability, or execution verdict.
