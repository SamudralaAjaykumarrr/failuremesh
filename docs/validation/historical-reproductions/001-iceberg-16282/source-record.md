# Historical Reproduction #001 — public source record

- Source: [Apache Iceberg issue #16282](https://github.com/apache/iceberg/issues/16282), opened 2026-05-11, titled “[Kafka Connect] Duplicate file registration across snapshots during coordinator recovery (append-only, IKC 1.10.1)”.
- Classification: public project bug report with reporter-supplied production observations and causal analysis; not an independent FailureMesh forensic audit. Existing [family-10 L0 source record](../../gate-a/incidents/10-worker-crash-after-write-before-ack.md) remains the canonical Gate A mapping.
- Reported environment: Apache Iceberg Kafka Connect 1.10.1, append-only table, no equality deletes, GCS storage. `catalog_type: UNKNOWN`; conflict: REST catalog in summary, JDBC catalog in evidence section.

## A. Reported observations

The reporter says the same physical parquet file appeared in two Iceberg snapshots, with duplicate rows, across two independent production incidents. For the dated 2026-05-07 incident, the report gives 4,873 duplicate groups, 9,746 duplicate rows, and 4,873 extra rows. The reported snapshot IDs are `8264179290764999750` and `314603727896153559`. The supplied timeline places worker timeout/coordinator loss and a first successful snapshot commit around 13:14 UTC, repeated commit timeouts afterward, new-coordinator offset discovery/reset around 13:35 UTC, and the second commit around 13:36 UTC. The metadata query reportedly found identical file path and record count in both snapshots. These are reporter-supplied observations, not raw trace verification by FailureMesh.

## B. Reporter causal analysis

The reporter describes C1 beginning commit cycle A; a worker writing X and emitting `DATA_WRITTEN`; C1 consuming the event and successfully committing X into S1 before advancing the durable control-topic offset; worker connection timeout, task crash and coordinator shutdown; C2 starting before that event, reconsuming it, and committing X into S2. The proposed root mechanism is non-idempotent file registration combined with non-atomic Iceberg commit and control-topic offset advancement. Exact shutdown and offset chronology are not independently established by the published excerpts.

## C. Reporter fix speculation

The reporter suggests deduplicating already registered file paths or filtering stale control events against committed offsets. Neither is a confirmed Apache Iceberg fix or an implementation in this milestone.

## D. FailureMesh inference

The causal prerequisites for a bounded experiment are separate durable table commit and control offset, replay of a stable event after coordinator recovery, append-only registration permitting duplicate file identity, and observability of the event, offset, snapshots, and file identity. Catalog implementation is not needed to express this mechanism, so its visible conflict does not force incident-specific applicability to UNKNOWN. This is an inference about the model, not a claim about the actual catalog.

## E. Unknown or conflicting details

REST versus JDBC catalog remains unresolved. The complete Kafka control-topic trace, exact crash point, actual coordinator concurrency, catalog transaction internals, both incidents' full similarity, and all upstream causal details remain unknown. The reported 20-minute contention period is not reproduced. Public raw file names and bucket paths are deliberately omitted.
