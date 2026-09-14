# HBase paused region server resumes during log recovery

- **Candidate family:** stale worker after lease expiry / ownership loss
- **Verification level:** L0 CANDIDATE. No FailureMesh reproduction, applicability, or execution verdict.
- **Organization/project/source:** Apache HBase, HBASE-2312 and related HBASE-2593.
- **Public references:** [HBASE-2312: possible data loss while rolling HLog](https://issues.apache.org/jira/browse/HBASE-2312); [HBASE-2593: race between log splitting and writing](https://issues.apache.org/jira/browse/HBASE-2593).
- **Source dates:** HBASE-2312 created 2010-03-12; HBASE-2593 created 2010-05-21.
- **Source type / quality:** Primary Apache project bug reports and technical race analysis. HBASE-2312 was resolved Fixed; HBASE-2593 was resolved Duplicate. The described failure window is not an independently documented customer production incident.

## Evidence boundary

- **SOURCE-ESTABLISHED FACT — defining mechanism:** HBASE-2312 describes a region server pausing during HLog roll, the master treating it as dead and starting log splitting, then the old server waking, creating a new HLog, and appending an edit. The issue describes that edit as lost in its possible-data-loss scenario. HBASE-2593 separately identifies the interval after the server is marked dead but before `fs.append` lease recovery: the paused server can resume while still holding the HDFS lease, so the master's append fails. It says recovery must break the lease so old-server writes stop passing.
- **SOURCE-ESTABLISHED FACT — status of the account:** These are project bug descriptions and engineering analyses of a possible race. They establish the causal failure window at the mechanism level; they do not supply a complete observed execution trace for a specific production incident.
- **UNKNOWN:** Whether the master successfully took the HDFS lease before a particular old-server append; whether any particular stale edit was durably accepted and externally visible; realized data loss, duplicate business effects, and customer impact are UNKNOWN. ZooKeeper session/ownership loss and HDFS write-lease transfer are distinct events and must not be collapsed.
- **INFERENCE:** The portable family is an old owner resuming write activity after ownership loss or recovery begins. The exact HDFS lease state at the attempted write changes whether that write succeeds.
- **HYPOTHESIS — controls-under-test:** Fencing old ownership at each effect sink, fail-stop behavior after ownership loss, and safe HLog lease recovery.

## Candidate PFC mapping

- **Architectural prerequisites supported by evidence:** Leased/ephemeral region-server ownership, master failover and log recovery, a paused owner capable of resuming, and an HLog write sink. Fencing effectiveness is a control-under-test.
- **Causal sequence:** **SOURCE-ESTABLISHED PROJECT ANALYSIS:** old region server owns work → long pause → master considers it dead and starts recovery → old server resumes and attempts HLog write; HBASE-2312 describes an appended edit after recovery begins, while HBASE-2593 identifies the need to break the old writer's HDFS lease.
- **Forbidden outcome (candidate):** An old ownership generation has a write accepted after a newer generation has acquired authority at the relevant sink. This is a proposed invariant violation, not an observed business outcome in these issues.
- **Candidate invariant:** No old-generation mutation is accepted at an effect sink after authoritative takeover for that sink.
- **Candidate abstract experiment:** In an isolated reference store, pause an owner, trigger ownership expiry and recovery, resume its pending write, and record whether the sink accepts or rejects it.
- **Required observations / proof needs:** Ownership/session transitions, recovery start and completion, HDFS/write-lease state, writer generation, attempted and accepted writes, effect-sink authority, and observer completeness.
- **Safety restrictions:** Level 1; synthetic records and bounded pauses, no live HBase cluster.
- **Provenance:** Apache HBase project bug reports; no FailureMesh execution.
- **Evidence limitations:** The exact race and stale write activity are documented, but lease-transfer ordering, accepted stale effects, and downstream impact for a specific incident remain UNKNOWN.
- **Portability question:** Which ownership epoch can be carried to and enforced at each effect sink across architectures?

The proposed experiment and invariant are **HYPOTHESIS**, not source observations. No customer AC was evaluated.

## Targeted exact-family audit

- **Exact-family match:** YES, at the defining stale-owner-after-ownership-loss mechanism level.
- **SOURCE-ESTABLISHED FACT — decisive criterion supported:** HBASE-2312 describes pause → server deemed dead → log splitting → old server resumes and appends; HBASE-2593 explains the lease-recovery race and why old writes must be stopped.
- **UNKNOWN — downstream criterion:** A specific accepted stale mutation after completed lease takeover, business-level corruption or duplication, and customer impact are UNKNOWN.
- **Previous source:** The [deer-flow issue #4414](https://github.com/bytedance/deer-flow/issues/4414) is superseded as primary evidence; it remains supplementary for old-task liveness after lease expiry, not accepted effects.
- **Verification level remains:** L0 CANDIDATE; no FailureMesh reproduction, customer applicability, or execution verdict.

## Canonical public-incident audit

- **Primary source class:** TECHNICAL_FAILURE_ANALYSIS.
- **Counts toward ten-public-incident milestone:** NO. The project race analyses lack a complete observed operational incident trace.
- **Mechanism-fit status:** YES, L0 candidate; this classification does not change the source-backed causal mapping.

## Final incident-search closeout

No separate qualifying exact-family public incident was established. [Kafka KAFKA-2729](https://issues.apache.org/jira/browse/KAFKA-2729) contains real production broker/session-expiry complaints and a maintainer's explanation that a zombie controller *can* continue updating ZooKeeper or sending broker requests after losing controller authority. It is **TECHNICAL_FAILURE_ANALYSIS** for this family because the decisive zombie action is offered as a possible cause, not tied by an observed trace to the reported production impact. [Apache Iceberg issue #16282](https://github.com/apache/iceberg/issues/16282) reports a real old coordinator racing after leadership loss and 20 minutes of commit timeouts; this is **PROJECT_BUG_WITH_PRODUCTION_OBSERVATION** and relevant supplementary stale-owner evidence. That same documented incident is counted for family 10's write-before-offset gap. The issue does not establish a lease/session-expiry chronology for the old coordinator, and it does not supply a separate fully described second incident for this family. It is not counted twice toward the canonical ten-public-incident milestone. HBase remains the primary exact mechanism mapping at L0; the family-3 incident gap remains open.
