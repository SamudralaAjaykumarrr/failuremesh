package corpus

import (
	"bytes"
	"testing"
)

func TestOversizedRefusalRecovery(t *testing.T) {
	for _, point := range []string{"after_refusal_intent", "after_object", "before_receipt", "after_refusal_receipt"} {
		t.Run(point, func(t *testing.T) {
			h := newHarness(t)
			a := h.add("a")
			h.s.fault = func(at string) error {
				if at == point {
					return fail("INJECTED_CRASH", 4)
				}
				return nil
			}
			_, e := h.s.RejectOversizedInput("oversized-review")
			reason(t, e, "INJECTED_CRASH")
			h.s.fault = nil
			recovery, e := h.s.Recover()
			if e != nil || len(recovery.Orphans) == 0 {
				t.Fatal("pending refusal lost", e)
			}
			if _, e = h.s.Head(); e == nil {
				t.Fatal("pending refusal promoted")
			}
			_, e = h.s.RejectOversizedInput("different-attempt")
			reason(t, e, "PENDING_REFUSAL")
			raw := bytes.Repeat([]byte("x"), MaxObjectBytes+1)
			r, e := h.s.ApplyBytes(raw, "oversized-review")
			reason(t, e, "OBJECT_LIMIT")
			again, e := h.s.RejectOversizedInput("oversized-review")
			reason(t, e, "OBJECT_LIMIT")
			if r != again || r.Sequence != 3 {
				t.Fatal("retry mismatch", r, again)
			}
			st, e := h.s.scan()
			if e != nil || st.heads["a"] != a || len(inventory(st)) != 2 || len(st.orphans) != 0 {
				t.Fatal("state changed or refusal missing", e)
			}
			ar := h.assign(Assignment{"a", "DEVELOPMENT", "synthetic"})
			fr := h.freeze(ar)
			report, e := h.s.Verify(fr.Object, h.head, now, true)
			if e != nil || count(report.Historical, "failed_attempts") != 1 {
				t.Fatal("freeze missing refusal", e)
			}
		})
	}
}
