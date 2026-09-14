# Distributed service and dependency graph

## Structural shape
Ingress → service A → service B/shared dependency → PostgreSQL or other transactional store; optional cache in front of origin. Client and service retries can compound, and critical-path latency may cascade.

## Facts an AC must carry
Runtime/build and service graph, each RPC deadline/timeout, retries/backoff/jitter, concurrency/admission, dependency capacity, cache miss/refill/repair behavior, DB/pool boundaries, fanout, fallback/circuit breakers, idempotency for mutating calls, reconciliation, per-hop traces and capacity/error observers.

## Likely applicability facts to investigate
Retry amplification, latency propagation, pool saturation and cache-origin stampede. These are questions, not `APPLICABLE` decisions; controls-under-test remain separate from prerequisites.

## Explicit UNKNOWN-capable facts
Provider capacity, dependency guarantee, exact retry layer, cache-key scope, critical-path deadlines, trace completeness, and whether a PostgreSQL boundary exists can be UNKNOWN. Queue/workers/leases/fencing and external effects are recorded as evidenced values or UNKNOWN.

## Safety boundary
Future experiments require isolated dependency, bounded load, abort on health limits; no production fault injection by default.

## AC evidence discipline
Every fact has operation/path scope, source and observation method, collection time, quality, sensitivity, and an explicit UNKNOWN value when unsupported. Conflicts remain visible. The AC is versioned and bound to build and environment. This shape asserts no actual deployment or applicability verdict. Raw source, traces, credentials, logs, identifiers, hostnames, IPs, and production rows remain local; only reviewed generalized structural facts may be exported with explicit approval.
