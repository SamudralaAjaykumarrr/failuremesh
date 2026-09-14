# PROPOSED BASELINE CHANGE

Status: **APPROVED on 2026-09-14** by the project owner: “Approve the bounded second executable PFC milestone.”

## What changes and why

The canonical baseline now records one additional, explicitly authorized synthetic executable slice: duplicate queue delivery. This permits a bounded test of whether the PFC/AC, safety, and proof model also represents a serial queue consumer mechanism. It does not change the company thesis or the ten-incident validation rule.

## Evidence and difference from the prior baseline

The prior authorization covered only timeout-after-external-commit. Gate A records 10/10 exact-family L0 mechanism mappings and 8/10 qualifying public observed incidents. Its family-2 Vercel Workflow record supports observed queue redelivery, while committed provider effects remain unknown. The first slice supplied a local synthetic PostgreSQL demonstration, not public incident evidence. These observations justify attempting a second bounded experiment, not claiming portability or public-source promotion.

## Approved envelope and unchanged boundaries

Only PFC #2 duplicate queue delivery is added: Go, a deterministic serial synthetic queue/reference adapter, one stable logical message delivered twice, PostgreSQL durable consumer effects and idempotency, vulnerable duplicate effect, identical remediated replay, and strict local evidence verification at safety level 1. The first bounded slice remains timeout-after-external-commit. Broader Phase 1, other PFCs, customer execution, real queue integrations, uncontrolled chaos, and production execution are **NOT AUTHORIZED**.

Gate A remains **REVISE**, mechanism fit 10/10, qualifying public observed incidents 8/10, families 1 and 3 unresolved, and generic incident search CLOSED. Synthetic results are separate from public-incident classification. The benefit of this narrow baseline addition is a reviewable second mechanism test while preserving the existing privacy, safety, UNKNOWN, and scoped-verdict rules.
