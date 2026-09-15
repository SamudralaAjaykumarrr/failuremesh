# Reproduction guide

## Prerequisites and boundary

Use a local checkout, the Go version in [go.mod](../../../go.mod) (currently 1.26.5), and an isolated PostgreSQL fixture. Docker with Compose can start the repository's PostgreSQL 16 fixture. Run all commands from the repository root; the CLI reads relative contract and source-record paths. Record `git rev-parse HEAD` and any local changes in your scorecard.

**Local synthetic/reference executions only. No production systems and no customer execution.** Results depend on bounded manifests, approved level-1 plans, and authoritative PostgreSQL evidence. Exact historical upstream execution is not claimed. Do not substitute customer architecture, data, or endpoints.

Follow [local fixture setup](../../phases/phase-1-local-run.md):

```sh
docker compose up -d --wait
export FAILUREMESH_DATABASE_URL='postgres://failuremesh:synthetic-only@127.0.0.1:55432/failuremesh?sslmode=disable'
```

The URL contains only the repository's public synthetic fixture values, not real credentials. The CLI requires the explicit fixture user/password, loopback literal, explicit port, database, and `sslmode=disable`; it rejects `PG*` environment settings and connection overrides. Use a clean shell without those settings. Never configure a real system here.

`init` creates the relevant tables; `reset` clears that fixture's experiment state. Use this disposable database only. Inspect/save output before resetting; do not run tests or other experiments concurrently against the same tables. `match` and `plan` show applicability and bounded safety scope; `run` also enforces the plan. A successful process exit alone is not proof: inspect the JSON decision and evidence.

## Historical #001

```sh
go run ./cmd/failuremesh historical iceberg-16282 match
go run ./cmd/failuremesh historical iceberg-16282 plan
go run ./cmd/failuremesh historical iceberg-16282 init
go run ./cmd/failuremesh historical iceberg-16282 reset
go run ./cmd/failuremesh historical iceberg-16282 run
go run ./cmd/failuremesh historical iceberg-16282 verify
```

Expected: **APPLICABLE**, comparison **MATCH**, recommendation **PASS**, scoped to duplicate file registration after recovery replay. `verify` re-reads PostgreSQL. Compare the [recorded result and limits](../historical-reproductions/001-iceberg-16282/comparison.md).

## Historical #002

```sh
go run ./cmd/failuremesh historical github-may4-cascade match
go run ./cmd/failuremesh historical github-may4-cascade plan
go run ./cmd/failuremesh historical github-may4-cascade init
go run ./cmd/failuremesh historical github-may4-cascade reset
go run ./cmd/failuremesh historical github-may4-cascade run
go run ./cmd/failuremesh historical github-may4-cascade verify
```

Expected: **APPLICABLE at SERVICE-LEVEL CASCADE**, comparison **MATCH**, recommendation **PASS**, for shared capacity saturation, cross-service deadline crossings, and recovery. Compare the [recorded result and limits](../historical-reproductions/002-github-may4-cascade/comparison.md). PASS in either historical run does not mean Gate A passed.

## PFC #1: timeout after external commit

```sh
go run ./cmd/failuremesh init
go run ./cmd/failuremesh match vulnerable
go run ./cmd/failuremesh plan vulnerable
go run ./cmd/failuremesh reset
go run ./cmd/failuremesh run vulnerable
go run ./cmd/failuremesh reset
go run ./cmd/failuremesh match remediated
go run ./cmd/failuremesh plan remediated
go run ./cmd/failuremesh run remediated
```

## PFC #2: duplicate queue delivery

```sh
go run ./cmd/failuremesh pfc2 init
go run ./cmd/failuremesh pfc2 match vulnerable
go run ./cmd/failuremesh pfc2 plan vulnerable
go run ./cmd/failuremesh pfc2 reset
go run ./cmd/failuremesh pfc2 run vulnerable
go run ./cmd/failuremesh pfc2 reset
go run ./cmd/failuremesh pfc2 match remediated
go run ./cmd/failuremesh pfc2 plan remediated
go run ./cmd/failuremesh pfc2 run remediated
```

## PFC #3: stale worker after lease expiry

```sh
go run ./cmd/failuremesh pfc3 init
go run ./cmd/failuremesh pfc3 match vulnerable
go run ./cmd/failuremesh pfc3 plan vulnerable
go run ./cmd/failuremesh pfc3 reset
go run ./cmd/failuremesh pfc3 run vulnerable
go run ./cmd/failuremesh pfc3 reset
go run ./cmd/failuremesh pfc3 match remediated
go run ./cmd/failuremesh pfc3 plan remediated
go run ./cmd/failuremesh pfc3 run remediated
```

For each PFC, both modes should be APPLICABLE; vulnerable → **EXPOSED**, remediated → scoped **PROVEN_RESILIENT**. Expected evidence is duplicate provider effects versus durable dedupe (#1), duplicate consumer effects versus message dedupe (#2), and an accepted stale effect versus transactional fencing (#3). See the [reviewer brief](reviewer-brief.md) for each proof's exact scope and limitations.

Missing/conflicting evidence must not earn a positive result. Record UNKNOWN, errors, and unexpected outcomes with SHA, command, local prerequisites, and public/synthetic evidence in the [scorecard](review-scorecard.md); do not rewrite expected results to hide failures. Preserve useful output locally before any reset. When finished, `docker compose down` stops the fixture without the volume-deletion option.
