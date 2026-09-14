# Privacy boundary

The local/customer-controlled boundary contains raw source code, secrets, credentials, raw logs, production rows, private postmortems, raw traces, customer identifiers, IPs, hostnames, and environment-specific topology. Scanning, experiment execution, raw evidence storage, redaction, and safety enforcement occur inside that boundary by default. No shared registry or commercial control plane receives those inputs merely because they were scanned or tested.

An Architecture Contract may expose generalized structural facts, such as retry semantics or transaction boundaries, only after classification and explicit approval for sharing. A PFC derived from a private incident is private unless a separate review approves a sufficiently generalized contribution. Approval is specific to an artifact/version and destination; it is not inferred from execution. Reject or hold ambiguous exports. Redaction cannot be treated as proof of anonymity without review.

Shared artifacts need provenance, version, classification, intended audience, and an export decision. Local proof bundles may contain sensitive references and remain local. Any summary sent outside must exclude raw evidence by default and retain a pointer or digest to local evidence rather than its contents. Future implementations must test that telemetry, errors, AI prompts, and debug paths obey the same boundary.
