"""Isolated regression tests for the repository-only checks."""

import shutil
import sys
import tempfile
import unittest
from pathlib import Path


REPOSITORY = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(REPOSITORY / "scripts"))
import check_markdown_links  # noqa: E402
import validate_gate_a  # noqa: E402


class GateATests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        for relative in ("docs/company/canonical-baseline-v1.0.md", "docs/company/baseline-change-001-bounded-phase1-entry.md", "docs/company/baseline-change-002-bounded-pfc2-entry.md", "docs/phases/phase-1-first-vertical-slice.md", "HANDOFF.md"):
            destination = self.root / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(REPOSITORY / relative, destination)
        shutil.copytree(REPOSITORY / "docs/validation/gate-a", self.root / "docs/validation/gate-a")

    def test_current_checkpoint_passes(self):
        self.assertEqual([], validate_gate_a.validate(self.root))

    def test_second_authorization_cannot_be_unbounded(self):
        self.replace_in("docs/company/baseline-change-002-bounded-pfc2-entry.md", "Broader Phase 1, other PFCs, customer execution, real queue integrations, uncontrolled chaos, and production execution are **NOT AUTHORIZED**", "Broader Phase 1 is AUTHORIZED")
        self.assertTrue(any("second bounded authorization" in error or "unscoped Phase 1 authorization" in error for error in validate_gate_a.validate(self.root)))

    def replace_in(self, relative, old, new):
        path = self.root / relative
        original = path.read_text()
        self.assertIn(old, original)
        path.write_text(original.replace(old, new))

    def test_historical_revise_cannot_be_changed_to_go(self):
        self.replace_in("docs/validation/gate-a/decision.md", "# Gate A decision: REVISE", "# Gate A decision: GO")
        self.assertTrue(any("REVISE decision" in error for error in validate_gate_a.validate(self.root)))

    def test_ten_of_ten_claim_without_records_fails(self):
        self.replace_in("HANDOFF.md", "canonical public-incident coverage is 8/10", "canonical public-incident coverage is 10/10")
        self.assertTrue(any("8/10 incident coverage" in error for error in validate_gate_a.validate(self.root)))

    def test_unresolved_families_cannot_be_removed(self):
        self.replace_in("docs/validation/gate-a/disposition.md", "Families 1 and 3 remain unresolved", "All families resolved")
        self.assertTrue(any("unresolved families" in error for error in validate_gate_a.validate(self.root)))

    def test_bounded_authorization_cannot_be_removed(self):
        self.replace_in("docs/phases/phase-1-first-vertical-slice.md", "AUTHORIZED — BOUNDED PROTOTYPE ONLY", "NOT AUTHORIZED")
        self.assertTrue(any("bounded authorization status" in error for error in validate_gate_a.validate(self.root)))

    def test_unscoped_authorization_fails(self):
        path = self.root / "HANDOFF.md"
        path.write_text(path.read_text() + "\nAll Phase 1 is AUTHORIZED. Production execution is AUTHORIZED.\n")
        errors = validate_gate_a.validate(self.root)
        self.assertTrue(any("unscoped Phase 1 authorization" in error for error in errors))
        self.assertTrue(any("customer or production authorization" in error for error in errors))

    def test_false_milestone_completion_fails(self):
        path = self.root / "HANDOFF.md"
        path.write_text(path.read_text() + "\nThe ten-incident milestone is complete.\n")
        self.assertTrue(any("false milestone completion" in error for error in validate_gate_a.validate(self.root)))

    def test_authorization_does_not_imply_reproduction_or_verdict(self):
        self.assertEqual([], validate_gate_a.validate(self.root))
        self.replace_in("HANDOFF.md", "Local synthetic applicability was APPLICABLE for both reference configurations; the vulnerable PostgreSQL run earned EXPOSED and the remediated bounded rerun earned scoped PROVEN_RESILIENT. No customer verdict has been earned", "FailureMesh issued an unscoped customer verdict")
        self.assertTrue(any("scoped synthetic results" in error for error in validate_gate_a.validate(self.root)))

    def test_missing_incident_fails(self):
        (self.root / "docs/validation/gate-a/incidents/01-timeout-after-external-commit.md").unlink()
        self.assertTrue(any("01-timeout-after-external-commit.md: missing" in error for error in validate_gate_a.validate(self.root)))

    def test_wrong_canonical_count_fails(self):
        path = self.root / "docs/validation/gate-a/incidents/03-stale-worker-after-lease-expiry.md"
        path.write_text(path.read_text().replace("milestone:** NO", "milestone:** YES"))
        errors = validate_gate_a.validate(self.root)
        self.assertTrue(any("must be NO" in error for error in errors))
        self.assertTrue(any("count is 9/10" in error for error in errors))

    def test_l1_promotion_fails(self):
        path = self.root / "docs/validation/gate-a/incidents/04-retry-amplification.md"
        path.write_text(path.read_text().replace("**Verification level:** L0 CANDIDATE", "**Verification level:** L1 REPRODUCED"))
        self.assertTrue(any("must remain L0 CANDIDATE" in error for error in validate_gate_a.validate(self.root)))

    def append_claim(self, claim):
        path = self.root / "docs/validation/gate-a/incidents/04-retry-amplification.md"
        path.write_text(path.read_text() + "\n" + claim + "\n")

    def test_positive_failuremesh_reproduction_claim_fails(self):
        self.append_claim("FailureMesh reproduced this incident. This record is L1 REPRODUCED.")
        errors = validate_gate_a.validate(self.root)
        self.assertTrue(any("contradictory current-state claim: FailureMesh reproduction" in error for error in errors))
        self.assertTrue(any("contradictory current-state claim: L1 promotion" in error for error in errors))

    def test_current_applicability_verdict_fails(self):
        self.append_claim("Customer applicability was determined APPLICABLE for this record.")
        self.assertTrue(any("contradictory current-state claim: customer applicability decision" in error for error in validate_gate_a.validate(self.root)))

    def test_current_execution_verdict_fails(self):
        self.append_claim("FailureMesh issued an execution verdict of PROVEN_RESILIENT.")
        self.assertTrue(any("contradictory current-state claim: FailureMesh execution verdict" in error for error in validate_gate_a.validate(self.root)))

    def test_direct_execution_verdict_claim_fails(self):
        self.append_claim("FailureMesh issued EXPOSED.")
        self.assertTrue(any("contradictory current-state claim: FailureMesh execution verdict" in error for error in validate_gate_a.validate(self.root)))

    def test_negative_explanatory_and_future_wording_passes(self):
        self.append_claim("No FailureMesh reproduction occurred. This does not promote the record to L1 REPRODUCED. EXPOSED is a future execution verdict. APPLICABLE is not established. Phase 1 may test this later.")
        self.assertEqual([], validate_gate_a.validate(self.root))

    def test_decision_row_drift_fails(self):
        path = self.root / "docs/validation/gate-a/decision.md"
        path.write_text(path.read_text().replace("| Yes | No | WordPress", "| Yes | Yes | WordPress"))
        self.assertTrue(any("family 1 decision row" in error for error in validate_gate_a.validate(self.root)))

    def test_unexpected_incident_fails(self):
        path = self.root / "docs/validation/gate-a/incidents/11-extra.md"
        path.write_text("# Extra\n")
        self.assertTrue(any("unexpected Gate A record" in error for error in validate_gate_a.validate(self.root)))

    def test_baseline_digest_mismatch_fails(self):
        path = self.root / "docs/company/canonical-baseline-v1.0.md"
        path.write_bytes(path.read_bytes() + b"\n")
        self.assertTrue(any("SHA-256 mismatch" in error for error in validate_gate_a.validate(self.root)))


class MarkdownTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)

    def test_broken_local_link_fails(self):
        (self.root / "README.md").write_text("[missing](docs/missing.md)\n")
        self.assertEqual(["README.md:1: broken local link: docs/missing.md"], check_markdown_links.validate(self.root))

    def test_directory_fragment_external_and_fence(self):
        (self.root / "docs").mkdir()
        (self.root / "README.md").write_text(
            "[docs](docs/) [section](docs/page.md#section) [web](https://example.org/no)\n"
            "`[inline](missing.md)` ![image](docs/page.md)\n"
            "```md\n[ignored](missing.md)\n```\n"
        )
        (self.root / "docs/page.md").write_text("# Section\n")
        self.assertEqual([], check_markdown_links.validate(self.root))


if __name__ == "__main__":
    unittest.main()
