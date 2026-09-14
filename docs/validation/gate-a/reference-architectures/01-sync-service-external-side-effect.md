# Synchronous service with external side effect

## Structural shape
Client → Go-style API service → PostgreSQL business transaction; service separately calls an external payment-like provider. A database transaction does not span provider commit. This is a conceptual shape, not a built payment simulator.

## Facts an AC must carry
Operation ID, build/runtime, request path, DB transaction start/commit, provider request/commit/response boundary, timeout and retry policy (including backoff/jitter), provider idempotency guarantee, application key persistence, reconciliation path, concurrent calls, and effect-ledger observer completeness.

## Likely applicability facts to investigate
Provider commit ordering, response-loss window, stable identity across retries, effect visibility and authority. These are questions, not `APPLICABLE` decisions; controls-under-test remain separate from prerequisites.

## Explicit UNKNOWN-capable facts
Provider guarantees, actual retry policy, idempotency effectiveness, ledger completeness, and response delivery can be UNKNOWN. Queue, worker, lease and fencing domains are UNKNOWN or absent only with evidence.

## Safety boundary
No external call or mutation without future approved safety plan; synthetic level 1 for Phase 1.

## AC evidence discipline
Every fact has operation/path scope, source and observation method, collection time, quality, sensitivity, and an explicit UNKNOWN value when unsupported. Conflicts remain visible. The AC is versioned and bound to build and environment. This shape asserts no actual deployment or applicability verdict. Raw source, traces, credentials, logs, identifiers, hostnames, IPs, and production rows remain local; only reviewed generalized structural facts may be exported with explicit approval.
