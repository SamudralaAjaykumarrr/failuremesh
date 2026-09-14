#!/usr/bin/env python3
"""Check the recorded Gate A checkpoint, not the truth of its source evidence."""

import hashlib
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
BASELINE_SHA256 = "f89ad28e6b9d63f1d08fedb33b91b8de894fc9d248e6096bacae92d909e9711b"
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
    disposition = read(root, str(gate / "disposition.md"), errors)
    phase = read(root, "docs/phases/phase-1-first-vertical-slice.md", errors)
    change = read(root, "docs/company/baseline-change-001-bounded-phase1-entry.md", errors)
    second_change = read(root, "docs/company/baseline-change-002-bounded-pfc2-entry.md", errors)
    for pattern, label in ((r"^# PROPOSED BASELINE CHANGE$", "second baseline change label"), (r"APPROVED on 2026-09-14", "second explicit approval"), (r"Gate A remains \*\*REVISE\*\*", "second change preserves REVISE"), (r"qualifying public observed incidents 8/10", "second change preserves incident count"), (r"families 1 and 3 unresolved", "second change preserves unresolved families"), (r"Broader Phase 1, other PFCs, customer execution, real queue integrations, uncontrolled chaos, and production execution are \*\*NOT AUTHORIZED\*\*", "second bounded authorization")):
        require(second_change, pattern, label, "docs/company/baseline-change-002-bounded-pfc2-entry.md", errors, re.IGNORECASE | re.MULTILINE)
    third_change = read(root, "docs/company/baseline-change-003-bounded-pfc3-entry.md", errors)
    for pattern, label in ((r"^# PROPOSED BASELINE CHANGE$", "third baseline change label"), (r"APPROVED on 2026-09-14", "third explicit approval"), (r"Gate A remains \*\*REVISE\*\*", "third change preserves REVISE"), (r"qualifying public observed incidents 8/10", "third change preserves incident count"), (r"families 1 and 3 unresolved", "third change preserves unresolved families"), (r"PFC #4\+, broader Phase 1, customer execution, production execution, real lease-system integration, and family-3 Gate A promotion are \*\*NOT AUTHORIZED\*\*", "third bounded authorization")):
        require(third_change, pattern, label, "docs/company/baseline-change-003-bounded-pfc3-entry.md", errors, re.IGNORECASE | re.MULTILINE)
    rows = re.findall(r"^\| \[(\d+) [^]]+\]\(incidents/[^)]+\) \|[^\n]+$", decision, re.MULTILINE)
    if [int(number) for number in rows] != list(range(1, 11)):
        errors.append(f"{gate / 'decision.md'}: expected decision rows for families 1 through 10 in order")
    for number, filename in enumerate(INCIDENTS, 1):
        pattern = rf"^\| \[{number} [^]]+\]\(incidents/{re.escape(filename)}\) \| [^|]+ \| Yes \| {'No' if number in NON_COUNTING else 'Yes'} \|"
        require(decision, pattern, f"family {number} decision row (mechanism Yes; canonical {'No' if number in NON_COUNTING else 'Yes'})", str(gate / "decision.md"), errors, re.MULTILINE)
    checks = (
        (decision, str(gate / "decision.md"), ((r"^# Gate A decision: REVISE\s*$", "REVISE decision"), (r"Mechanism-fit coverage: 10/10", "10/10 mechanism fit"), (r"Canonical public-incident coverage: 8/10", "8/10 incident coverage"), (r"Phase 1: NOT AUTHORIZED", "Phase 1 NOT AUTHORIZED"), (r"families \*\*1 and 3\*\*", "unresolved families 1 and 3"))),
        (handoff, "HANDOFF.md", ((r"Gate result: \*\*REVISE\*\*", "REVISE gate result"), (r"Mechanism-fit coverage is 10/10", "10/10 mechanism fit"), (r"canonical public-incident coverage is 8/10", "8/10 incident coverage"), (r"Families 1 and 3 lack", "unresolved families 1 and 3"), (r"authorizes \*\*only\*\* the \[Phase 1 first vertical slice\]", "bounded prototype authorization"), (r"Broader Phase 1, customer execution, and production execution remain \*\*NOT AUTHORIZED\*\*", "broader execution prohibited"), (r"ten-incident milestone is incomplete", "incomplete incident milestone"), (r"Local synthetic applicability was APPLICABLE for both reference configurations; the vulnerable PostgreSQL run earned EXPOSED and the remediated bounded rerun earned scoped PROVEN_RESILIENT. No customer verdict has been earned", "scoped synthetic results"))),
        (disposition, str(gate / "disposition.md"), ((r"Gate A remains REVISE", "current REVISE"), (r"Mechanism fit remains 10/10", "current mechanism fit"), (r"coverage remains 8/10", "current incident count"), (r"Families 1 and 3 remain unresolved", "unresolved families"), (r"Bounded Phase 1 prototype: AUTHORIZED", "bounded authorization"), (r"Broader Phase 1, customer execution, and production execution: NOT AUTHORIZED", "unscoped execution prohibited"), (r"ten-incident validation milestone remains incomplete", "incomplete milestone"))),
        (phase, "docs/phases/phase-1-first-vertical-slice.md", ((r"Authorization status: \*\*AUTHORIZED — BOUNDED PROTOTYPE ONLY\*\*", "bounded authorization status"), (r"Gate A remains REVISE", "REVISE status"), (r"8/10 qualifying public observed incidents", "incident count"), (r"families 1 and 3 remain unresolved", "unresolved families"), (r"ten-incident validation milestone is incomplete", "incomplete milestone"), (r"No prototype result by itself establishes", "claims boundary"))),
        (change, "docs/company/baseline-change-001-bounded-phase1-entry.md", ((r"^# PROPOSED BASELINE CHANGE$", "baseline change label"), (r"APPROVED on 2026-09-14", "explicit approval"), (r"BOUNDED PHASE 1 PROTOTYPE: AUTHORIZED", "authorization"), (r"Gate A remains REVISE", "REVISE preserved"), (r"coverage remains 8/10", "incident count preserved"), (r"families 1 and 3 remain unresolved", "unresolved families"))),
    )
    for content, relative, patterns in checks:
        for pattern, label in patterns:
            require(content, pattern, label, relative, errors, re.IGNORECASE | re.MULTILINE)
    current = "\n".join((handoff, disposition, phase, change, second_change, third_change))
    forbidden = (
        (r"\bGate A\s*(?::|remains|is|=)\s*GO\b", "false Gate A GO"),
        (r"\b(?:qualifying|canonical)\s+(?:public\s+)?(?:observed\s+)?incidents?\s*(?::|is|are|remains|coverage is|coverage remains)\s*10/10\b", "false incident coverage"),
        (r"\b(?:ten[- ](?:public[- ]observed[- ])?incident|incident-validation) milestone (?:is|remains) complete\b", "false milestone completion"),
        (r"\b(?:all|unbounded|unscoped|broader) Phase 1 (?:is|:|remains) AUTHORIZED\b", "unscoped Phase 1 authorization"),
        (r"\b(?:customer|production) execution (?:is|:|remains) AUTHORIZED\b", "customer or production authorization"),
    )
    for pattern, label in forbidden:
        if re.search(pattern, current, re.IGNORECASE):
            errors.append(f"current authorization: contradictory claim: {label}")
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
