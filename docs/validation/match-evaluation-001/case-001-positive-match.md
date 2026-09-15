# Case 001 — positive match

## A. Source facts

All cases use `pfc.timeout-after-external-commit@1.0.0`, still L0 CANDIDATE. The source is the repository's synthetic payment reference, not a qualifying public incident. See [contract](../../../contracts/timeout-after-external-commit.v1.json), [AC and evaluator](../../../internal/phase1/contracts.go), [adapter](../../../internal/phase1/run.go), and [schema](../../../internal/phase1/schema.sql). These paths are relative to this package through the repository root only where indicated below.

## B. Architecture facts

The existing `ReferenceAC(false)` describes Go, PostgreSQL business state, a separate simulator transaction committing provider effects, one immediate retry after response loss with the same run/payment identity, and an observable append-only ledger. Idempotency and reconciliation are absent. The adapter's `provider` commits separately; `Run` holds success after commit, records response loss, and continues to attempt 2; `ledger` reads effects. This is a real existing synthetic implementation, inspected statically here.

All five prerequisites are evidenced true: `external_effect`, `post_commit_uncertainty`, `post_commit_fault_placeable`, `same_operation_retry`, `effect_ledger_observable`. No specified exclusion defeats this mechanism. Missing idempotency is a control-under-test, not a prerequisite.

## C. Assumptions

Synthetic declarations define the target within one logical payment operation. Facts inherited from the reference concern provider capabilities; no customer architecture, automatic extraction, runtime discovery, or upstream behavior is inferred. No execution proof is supplied by this matching evaluation.

## D. Pre-registered expected judgment

Recorded before this evaluation's first matcher invocation: `2026-09-15T23:27:07.156925+00:00`.

**APPLICABLE** — Every causal prerequisite is supported in the existing bounded reference. Commit → response loss → same-operation retry is meaningful and observable. APPLICABLE predicts neither duplicate effects nor resilience.

This is an agent-prepared expected human-review judgment requested by the task, not an actual independent human label or authoritative truth. Source/code inspection preceded registration; this is not a blinded experiment.

## E. FailureMesh judgment

Observed **APPLICABLE** from unchanged `phase1.Evaluate`, called through the [replay harness](grading-method.md). Full supplied AC and raw returned Match follow; the harness adds only the case identifier and input envelope. No compilation, safety-plan creation, database access or fault execution occurs in this call.

```json
{
  "case": "001",
  "input": {
    "schema_version": "1",
    "id": "reference-payment",
    "version": "1",
    "build": "reference-vulnerable-v1",
    "environment": "synthetic-reference",
    "runtime": "Go",
    "database_boundary": "business state in PostgreSQL; external provider commit is independent",
    "provider_boundary": "separate simulator transaction commits provider effect",
    "retry_policy": "one immediate retry of same logical ID after lost response",
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
      },
      "same_operation_retry": {
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
    "verdict": "APPLICABLE",
    "pfc": "pfc.timeout-after-external-commit@1.0.0",
    "ac": "8bba3f5e85e1f9f741cba661894968697514fc17557293be667c5b87ba59b9d8",
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
        "state": "TRUE",
        "evidence": "reference adapter and PostgreSQL schema",
        "reason": "evidenced prerequisite"
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

Agreement: **correct positive** relative to section D. All five fact decisions are TRUE, each has nonempty evidence/source and direct quality, and all required AC identity/boundary fields are present. `falseFound` and `unknown` remain false, so `phase1.Evaluate` returns APPLICABLE. No control fact drives this decision.

## G. Disagreements / ambiguity

A reviewer may question whether the broad built-in evidence string is adequate corroboration. The linked adapter and schema provide the concrete support; this is not an independently scanned architecture.

## H. Residual uncertainty

Only static applicability is tested. No exposure, resilience, portability, customer applicability or production-readiness verdict follows. The matcher checks supplied fact values and metadata rather than validating their underlying truth. Independent expert grading remains outstanding.
