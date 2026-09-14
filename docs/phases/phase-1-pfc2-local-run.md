# Bounded PFC #2 local synthetic run

This separately approved second executable slice tests `pfc.duplicate-queue-delivery@1.0.0` in a deterministic, serial synthetic queue. It is not Kafka, SQS, RabbitMQ, or any real provider. The queue adapter publishes one `message-X`, delivers it as `delivery-1`, waits for its PostgreSQL consumer effect and acknowledgement, then injects the V0 duplicate-operation/message fault and delivers the **same** logical message as `delivery-2`. No sleep or random delivery behavior places the fault.

The Architecture Contract records evidenced queue delivery, duplicate possibility, stable message identity, durable side effect, repeat reachability, ledger observability, and durable consumer idempotency as a control under test. Missing, stale, conflicting, or insufficient facts yield `UNKNOWN`; an evidenced absent necessary prerequisite yields `NOT_APPLICABLE`. The `consumer_effects` table is the authoritative durable-effect ledger. The vulnerable mode commits a distinct row for each delivery. The remediated mode uses `consumer_idempotency` with primary key `(run_id, message_id)` and a reference to the first effect. Delivery 2 still reaches the consumer and records a dedupe decision referring to effect A. Both modes use the same serial two-delivery experiment bounds, message ID, fault, and observer plan.

Use the same loopback-only synthetic PostgreSQL fixture described in [the first local run](phase-1-local-run.md). From the repository root:

```sh
export FAILUREMESH_DATABASE_URL='postgres://failuremesh:synthetic-only@127.0.0.1:55432/failuremesh?sslmode=disable'
go run ./cmd/failuremesh pfc2 init
go run ./cmd/failuremesh pfc2 match vulnerable
go run ./cmd/failuremesh pfc2 plan vulnerable
go run ./cmd/failuremesh pfc2 reset
go run ./cmd/failuremesh pfc2 run vulnerable > /tmp/failuremesh-pfc2-exposed.json
go run ./cmd/failuremesh pfc2 reset
go run ./cmd/failuremesh pfc2 match remediated
go run ./cmd/failuremesh pfc2 plan remediated
go run ./cmd/failuremesh pfc2 run remediated > /tmp/failuremesh-pfc2-resilient.json
```

The verifier requires the exact ten-event trace, both delivery identities and attempts, the duplicate after first acknowledgement, and exact PostgreSQL ledger rows. The CLI verifier re-reads the effect ledger and durable idempotency state before issuing a verdict. It returns `EXPOSED` only for two distinct committed effects from one message. It returns scoped `PROVEN_RESILIENT` only when both deliveries occurred, the second was deduped to the first committed effect, and the complete ledger has exactly that one effect. Missing or inconsistent evidence returns `UNKNOWN`. Reset is required before each run. The evidence digest identifies local bundle bytes; it is not independent attestation.

This result covers only the pinned PFC, build, synthetic environment, at-most-one invariant, and one serial two-delivery experiment. It does not prove concurrent consumer behavior, real queue semantics, customer architecture applicability, portability, production execution, or broader resilience. Gate A remains REVISE with 8/10 qualifying public observed incidents; families 1 and 3 remain unresolved. Synthetic execution does not change source classification.
