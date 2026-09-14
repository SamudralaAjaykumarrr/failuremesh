# Architecture Contract v1 (conceptual)

An AC is a privacy-preserving, versioned description of the system facts relevant to a PFC. It is not a source-code dump or assurance claim. This is a conceptual model, not a Go type or fixed wire schema.

Each fact has a semantic value or explicit unknown, scope (service/path/operation), source class, observation or attestation method, collection time, confidence/quality, and sensitivity classification. Conflicting sources remain visible; they cannot be silently resolved. The AC carries its schema version, application/build identity, environment identity, and freshness boundary. AC facts may be local even when a generalized structural summary is approved for sharing.

The baseline's required fact domains are runtime/language; databases and transaction boundaries; queues and delivery semantics; workers; leases; concurrency and fencing; retries, backoff, jitter, and timeouts; idempotency; reconciliation; external side effects and provider guarantees; caches; dependencies; and observability required for proof. Absence of a domain is `unknown`, not `false`. A PFC may require only a subset, but the relevant transaction and effect boundaries must be explicit.

For Phase 1 the AC must identify the Go operation and build, PostgreSQL write boundary, external payment request/effect boundary, response timeout and retry policy, logical-operation identity, idempotency and reconciliation behavior, provider guarantee, and observation access to a complete effect ledger. Values unsupported by evidence remain unknown. An AC update or changed build/environment triggers reevaluation; prior applicability and execution verdicts do not automatically transfer.
