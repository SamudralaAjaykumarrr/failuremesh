package corpus

import "testing"

func TestCorrectionAfterCutoff(t *testing.T) {
	h := newHarness(t)
	a := h.add("a")
	ar := h.assign(Assignment{"a", "DEVELOPMENT", "fixture"})
	h.freeze(ar)
	v, e := h.s.Inspect(a.Digest)
	if e != nil {
		t.Fatal(e)
	}
	o := v.Object
	o.Revision++
	o.Parent = &a.Digest
	o.ReportedAt = "2026-09-20T12:00:00Z"
	o.Candidate.Family.Rationale = "correction after selection cutoff"
	result, e := h.s.Apply(h.request("candidate revise", o))
	if e != nil {
		t.Fatalf("post-cutoff correction refused: %v", e)
	}
	h.head = &result.Head
	late := candidateObject("late", h.p)
	late.ReportedAt = o.ReportedAt
	_, e = h.s.Apply(h.request("candidate add", late))
	reason(t, e, "OUTSIDE_POLICY_WINDOW")
}
func TestCandidateRevisionCycleRefused(t *testing.T) {
	h := newHarness(t)
	a := h.add("a")
	b := h.add("b")
	a = h.revise(a, func(o *Object) {
		o.Candidate.Parents = []Ref{b}
		o.Candidate.Transformation = "synthetic ancestry test"
	})
	v, e := h.s.Inspect(b.Digest)
	if e != nil {
		t.Fatal(e)
	}
	o := v.Object
	o.Revision++
	o.Parent = &b.Digest
	o.Candidate.Parents = []Ref{a}
	o.Candidate.Transformation = "synthetic ancestry test"
	before := *h.head
	_, e = h.s.Apply(h.request("candidate revise", o))
	reason(t, e, "ANCESTRY_CYCLE")
	after, e := h.s.Head()
	if e != nil || *after == before {
		t.Fatal("refusal missing audit event", e)
	}
	st, e := h.s.scan()
	if e != nil || st.heads["b"] != b {
		t.Fatal("refusal changed candidate", e)
	}
}
func TestWithdrawalCannotBeLaunderedThroughSuccessor(t *testing.T) {
	h := newHarness(t)
	h.add("a")
	ar := h.assign(Assignment{"a", "DEVELOPMENT", "fixture"})
	first := h.freeze(ar)
	v, e := h.s.Inspect(first.Object)
	if e != nil {
		t.Fatal(e)
	}
	o := v.Object
	o.Revision++
	o.Parent = &first.Object
	o.Manifest.ParentManifest = &first.Object
	o, e = h.s.PrepareFreeze(o)
	if e != nil {
		t.Fatal(e)
	}
	second := h.apply("freeze", o)
	notice := envelope("withdrawal", "withdraw-original", h.p.Digest)
	notice.Notice = &Notice{first.Object, []string{"a"}, "source withdrawn"}
	h.apply("withdraw", notice)
	report, e := h.s.Verify(second.Object, h.head, now, true)
	if e != nil {
		t.Fatal(e)
	}
	if report.CurrentUseStatus != "BLOCKED" {
		t.Fatal("successor bypassed withdrawal")
	}
	v, e = h.s.Inspect(second.Object)
	if e != nil {
		t.Fatal(e)
	}
	o = v.Object
	o.Revision++
	o.Parent = &second.Object
	o.Manifest.ParentManifest = &second.Object
	o, e = h.s.PrepareFreeze(o)
	if e != nil {
		return
	}
	_, e = h.s.Apply(h.request("freeze", o))
	reason(t, e, "WITHDRAWN_COMPONENT")
}

func TestPublicationLimitBoundaries(t *testing.T) {
	// Test the exact guard used by Apply before any publication, without writing
	// tens of thousands of redundant fixtures. Store-level size/refusal tests are
	// retained in TestHardLimitBoundaries and recovery integration tests.
	st := emptyState()
	st.bytes = MaxClosureBytes - 3
	if e := publicationLimits(st, 3); e != nil {
		t.Fatal(e)
	}
	reason(t, publicationLimits(st, 4), "CLOSURE_LIMIT")
	st = emptyState()
	for i := 0; i < MaxClosureObjects-1; i++ {
		st.objects[Hash([]byte(string(rune(i))))] = Object{}
	}
	if e := publicationLimits(st, 1); e != nil {
		t.Fatal(e)
	}
	st.objects[Hash([]byte("additional-object"))] = Object{}
	reason(t, publicationLimits(st, 1), "CLOSURE_LIMIT")
	st = emptyState()
	st.storeFiles = MaxStoreFiles - 4
	if e := publicationLimits(st, 1); e != nil {
		t.Fatal(e)
	}
	st.storeFiles++
	reason(t, publicationLimits(st, 1), "STORE_LIMIT")
	st = emptyState()
	st.storeBytes = MaxStoreBytes - 2
	if e := publicationLimits(st, 1); e != nil {
		t.Fatal(e)
	}
	reason(t, publicationLimits(st, 2), "STORE_LIMIT")
}

func TestExposureContextDelimiterRefused(t *testing.T) {
	// Tuple keys must not let a new actor/context masquerade as a context whose
	// NO_WITH_BASIS declarations already exist.
	h := newHarness(t)
	o := candidateObject("a", h.p)
	o.Producer.Actor = "actor\x00context"
	o.Producer.Context = "session"
	_, e := h.s.Apply(h.request("candidate add", o))
	if e == nil {
		t.Fatal("ambiguous exposure context accepted")
	}
}
