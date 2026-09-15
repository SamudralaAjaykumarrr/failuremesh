# Match Evaluation #001 results

These classifications are relative to the [pre-registered expected judgments](grading-method.md), not authoritative human truth. Each linked case contains the complete AC, raw matcher decision, evidence rationale and ambiguity.

| Case | Expected judgment | FailureMesh judgment | Agreement? | Observed classification | Most important reason | Disputable grading point |
| --- | --- | --- | --- | --- | --- | --- |
| [001](case-001-positive-match.md) | APPLICABLE | APPLICABLE | Yes | correct positive | All five causal prerequisites have direct synthetic support | Broad reference evidence strings require source inspection |
| [002](case-002-near-match-challenge.md) | NOT_APPLICABLE | NOT_APPLICABLE | Yes | correct negative | Same-operation retry is explicitly absent in the bounded design | Synthetic declaration is not runtime proof of all real retry paths |
| [003](case-003-insufficient-evidence.md) | UNKNOWN | UNKNOWN | Yes | correct UNKNOWN | Retry fact is omitted; nothing proves it true or false | Packet withholding tests evidence scope, not a naturally discovered gap |

No false positive occurred in these three cases. No false negative occurred. Case 003 returned UNKNOWN without silently assuming retry. No invocation failed. No label disagreement was observed; the disputable grading points above remain open for independent review. These selected, code-informed cases do not estimate population accuracy, real-world false-positive rates, or general matching quality.

## Quality and governance verification

Checks run on 2026-09-15 against base revision `bc8d2956b1f4ac1cd20c8d5c35eaba5c961d32a5` plus this documentation package:

| Check | Result |
| --- | --- |
| Documented replay command | PASS: all three complete input/result records equal recorded JSON |
| Original A–D registration digest reconstruction | PASS for all cases; only documented link-depth correction after registration |
| `python3 scripts/check_markdown_links.py` | PASS: repository-local links (not remote URLs/fragments) |
| `python3 scripts/validate_gate_a.py` | PASS: baseline byte integrity and Gate A structural/checkpoint integrity |
| `python3 -B -m unittest discover -s scripts/tests -v` | PASS: 25 tooling tests |
| `gofmt -l .` | PASS: no output |
| `go mod verify` | PASS: all modules verified |
| `go vet ./...` | PASS |
| `go test ./...` and fresh `go test -count=1 ./...` | PASS: all five library packages; CLI has no test files |
| `go test -race -count=1 ./...` | PASS: all five library packages; CLI has no test files |
| `git diff --check` plus explicit untracked-file whitespace inspection | PASS |
| Complete changed/untracked-file inspection | PASS: six Markdown files plus one informational HANDOFF line; no Go source changes |

Go commands used `GOCACHE=/tmp/failuremesh-go-cache`. Normal and race suites used the existing isolated loopback synthetic PostgreSQL fixture via `FAILUREMESH_TEST_DATABASE_URL` as documented in the phase run guides, with database access outside the filesystem/network sandbox. Fresh runs exercised database-backed tests; matching replay itself uses no database. These quality tests are separate from the three matching observations and do not grant new execution authority.

Explicit governance audit: canonical baseline, Gate A decision, engine/verifier, contracts, phase specifications and approval documents have no changes. Every pre-existing HANDOFF line is preserved. Gate A remains **REVISE**; mechanism fit remains **10/10**; qualifying public incidents remain **8/10**; families **1 and 3 remain unresolved**. PFC #4 and Historical #003 remain **unauthorized**. Customer execution, production execution and broader Phase 1 remain **unauthorized**. No customer/private evidence or production system was used. Validator success is not source-evidence truth or external validation.

Final worktree scope: modified `HANDOFF.md`; six untracked Markdown files in this package. No staging, commit, push, PR or merge was performed. The temporary Go harness was removed; its complete source remains in the grading method.

## Recommendation

**READY FOR INDEPENDENT REVIEW.** The cases expose full inputs, exact rules and results, preserved expected judgments, replay instructions and disputable grading points; required checks pass and governance boundaries remain intact. This recommendation concerns artifact reviewability only. Independent human grading, natural architecture samples, and broader matching-quality evidence remain outstanding.
