# Expert review session record template

Create one copy per session. Preserve the reviewer's words and the limits of their account. This is a record of technical criticism, not a Gate A decision, canonical incident record, product verdict, or completed validation milestone. Use the [brief](expert-review-brief.md), [guide](expert-review-interview-guide.md), and [scorecard](expert-review-scorecard.md).

| Session field | Record |
| --- | --- |
| Date, reviewer role/domain, organization type if safely generalizable | |
| Review scope, materials seen, questions discussed/skipped | |
| Major supports, with question number and claim type | |
| Major objections and counterexamples, with question number and claim type | |
| Critical blockers, affected premise, and reviewer context | |
| Unknowns, conflicts, and limits of direct experience | |
| Voluntarily presented public source candidates; link to separate [submission form](expert-source-submission.md) | |
| Follow-up owner/action, if any; no broad-search mandate | |
| Thesis impact: supports / weakens / does not affect / unknown, with reason | |
| Architecture impact: none / conceptual clarification / possible defect / unknown, with reason | |
| Gate A impact: none unless separately reviewed under the disposition | |
| Baseline-change implication: none / potential, requiring separate process | |
| Scorecard location; omitted dimensions and reasons | |

Tag each substantive assertion as **CURRENT FACT** (cite an authoritative repository record), **REVIEWER OPINION** (judgment or forecast), **OBSERVED EXPERIENCE** (firsthand account with scope and limits), **FAILUREMESH HYPOTHESIS** (untested proposition), **UNKNOWN** (unresolved), or **PUBLIC SOURCE CLAIM** (specific public URL pending source review). An observed experience is not automatically a public incident, and a public source claim is not automatically eligible Gate A evidence. Mark direct quotes versus paraphrases. Do not turn notes or scores into canonical facts by repetition.

If feedback suggests altering the thesis, open a separate item labeled exactly **PROPOSED BASELINE CHANGE** and supply what would change, why, supporting evidence, difference from the current baseline, and why the benefit justifies it. Do not edit the [canonical baseline](../company/canonical-baseline-v1.0.md) or infer approval from this log. Gate A remains REVISE, 10/10 mechanism fit, 8/10 qualifying incidents, and families 1 and 3 unresolved. The [approved bounded-entry change](../company/baseline-change-001-bounded-phase1-entry.md) authorizes only the first prototype; broader execution needs separate authorization.

Do not record credentials, customer secrets, private production logs, personal data, confidential postmortems, proprietary code, or restricted customer information. Generalize experience; use public sources where possible. Raw customer/private evidence is not required.
