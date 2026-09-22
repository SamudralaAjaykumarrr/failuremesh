package corpus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/evidence"
)

const now = "2026-09-17T12:00:00Z"

func known(s string) Fact { return Fact{"KNOWN", &s, "authored synthetic test metadata", []string{}} }
func unknown() Fact       { return Fact{"UNKNOWN", nil, "not established in synthetic fixture", []string{}} }
func author() Identity {
	return Identity{"fixture-author", "tool", "curator", "corpus-test", "1", "", "synthetic-session"}
}
func envelope(kind, id string, p Digest) Object {
	return Object{Schema: Schema, Encoding: Encoding, Kind: kind, ID: id, Revision: 1, Producer: author(), PolicyDigest: p, ReportedAt: now, Sensitivity: "synthetic", Reasons: []string{"SYNTHETIC_TEST"}, Rationale: "authored synthetic metadata only"}
}
func policyObject() Object {
	o := envelope("policy", "policy", VocabularyDigest())
	p := Policy{Round: "round", Vocabulary: VocabularyDigest(), Scope: "synthetic test of public-metadata paths, no actual incidents", Channels: []string{"manual-synthetic-fixtures"}, Start: "2026-09-17T00:00:00Z", Cutoff: "2026-09-18T00:00:00Z", Eligibility: []string{"metadata-permitted"}, Challenges: []string{"ambiguous"}, PublisherMultiplier: 3, DiversityException: "pre-result synthetic demonstration shortfall", SplitMethod: "explicit-table-v1", Seed: "NOT_USED", RightsRequirements: []string{"metadata-allow"}, ExposureExclusions: []string{"rules-results-target"}, StopRules: []string{"finite-budget"}, Targets: []Target{{"natural", 25}, {"split:DEVELOPMENT", 15}, {"split:VALIDATION", 5}, {"split:HOLDOUT", 5}}}
	for _, f := range families {
		p.Budgets = append(p.Budgets, Budget{f, 100, 60})
		p.Targets = append(p.Targets, Target{"family:" + f, 5})
	}
	o.Policy = &p
	return o
}
func rights() evidence.Rights {
	return evidence.Rights{License: "synthetic-test", Basis: "authored metadata fixture", Review: "explicit", Metadata: "allow", Digest: "allow", Excerpt: "unknown", WholeBody: "unknown", Export: "unknown"}
}
func candidateObject(id string, p Ref) Object {
	o := envelope("candidate", id, p.Digest)
	o.Sensitivity = "public"
	c := Candidate{Discovery: Discovery{"manual-local", author(), now, unknown(), "supplied synthetic lead", "round", "unknown", "no retrieval requested"}, Sources: []Citation{{"citation", "https://example.invalid/" + id, "authored metadata", "synthetic-test-report", "origin", "citation-only", "uncaptured citation, no semantic evidence", rights(), "unknown", now}}, Captured: []Captured{}, Publisher: Attribution{known("publisher-" + id), []string{"publisher-" + id}, "authored ownership fixture"}, Project: known("project-" + id), ProjectLineage: known("project-" + id), IncidentTime: Date{unknown(), "unknown", "UNKNOWN"}, PublicationTime: Date{unknown(), "unknown", "UNKNOWN"}, System: unknown(), Stack: unknown(), ArchitectureRefs: []string{}, Family: Family{VocabularyDigest(), "CANDIDATE", known(families[0]), []string{}, "synthetic tentative mechanism", "synthetic scope", "manual hypothesis, not truth"}, OriginClass: "natural_incident", Origin: known("origin-" + id), Episode: known("episode-" + id), Independence: known("bounded manual assessment"), ExposureRefs: []Digest{}, Pretraining: "UNKNOWN", Decision: Decision{"INCLUDED", []string{"SYNTHETIC_TEST"}, "synthetic inclusion", p, "fixture-author", "round", "no actual search; alternatives not applicable"}, Split: "UNASSIGNED", FreezeState: "SELECTED", Parents: []Ref{}}
	for _, dim := range []string{"causal_clarity", "architecture_context", "reproduction", "directness"} {
		c.Assessments = append(c.Assessments, Assessment{dim, "unknown", "synthetic metadata only", []string{}, "fixture-author"})
	}
	o.Candidate = &c
	return o
}

type harness struct {
	t    *testing.T
	s    Store
	head *Digest
	n    int
	p    Ref
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{t: t, s: Store{Root: filepath.Join(t.TempDir(), "store")}}
	r := h.apply("policy register", policyObject())
	h.p = Ref{"policy", 1, r.Object}
	return h
}
func (h *harness) request(action string, o Object) Request {
	if head, e := h.s.Head(); e == nil {
		h.head = head
	}
	h.n++
	return Request{fmt.Sprintf("op-%d", h.n), h.head, action, o}
}
func (h *harness) apply(action string, o Object) Result {
	h.t.Helper()
	r, e := h.s.Apply(h.request(action, o))
	if e != nil {
		h.t.Fatal(action, e)
	}
	h.head = &r.Head
	return r
}
func (h *harness) add(id string) Ref {
	r := h.apply("candidate add", candidateObject(id, h.p))
	return Ref{id, 1, r.Object}
}
func (h *harness) revise(r Ref, fn func(*Object)) Ref {
	o, e := h.s.Inspect(r.Digest)
	if e != nil {
		h.t.Fatal(e)
	}
	v := o.Object
	v.Revision++
	v.Parent = &r.Digest
	fn(&v)
	result := h.apply("candidate revise", v)
	return Ref{v.ID, v.Revision, result.Object}
}
func (h *harness) relations(rs ...Relation) {
	o := envelope("cluster", fmt.Sprintf("cluster-%d", h.n), h.p.Digest)
	o.Relations = &rs
	h.apply("cluster", o)
}
func relation(a, b Ref, kind, status string) Relation {
	return Relation{a.ID + "-" + b.ID + "-" + kind, a, b, kind, status, "synthetic provenance fixture", "explicit relation proposal"}
}
func (h *harness) expose(id, cat, state string) {
	o := envelope("exposure", fmt.Sprintf("exposure-%d", h.n), h.p.Digest)
	o.Exposure = &Exposure{[]string{id}, "round", cat, state, state, "synthetic access declaration", "synthetic isolated context", unknown(), now, []Digest{}}
	h.apply("exposure append", o)
}
func (h *harness) assign(rows ...Assignment) Ref {
	o := envelope("assignment", fmt.Sprintf("assignment-%d", h.n), h.p.Digest)
	o.Assignments = &rows
	r := h.apply("assign-split", o)
	return Ref{o.ID, 1, r.Object}
}
func (h *harness) freeze(a Ref) Result {
	o := envelope("manifest", "corpus", h.p.Digest)
	o.Schema = ManifestSchema
	o.Manifest = &Manifest{Policy: h.p, Assignment: a, CheckTime: now, Build: Hash([]byte("synthetic build")), SourceCommit: "synthetic-uncommitted"}
	v, e := h.s.PrepareFreeze(o)
	if e != nil {
		h.t.Fatal(e)
	}
	return h.apply("freeze", v)
}
func count(s Summary, name string) uint32 {
	for _, c := range s.Counts {
		if c.Name == name {
			return c.Count
		}
	}
	return 0
}
func (h *harness) summary() Summary {
	f, e := h.s.lock(false)
	if e != nil {
		h.t.Fatal(e)
	}
	defer f.Close()
	st, e := h.s.scan()
	if e != nil {
		h.t.Fatal(e)
	}
	v, e := h.s.summarize(st, h.p.Digest, now, false)
	if e != nil {
		h.t.Fatal(e)
	}
	return v
}
func reason(t *testing.T, e error, want string) {
	t.Helper()
	if e == nil || e.Error() != want {
		t.Fatalf("want %s, got %v", want, e)
	}
}

func TestDemonstration(t *testing.T) {
	h := newHarness(t)
	a := h.add("a")
	b := h.add("b")
	b = h.revise(b, func(o *Object) {
		o.Candidate.Decision.State = "EXCLUDED"
		o.Candidate.Decision.Reasons = []string{"DUPLICATE"}
	})
	h.relations(relation(b, a, "MIRROR_OF", "ESTABLISHED"))
	a = h.revise(a, func(o *Object) {
		o.Candidate.Family.Primary = known(families[1])
		o.Candidate.Family.Rationale = "alternative tentative family"
	})
	h.expose("a", "matcher_rules", "YES")
	assignment := h.assign(Assignment{"a", "DEVELOPMENT", "matcher informed"}, Assignment{"b", "NOT_SELECTED", "duplicate excluded"})
	result := h.freeze(assignment)
	t.Logf("candidate=%s manifest=%s receipt=%s", a.Digest.Hex, result.Object.Hex, result.Head.Hex)
	report, e := h.s.Verify(result.Object, h.head, now, true)
	if e != nil {
		t.Fatal(e)
	}
	if report.Integrity != "INTACT" || report.SelectionTargetStatus != "SHORTFALL" || count(report.Historical, "natural") != 1 {
		t.Fatalf("%+v", report)
	}
	old, e := os.ReadFile(filepath.Join(h.s.Root, objectPath(result.Object)))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = h.s.Inspect(result.Object); e != nil {
		t.Fatal(e)
	}
	h.expose("a", "prior_case_outcome", "YES")
	h.revise(a, func(o *Object) {
		o.Candidate.Family.State = "DISPUTED"
		o.Candidate.Family.Alternatives = []string{families[0]}
	})
	historical, e := h.s.Verify(result.Object, &result.Head, now, false)
	if e != nil {
		t.Fatal(e)
	}
	current, e := h.s.Verify(result.Object, h.head, now, true)
	if e != nil {
		t.Fatal(e)
	}
	if historical.CurrentUseStatus != "HISTORICAL_ONLY" || current.Freshness != "EXPOSED" || len(current.Notices) < 2 {
		t.Fatalf("historical=%+v current=%+v", historical, current)
	}
	after, _ := os.ReadFile(filepath.Join(h.s.Root, objectPath(result.Object)))
	if !bytes.Equal(old, after) {
		t.Fatal("historical bytes changed")
	}
}
func TestOriginCountingIntegration(t *testing.T) {
	for _, kind := range []string{"SAME_INCIDENT", "MIRROR_OF", "COPIED_SUMMARY_OF", "FOLLOWUP_OF", "ISSUE_REPORT_OF", "SAME_CAUSAL_EPISODE"} {
		t.Run(kind, func(t *testing.T) {
			h := newHarness(t)
			a := h.add("a")
			b := h.add("b")
			c := h.add("c")
			h.relations(relation(b, a, kind, "ESTABLISHED"), relation(c, a, kind, "ESTABLISHED"))
			sum := h.summary()
			if count(sum, "natural") != 1 || count(sum, "submissions") != 3 {
				t.Fatalf("inflated count %+v", sum)
			}
		})
	}
	t.Run("possible overlap", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		b := h.add("b")
		h.relations(relation(b, a, "SAME_INCIDENT", "POSSIBLE"))
		if count(h.summary(), "natural") != 0 {
			t.Fatal("possible established independence")
		}
	})
	t.Run("secondary quotes two unrelated incidents", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		b := h.add("b")
		c := h.add("secondary")
		h.revise(c, func(o *Object) { o.Candidate.Decision.State = "EXCLUDED" })
		h.relations(relation(c, a, "QUOTES_ORIGIN", "ESTABLISHED"), relation(c, b, "QUOTES_ORIGIN", "ESTABLISHED"))
		if count(h.summary(), "natural") != 2 {
			t.Fatal("citation fused distinct origins")
		}
	})
	t.Run("unknown episode", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		h.revise(a, func(o *Object) { o.Candidate.Episode = unknown() })
		if count(h.summary(), "natural") != 0 {
			t.Fatal("unknown episode counted")
		}
	})
	t.Run("synthetic descendant falsely natural", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		b := h.add("b")
		h.relations(relation(b, a, "SYNTHETIC_DESCENDANT_OF", "ESTABLISHED"))
		sum := h.summary()
		if count(sum, "natural") != 1 || count(sum, "synthetic") != 1 {
			t.Fatal("synthetic inflation")
		}
	})
	t.Run("rejected proposal retained", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		b := h.add("b")
		h.relations(relation(b, a, "MIRROR_OF", "REJECTED"))
		if count(h.summary(), "natural") != 2 {
			t.Fatal("rejected merged")
		}
	})
}
func TestAuthorityPrivacyAndEncoding(t *testing.T) {
	h := newHarness(t)
	o := candidateObject("a", h.p)
	for _, who := range []string{"human", "ai"} {
		v := o
		v.Producer.Class = who
		v.Producer.Model = "synthetic-model"
		v.Candidate = &Candidate{}
		*v.Candidate = *o.Candidate
		v.Candidate.Family.State = "EVIDENCE_SUPPORTED"
		_, e := h.s.Apply(h.request("candidate add", v))
		reason(t, e, "FAMILY_AUTHORITY_REFUSED")
	}
	for _, kind := range []string{"privacy", "rights"} {
		v := candidateObject(kind, h.p)
		if kind == "privacy" {
			v.Candidate.Family.Rationale = "[private] synthetic prohibited marker"
		} else {
			v.Candidate.Sources[0].Rights.Metadata = "unknown"
		}
		r, e := h.s.Apply(h.request("candidate add", v))
		if e == nil || r.Sequence == 0 {
			t.Fatal("no durable safe refusal", e)
		}
		h.head = &r.Head
		filepath.WalkDir(h.s.Root, func(p string, d os.DirEntry, e error) error {
			if e == nil && !d.IsDir() {
				b, _ := os.ReadFile(p)
				if bytes.Contains(b, []byte("synthetic prohibited marker")) {
					t.Fatal("private payload persisted")
				}
			}
			return e
		})
	}
	b, d, e := Encode(o)
	if e != nil {
		t.Fatal(e)
	}
	for name, raw := range map[string][]byte{"duplicate": bytes.Replace(b, []byte(`"schema_version":"corpus-v1"`), []byte(`"schema_version":"corpus-v1","schema_version":"corpus-v1"`), 1), "unknown": bytes.Replace(b, []byte(`"schema_version":"corpus-v1"`), []byte(`"schema_version":"future"`), 1), "newline": append(slices.Clone(b), '\n'), "unknown field": bytes.Replace(b, []byte(`"encoding":`), []byte(`"unknown":null,"encoding":`), 1), "tamper": append(slices.Clone(b), 0)} {
		t.Run(name, func(t *testing.T) {
			if _, e := Decode(raw, Hash(raw)); e == nil {
				t.Fatal("ambiguous encoding accepted")
			}
		})
	}
	if _, e = Decode(b, Hash([]byte("other"))); e == nil {
		t.Fatal("wrong digest")
	}
	v, e := Decode(b, d)
	if e != nil {
		t.Fatal(e)
	}
	b2, d2, e := Encode(v)
	if e != nil || !bytes.Equal(b, b2) || d != d2 {
		t.Fatal("roundtrip")
	}
}
func TestFreezeInventoryAndSummary(t *testing.T) {
	h := newHarness(t)
	a := h.add("a")
	assignment := h.assign(Assignment{"a", "DEVELOPMENT", "fixture"})
	skeleton := envelope("manifest", "corpus", h.p.Digest)
	skeleton.Schema = ManifestSchema
	skeleton.Manifest = &Manifest{Policy: h.p, Assignment: assignment, CheckTime: now, Build: Hash([]byte("build")), SourceCommit: "synthetic"}
	good, e := h.s.PrepareFreeze(skeleton)
	if e != nil {
		t.Fatal(e)
	}
	for _, attack := range []string{"summary", "inventory", "closure", "head"} {
		b, _ := json.Marshal(good)
		var v Object
		json.Unmarshal(b, &v)
		switch attack {
		case "summary":
			v.Manifest.Summary.Counts[0].Count++
		case "inventory":
			v.Manifest.Inventory = nil
		case "closure":
			v.Manifest.Closure = nil
		case "head":
			v.Manifest.LedgerSequence++
		}
		_, e := h.s.Apply(h.request("freeze", v))
		if e == nil {
			t.Fatal("accepted", attack)
		}
	}
	h.revise(a, func(o *Object) { o.Candidate.Decision.State = "PENDING" })
	_, e = h.s.PrepareFreeze(skeleton)
	reason(t, e, "PENDING_INVENTORY")
}
func TestSplitAndExposureIntegration(t *testing.T) {
	t.Run("YES dominates NO across aliases and model", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		b := h.add("renamed")
		h.expose("a", "matcher_rules", "YES")
		h.expose("a", "matcher_rules", "NO_WITH_BASIS")
		h.relations(relation(b, a, "SHARED_PROJECT_LINEAGE", "ESTABLISHED"))
		sum := h.summary()
		for _, m := range sum.Members {
			if m.Freshness != "EXPOSED" {
				t.Fatal("reset exposure")
			}
		}
	})
	t.Run("undeclared context blocks holdout", func(t *testing.T) {
		h := newHarness(t)
		h.add("a")
		rows := []Assignment{{"a", "HOLDOUT", "claimed blind"}}
		o := envelope("assignment", "assignment", h.p.Digest)
		o.Assignments = &rows
		_, e := h.s.Apply(h.request("assign-split", o))
		reason(t, e, "FRESH_SPLIT_REFUSED")
	})
	t.Run("development cannot promote", func(t *testing.T) {
		h := newHarness(t)
		h.add("a")
		h.assign(Assignment{"a", "DEVELOPMENT", "used"})
		rows := []Assignment{{"a", "HOLDOUT", "promotion"}}
		o := envelope("assignment", "assignment-new", h.p.Digest)
		o.Assignments = &rows
		_, e := h.s.Apply(h.request("assign-split", o))
		reason(t, e, "SPLIT_HISTORY_CONFLICT")
	})
	t.Run("holdout reveal revision", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		for _, c := range categories {
			h.expose("a", c, "NO_WITH_BASIS")
		}
		ar := h.assign(Assignment{"a", "HOLDOUT", "scoped self-reported separation"})
		f := h.freeze(ar)
		h.expose("a", "prior_case_outcome", "YES")
		h.revise(a, func(o *Object) { o.Candidate.Family.Rationale = "post-reveal correction" })
		r, e := h.s.Verify(f.Object, h.head, now, true)
		if e != nil {
			t.Fatal(e)
		}
		if r.CurrentUseStatus != "BLOCKED" {
			t.Fatal("freshness restored")
		}
		old, e := h.s.Verify(f.Object, &f.Head, now, false)
		if e != nil || old.Historical.Members[0].Split != "HOLDOUT" {
			t.Fatal("lost historical holdout", e)
		}
	})
	t.Run("late project merge", func(t *testing.T) {
		h := newHarness(t)
		a := h.add("a")
		b := h.add("b")
		for _, c := range categories {
			h.expose("b", c, "NO_WITH_BASIS")
		}
		ar := h.assign(Assignment{"a", "DEVELOPMENT", "fixture"}, Assignment{"b", "HOLDOUT", "separate fixture"})
		f := h.freeze(ar)
		h.relations(relation(b, a, "SHARED_PROJECT_LINEAGE", "ESTABLISHED"))
		r, e := h.s.Verify(f.Object, h.head, now, true)
		if e != nil {
			t.Fatal(e)
		}
		if r.CurrentUseStatus != "BLOCKED" || !slices.Contains(r.Limitations, "COMPONENT_SPLIT_CONFLICT") {
			t.Fatalf("%+v", r)
		}
	})
}
func TestPublicationRecoveryAndConcurrency(t *testing.T) {
	t.Run("stale writers exact retry", func(t *testing.T) {
		h := newHarness(t)
		q := h.request("candidate add", candidateObject("a", h.p))
		q2 := h.request("candidate add", candidateObject("b", h.p))
		var wg sync.WaitGroup
		results := make(chan error, 2)
		for _, q := range []Request{q, q2} {
			wg.Add(1)
			go func(q Request) { defer wg.Done(); _, e := h.s.Apply(q); results <- e }(q)
		}
		wg.Wait()
		close(results)
		ok, bad := 0, 0
		for e := range results {
			if e == nil {
				ok++
			} else {
				reason(t, e, "STALE_HEAD")
				bad++
			}
		}
		if ok != 1 || bad != 1 {
			t.Fatal(ok, bad)
		}
		for _, q := range []Request{q, q2} {
			r, e := h.s.Apply(q)
			if e == nil {
				again, e := h.s.Apply(q)
				if e != nil || again != r {
					t.Fatal("nonidempotent")
				}
			}
		}
	})
	for _, point := range []string{"after_object", "before_receipt"} {
		t.Run(point, func(t *testing.T) {
			h := newHarness(t)
			h.add("a")
			ar := h.assign(Assignment{"a", "DEVELOPMENT", "fixture"})
			o := envelope("manifest", "corpus", h.p.Digest)
			o.Schema = ManifestSchema
			o.Manifest = &Manifest{Policy: h.p, Assignment: ar, CheckTime: now, Build: Hash([]byte("build")), SourceCommit: "synthetic"}
			o, e := h.s.PrepareFreeze(o)
			if e != nil {
				t.Fatal(e)
			}
			q := h.request("freeze", o)
			h.s.fault = func(p string) error {
				if p == point {
					return fail("SIMULATED_CRASH", 4)
				}
				return nil
			}
			_, e = h.s.Apply(q)
			reason(t, e, "SIMULATED_CRASH")
			h.s.fault = nil
			r, e := h.s.Recover()
			if e != nil || len(r.Orphans) == 0 || !same(r.Head, h.head) {
				t.Fatal("orphan promoted", r, e)
			}
			result, e := h.s.Apply(q)
			if e != nil {
				t.Fatal(e)
			}
			retry, e := h.s.Apply(q)
			if e != nil || retry != result {
				t.Fatal("retry changed")
			}
		})
	}
	for _, attack := range []string{"temp", "binding", "object", "event", "symlink", "directory"} {
		t.Run(attack, func(t *testing.T) {
			h := newHarness(t)
			a := h.add("a")
			switch attack {
			case "temp":
				os.WriteFile(filepath.Join(h.s.Root, "events/.pending-crash"), []byte("x"), 0600)
			case "binding":
				os.Remove(filepath.Join(h.s.Root, bindingPath(a)))
			case "object":
				os.WriteFile(filepath.Join(h.s.Root, objectPath(a.Digest)), []byte("{}"), 0600)
			case "event":
				os.Remove(filepath.Join(h.s.Root, eventPath(1)))
			case "symlink":
				os.Remove(filepath.Join(h.s.Root, objectPath(a.Digest)))
				os.Symlink("/etc/passwd", filepath.Join(h.s.Root, objectPath(a.Digest)))
			case "directory":
				os.Remove(filepath.Join(h.s.Root, objectPath(a.Digest)))
				os.Mkdir(filepath.Join(h.s.Root, objectPath(a.Digest)), 0700)
			}
			if _, e := h.s.Recover(); e == nil {
				t.Fatal("recovery ignored", attack)
			}
		})
	}
}
func TestLimitsAndDeterministicReplay(t *testing.T) {
	o := policyObject()
	o.Rationale = strings.Repeat("x", MaxObjectBytes)
	if _, _, e := Encode(o); e == nil {
		t.Fatal("object limit")
	}
	h := newHarness(t)
	c := candidateObject("too-many", h.p)
	for i := 0; i < MaxSources; i++ {
		v := c.Candidate.Sources[0]
		v.ID = fmt.Sprint(i)
		c.Candidate.Sources = append(c.Candidate.Sources, v)
	}
	_, _, e := Encode(c)
	reason(t, e, "SOURCE_LIMIT")
	var digests []Digest
	for i := 0; i < 2; i++ {
		h := newHarness(t)
		h.add("a")
		h.add("b")
		ar := h.assign(Assignment{"b", "DEVELOPMENT", "fixture"}, Assignment{"a", "DEVELOPMENT", "fixture"})
		f := h.freeze(ar)
		digests = append(digests, f.Object)
	}
	if digests[0] != digests[1] {
		t.Fatal("clean-store replay drift")
	}
}
func TestGoldenBytes(t *testing.T) {
	o := envelope("attempt", "opaque", VocabularyDigest())
	why := "RIGHTS_BLOCKED"
	o.AttemptReason = &why
	b, d, e := Encode(o)
	if e != nil {
		t.Fatal(e)
	}
	gold, e := os.ReadFile("testdata/attempt.golden.json")
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(b, gold) {
		t.Fatalf("golden differs: %s", b)
	}
	digest, e := os.ReadFile("testdata/attempt.sha256")
	if e != nil || string(digest) != d.Hex {
		t.Fatalf("golden digest differs: %s", d.Hex)
	}
	changed := o
	changed.Rationale = "meaningful correction"
	_, d2, _ := Encode(changed)
	if d2 == d {
		t.Fatal("change did not change digest")
	}
}
func TestSetOrderDoesNotAffectIdentity(t *testing.T) {
	o := policyObject()
	b, d, e := Encode(o)
	if e != nil {
		t.Fatal(e)
	}
	slices.Reverse(o.Policy.Budgets)
	slices.Reverse(o.Policy.Targets)
	b2, d2, e := Encode(o)
	if e != nil || !bytes.Equal(b, b2) || !reflect.DeepEqual(d, d2) {
		t.Fatal("set ordering nondeterministic")
	}
}
