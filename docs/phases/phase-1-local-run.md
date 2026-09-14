# Phase 1 local synthetic run

This reference executable combines a Go caller and payment simulator in one local process. They use separate PostgreSQL transactions: the business-state write cannot atomically include the simulator's external-effect commit. The simulator's append-only `provider_effects` table is the authoritative ledger. Its remediated path uses `provider_idempotency` with a durable primary key on `(run_id, logical_id)`. This is a synthetic level-1 adapter, not a customer or production integration.

From the repository root, start PostgreSQL with `docker compose up -d --wait`. The `synthetic-only` password is a development fixture, not a credential for any real system. Set the local connection string:

The CLI accepts only a PostgreSQL URL with the explicit synthetic fixture user `failuremesh` and password `synthetic-only`, a literal loopback host (`127.0.0.1` or `::1`), an explicit port, database `failuremesh`, and `sslmode=disable`. The password is fixture data, not a real credential. URL query overrides, service/pass/certificate files, fallback hosts, runtime parameters, and `PG*` environment settings are rejected before opening the parsed connection configuration. The adapter also suppresses pgx's default `.pgpass` lookup during parsing.

```sh
export FAILUREMESH_DATABASE_URL='postgres://failuremesh:synthetic-only@127.0.0.1:55432/failuremesh?sslmode=disable'
go run ./cmd/failuremesh init
go run ./cmd/failuremesh match vulnerable
go run ./cmd/failuremesh plan vulnerable
go run ./cmd/failuremesh reset
go run ./cmd/failuremesh run vulnerable > /tmp/failuremesh-exposed.json
go run ./cmd/failuremesh reset
go run ./cmd/failuremesh match remediated
go run ./cmd/failuremesh plan remediated
go run ./cmd/failuremesh run remediated > /tmp/failuremesh-resilient.json
docker compose down -v
```

Inspect the two JSON files. `verdict` is `EXPOSED` only when the evidence records the first durable commit, barrier hold, lost response, observed caller loss, retry of the same `payment-X`, and two distinct ledger effects. The remediated run preserves that fault and retry but records `provider_dedupe` and one committed effect; its `PROVEN_RESILIENT` verdict covers only PFC 1.0.0, the pinned remediated reference build, this synthetic environment, the at-most-one invariant, and one two-attempt serial experiment. Missing or conflicting evidence yields `UNKNOWN`. Reset is required before each run; a dirty starting state fails closed. The JSON digest identifies the local evidence bytes and is not independent attestation of their truth.

The verifier checks the exact nine-event sequence, each attempt and effect ID, and the authoritative ledger rows against the pinned build mode. It independently enforces the manifest and safety-plan scope even when their digest is recomputed. Contract loading rejects semantic changes to the PFC 1.0.0 mechanism, controls, causal sequence, or proof requirements.

The PFC lives in [the versioned contract](../../contracts/timeout-after-external-commit.v1.json), the schema in [the migration](../../migrations/001_reference.sql), and the reference adapter/verifier in `internal/phase1`. The fault is a synchronous test-only barrier: after the provider transaction commits, the adapter records commit and holds successful delivery, then marks the response lost. No sleep or network timing places the fault. The caller then retries the same logical operation. The AC records the Go/build, PostgreSQL and independent provider boundaries, retry semantics, logical identity, provider guarantee, control state, and ledger observability. Applicability evaluates evidenced prerequisites; control state is a control-under-test. The safety plan binds the manifest and rejects every environment except `synthetic-reference` at level 1.

For the PostgreSQL integration suite, set `FAILUREMESH_TEST_DATABASE_URL` to the same URL and run `go test ./...` and `go test -race ./...`. Tests reset the synthetic tables. This prototype uses an in-process simulator and one serial logical operation; it does not test concurrent idempotency races, process crashes during the barrier, independent observer tamper resistance, customer architecture mapping, portability, or production execution. Gate A's public-incident count remains 8/10, with families 1 and 3 unresolved.
