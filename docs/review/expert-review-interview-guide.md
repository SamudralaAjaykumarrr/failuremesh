# Expert technical interview guide

Use for a 30–45 minute conversation with an SRE, platform, distributed-systems, backend-infrastructure, or reliability engineer. Send the [brief](expert-review-brief.md) first. Ask for counterexamples and concrete experiences before explaining a preferred answer. Do not seek confidential artifacts; public sources and generalized experience suffice. Record skipped questions and negative answers in the [scorecard](expert-review-scorecard.md) and [decision log](expert-review-decision-log.md). An interview is not a Gate A decision or completed market validation.

## Opening and thesis tests (about 20–25 minutes)

1. What recurring failure has your team learned from elsewhere, if any? How did that knowledge reach your team, and what remained untested? **[THESIS: problem validity]**
2. Describe a failure mechanism that did or did not transfer between two systems. Which architecture differences changed the outcome? **[THESIS: transferability]**
3. Given a causal PFC with prerequisites, sequence, invariant, experiment, and proof needs, which parts clarify the failure and which obscure it? What cannot be represented? **[THESIS: PFC abstraction]**
4. What facts would a matcher need to decide whether a specific operation can exercise that mechanism? Which AC facts are unavailable, stale, ambiguous, or expensive to establish? **[THESIS: AC feasibility]**
5. Where would an applicability result be wrong? Give a false positive and a false negative that would cause you to distrust the tool. What explanation or `UNKNOWN` behavior would you require? **[THESIS: applicability trust]**
6. What controlled experiment, if any, would your team permit on a developer machine, CI, or staging system? What approvals, abort controls, and operational cost would make local/customer-controlled execution acceptable or unacceptable? **[THESIS: execution acceptability]**
7. For a finite experiment claiming scoped `PROVEN_RESILIENT`, what observations and completeness checks would you demand? What would make the result inconclusive despite no observed failure? **[THESIS: evidence trust]**
8. Which of the [ten families](expert-review-brief.md) address costly problems in your domain? Which are incorrectly scoped, too overlapping, or missing a necessary distinction? **[THESIS: corpus fit]**

## Adoption and alternatives (about 8–12 minutes)

9. What privacy or security concern would prevent evaluation? What generalized structural facts could be shared, if any? **[THESIS: privacy acceptability]**
10. Who would own installation, review of applicability, approval of experiments, and follow-up? Where would the workflow stall? **[THESIS: operational adoption]**
11. Which current tools or practices would this replace, compete with, or complement? What capability would remain distinct, if any? **[THESIS: differentiation]**
12. What would make you refuse to use it? What minimum successful, bounded demonstration would make you interested enough to examine it further? **[THESIS: adoption likelihood]**

## Preferences and close (about 5–8 minutes)

13. Which output format, integration point, or review cadence would suit your team? **[PREFERENCE: does not validate the thesis]**
14. What terminology or ordering would make the PFC/AC explanation easier to inspect? **[PREFERENCE: does not validate the thesis]**
15. What did this conversation miss? Did any answer reflect direct experience, opinion, or a public source we can cite? **[CLASSIFICATION CHECK]**

If a reviewer volunteers a concrete public source for family 1 or 3, record it separately with the [source submission form](expert-source-submission.md). Do not ask the reviewer to search broadly. Do not request credentials, customer secrets, private production logs, personal data, confidential postmortems, proprietary code, or restricted customer information. Do not convert interview notes into canonical incident evidence.
