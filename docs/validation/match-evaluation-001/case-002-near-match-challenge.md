# Case 002 — near-match challenge

## A. Source facts

All cases use `pfc.timeout-after-external-commit@1.0.0`, still L0 CANDIDATE. The source is the repository's synthetic payment reference, not a qualifying public incident. See [contract](../../../contracts/timeout-after-external-commit.v1.json), [AC and evaluator](../../../internal/phase1/contracts.go), [adapter](../../../internal/phase1/run.go), and [schema](../../../internal/phase1/schema.sql). These paths are relative to this package through the repository root only where indicated below.

## B. Architecture facts

Synthetic design specification: retain the separate payment provider, post-commit response-loss barrier, stable payment identity and observable ledger, but the caller makes exactly one request and terminates on loss. No SDK retry, proxy retry, background recovery or manual resubmission exists within this operation's scope. This is a declared counterfactual architecture, not an implemented runner variant. Build is `match-evaluation-001-no-retry-v1`, AC ID is `match-evaluation-001-no-retry`, and RetryPolicy is `one request only; terminal on loss; no retry path`.

`external_effect`, `post_commit_uncertainty`, `post_commit_fault_placeable`, and `effect_ledger_observable` remain true by synthetic design. `same_operation_retry` is evidenced false by the complete bounded design above. Controls remain absent. The absent retry is a material causal prerequisite, not a protective-control exclusion: the contract's second attempt cannot happen.

## C. Assumptions

Synthetic declarations define the target within one logical payment operation. Facts inherited from the reference concern provider capabilities; no customer architecture, automatic extraction, runtime discovery, or upstream behavior is inferred. No execution proof is supplied by this matching evaluation.

## D. Pre-registered expected judgment

Recorded before this evaluation's first matcher invocation: `2026-09-15T23:27:07.156925+00:00`.

**NOT_APPLICABLE** — A proved false necessary prerequisite defeats this scoped two-attempt mechanism even though payments, PostgreSQL, an external commit and lost responses look similar. Therefore NOT_APPLICABLE, not a claim that payment processing is safe.

This is an agent-prepared expected human-review judgment requested by the task, not an actual independent human label or authoritative truth. Source/code inspection preceded registration; this is not a blinded experiment.

## E. FailureMesh judgment

Observed **NOT_APPLICABLE** from unchanged `phase1.Evaluate`, called through the [replay harness](grading-method.md). Full supplied AC and raw returned Match follow; the harness adds only the case identifier and input envelope. No compilation, safety-plan creation, database access or fault execution occurs in this call.

```json
{
  "case": "002",
  "input": {
    "schema_version": "1",
    "id": "match-evaluation-001-no-retry",
    "version": "1",
    "build": "match-evaluation-001-no-retry-v1",
    "environment": "synthetic-reference",
    "runtime": "Go",
    "database_boundary": "business state in PostgreSQL; external provider commit is independent",
    "provider_boundary": "separate simulator transaction commits provider effect",
    "retry_policy": "one request only; terminal on loss; no retry path",
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
        "value": false,
        "evidence": "Synthetic design: one request; terminal on loss; no SDK, proxy, background or manual retry in scope",
        "source": "match-evaluation-001 case-002 section B",
        "scope": "payment operation",
        "quality": "direct",
        "sensitivity": "synthetic"
      }
    }
  },
  "result": {
    "verdict": "NOT_APPLICABLE",
    "pfc": "pfc.timeout-after-external-commit@1.0.0",
    "ac": "c00598e2896c391de1b2c910edc9e69e1010249f403eaf0d2bef60d8f0341f24",
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
        "state": "FALSE",
        "evidence": "Synthetic design: one request; terminal on loss; no SDK, proxy, background or manual retry in scope",
        "reason": "necessary causal prerequisite absent"
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

Agreement: **correct negative** relative to section D. `same_operation_retry` is FALSE with nonempty direct synthetic evidence. `falseFound` is true, so `phase1.Evaluate` returns NOT_APPLICABLE. The other four prerequisites are TRUE. Similar vocabulary does not override the absent causal step.

## G. Disagreements / ambiguity

The grade depends on accepting the synthetic specification as direct evidence of no retry anywhere in scope. If this described a real system without corroboration, UNKNOWN would be defensible. The current evaluator trusts evidence metadata; it does not prove that the declared caller exists.

## H. Residual uncertainty

Only static applicability is tested. No exposure, resilience, portability, customer applicability or production-readiness verdict follows. The matcher checks supplied fact values and metadata rather than validating their underlying truth. Independent expert grading remains outstanding.
