package corpus

import (
	"bytes"
	"encoding/json"
	"github.com/SamudralaAjaykumarrr/failuremesh/internal/evidence"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func negativeDeclarations(h *harness, id string) {
	for _, c := range categories {
		h.expose(id, c, "NO_WITH_BASIS")
	}
}
func assignmentObject(h *harness, rows ...Assignment) Object {
	o := envelope("assignment", "review-assignment", h.p.Digest)
	o.Assignments = &rows
	return o
}
func freezeObject(h *harness, a Ref) Object {
	o := envelope("manifest", "review-manifest", h.p.Digest)
	o.Schema = ManifestSchema
	o.Manifest = &Manifest{Policy: h.p, Assignment: a, CheckTime: now, Build: Hash([]byte("review-build")), SourceCommit: "synthetic", Method: "explicit-table-v1", Seed: "NOT_USED", Chronology: "SELF_REPORTED", Pretraining: "UNKNOWN"}
	return o
}
func TestReviewB1(t *testing.T) {
	for _, mode := range []string{"decision actor", "late merge"} {
		t.Run(mode, func(t *testing.T) {
			h := newHarness(t)
			o := candidateObject("a", h.p)
			if mode == "decision actor" {
				o.Candidate.Decision.Actor = "alternate-selector"
			}
			r := h.apply("candidate add", o)
			a := Ref{"a", 1, r.Object}
			negativeDeclarations(h, "a")
			rows := []Assignment{{"a", "HOLDOUT", "review"}}
			if mode == "late merge" {
				b := h.add("b")
				h.relations(relation(b, a, "SHARED_PROJECT_LINEAGE", "ESTABLISHED"))
				rows = append(rows, Assignment{"b", "HOLDOUT", "review"})
			}
			_, e := h.s.Apply(h.request("assign-split", assignmentObject(h, rows...)))
			if e == nil {
				t.Fatal("review attack accepted: missing declarations erased")
			}
			reason(t, e, "FRESH_SPLIT_REFUSED")
		})
	}
}
func TestReviewB2(t *testing.T) {
	h := newHarness(t)
	root := t.TempDir()
	r, d, _ := existingEvidence(t, root, "review-capture", false)
	h.s.EvidenceRoot = root
	h.s.EvidenceStoreID = "evidence-fixture"
	o := capturedCandidate(h, "a", r, d)
	o.Sensitivity = "public"
	o.Candidate.OriginClass = "natural_incident"
	b, _ := json.Marshal(h.request("candidate add", o))
	q, e := DecodeRequest(b)
	if e != nil {
		t.Fatal(e)
	}
	_, e = h.s.Apply(q)
	if e == nil {
		t.Fatal("review attack accepted: synthetic capture relabeled natural")
	}
	reason(t, e, "CAPTURE_CLASSIFICATION_MISMATCH")
}
func TestReviewB3(t *testing.T) {
	for _, mode := range []string{"classification", "cutoff", "duplicate", "parent", "relation", "holdout", "stale"} {
		t.Run(mode, func(t *testing.T) {
			h := newHarness(t)
			a := h.add("a")
			o := candidateObject("b", h.p)
			action := "candidate add"
			want := ""
			switch mode {
			case "classification":
				o.Candidate.OriginClass = "invalid"
				want = "INVALID_CLASSIFICATION"
			case "cutoff":
				o.ReportedAt = "2026-09-19T00:00:00Z"
				want = "OUTSIDE_POLICY_WINDOW"
			case "duplicate":
				o.ID = "a"
				want = "REVISION_CONFLICT"
			case "parent":
				o.ID = "a"
				o.Revision = 2
				d := Hash([]byte("wrong"))
				o.Parent = &d
				action = "candidate revise"
				want = "REVISION_CONFLICT"
			case "relation":
				o = envelope("cluster", "bad-relation", h.p.Digest)
				rs := []Relation{relation(a, Ref{"missing", 1, Hash([]byte("missing"))}, "SAME_INCIDENT", "ESTABLISHED")}
				o.Relations = &rs
				action = "cluster"
				want = "DANGLING_RELATION"
			case "holdout":
				o = assignmentObject(h, Assignment{"a", "HOLDOUT", "review"})
				action = "assign-split"
				want = "FRESH_SPLIT_REFUSED"
			case "stale":
				want = "STALE_HEAD"
			}
			q := h.request(action, o)
			if mode == "stale" {
				q.ExpectedHead = nil
			}
			r, e := h.s.Apply(q)
			reason(t, e, want)
			if r.Sequence == 0 {
				t.Fatal("review attack: refusal missing durable attempt receipt")
			}
			h.head = &r.Head
			st, e := h.s.scan()
			if e != nil {
				t.Fatal(e)
			}
			if len(inventory(st)) != 2 {
				t.Fatal("missing attempt inventory")
			}
			if st.heads["a"] != a {
				t.Fatal("refusal changed candidate")
			}
			again, e := h.s.Apply(q)
			reason(t, e, want)
			if again != r {
				t.Fatal("retry changed receipt")
			}
			ar := h.assign(Assignment{"a", "DEVELOPMENT", "review"})
			f := h.freeze(ar)
			report, e := h.s.Verify(f.Object, h.head, now, true)
			if e != nil || count(report.Historical, "failed_attempts") != 1 {
				t.Fatal("freeze omitted required refusal", e)
			}
		})
	}
}
func TestReviewB4(t *testing.T) {
	t.Run("stale revision", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		negativeDeclarations(h, "a")
		ar := h.assign(Assignment{"a", "HOLDOUT", "review"})
		h.revise(a, func(o *Object) {
			o.Candidate.Family.Primary = known(families[1])
			o.Candidate.Sources[0].Locator = "https://example.invalid/revised"
		})
		o, e := h.s.PrepareFreeze(freezeObject(h, ar))
		if e == nil {
			_, e = h.s.Apply(h.request("freeze", o))
		}
		if e == nil {
			t.Fatal("review attack: stale assignment frozen")
		}
		reason(t, e, "STALE_ASSIGNMENT_STATE")
	})
	t.Run("cross policy", func(t *testing.T) {
		h := newHarness(t)
		h.add("a")
		negativeDeclarations(h, "a")
		h.assign(Assignment{"a", "HOLDOUT", "review"})
		p := policyObject()
		p.ID = "p2"
		p.Policy.Round = "round2"
		r := h.apply("policy register", p)
		o := assignmentObject(h, Assignment{"a", "DEVELOPMENT", "review"})
		o.ID = "p2-assignment"
		o.PolicyDigest = r.Object
		_, e := h.s.Apply(h.request("assign-split", o))
		if e == nil {
			t.Fatal("review attack: cross-policy table accepted")
		}
		reason(t, e, "ASSIGNMENT_POLICY_MISMATCH")
	})
}

func TestReviewB3Durability(t *testing.T) {
	for _, point := range []string{"after_refusal_intent", "after_object", "before_receipt", "after_refusal_receipt"} {
		t.Run(point, func(t *testing.T) {
			h := newHarness(t)
			a := h.add("a")
			q := h.request("candidate add", candidateObject("never-retain", h.p))
			q.Object.Candidate.Family.Rationale = "[private] prohibited-review-content"
			payload, _ := json.Marshal(q.Object)
			h.s.fault = func(at string) error {
				if at == point {
					return fail("INJECTED_CRASH", 4)
				}
				return nil
			}
			_, e := h.s.Apply(q)
			reason(t, e, "INJECTED_CRASH")
			h.s.fault = nil
			recovery, e := h.s.Recover()
			if e != nil || len(recovery.Orphans) == 0 {
				t.Fatal("missing pending refusal", e)
			}
			_, e = h.s.Head()
			if e == nil {
				t.Fatal("pending refusal silently promoted")
			}
			another := q
			another.Operation = "another"
			_, e = h.s.Apply(another)
			reason(t, e, "PENDING_REFUSAL")
			// Even changed forbidden bytes cannot replace an already pending opaque attempt.
			q.Object.Candidate.Family.Rationale = "different prohibited payload"
			r, e := h.s.Apply(q)
			reason(t, e, "PRIVACY_QUARANTINE")
			if r.Sequence != 3 {
				t.Fatal(r)
			}
			again, e := h.s.Apply(q)
			reason(t, e, "PRIVACY_QUARANTINE")
			if again != r {
				t.Fatal("nonidempotent refusal")
			}
			st, e := h.s.scan()
			if e != nil || st.heads["a"] != a || len(inventory(st)) != 2 || len(st.orphans) != 0 {
				t.Fatal("candidate mutation or lost receipt", e)
			}
			filepath.WalkDir(h.s.Root, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					b, _ := os.ReadFile(path)
					for _, forbidden := range [][]byte{[]byte("never-retain"), []byte("prohibited-review-content"), []byte(Hash(payload).Hex)} {
						if bytes.Contains(b, forbidden) {
							t.Fatal("unsafe refusal bytes", path)
						}
					}
				}
				return nil
			})
			ar := h.assign(Assignment{"a", "DEVELOPMENT", "review"})
			f := h.freeze(ar)
			report, e := h.s.Verify(f.Object, h.head, now, true)
			if e != nil || count(report.Historical, "failed_attempts") != 1 {
				t.Fatal("freeze omitted refusal", e)
			}
			if e = os.Remove(filepath.Join(h.s.Root, objectPath(r.Object))); e != nil {
				t.Fatal(e)
			}
			_, e = h.s.Verify(f.Object, h.head, now, true)
			reason(t, e, "COMMITTED_CLOSURE_MISSING")
		})
	}
}
func TestReviewB3MalformedWire(t *testing.T) {
	h := newHarness(t)
	h.add("a")
	q := h.request("candidate add", candidateObject("forbidden-id", h.p))
	b, _ := json.Marshal(q)
	b = bytes.Replace(b, []byte(`"candidate":`), []byte(`"source_body":"[private] forbidden-body","candidate":`), 1)
	r, e := h.s.ApplyBytes(b, "")
	if e == nil || r.Sequence == 0 {
		t.Fatal("malformed wire lost attempt", e)
	}
	again, e := h.s.ApplyBytes(b, "")
	if e == nil || again != r {
		t.Fatal("malformed retry", e)
	}
	r, e = h.s.ApplyBytes([]byte("{invalid [private] body"), "opaque-malformed")
	if e == nil || r.Sequence == 0 {
		t.Fatal("unreadable envelope lost safe attempt", e)
	}
	st, e := h.s.scan()
	if e != nil || len(inventory(st)) != 3 {
		t.Fatal("inventory", e)
	}
	for _, o := range st.objects {
		b, _ := json.Marshal(o)
		if bytes.Contains(b, []byte("forbidden")) || bytes.Contains(b, []byte("[private]")) {
			t.Fatal("payload retained")
		}
	}
}

func TestReviewB1Adjacent(t *testing.T) {
	for _, kind := range []string{"producer", "discoverer", "decision", "assessor", "new session", "new model"} {
		t.Run(kind, func(t *testing.T) {
			h := newHarness(t)
			o := candidateObject("a", h.p)
			switch kind {
			case "producer":
				o.Producer.Actor = "other-producer"
			case "discoverer":
				o.Candidate.Discovery.Discoverer.Actor = "other-discoverer"
			case "decision":
				o.Candidate.Decision.Actor = "other-selector"
			case "assessor":
				o.Candidate.Assessments[0].Assessor = "other-assessor"
			case "new session":
				o.Producer.Context = "new-session"
			case "new model":
				o.Producer.Model = "new-model"
			}
			h.apply("candidate add", o)
			negativeDeclarations(h, "a")
			if h.summary().Members[0].Freshness != "UNKNOWN" {
				t.Fatal("missing actor/context erased")
			}
			h.expose("a", "matcher_rules", "YES")
			if h.summary().Members[0].Freshness != "EXPOSED" {
				t.Fatal("YES did not dominate UNKNOWN")
			}
			h.expose("a", "matcher_rules", "NO_WITH_BASIS")
			if h.summary().Members[0].Freshness != "EXPOSED" {
				t.Fatal("NO reset YES")
			}
		})
	}
	for _, kind := range []string{"alias", "revision", "new round"} {
		t.Run(kind, func(t *testing.T) {
			h := newHarness(t)
			a := h.add("a")
			p := h.p
			if kind == "new round" {
				o := policyObject()
				o.ID = "p2"
				o.Policy.Round = "round2"
				r := h.apply("policy register", o)
				p = Ref{o.ID, 1, r.Object}
			}
			o := candidateObject("b", p)
			if kind == "new round" {
				o.Candidate.Discovery.Round = "round2"
				o.Candidate.Decision.Round = "round2"
			}
			r := h.apply("candidate add", o)
			b := Ref{"b", 1, r.Object}
			negativeDeclarations(h, "b")
			if kind == "revision" {
				b = h.revise(b, func(o *Object) { o.Candidate.Family.Rationale = "revised" })
			}
			h.relations(relation(b, a, "SHARED_PROJECT_LINEAGE", "ESTABLISHED"))
			for _, m := range h.summary().Members {
				if m.Freshness != "UNKNOWN" {
					t.Fatal("merge erased missing declaration", kind)
				}
			}
			h.expose("a", "matcher_rules", "YES")
			for _, m := range h.summary().Members {
				if m.Freshness != "EXPOSED" {
					t.Fatal("alias lost YES")
				}
			}
		})
	}
	t.Run("freeze selector", func(t *testing.T) {
		h := newHarness(t)
		h.add("a")
		negativeDeclarations(h, "a")
		ar := h.assign(Assignment{"a", "HOLDOUT", "review"})
		o := freezeObject(h, ar)
		o.Producer.Actor = "new-freezer"
		_, e := h.s.PrepareFreeze(o)
		reason(t, e, "FRESH_SPLIT_REFUSED")
	})
}
func TestReviewB2Adjacent(t *testing.T) {
	for _, kind := range []string{"missing record", "stale artifact", "unknown rights", "unsupported target", "invalid classification", "revision"} {
		t.Run(kind, func(t *testing.T) {
			h := newHarness(t)
			root := t.TempDir()
			r, d, _ := existingEvidence(t, root, "review-capture", false)
			h.s.EvidenceRoot = root
			h.s.EvidenceStoreID = "evidence-fixture"
			o := capturedCandidate(h, "a", r, d)
			action := "candidate add"
			want := ""
			switch kind {
			case "missing record":
				o.Candidate.Captured[0].Record = Hash([]byte("absent"))
				want = "CAPTURE_UNAVAILABLE"
			case "stale artifact":
				o.Candidate.Captured[0].Source.Revision++
				want = "CAPTURE_IDENTITY_MISMATCH"
			case "unknown rights":
				o.Candidate.Sources[0].Rights.Metadata = "unknown"
				want = "RIGHTS_BLOCKED"
			case "unsupported target":
				o.Candidate.Captured[0].Source.ArtifactID = "unsupported"
				want = "CAPTURE_IDENTITY_MISMATCH"
			case "invalid classification":
				o.Candidate.OriginClass = "authoritative-natural"
				want = "INVALID_CLASSIFICATION"
			case "revision":
				added := h.apply(action, o)
				o.Revision = 2
				o.Parent = &added.Object
				o.Sensitivity = "public"
				o.Candidate.OriginClass = "natural_incident"
				action = "candidate revise"
				want = "ORIGIN_CLASS_IMMUTABLE"
			}
			raw, _ := json.Marshal(h.request(action, o))
			res, e := h.s.ApplyBytes(raw, "")
			reason(t, e, want)
			if res.Sequence == 0 {
				t.Fatal("capture refusal unauditable")
			}
		})
	}
}
func TestReviewB4Adjacent(t *testing.T) {
	for _, kind := range []string{"mirror", "project merge", "alias", "synthetic", "exposure", "session", "model", "revision rename"} {
		t.Run(kind, func(t *testing.T) {
			h := newHarness(t)
			a := h.add("a")
			b := h.add("b")
			negativeDeclarations(h, "a")
			negativeDeclarations(h, "b")
			ar := h.assign(Assignment{"a", "HOLDOUT", "review"}, Assignment{"b", "HOLDOUT", "review"})
			switch kind {
			case "mirror":
				h.relations(relation(b, a, "MIRROR_OF", "ESTABLISHED"))
			case "project merge":
				h.relations(relation(b, a, "SHARED_PROJECT_LINEAGE", "ESTABLISHED"))
			case "alias":
				h.revise(b, func(o *Object) { o.Candidate.Sources[0].Locator = "https://example.invalid/a" })
			case "synthetic":
				h.revise(b, func(o *Object) { o.Candidate.Parents = []Ref{a}; o.Candidate.Transformation = "synthetic change" })
			case "exposure":
				h.expose("a", "matcher_rules", "YES")
			case "session":
				h.revise(a, func(o *Object) { o.Producer.Context = "new-session" })
			case "model":
				h.revise(a, func(o *Object) { o.Producer.Model = "new-model" })
			case "revision rename":
				v, e := h.s.Inspect(a.Digest)
				if e != nil {
					t.Fatal(e)
				}
				o := v.Object
				o.ID = "renamed"
				o.Revision = 2
				o.Parent = &a.Digest
				_, e = h.s.Apply(h.request("candidate revise", o))
				reason(t, e, "DANGLING_PARENT")
				return
			}
			o, e := h.s.PrepareFreeze(freezeObject(h, ar))
			if e == nil {
				_, e = h.s.Apply(h.request("freeze", o))
			}
			if e == nil {
				t.Fatal("stale assignment accepted", kind)
			}
			if !one(e.Error(), "STALE_ASSIGNMENT_STATE", "FRESH_SPLIT_REFUSED") {
				t.Fatal(e)
			}
		})
	}
	t.Run("unrelated policy table", func(t *testing.T) {
		h := newHarness(t)
		h.add("a")
		p := policyObject()
		p.ID = "p2"
		p.Policy.Round = "round2"
		r := h.apply("policy register", p)
		p2 := Ref{p.ID, 1, r.Object}
		b := candidateObject("b", p2)
		b.Candidate.Discovery.Round = "round2"
		b.Candidate.Decision.Round = "round2"
		h.apply("candidate add", b)
		negativeDeclarations(h, "a")
		ar := h.assign(Assignment{"a", "HOLDOUT", "review"})
		o := assignmentObject(h, Assignment{"b", "DEVELOPMENT", "other round"})
		o.PolicyDigest = p2.Digest
		h.apply("assign-split", o)
		f := h.freeze(ar)
		report, e := h.s.Verify(f.Object, h.head, now, true)
		if e != nil || len(report.Historical.Members) != 1 || report.Historical.Members[0].Split != "HOLDOUT" {
			t.Fatal("cross-policy summary pollution", e)
		}
	})
	t.Run("false split summary", func(t *testing.T) {
		h := newHarness(t)
		h.add("a")
		ar := h.assign(Assignment{"a", "DEVELOPMENT", "review"})
		o, e := h.s.PrepareFreeze(freezeObject(h, ar))
		if e != nil {
			t.Fatal(e)
		}
		o.Manifest.Summary.Members[0].Split = "HOLDOUT"
		_, e = h.s.Apply(h.request("freeze", o))
		reason(t, e, "FALSE_MANIFEST_SUMMARY")
	})
}

func TestReviewB3Golden(t *testing.T) {
	h := newHarness(t)
	q := h.request("candidate add", candidateObject("unretained", h.p))
	q.Operation = "refusal-golden"
	q.Object.Candidate.OriginClass = "invalid"
	h.s.fault = func(point string) error {
		if point == "after_refusal_intent" {
			return fail("INJECTED_CRASH", 4)
		}
		return nil
	}
	_, e := h.s.Apply(q)
	reason(t, e, "INJECTED_CRASH")
	pending, e := os.ReadFile(filepath.Join(h.s.Root, "pending-refusal.json"))
	if e != nil {
		t.Fatal(e)
	}
	h.s.fault = nil
	r, e := h.s.Apply(q)
	reason(t, e, "INVALID_CLASSIFICATION")
	object, e := os.ReadFile(filepath.Join(h.s.Root, objectPath(r.Object)))
	if e != nil {
		t.Fatal(e)
	}
	event, e := os.ReadFile(filepath.Join(h.s.Root, eventPath(r.Sequence)))
	if e != nil {
		t.Fatal(e)
	}
	actual, _ := json.Marshal([]json.RawMessage{pending, object, event})
	gold, e := os.ReadFile("testdata/refusal.golden.json")
	if e != nil || !bytes.Equal(actual, gold) {
		t.Fatalf("refusal wire drift %s: %v", Hash(actual).Hex, e)
	}
}
func TestReviewB3ConcurrentRefusals(t *testing.T) {
	h := newHarness(t)
	a := h.add("a")
	queries := []Request{}
	for i := 0; i < 3; i++ {
		q := h.request("candidate add", candidateObject("unretained", h.p))
		q.Object.Candidate.OriginClass = "invalid"
		queries = append(queries, q)
	}
	var wg sync.WaitGroup
	results := make(chan Result, 3)
	errs := make(chan error, 3)
	for _, q := range queries {
		wg.Add(1)
		go func(q Request) { defer wg.Done(); r, e := h.s.Apply(q); results <- r; errs <- e }(q)
	}
	wg.Wait()
	close(results)
	close(errs)
	for e := range errs {
		reason(t, e, "INVALID_CLASSIFICATION")
	}
	seq := map[uint32]bool{}
	for r := range results {
		if r.Sequence == 0 || seq[r.Sequence] {
			t.Fatal("lost/duplicate refusal")
		}
		seq[r.Sequence] = true
	}
	st, e := h.s.scan()
	if e != nil || len(inventory(st)) != 4 || st.heads["a"] != a {
		t.Fatal("concurrent accounting", e)
	}
	for _, q := range queries {
		r, e := h.s.Apply(q)
		reason(t, e, "INVALID_CLASSIFICATION")
		if !seq[r.Sequence] {
			t.Fatal("retry duplicated attempt")
		}
	}
}

func TestReviewB2PublicIsNotNaturalAuthority(t *testing.T) {
	h := newHarness(t)
	r, _, _ := existingEvidence(t, t.TempDir(), "public-capture", false)
	// This is still authored test content. Public visibility is not an authoritative
	// natural-incident flag in Phase 2A; a conservative synthetic corpus label does
	// not gain a natural case merely because the referenced artifact is public.
	r.Artifacts[0].Sensitivity = "public"
	r.Artifacts[0].QualityClass = "public_metadata"
	r.Artifacts[0].Locations[0].Sensitivity = "public"
	root := t.TempDir()
	d, _, e := evidence.Save(root, r, now)
	if e != nil {
		t.Fatal(e)
	}
	h.s.EvidenceRoot = root
	h.s.EvidenceStoreID = "evidence-fixture"
	o := capturedCandidate(h, "a", r, d)
	h.apply("candidate add", o)
	ar := h.assign(Assignment{"a", "DEVELOPMENT", "synthetic remains synthetic"})
	f := h.freeze(ar)
	v, e := h.s.Verify(f.Object, h.head, now, true)
	if e != nil || count(v.Historical, "natural") != 0 || count(v.Historical, "synthetic") != 1 || v.CurrentUseStatus != "METADATA_REGISTER_ONLY" {
		t.Fatal("public capture became natural authority", e, v)
	}
}
func TestReviewB2InvalidTargetRecord(t *testing.T) {
	h := newHarness(t)
	root := t.TempDir()
	r, d, _ := existingEvidence(t, root, "capture", false)
	h.s.EvidenceRoot = root
	h.s.EvidenceStoreID = "evidence-fixture"
	raw, e := os.ReadFile(filepath.Join(root, "sha256-"+d.Hex+".json"))
	if e != nil {
		t.Fatal(e)
	}
	var invalid evidence.Record
	if json.Unmarshal(raw, &invalid) != nil {
		t.Fatal("record")
	}
	invalid.TargetClaim = "missing-target"
	raw, _ = json.Marshal(invalid)
	bad := Hash(raw)
	if e = os.WriteFile(filepath.Join(root, "sha256-"+bad.Hex+".json"), raw, 0600); e != nil {
		t.Fatal(e)
	}
	o := capturedCandidate(h, "a", r, bad)
	res, e := h.s.Apply(h.request("candidate add", o))
	reason(t, e, "PHASE2A_VALIDATION_REFUSED")
	if res.Sequence == 0 {
		t.Fatal("unsupported target lost refusal")
	}
}

func TestReviewB4HistoricalCurrentBinding(t *testing.T) {
	h := newHarness(t)
	a := h.add("a")
	negativeDeclarations(h, "a")
	ar := h.assign(Assignment{"a", "HOLDOUT", "review"})
	f := h.freeze(ar)
	old, e := h.s.Verify(f.Object, &f.Head, now, false)
	if e != nil {
		t.Fatal(e)
	}
	h.revise(a, func(o *Object) { o.Candidate.Sources[0].Locator = "https://example.invalid/correction" })
	current, e := h.s.Verify(f.Object, h.head, now, true)
	if e != nil || current.CurrentUseStatus != "BLOCKED" || current.Freshness != "UNKNOWN" {
		t.Fatal("stale assignment reported current eligibility", e, current)
	}
	historical, e := h.s.Verify(f.Object, &f.Head, now, false)
	if e != nil || !reflect.DeepEqual(old.Historical, historical.Historical) {
		t.Fatal("historical assignment changed", e)
	}
}

func TestReviewB3OperationConflict(t *testing.T) {
	h := newHarness(t)
	q := h.request("candidate add", candidateObject("a", h.p))
	original, e := h.s.Apply(q)
	if e != nil {
		t.Fatal(e)
	}
	altered := q
	altered.Object.Rationale = "[private] different prohibited bytes"
	r, e := h.s.Apply(altered)
	reason(t, e, "OPERATION_CONFLICT")
	if r.Sequence == 0 {
		t.Fatal("conflict disappeared")
	}
	retry, e := h.s.Apply(altered)
	reason(t, e, "OPERATION_CONFLICT")
	if retry != r {
		t.Fatal("conflict retry drift")
	}
	replay, e := h.s.Apply(q)
	if e != nil || replay != original {
		t.Fatal("original operation replaced", e)
	}
	st, e := h.s.scan()
	if e != nil || len(inventory(st)) != 2 {
		t.Fatal("missing protocol refusal", e)
	}
}
func TestReviewB3AnonymousEnvelope(t *testing.T) {
	h := newHarness(t)
	r, e := h.s.ApplyBytes([]byte("[private] unreadable submission"), "")
	if e == nil || r.Sequence == 0 {
		t.Fatal("anonymous attempt missing", e)
	}
	st, e := h.s.scan()
	if e != nil || len(inventory(st)) != 1 {
		t.Fatal("anonymous inventory", e)
	}
	for _, o := range st.objects {
		raw, _ := json.Marshal(o)
		if bytes.Contains(raw, []byte("unreadable submission")) {
			t.Fatal("anonymous payload retained")
		}
	}
}

func TestReviewB3ManualAttemptPolicy(t *testing.T) {
	h := newHarness(t)
	o := envelope("attempt", "manual", Hash([]byte("unregistered-policy")))
	why := "MALFORMED_INTAKE"
	o.AttemptReason = &why
	r, e := h.s.Apply(h.request("attempt", o))
	reason(t, e, "POLICY_NOT_REGISTERED")
	if r.Sequence == 0 {
		t.Fatal("missing safe refusal")
	}
	if _, e = h.s.Head(); e != nil {
		t.Fatal("manual attempt corrupted ledger", e)
	}
}

func TestReviewB3UnsafeOperationIdentity(t *testing.T) {
	for _, op := range []string{"private_key_fixture", string([]byte{0xff, 0xfe})} {
		t.Run("opaque", func(t *testing.T) {
			h := newHarness(t)
			q := h.request("candidate add", candidateObject("unretained", h.p))
			q.Operation = op
			r, e := h.s.Apply(q)
			reason(t, e, "UNSAFE_ATTEMPT_ID")
			if r.Sequence == 0 {
				t.Fatal("unsafe identity lost attempt")
			}
			st, e := h.s.scan()
			if e != nil {
				t.Fatal(e)
			}
			for _, ev := range st.events {
				if ev.Operation == op {
					t.Fatal("unsafe ID retained")
				}
			}
		})
	}
}

func TestReviewB2UnsafeCaptureClassification(t *testing.T) {
	for _, class := range []string{"private", "secret", "unknown", "private-location"} {
		t.Run(class, func(t *testing.T) {
			h := newHarness(t)
			r, _, _ := existingEvidence(t, t.TempDir(), "classified-capture", false)
			r.Artifacts[0].Sensitivity = class
			r.Artifacts[0].QualityClass = "metadata"
			r.Artifacts[0].Locations[0].Sensitivity = class
			if class == "private-location" {
				r.Artifacts[0].Sensitivity = "public"
				r.Artifacts[0].Locations[0].Sensitivity = "private"
			}
			root := t.TempDir()
			d, _, e := evidence.Save(root, r, now)
			if e != nil {
				t.Fatal("independent Phase 2A record", e)
			}
			h.s.EvidenceRoot = root
			h.s.EvidenceStoreID = "evidence-fixture"
			o := capturedCandidate(h, "a", r, d)
			o.Sensitivity = "public"
			o.Candidate.OriginClass = "natural_incident"
			res, e := h.s.Apply(h.request("candidate add", o))
			if e == nil {
				t.Fatal("unsafe capture classification accepted")
			}
			reason(t, e, "CAPTURE_CLASSIFICATION_MISMATCH")
			if res.Sequence == 0 {
				t.Fatal("missing refusal")
			}
		})
	}
}
