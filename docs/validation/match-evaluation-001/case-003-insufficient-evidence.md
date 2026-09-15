# Case 003 — insufficient evidence

## A. Source facts

All cases use `pfc.timeout-after-external-commit@1.0.0`, still L0 CANDIDATE. The source is the repository's synthetic payment reference, not a qualifying public incident. See [contract](../../../contracts/timeout-after-external-commit.v1.json), [AC and evaluator](../../../internal/phase1/contracts.go), [adapter](../../../internal/phase1/run.go), and [schema](../../../internal/phase1/schema.sql). These paths are relative to this package through the repository root only where indicated below.

## B. Architecture facts

Bounded evidence-withholding fixture: retain the reference provider transaction, response-loss barrier, identity and observable ledger, but supply no retry fact and set RetryPolicy to `UNKNOWN: caller retry evidence not supplied`. Build is `match-evaluation-001-retry-unknown-v1`; AC ID is `match-evaluation-001-retry-unknown`. The evaluated evidence packet deliberately omits the `same_operation_retry` key entirely. It does not assert retry is absent.

Four prerequisites have direct synthetic support. `same_operation_retry` has no value, evidence, source or quality because the entire fact is omitted. No prerequisite is proved false and no mechanism-defeating exclusion is established. Idempotency and reconciliation remain absent as declared fixture context.

## C. Assumptions

Synthetic declarations define the target within one logical payment operation. Facts inherited from the reference concern provider capabilities; no customer architecture, automatic extraction, runtime discovery, or upstream behavior is inferred. No execution proof is supplied by this matching evaluation.

## D. Pre-registered expected judgment

Recorded before this evaluation's first matcher invocation: `2026-09-15T23:27:07.156925+00:00`.

**UNKNOWN** — The causal mapping requires a same-operation retry, but this packet provides no evidence either way. Missing information must stay UNKNOWN; neither NOT_APPLICABLE nor APPLICABLE follows.

This is an agent-prepared expected human-review judgment requested by the task, not an actual independent human label or authoritative truth. Source/code inspection preceded registration; this is not a blinded experiment.

## E. FailureMesh judgment

Observed **UNKNOWN** from unchanged `phase1.Evaluate`, called through the [replay harness](grading-method.md). Full supplied AC and raw returned Match follow; the harness adds only the case identifier and input envelope. No compilation, safety-plan creation, database access or fault execution occurs in this call.

```json
{
  "case": "003",
  "input": {
    "schema_version": "1",
    "id": "match-evaluation-001-retry-unknown",
    "version": "1",
    "build": "match-evaluation-001-retry-unknown-v1",
    "environment": "synthetic-reference",
    "runtime": "Go",
    "database_boundary": "business state in PostgreSQL; external provider commit is independent",
    "provider_boundary": "separate simulator transaction commits provider effect",
    "retry_policy": "UNKNOWN: caller retry evidence not supplied",
    "provider_guarantee": "append-only effect ledger; optional durable idempotency key",
    "logical_identity": "run ID plus logical payment ID",
    "idempotency": "absent",
    "reconciliation": "absent",
    "facts": {
      "effect_ledger_observable": {
        "value": true,
        "evidence": "reference adapter and PostgreSQL schema",
        "source": "reference implementation",
        "scope": "payment operation",
        "quality": "direct",
        "sensitivity": "synthetic"
      },
      "external_effect": {
        "value": true,
        "evidence": "reference adapter and PostgreSQL schema",
        "source": "reference implementation",
        "scope": "payment operation",
        "quality": "direct",
        "sensitivity": "synthetic"
      },
      "post_commit_fault_placeable": {
        "value": true,
        "evidence": "reference adapter and PostgreSQL schema",
        "source": "reference implementation",
        "scope": "payment operation",
        "quality": "direct",
        "sensitivity": "synthetic"
      },
      "post_commit_uncertainty": {
        "value": true,
        "evidence": "reference adapter and PostgreSQL schema",
        "source": "reference implementation",
        "scope": "payment operation",
        "quality": "direct",
        "sensitivity": "synthetic"
      }
    }
  },
  "result": {
    "verdict": "UNKNOWN",
    "pfc": "pfc.timeout-after-external-commit@1.0.0",
    "ac": "ad035c43935852ea9a6ced3c12a69a00c09e3b871fee9d621f3a900b68a6d3b9",
    "facts": [
      {
        "name": "external_effect",
        "state": "TRUE",
        "evidence": "reference adapter and PostgreSQL schema",
        "reason": "evidenced prerequisite"
      },
      {
        "name": "post_commit_uncertainty",
        "state": "TRUE",
        "evidence": "reference adapter and PostgreSQL schema",
        "reason": "evidenced prerequisite"
      },
      {
        "name": "post_commit_fault_placeable",
        "state": "TRUE",
        "evidence": "reference adapter and PostgreSQL schema",
        "reason": "evidenced prerequisite"
      },
      {
        "name": "same_operation_retry",
        "state": "UNKNOWN",
        "reason": "missing, stale, conflicting, or unsupported fact"
      },
      {
        "name": "effect_ledger_observable",
        "state": "TRUE",
        "evidence": "reference adapter and PostgreSQL schema",
        "reason": "evidenced prerequisite"
      }
    ]
  }
}
```

## F. Grading rationale

Agreement: **correct UNKNOWN** relative to section D. The missing `same_operation_retry` map entry triggers the first switch branch: `missing, stale, conflicting, or unsupported fact`. `unknown` is true and `falseFound` is false. The other four facts cannot establish applicability alone.

## G. Disagreements / ambiguity

This is deliberately withheld evidence, not a naturally discovered incident gap. A reviewer who imports the original reference caller's retry implementation could label that original architecture APPLICABLE; that is a different evidence scope. This case tests the packet actually supplied.

## H. Residual uncertainty

Only static applicability is tested. No exposure, resilience, portability, customer applicability or production-readiness verdict follows. The matcher checks supplied fact values and metadata rather than validating their underlying truth. Independent expert grading remains outstanding.
