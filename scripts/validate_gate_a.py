#!/usr/bin/env python3
"""Check the recorded Gate A checkpoint, not the truth of its source evidence."""

import hashlib
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
BASELINE_SHA256 = "3f204ff171a90df237bef46dda7e24fc72207edc46a2deb9e3e2fc62de686b35"
INCIDENTS = (
    "01-timeout-after-external-commit.md",
    "02-duplicate-queue-delivery.md",
    "03-stale-worker-after-lease-expiry.md",
    "04-retry-amplification.md",
    "05-postgresql-connection-pool-exhaustion.md",
    "06-poison-message-retry-loop.md",
    "07-partial-transaction-divergence.md",
    "08-dependency-latency-cascade.md",
    "09-cache-stampede.md",
    "10-worker-crash-after-write-before-ack.md",
)
ARCHITECTURES = (
    "01-sync-service-external-side-effect.md",
    "02-queue-worker-delivery.md",
    "03-distributed-service-dependency.md",
)
NON_COUNTING = {1, 3}
CURRENT_CLAIMS = (
    ("FailureMesh reproduction", re.compile(r"\bFailureMesh\s+(?:has\s+)?reproduced\b|\bFailureMesh reproduction\s+(?:was|is|has been)\s+(?:completed|achieved|confirmed)\b", re.IGNORECASE)),
    ("L1 promotion", re.compile(r"\b(?:this|the)\s+(?:record|candidate|incident)\s+(?:is|was|has been|was promoted to|has been promoted to)\s+L1\s+REPRODUCED\b|\*\*(?:current )?verification level:\*\*\s*L1\s+REPRODUCED\b", re.IGNORECASE)),
    ("customer applicability decision", re.compile(r"\b(?:customer )?applicability\s+(?:(?:decision|verdict)\s*:\s*|(?:was|is|has been)\s+(?:determined\s+|decided\s+)?)(?:APPLICABLE|NOT_APPLICABLE)\b|\*\*current (?:customer )?applicability:\*\*\s*(?:APPLICABLE|NOT_APPLICABLE)\b|\b(?:this|the)\s+record\s+(?:has received|received)\s+(?:a\s+)?customer applicability (?:decision|verdict)\s+of\s+(?:APPLICABLE|NOT_APPLICABLE)\b", re.IGNORECASE)),
    ("FailureMesh execution verdict", re.compile(r"\bFailureMesh\s+(?:issued|returned|assigned|emitted)\s+(?:an?\s+)?(?:(?:execution\s+)?verdict\s+(?:of\s+)?)?(?:EXPOSED|PROVEN_RESILIENT)\b|\b(?:current )?execution verdict\s*:\s*(?:EXPOSED|PROVEN_RESILIENT)\b|\b(?:this|the)\s+record\s+(?:is|was)\s+(?:EXPOSED|PROVEN_RESILIENT)\b", re.IGNORECASE)),
)


def read(root, relative, errors):
    path = root / relative
    if not path.is_file():
        errors.append(f"{relative}: missing required file")
        return ""
    try:
        return path.read_text(encoding="utf-8")
    except (OSError, UnicodeError) as exc:
        errors.append(f"{relative}: cannot read UTF-8 text: {exc}")
        return ""


def require(text, pattern, label, relative, errors, flags=re.IGNORECASE):
    if not re.search(pattern, text, flags):
        errors.append(f"{relative}: missing or inconsistent {label}")


def field(text, name):
    match = re.search(rf"^- \*\*{re.escape(name)}:\*\*\s*(.+)$", text, re.MULTILINE)
    return match.group(1) if match else None


def validate(root=ROOT):
    errors = []
    baseline = "docs/company/canonical-baseline-v1.0.md"
    baseline_path = root / baseline
    if not baseline_path.is_file():
        errors.append(f"{baseline}: missing required file")
    else:
        actual = hashlib.sha256(baseline_path.read_bytes()).hexdigest()
        if actual != BASELINE_SHA256:
            errors.append(f"{baseline}: SHA-256 mismatch (expected {BASELINE_SHA256}, got {actual}); an approved baseline change requires a deliberate digest update")

    gate = Path("docs/validation/gate-a")
    incident_dir = root / gate / "incidents"
    architecture_dir = root / gate / "reference-architectures"
    for directory, expected in ((incident_dir, INCIDENTS), (architecture_dir, ARCHITECTURES)):
        actual = {p.name for p in directory.glob("*.md")} if directory.is_dir() else set()
        for missing in sorted(set(expected) - actual):
            errors.append(f"{directory.relative_to(root) / missing}: missing required file")
        for extra in sorted(actual - set(expected)):
            errors.append(f"{directory.relative_to(root) / extra}: unexpected Gate A record")

    counts = []
    for number, filename in enumerate(INCIDENTS, 1):
        relative = str(gate / "incidents" / filename)
        content = read(root, relative, errors)
        if not content:
            continue
        for name in ("Candidate family", "Verification level", "Exact-family match", "Verification level remains", "Counts toward ten-public-incident milestone", "Mechanism-fit status"):
            if field(content, name) is None:
                errors.append(f"{relative}: missing {name} field")
        for heading in ("Evidence boundary", "Candidate PFC mapping", "Canonical public-incident audit"):
            require(content, rf"^## {re.escape(heading)}\s*$", heading + " section", relative, errors, re.MULTILINE)
        require(content, r"^- \*\*Public references?:\*\*\s*\[[^]]+\]\(https?://[^)]+\)", "public source/provenance link", relative, errors, re.MULTILINE)
        require(content, r"^- \*\*UNKNOWN(?:\s|:| —)", "explicit UNKNOWN evidence", relative, errors, re.MULTILINE)
        for name in ("Verification level", "Verification level remains"):
            value = field(content, name)
            if value is not None and not value.startswith("L0 CANDIDATE"):
                errors.append(f"{relative}: {name} must remain L0 CANDIDATE")
        for name in ("Exact-family match", "Mechanism-fit status"):
            value = field(content, name)
            if value is not None and not re.match(r"YES\b", value):
                errors.append(f"{relative}: {name} must remain YES at this checkpoint")
        value = field(content, "Counts toward ten-public-incident milestone")
        expected = "NO" if number in NON_COUNTING else "YES"
        if value is not None:
            observed = re.match(r"(YES|NO)\b", value)
            if not observed or observed.group(1) != expected:
                errors.append(f"{relative}: canonical incident milestone must be {expected}")
            if observed and observed.group(1) == "YES":
                counts.append(number)
        if not re.search(r"\bno FailureMesh reproduction\b|\bno reproduction\b", content, re.IGNORECASE):
            errors.append(f"{relative}: missing explicit no-reproduction boundary")
        if not re.search(r"\bno (?:customer )?applicability\b|\bno customer AC was evaluated\b", content, re.IGNORECASE):
            errors.append(f"{relative}: missing explicit no-applicability boundary")
        if not re.search(r"\bno [^\n.]*\b(?:execution )?verdict\b", content, re.IGNORECASE):
            errors.append(f"{relative}: missing explicit no-verdict boundary")
        for line_number, line in enumerate(content.splitlines(), 1):
            for label, pattern in CURRENT_CLAIMS:
                if pattern.search(line):
                    errors.append(f"{relative}:{line_number}: contradictory current-state claim: {label}")
    if len(counts) != 8:
        errors.append(f"Gate A: canonical incident count is {len(counts)}/10, expected 8/10")

    for filename in ARCHITECTURES:
        relative = str(gate / "reference-architectures" / filename)
        content = read(root, relative, errors)
        for heading in ("Structural shape", "Facts an AC must carry", "Likely applicability facts to investigate", "Explicit UNKNOWN-capable facts", "Safety boundary", "AC evidence discipline"):
            require(content, rf"^## {re.escape(heading)}\s*$", heading + " section", relative, errors, re.MULTILINE)
        require(content, r"\bUNKNOWN\b", "explicit UNKNOWN value", relative, errors)

    for filename in ("README.md", "abstraction-fit.md", "decision.md"):
        read(root, str(gate / filename), errors)
    decision = read(root, str(gate / "decision.md"), errors)
    handoff = read(root, "HANDOFF.md", errors)
    rows = re.findall(r"^\| \[(\d+) [^]]+\]\(incidents/[^)]+\) \|[^\n]+$", decision, re.MULTILINE)
    if [int(number) for number in rows] != list(range(1, 11)):
        errors.append(f"{gate / 'decision.md'}: expected decision rows for families 1 through 10 in order")
    for number, filename in enumerate(INCIDENTS, 1):
        pattern = rf"^\| \[{number} [^]]+\]\(incidents/{re.escape(filename)}\) \| [^|]+ \| Yes \| {'No' if number in NON_COUNTING else 'Yes'} \|"
        require(decision, pattern, f"family {number} decision row (mechanism Yes; canonical {'No' if number in NON_COUNTING else 'Yes'})", str(gate / "decision.md"), errors, re.MULTILINE)
    checks = (
        (decision, str(gate / "decision.md"), ((r"^# Gate A decision: REVISE\s*$", "REVISE decision"), (r"Mechanism-fit coverage: 10/10", "10/10 mechanism fit"), (r"Canonical public-incident coverage: 8/10", "8/10 incident coverage"), (r"Phase 1: NOT AUTHORIZED", "Phase 1 NOT AUTHORIZED"), (r"families \*\*1 and 3\*\*", "unresolved families 1 and 3"))),
        (handoff, "HANDOFF.md", ((r"Current phase: \*\*Pre-Phase-1 validation \(Gate A\)\*\*", "current Gate A phase"), (r"Gate result: \*\*REVISE\*\*", "REVISE gate result"), (r"Mechanism-fit coverage is 10/10", "10/10 mechanism fit"), (r"canonical public-incident coverage is 8/10", "8/10 incident coverage"), (r"Families 1 and 3 lack", "unresolved families 1 and 3"), (r"Phase 1 is not authorized", "Phase 1 not authorized"))),
    )
    for content, relative, patterns in checks:
        for pattern, label in patterns:
            require(content, pattern, label, relative, errors, re.IGNORECASE | re.MULTILINE)
    return sorted(set(errors))


def main():
    errors = validate()
    if errors:
        for error in errors:
            print(f"ERROR: {error}", file=sys.stderr)
        return 1
    print("baseline integrity passed (byte digest only; not semantic correctness)")
    print("Gate A structural/checkpoint integrity passed (not source-evidence verification)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
