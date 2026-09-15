# PROPOSED BASELINE CHANGE

Status: **APPROVED on 2026-09-15** by the project owner:

> Approve Historical Reproduction #002 as the next bounded validation milestone.
> It must use a materially different external failure mechanism from
> Historical Reproduction #001, remain local and deterministic, use public
> evidence only, and must not authorize PFC #4, customer execution, or
> production execution.

## What changes

Historical Reproduction #002 may proceed as the second local bounded historical transfer test, in parallel with the still-incomplete expert-review track. #001 tested recovery replay and duplicate durable registration. #002 tests shared dependency capacity saturation, latency, and timeout propagation. This record extends the approved non-blocking validation sequence by one named incident-specific milestone; it does not authorize a generalized executable contract.

## Why, evidence, and difference from prior baseline

The owner explicitly approved one materially different external mechanism test. GitHub's public May 2026 availability report describes the May 4 migration/traffic overlap, database connection saturation, primary database contention, cross-service timeouts, and recovery after migration pause. The report lacks exact internal telemetry, so this milestone uses a labeled synthetic Go/PostgreSQL service-level model. Testing another mechanism can expose transfer and verifier weaknesses sooner, but cannot establish general portability. The previous approved sequence only named Historical Reproduction #001; this change names #002 without changing the canonical thesis or validation gate.

## Benefit and unchanged boundaries

This bounded comparison can test capacity, graph, deadline, and recovery evidence with a DB-backed verifier. Gate A remains REVISE, mechanism fit 10/10, qualifying public incidents 8/10, families 1 and 3 unresolved, and generic incident search CLOSED. It neither changes the Gate A count nor authorizes PFC #4, a fourth generalized executable PFC, Historical #003, broader Phase 1, customer or production execution, live GitHub infrastructure, real traffic, Gate A promotion, or expert-validation completion. Historical #001 remains landed and independently reviewed. No general portability, exact upstream reproduction, customer readiness, or market claim follows.
