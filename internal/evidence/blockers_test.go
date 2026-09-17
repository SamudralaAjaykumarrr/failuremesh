package evidence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// Each capture retains only its range of the same exact synthetic content.
// Claims deliberately have no evidence: these tests concern retention, not truth.
func excerptCapture(t *testing.T, start, end int64, revision int, source, whole string) Record {
	t.Helper()
	r := fixture(t, false)
	a := &r.Artifacts[0]
	a.SourceID = source
	a.ArtifactID = fmt.Sprintf("capture.%s.%d", source, revision)
	a.Revision = revision
	raw := bytes.Clone(a.Body)
	l := &a.Locations[0]
	l.ID = a.ArtifactID + ".excerpt"
	l.ArtifactID = a.ArtifactID
	l.Selector.Start = start
	l.Selector.End = end
	l.Access = "excerpt"
	l.Excerpt = raw[start:end]
	d := Hash(l.Excerpt)
	l.ExcerptDigest = &d
	a.Retained = false
	a.Body = nil
	a.Rights.WholeBody = whole
	a.Rights.Excerpt = "allow"
	a.Rights.ExcerptScope = "explicit synthetic byte range"
	a.Rights.Export = "allow"
	c := &r.Claims[0]
	c.ID = a.ArtifactID + ".claim"
	c.SourceRefs = nil
	c.EvidenceRefs = nil
	c.Extraction.InputDigests = nil
	r.TargetClaim = c.ID
	return r
}
func composedRecords(rs ...Record) Record {
	out := rs[0]
	out.Artifacts = nil
	out.Claims = nil
	for _, r := range rs {
		out.Artifacts = append(out.Artifacts, r.Artifacts...)
		out.Claims = append(out.Claims, r.Claims...)
	}
	return out
}
func exportRecord(r Record, parents ...[]byte) ([]byte, error) {
	// Compute the approval independently of Encode so a rejected encoding cannot
	// accidentally turn this into a test of invalid approval instead of rights.
	b, _ := json.Marshal(r)
	return Export(r, ExportApproval{Hash(b), "synthetic-reviewer", true}, fixtureTime, parents...)
}

func TestB1OriginalExploit(t *testing.T) {
	parent := fixture(t, false).Artifacts[0].Body
	n := fixture(t, false).Artifacts[0].ByteLength
	first := excerptCapture(t, 0, n/2, 1, "source.same", "deny")
	second := excerptCapture(t, n/2, n, 2, "source.same", "deny")
	for _, part := range []Record{first, second} {
		if err := Validate(part, parent); err != nil {
			t.Fatal("individually valid excerpt rejected", err)
		}
	}
	combined := composedRecords(first, second)
	if err := Validate(combined, parent); err == nil || err.Error() != "cumulative_whole_body_permission_required" {
		t.Error("B1: Validate accepted complementary revisions with whole-body denied")
	}
	if _, _, err := Save(t.TempDir(), combined, fixtureTime, parent); err == nil || err.Error() != "cumulative_whole_body_permission_required" {
		t.Error("B1: Save accepted complementary revisions with whole-body denied")
	}
	if b, err := exportRecord(combined, parent); err == nil || err.Error() != "cumulative_whole_body_permission_required" || b != nil {
		t.Error("B1: Export released complementary revisions with whole-body denied")
	}
	root := t.TempDir()
	if _, _, err := Save(root, first, fixtureTime, parent); err != nil {
		t.Fatal("bounded first excerpt", err)
	}
	if _, _, err := Save(root, second, fixtureTime, parent); err == nil || err.Error() != "cumulative_whole_body_permission_required" {
		t.Error("B1: successive Save reconstructed whole body")
	}
}

func TestB2OriginalExploit(t *testing.T) {
	r := fixture(t, false)
	text := "true"
	r.Claims[0].Value = Value{Type: "string", Text: &text}
	addChallenge(&r, fixture(t, false))
	got := Assess(r)
	if got.State != Unsupported || got.State.Projection() != "UNKNOWN" || !slices.Contains(got.Reasons, "predicate_requires_boolean_value") {
		t.Fatalf("B2: ineligible target inherited authority: state=%s projection=%s reasons=%v decisions=%+v", got.State, got.State.Projection(), got.Reasons, got.Decisions)
	}
}

func TestB1CoverageComposition(t *testing.T) {
	parent := fixture(t, false).Artifacts[0].Body
	n := fixture(t, false).Artifacts[0].ByteLength
	tests := []struct {
		name    string
		ranges  [][2]int64
		whole   []string
		allowed bool
	}{
		{"aliases complementary halves", [][2]int64{{0, n / 2}, {n / 2, n}}, []string{"deny", "deny"}, false},
		{"unknown is not whole permission", [][2]int64{{0, n / 2}, {n / 2, n}}, []string{"unknown", "unknown"}, false},
		{"out of order thirds", [][2]int64{{2 * n / 3, n}, {0, n / 3}, {n / 3, 2 * n / 3}}, []string{"deny", "deny", "deny"}, false},
		{"overlap is not unique coverage", [][2]int64{{0, 3 * n / 4}, {n / 4, 3 * n / 4}}, []string{"deny", "deny"}, true},
		{"duplicates are not unique coverage", [][2]int64{{0, 3 * n / 4}, {0, 3 * n / 4}}, []string{"deny", "deny"}, true},
		{"one byte gap remains", [][2]int64{{0, n/2 - 1}, {n / 2, n}}, []string{"deny", "deny"}, true},
		{"overlapping full coverage", [][2]int64{{n / 3, n}, {0, 2 * n / 3}}, []string{"deny", "deny"}, false},
		{"all contributors explicitly permitted", [][2]int64{{n / 2, n}, {0, n / 2}}, []string{"allow", "allow"}, true},
		{"allow cannot launder denied contributor", [][2]int64{{0, n / 2}, {n / 2, n}}, []string{"deny", "allow"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var records []Record
			for i, span := range tt.ranges {
				records = append(records, excerptCapture(t, span[0], span[1], 1, fmt.Sprintf("alias.%d", i), tt.whole[i]))
			}
			r := composedRecords(records...)
			err := Validate(r, parent)
			if (err == nil) != tt.allowed {
				t.Fatalf("Validate: %v, allowed=%v", err, tt.allowed)
			}
			_, _, err = Save(t.TempDir(), r, fixtureTime, parent)
			if (err == nil) != tt.allowed {
				t.Fatalf("Save: %v, allowed=%v", err, tt.allowed)
			}
			b, err := exportRecord(r, parent)
			if (err == nil) != tt.allowed || (!tt.allowed && b != nil) {
				t.Fatalf("Export: %v, allowed=%v", err, tt.allowed)
			}
			// Each candidate is independently permitted; the store must account for
			// already persisted contributors across every successive Save.
			root := t.TempDir()
			for i, part := range records {
				if err := Validate(part, parent); err != nil {
					t.Fatal("partial capture invalid", err)
				}
				before, err := os.ReadDir(root)
				if err != nil {
					t.Fatal(err)
				}
				_, _, err = Save(root, part, fixtureTime, parent)
				allowed := tt.allowed || i < len(records)-1
				if (err == nil) != allowed {
					t.Fatalf("successive Save %d: %v, allowed=%v", i, err, allowed)
				}
				if !allowed {
					after, err := os.ReadDir(root)
					if err != nil {
						t.Fatal(err)
					}
					if len(after) != len(before) {
						t.Fatal("rejected save retained new material")
					}
				}
			}
		})
	}
}

func TestB1ExportStillRequiresEveryPermission(t *testing.T) {
	parent := fixture(t, false).Artifacts[0].Body
	n := fixture(t, false).Artifacts[0].ByteLength
	for _, permission := range []string{"deny", "unknown"} {
		r := composedRecords(excerptCapture(t, 0, n/2, 1, "alias.a", "allow"), excerptCapture(t, n/2, n, 1, "alias.b", "allow"))
		r.Artifacts[1].Rights.Export = permission
		if err := Validate(r, parent); err != nil {
			t.Fatal(err)
		}
		if b, err := exportRecord(r, parent); err == nil || b != nil {
			t.Fatal("whole-body permission bypassed export policy")
		}
	}
	// Metadata-only permission does not authorize retained text from denied captures.
	r := composedRecords(excerptCapture(t, 0, n/2, 1, "alias.a", "deny"), excerptCapture(t, n/2, n, 1, "alias.b", "deny"))
	meta := fixture(t, false).Artifacts[0]
	meta.ArtifactID = "metadata-only"
	meta.SourceID = "metadata-only"
	meta.Retained = false
	meta.Body = nil
	meta.Locations = nil
	r.Artifacts = append(r.Artifacts, meta)
	if Validate(r, parent) == nil {
		t.Fatal("metadata-only allow granted cumulative permission")
	}
}

func TestB1InvalidCoverageFailsClosed(t *testing.T) {
	parent := fixture(t, false).Artifacts[0].Body
	n := fixture(t, false).Artifacts[0].ByteLength
	for _, span := range [][2]int64{{0, 0}, {2, 1}, {-1, 2}, {0, n + 1}} {
		r := excerptCapture(t, 0, n/2, 1, "range", "deny")
		r.Artifacts[0].Locations[0].Selector.Start = span[0]
		r.Artifacts[0].Locations[0].Selector.End = span[1]
		if Validate(r, parent) == nil {
			t.Fatal("invalid range accepted", span)
		}
	}
	r := composedRecords(excerptCapture(t, 0, n/2, 1, "alias.a", "deny"), excerptCapture(t, n/2, n, 1, "alias.b", "deny"))
	r.Artifacts[1].ByteLength++
	if Validate(r, parent) == nil {
		t.Fatal("same digest with conflicting length accepted")
	}
}

func TestB1ConcurrentSavesAndPersistedBoundary(t *testing.T) {
	parent := fixture(t, false).Artifacts[0].Body
	n := fixture(t, false).Artifacts[0].ByteLength
	first := excerptCapture(t, 0, n/2, 1, "alias.a", "deny")
	second := excerptCapture(t, n/2, n, 1, "alias.b", "deny")
	root := t.TempDir()
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, r := range []Record{first, second} {
		go func(r Record) { <-start; _, _, err := Save(root, r, fixtureTime, parent); results <- err }(r)
	}
	close(start)
	passed := 0
	for i := 0; i < 2; i++ {
		if <-results == nil {
			passed++
		}
	}
	if passed != 1 {
		t.Fatalf("concurrent complementary saves: %d passed", passed)
	}
	// A failed save may have already published an artifact binding. It must still
	// count, even if there is no completed source record or assessment object.
	root = t.TempDir()
	a := first.Artifacts[0]
	raw, _ := json.Marshal(a)
	if err := publish(root, artifactBindingName(a.ArtifactID), raw); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Save(root, second, fixtureTime, parent); err == nil {
		t.Fatal("orphaned binding ignored")
	}
	for _, name := range []string{artifactBindingName(a.ArtifactID), ".evidence-interrupted"} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, name), []byte("unreadable publication"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := Save(root, second, fixtureTime, parent); err == nil {
			t.Fatal("unreadable persisted state ignored", name)
		}
	}
}

func TestB2TargetEligibilityAndRelatedClaims(t *testing.T) {
	for _, proposer := range []string{"tool", "human", "ai"} {
		for _, contradictory := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/conflict=%v", proposer, contradictory), func(t *testing.T) {
				r := fixture(t, false)
				text := "true"
				r.Claims[0].Value = Value{Type: "string", Text: &text}
				r.Claims[0].Extraction.ProposerClass = proposer
				if proposer == "ai" {
					ai(&r.Claims[0])
				}
				r.Claims[0].AdvisoryConfidence = &Confidence{1, "synthetic advisory confidence"}
				for i := 0; i < 4; i++ {
					addChallenge(&r, fixture(t, false))
				}
				if contradictory {
					addChallenge(&r, fixture(t, true))
				}
				got := state(t, r, Unsupported)
				if !slices.Contains(got.Reasons, "predicate_requires_boolean_value") {
					t.Fatal("lost target rejection", got)
				}
				// An independently selected eligible target can still use related support,
				// and contradictory admissible evidence still blocks that eligible target.
				r.TargetClaim = r.Claims[1].ID
				want := SupportedTrue
				if contradictory {
					want = Conflicting
				}
				state(t, r, want)
			})
		}
	}
	r := fixture(t, false)
	r.Claims[0].Predicate = "unsupported.predicate"
	addChallenge(&r, fixture(t, false))
	got := state(t, r, Unsupported)
	if !slices.Contains(got.Reasons, "unsupported_predicate") {
		t.Fatal("lost unsupported predicate reason")
	}
	for _, negative := range []bool{false, true} {
		r := fixture(t, negative)
		r.Claims[0].EvidenceRefs = nil
		addChallenge(&r, fixture(t, negative))
		want := SupportedTrue
		if negative {
			want = SupportedFalse
		}
		state(t, r, want)
	}
	r = fixture(t, false)
	r.Claims[0].Value.Boolean = nil
	addChallenge(&r, fixture(t, false))
	state(t, r, Unsupported)
}
