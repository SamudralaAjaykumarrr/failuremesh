# External review scorecard

Blank template; one record per reviewer/session. Use a generalized reviewer identifier and public/synthetic evidence only. No personal contact data, confidential incidents, private logs, customer data, credentials, or secrets. Reuse the existing [session record](../../review/expert-review-decision-log.md) for detailed notes and claim classification; do not count the same review twice.

| Field | Response |
| --- | --- |
| Reviewer | |
| Role/background | |
| Date | |
| Compensated review? yes/no | |
| Repository commit/SHA reviewed | |
| Artifacts reviewed (include local modifications) | |
| Did reviewer execute code? yes/no | |
| Did reviewer inspect verifier? yes/no | |

## Ratings

Rate 1–5: **1** major objection/unusable in reviewed context; **2** substantial concerns; **3** mixed/conditional; **4** credible with stated limits; **5** strongly supported in reviewed context. Use **N/A** for unreviewed dimensions, with a reason. Supply evidence or reasoning, confidence, and any blocker. These are reviewer judgments, not verifier verdicts. There is no aggregate pass threshold; high scores cannot cancel critical objections or change Gate A.

| Criterion | Rating 1–5 or N/A | Evidence/reason; confidence; blocker |
| --- | --- | --- |
| Problem importance | | |
| Causal-model clarity | | |
| PFC abstraction quality | | |
| AC abstraction quality | | |
| Applicability trust | | |
| Proof integrity | | |
| Privacy model | | |
| Reproducibility | | |
| Practical usefulness | | |
| Willingness to test at work | | |

## Open questions

| Prompt | Response |
| --- | --- |
| Strongest criticism | |
| Most likely false positive | |
| Most likely false negative | |
| Weakest abstraction | |
| Missing evidence | |
| Missing failure family | |
| What would prevent adoption? | |
| What would increase trust? | |
| Would you run a bounded non-production pilot? yes/no/maybe | |
| Would you introduce FailureMesh to another engineer? yes/no/maybe | |

## Final classification

Select **SERIOUS REVIEW / PARTIAL REVIEW / NOT COUNTED**: ___

Classification rationale and supporting criticism or execution/inspection evidence: ___

- **SERIOUS REVIEW:** substantive technical criticism or concrete execution/inspection evidence tied to reviewed artifacts and scope. Capture what was challenged or checked, the result, and limits. A negative result can qualify.
- **PARTIAL REVIEW:** relevant engagement with incomplete scope or insufficient detail to meet the serious-review standard; record what remains missing.
- **NOT COUNTED:** social-media like, generic praise, unsupported score-only response, or no substantive review material. These must never count as serious reviews.

Preserve dissent, unsuccessful attempts, skipped questions, and UNKNOWNs. Compensation must remain disclosed; it is not evidence of endorsement. Interest in testing is neither pilot authorization nor an existing pilot. No classification alone completes expert validation, establishes a partner/payment, or changes execution authority.
