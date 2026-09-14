# Queue with workers and delivery semantics

## Structural shape
Producer → durable queue → multiple workers → PostgreSQL transaction and optional external effect. Acknowledgement is separate from business commit. Worker ownership may have lease/visibility timeout and a fencing epoch.

## Facts an AC must carry
Runtime and worker versions, producer identity, queue durability and at-least-once/redelivery semantics, ack point, visibility/lease renewal and takeover, fencing enforcement at each sink, message and logical operation IDs, retry/dead-letter policy, concurrency, DB transaction and external-effect boundaries, idempotency and reconciliation, complete queue/durable-write observers.

## Likely applicability facts to investigate
Duplicate delivery, poison retry, stale-owner overlap, write-before-ack crash window. These are questions, not `APPLICABLE` decisions; controls-under-test remain separate from prerequisites.

## Explicit UNKNOWN-capable facts
Broker redelivery guarantee, ack durability, lease clock, old-worker cancellation, sink fencing, crash injection point, and observer completeness can be UNKNOWN. Timeout/backoff/jitter and external effect presence must be evidenced or UNKNOWN.

## Safety boundary
Use only synthetic queue/workers and bounded poison/retry/crash tests in future work.

## AC evidence discipline
Every fact has operation/path scope, source and observation method, collection time, quality, sensitivity, and an explicit UNKNOWN value when unsupported. Conflicts remain visible. The AC is versioned and bound to build and environment. This shape asserts no actual deployment or applicability verdict. Raw source, traces, credentials, logs, identifiers, hostnames, IPs, and production rows remain local; only reviewed generalized structural facts may be exported with explicit approval.
