# Voluntarily presented public source: family 1 or 3

**Submission does not change Gate A.** This form records a concrete source a reviewer already knows; it is not a request for broad or recurring incident searches. Gate A remains REVISE, mechanism fit 10/10, qualifying incidents 8/10, families 1 and 3 unresolved, and Phase 1 NOT AUTHORIZED until a separate reviewed decision changes them.

Use public material only. Do not submit credentials, customer secrets, private production logs, personal data, confidential postmortems, proprietary code, or restricted customer information. Raw customer/private evidence is not required. A URL is a **PUBLIC SOURCE CLAIM**, not yet a repository-established fact.

| Field | Reviewer submission |
| --- | --- |
| Source URL and specific section/anchor | |
| Source owner/publisher | |
| Source publication or event date; uncertainty | |
| Family: 1 timeout after external commit / 3 stale worker after lease expiry | |
| Why this is an exact-family match, not merely a label match | |
| Observed operational context and scope | |
| Observed chronology, in source order | |
| Exact mechanism steps and source location for each | |
| What is directly evidenced | |
| What remains `UNKNOWN` (including causal links, effects, production scope) | |
| Source class: synthetic / formal / technical analysis / observed production / other observed operational | |
| Appears independent of existing records? Basis | |
| Possible double-counting of one event or family | |
| Conflicts or overlaps with current incident records | |
| Reviewer confidence and basis | |

Later review must follow the [REVISE disposition](../validation/gate-a/disposition.md): inspect the original source and provenance; establish a real observed operational failure and source class; check exact-family defining chronology and eligibility; assess independence, overlap, ambiguity, and [abstraction fit](../validation/gate-a/abstraction-fit.md); compare the existing family record; retain every unsupported step as `UNKNOWN`; then explicitly review any evidence-record, classification, fit-audit, and [Gate A decision](../validation/gate-a/decision.md) update. Family 1 requires a committed effect, post-commit caller uncertainty, retry, and a second committed logical effect. Family 3 requires lease/authority expiry or takeover followed by an observed stale actor's action. A synthetic run, formal model, technical possibility, advice, or near-match does not fill the missing incident slot. Presentation alone cannot increase 8/10, promote a PFC, authorize execution, or establish a FailureMesh verdict.
