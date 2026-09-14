# PROPOSED BASELINE CHANGE

Status: **APPROVED on 2026-09-14** by the project owner: “Approve the bounded third executable PFC milestone: stale worker after lease expiry with fencing.”

## What changes

The third canonical family, `pfc.stale-worker-after-lease-expiry@1.0.0`, may be implemented as a level-1 deterministic synthetic Go/PostgreSQL reference experiment with two logical workers, logical expiry, durable token generation, and a protected resource enforcing fencing in the remediated build. PFC #1 timeout-after-external-commit, PFC #2 duplicate-queue-delivery, and PFC #3 stale-worker-after-lease-expiry are separately **AUTHORIZED bounded synthetic slices**. PFC #4+, broader Phase 1, customer execution, production execution, real lease-system integration, and family-3 Gate A promotion are **NOT AUTHORIZED**.

## Why and evidence

The owner expressly approved this narrow engineering milestone. Gate A family 3 has a mechanism-level L0 mapping, sufficient to motivate a synthetic experiment, but no qualifying public observed incident. The experiment tests the invariant at a PostgreSQL effect sink without representing a customer architecture or a real lease provider. Gate A remains **REVISE**, mechanism fit 10/10, qualifying public observed incidents 8/10, families 1 and 3 unresolved, and generic incident search CLOSED.

## Difference from the current baseline

The baseline previously named the third family and the three-executable-PFC validation sequence, while explicitly authorizing only the first two bounded slices. This change authorizes only the third slice. The core thesis, PFC lifecycle, source classifications, safety levels, privacy policy, and validation gate are unchanged.

## Why the benefit justifies the change

A controlled stale-owner replay can test whether the PFC, AC, verifier, and proof boundary express a distinct lease/fencing mechanism. The narrow local scope keeps that engineering evidence separate from public incident validation and any portability or production claim. After this milestone is independently reviewed and landed, the planned next mode is external validation, not automatic PFC #4 implementation.
