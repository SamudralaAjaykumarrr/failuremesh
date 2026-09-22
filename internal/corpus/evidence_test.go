package corpus

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/evidence"
)

// This is an independently authored Phase 2A source/claim fixture. Corpus intake
// only receives its already-published identity; it never manufactures claims.
func existingEvidence(t *testing.T, root, id string, excerpt bool) (evidence.Record, Digest, []byte) {
	t.Helper()
	body, e := os.ReadFile("../evidence/testdata/retry.json")
	if e != nil {
		t.Fatal(e)
	}
	var obs struct {
		Subject   string             `json:"subject"`
		Scope     evidence.Scope     `json:"scope"`
		Freshness evidence.Freshness `json:"freshness"`
	}
	if json.Unmarshal(body, &obs) != nil {
		t.Fatal("fixture JSON")
	}
	a := evidence.SourceArtifact{SchemaVersion: evidence.Schema, SourceID: id, ArtifactID: id + "-artifact", Revision: 1, CanonicalURL: "https://example.invalid/fixture", RetrievedURL: "https://example.invalid/fixture", Publisher: "fixture-author", PublisherIdentityBasis: "authored fixture", PublicationTimestampBasis: "unknown", RetrievedAt: now, SourceType: "synthetic_observation", QualityClass: "synthetic", QualityPolicyVersion: evidence.Policy, Quality: evidence.Quality{Attribution: "fixture-author", Directness: "structured_fixture", ScopeFit: "exact", VersionFit: "exact", Integrity: "exact_bytes", Completeness: "bounded_observer", Independence: "synthetic_not_independent"}, Retrieval: evidence.Retrieval{Method: "manual-local", Collector: "fixture-author", Version: "1", IdentityClass: "tool", Status: "supplied", ContentType: "application/json"}, Rights: rights(), Mutability: "revision-pinned", Availability: "available", CheckedAt: now, Sensitivity: "synthetic"}
	a.Rights.WholeBody = "allow"
	a, e = evidence.Capture(a, body, true)
	if e != nil {
		t.Fatal(e)
	}
	a.Locations = []evidence.Location{{ID: id + "-location", ArtifactID: a.ArtifactID, ArtifactDigest: a.ContentDigest, Selector: evidence.Selector{Kind: "byte", Representation: a.Representation, End: a.ByteLength}, Sensitivity: "synthetic", Access: "body"}}
	yes := true
	c := evidence.AtomicClaim{SchemaVersion: evidence.Schema, ID: id + "-claim", Revision: 1, Predicate: "synthetic.same_operation_retry", Value: evidence.Value{Type: "boolean", Boolean: &yes}, Subject: obs.Subject, Scope: obs.Scope, EpistemicBasis: "experiment_observed", SourceRefs: []evidence.SourceRef{{SourceID: a.SourceID, ArtifactID: a.ArtifactID, Revision: 1, Digest: a.ContentDigest}}, EvidenceRefs: []string{a.Locations[0].ID}, Extraction: evidence.Extraction{Method: "structured", Tool: "fixture-reader", Version: "1", InputDigests: []Digest{a.ContentDigest}, Transformation: "exact typed fixture", ProposerClass: "tool", Context: "synthetic-only"}, Freshness: obs.Freshness, Sensitivity: "synthetic", Boundary: "local", SupportStatus: "candidate", ContradictionStatus: "none"}
	if excerpt {
		a.Retained = false
		a.Body = nil
		a.Rights.WholeBody = "deny"
		a.Rights.Excerpt = "allow"
		a.Rights.ExcerptScope = "synthetic bounded half"
		a.Locations[0].Access = "excerpt"
		a.Locations[0].Selector.End = int64(len(body) / 2)
		a.Locations[0].Excerpt = bytes.Clone(body[:len(body)/2])
		d := Hash(a.Locations[0].Excerpt)
		a.Locations[0].ExcerptDigest = &d
	}
	r := evidence.Record{SchemaVersion: evidence.Schema, Encoding: evidence.Encoding, PolicyVersion: evidence.Policy, EvaluationTime: now, TargetClaim: c.ID, Artifacts: []evidence.SourceArtifact{a}, Claims: []evidence.AtomicClaim{c}}
	d, _, e := evidence.Save(root, r, now, body)
	if e != nil {
		t.Fatal(e)
	}
	return r, d, body
}
func capturedCandidate(h *harness, id string, r evidence.Record, d Digest) Object {
	o := candidateObject(id, h.p)
	o.Sensitivity = "synthetic"
	o.Candidate.OriginClass = "synthetic_original"
	a := r.Artifacts[0]
	o.Candidate.Sources[0].Rights = a.Rights
	o.Candidate.Captured = []Captured{{CitationID: "citation", StoreID: "evidence-fixture", Record: d, Source: evidence.SourceRef{SourceID: a.SourceID, ArtifactID: a.ArtifactID, Revision: a.Revision, Digest: a.ContentDigest}, Representation: a.Representation, Locations: []string{a.Locations[0].ID}, ParentDependency: "permitted transient bytes required for excerpt replay"}}
	return o
}
func TestPhase2AIntegration(t *testing.T) {
	for _, excerpt := range []bool{false, true} {
		t.Run(map[bool]string{false: "public body export unknown", true: "transient excerpt parent"}[excerpt], func(t *testing.T) {
			h := newHarness(t)
			root := t.TempDir()
			r, d, parent := existingEvidence(t, root, "capture", excerpt)
			h.s.EvidenceRoot = root
			h.s.EvidenceStoreID = "evidence-fixture"
			h.s.Parents = [][]byte{parent}
			o := capturedCandidate(h, "a", r, d)
			h.apply("candidate add", o)
			ar := h.assign(Assignment{"a", "DEVELOPMENT", "synthetic captured fixture"})
			f := h.freeze(ar)
			report, e := h.s.Verify(f.Object, h.head, now, true)
			if e != nil {
				t.Fatal(e)
			}
			if excerpt && report.AuditCapability != "CONDITIONAL_AUDIT" {
				t.Fatal("excerpt overclaim")
			}
			if !excerpt && report.AuditCapability != "LOCAL_AUDIT_AVAILABLE" {
				t.Fatal("local audit unavailable")
			}
			// No content bytes or their base64 encoding leak into corpus records.
			encoded, _ := json.Marshal(parent)
			filepath.WalkDir(h.s.Root, func(p string, en os.DirEntry, e error) error {
				if e == nil && !en.IsDir() {
					b, _ := os.ReadFile(p)
					if bytes.Contains(b, parent) || bytes.Contains(b, encoded) {
						t.Fatal("source body copied")
					}
				}
				return e
			})
			if excerpt {
				h.s.Parents = nil
			} else {
				if os.Remove(filepath.Join(root, "sha256-"+d.Hex+".json")) != nil {
					t.Fatal("remove capture")
				}
			}
			current, e := h.s.Verify(f.Object, h.head, now, true)
			if e != nil || current.AuditCapability != "CONDITIONAL_AUDIT" {
				t.Fatal("lost source overclaimed", current, e)
			}
			historical, e := h.s.Verify(f.Object, &f.Head, now, false)
			if e != nil || historical.Integrity != "INTACT" {
				t.Fatal("historical destroyed", e)
			}
			o = capturedCandidate(h, "new", r, d)
			_, e = h.s.Apply(h.request("candidate add", o))
			if e == nil {
				t.Fatal("missing capture/parent accepted")
			}
		})
	}
	t.Run("exact captured identities", func(t *testing.T) {
		h := newHarness(t)
		root := t.TempDir()
		r, d, parent := existingEvidence(t, root, "capture", false)
		h.s.EvidenceRoot = root
		h.s.EvidenceStoreID = "evidence-fixture"
		h.s.Parents = [][]byte{parent}
		for _, attack := range []string{"source", "artifact", "representation", "location", "rights"} {
			o := capturedCandidate(h, attack, r, d)
			switch attack {
			case "source":
				o.Candidate.Captured[0].Source.SourceID = "wrong"
			case "artifact":
				o.Candidate.Captured[0].Source.ArtifactID = "wrong"
			case "representation":
				o.Candidate.Captured[0].Representation = "wrong"
			case "location":
				o.Candidate.Captured[0].Locations = []string{"missing"}
			case "rights":
				o.Candidate.Sources[0].Rights.WholeBody = "deny"
			}
			if _, e := h.s.Apply(h.request("candidate add", o)); e == nil {
				t.Fatal("identity/rights bypass", attack)
			}
		}
	})
	t.Run("complementary captures and forged parent range", func(t *testing.T) {
		h := newHarness(t)
		root := t.TempDir()
		r, d, parent := existingEvidence(t, root, "capture", true)
		h.s.EvidenceRoot = root
		h.s.EvidenceStoreID = "evidence-fixture"
		h.s.Parents = [][]byte{parent}
		h.apply("candidate add", capturedCandidate(h, "a", r, d))
		// Simulate a hand-written malicious Phase 2A binding, never a valid Save.
		a := r.Artifacts[0]
		a.SourceID = "other"
		a.ArtifactID = "other-artifact"
		a.Locations[0].ID = "other-location"
		a.Locations[0].ArtifactID = a.ArtifactID
		a.Locations[0].Selector.Start = int64(len(parent) / 2)
		a.Locations[0].Selector.End = int64(len(parent))
		a.Locations[0].Excerpt = bytes.Clone(parent[len(parent)/2:])
		ed := Hash(a.Locations[0].Excerpt)
		a.Locations[0].ExcerptDigest = &ed
		writeBindings := func() {
			b, _ := json.Marshal(a)
			key, _ := json.Marshal([]string{a.ArtifactID})
			os.WriteFile(filepath.Join(root, "artifact-"+Hash(key).Hex+".json"), b, 0600)
			rev, _ := json.Marshal(struct {
				Source   string
				Revision int
			}{a.SourceID, a.Revision})
			os.WriteFile(filepath.Join(root, "source-revision-"+Hash(rev).Hex+".json"), b, 0600)
		}
		writeBindings()
		_, e := h.s.Apply(h.request("candidate add", capturedCandidate(h, "second", r, d)))
		reason(t, e, "PHASE2A_STORE_COVERAGE_REFUSED")
		a.Locations[0].Selector.Start = 0
		a.Locations[0].Selector.End = int64(len(a.Locations[0].Excerpt))
		writeBindings()
		_, e = h.s.Apply(h.request("candidate add", capturedCandidate(h, "third", r, d)))
		reason(t, e, "PHASE2A_STORE_COVERAGE_REFUSED")
	})
}
