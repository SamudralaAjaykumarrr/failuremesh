# Grading method and pre-registration

Registration time: `2026-09-15T23:27:07.156925+00:00`. Base revision: `bc8d2956b1f4ac1cd20c8d5c35eaba5c961d32a5`.

Expected judgments, recorded in each case section D before invoking the matcher:

| Case | Expected | Decisive rationale |
| --- | --- | --- |
| 001 | APPLICABLE | All five necessary prerequisites supported |
| 002 | NOT_APPLICABLE | Same-operation retry demonstrably absent in the synthetic design |
| 003 | UNKNOWN | Retry evidence omitted, with no proved false prerequisite |

These are pre-registered expected judgments for human review, prepared by the coding agent from repository semantics. No independent human endorsement, reviewer transcript, identity, or completed review is claimed. The task supplies the generalized external-feedback motivation; no private feedback is imported. Inspection of existing rules before registration means these are transparent designed cases, not blind holdouts.

Use [applicability semantics](../../architecture/applicability.md): evidenced false necessary facts yield NOT_APPLICABLE; otherwise every required fact must be supported for APPLICABLE; otherwise UNKNOWN. Control presence is not a generic exclusion. Applicability and execution verdicts are separate.

Freeze the case design and expected judgment before the first result; preserve disagreement instead of editing a label to fit output. Full ACs, raw matcher outputs and replay code will be appended. Digests identify bytes, not independent attestation of chronology. Local tool history records write-before-run ordering; no timestamped third-party registration or signed provenance is claimed.

Classification rules: expected/actual APPLICABLE is correct positive; NOT_APPLICABLE/NOT_APPLICABLE is correct negative; UNKNOWN/UNKNOWN is correct UNKNOWN. APPLICABLE against expected NOT_APPLICABLE is a false positive; NOT_APPLICABLE against expected APPLICABLE is a false negative. Unjustified certainty against expected UNKNOWN, or UNKNOWN against a determinate expectation, must be described explicitly and graded disputed rather than hidden. Invocation failure is execution not possible. Any defensible label challenge remains visible even when labels agree. No aggregate accuracy metric is the deliverable.

## Provenance and audit procedure

The three section A–D prefixes were written before any evaluation invocation with expected judgments and architecture designs unchanged afterward. A subsequent mechanical correction changed root-relative links from four parent traversals to three; no evidence or label changed. SHA-256 below covers UTF-8 bytes before the literal `## E. FailureMesh judgment` heading (including preceding blank lines). To reproduce the original hashes from current case files, replace `../../../contracts/` with `../../../../contracts/` and `../../../internal/` with `../../../../internal/` in the prefix before hashing (the original link-depth typo). This allows checking the preserved registration content; it cannot independently prove when it was written.

| Case | A–D SHA-256 |
| --- | --- |
| case-001-positive-match.md | `5e0bb79d95a2f2f44a7f1e697c39b5055ed7685870ffc2b500ca9c540b56ffed` |
| case-002-near-match-challenge.md | `366504872c44c640eb52c9e78fdfcb295e00912c4382b9c87ba80e1add65b16a` |
| case-003-insufficient-evidence.md | `6001a0ebe5bc572d8d3f857aff777049151d29de6b9c1daa89c2b68d139650ef` |

Results captured on 2026-09-15T23:29:50.159503+00:00 after registration. Host toolchain: `go version go1.26.5 linux/amd64`. The first successful invocation was `GOCACHE=/tmp/failuremesh-go-cache go run match_evaluation_001_tmp.go`; exit 0. Complete stdout SHA-256: `3773be6e6e983eb5b0df763ae1ff167ccf5395da9c355d1da7e24fafe6610b4d`. Each case embeds its complete JSON record. No result-driven input or expected-label edits occurred.

The base revision pins engine, contract, reference adapter and schema. Their bytes are unchanged by this package. Match field `ac` is SHA-256 of Go `json.Marshal(AC)` through `phase1.Digest`, not of indented JSON. The match includes PFC version and every prerequisite decision. No engine rule version field exists here; use the pinned revision.

The CLI supports only built-in vulnerable/remediated ACs. The following harness uses exactly its `LoadPFC` → `ReferenceAC` → `Evaluate` path, changing only the documented case inputs. It does not implement matching rules or grade its own results. It deliberately never calls Compile, Plan, Run or Verify. Cases 002/003 are static synthetic AC designs, not new executable adapters. A temporary file must live inside the module to import its existing internal package.

## Replay harness

```go
package main

import (
	"encoding/json"
	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
	"os"
)

func main() {
	p, err := phase1.LoadPFC("contracts/timeout-after-external-commit.v1.json")
	if err != nil {
		panic(err)
	}
	type record struct {
		Case   string       `json:"case"`
		Input  phase1.AC    `json:"input"`
		Result phase1.Match `json:"result"`
	}
	records := []record{}
	for _, id := range []string{"001", "002", "003"} {
		a := phase1.ReferenceAC(false)
		if id == "002" {
			a.ID = "match-evaluation-001-no-retry"
			a.Build = "match-evaluation-001-no-retry-v1"
			a.RetryPolicy = "one request only; terminal on loss; no retry path"
			a.Facts["same_operation_retry"] = phase1.Fact{
				Value:    phase1.Bool(false),
				Evidence: "Synthetic design: one request; terminal on loss; no SDK, proxy, background or manual retry in scope",
				Source:   "match-evaluation-001 case-002 section B",
				Scope:    "payment operation", Quality: "direct", Sensitivity: "synthetic",
			}
		}
		if id == "003" {
			a.ID = "match-evaluation-001-retry-unknown"
			a.Build = "match-evaluation-001-retry-unknown-v1"
			a.RetryPolicy = "UNKNOWN: caller retry evidence not supplied"
			delete(a.Facts, "same_operation_retry")
		}
		records = append(records, record{id, a, phase1.Evaluate(p, a)})
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	if err := e.Encode(records); err != nil {
		panic(err)
	}
}
```

## Reproduce and compare

From repository root, run this standard-library Python command. It extracts the exact harness above into a temporary directory inside the module, executes it, compares every input and output field with each case's recorded JSON, and removes its own temporary directory. It does not edit existing Go sources. Go may need its normal module/build cache access. No PostgreSQL or external service is needed for matching.

```sh
python3 - <<'PYREPLAY'
from pathlib import Path
import json, os, subprocess, tempfile
p = Path('docs/validation/match-evaluation-001')
code = (p / 'grading-method.md').read_text().split('```go\n', 1)[1].split('```', 1)[0]
with tempfile.TemporaryDirectory(prefix='match-eval-replay-', dir='.') as tmp:
    source = Path(tmp) / 'main.go'
    source.write_text(code)
    env = dict(os.environ, GOCACHE='/tmp/failuremesh-go-cache')
    result = subprocess.run(['go', 'run', str(source)], env=env,
                            check=True, text=True, capture_output=True)
    actual = json.loads(result.stdout)
expected = [json.loads(f.read_text().split('```json\n', 1)[1].split('```', 1)[0])
            for f in sorted(p.glob('case-*.md'))]
assert actual == expected, 'Replay differs: inspect inputs and results; do not relabel'
print('All three complete input/result records match')
PYREPLAY
```

## Evidence limits and independent grading

The evaluator consumes typed assertions and evidence strings; it does not inspect the referenced source, verify metadata honesty, or discover retry semantics from prose. Its checks in `phase1.Evaluate` are: absent/nil fact, empty evidence/source, stale/conflict flags or non-direct quality → UNKNOWN fact; evidenced false → FALSE; otherwise TRUE. It also checks schema 1 and nonempty build/environment/provider/database/logical identity. A false prerequisite wins over unknown; all supported true yields APPLICABLE. It does not check the semantic consistency of `RetryPolicy` against facts. These three cases do not test dishonest facts, contradictory prose, stale evidence, or all possible exclusions.

For each disputed grade, identify the case, exact prerequisite, evidence scope, proposed judgment and supporting reason. Preserve the original section D and observed section E. Add a dated independent assessment in subsequent review work rather than overwrite the record. No review is counted by preparing this package.
