# ADR 0004: Local-first privacy boundary

Status: Accepted for Phase 0. Date: 2026-09-13.

Decision: Raw customer-sensitive data and execution remain local/customer-controlled by default. Only sufficiently generalized structural information with explicit approval crosses into shared intelligence. See [privacy](../privacy.md).

Reason: The system must learn transferable mechanisms without requiring raw customer code, secrets, logs, rows, traces, identifiers, or private postmortems in a central service. Consequence: export is a separate reviewed operation; missing classification or approval blocks export.
