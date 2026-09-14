package pfc3

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

func testPFC(t *testing.T) PFC {
	t.Helper()
	p, e := Load("../../contracts/stale-worker-after-lease-expiry.v1.json")
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestContractAndApplicability(t *testing.T) {
	p := testPFC(t)
	for _, rem := range []bool{false, true} {
		a := ReferenceAC(rem)
		if Evaluate(p, a).Verdict != "APPLICABLE" {
			t.Fatal("applicable")
		}
		f := a.Facts[prerequisites[0]]
		f.Value = phase1.Bool(false)
		a.Facts[prerequisites[0]] = f
		if Evaluate(p, a).Verdict != "NOT_APPLICABLE" {
			t.Fatal("not applicable")
		}
		a = ReferenceAC(rem)
		delete(a.Facts, prerequisites[0])
		if Evaluate(p, a).Verdict != "UNKNOWN" {
			t.Fatal("missing")
		}
		a = ReferenceAC(rem)
		f = a.Facts[prerequisites[0]]
		f.Stale = true
		a.Facts[prerequisites[0]] = f
		if Evaluate(p, a).Verdict != "UNKNOWN" {
			t.Fatal("stale")
		}
		a = ReferenceAC(rem)
		f = a.Facts[prerequisites[0]]
		f.Conflict = true
		a.Facts[prerequisites[0]] = f
		if Evaluate(p, a).Verdict != "UNKNOWN" {
			t.Fatal("conflict")
		}
		a = ReferenceAC(rem)
		f = a.Facts["fencing_enforced"]
		f.Evidence = ""
		a.Facts["fencing_enforced"] = f
		if Evaluate(p, a).Verdict != "UNKNOWN" {
			t.Fatal("insufficient")
		}
	}
	b, _ := json.Marshal(p)
	var altered PFC
	json.Unmarshal(b, &altered)
	altered.CausalSequence[0] = "other"
	if Valid(altered) {
		t.Fatal("semantic drift")
	}
}
func fixture(t *testing.T, rem bool) Evidence {
	t.Helper()
	p := testPFC(t)
	a := ReferenceAC(rem)
	m, e := Compile(p, a, Evaluate(p, a))
	if e != nil {
		t.Fatal(e)
	}
	ev := Evidence{Manifest: m, Safety: Plan(p, m), Events: expected(m), History: []History{{1, m.WorkerA, 1, "acquired", 0}, {4, m.WorkerA, 1, "expired", 1}, {5, m.WorkerB, 2, "acquired", 1}}, Effects: []Effect{{7, m.WorkerB, 2, "B-legitimate"}}, Attempts: []Attempt{{7, m.WorkerB, 2, true}, {10, m.WorkerA, 1, !rem}}, Resource: Resource{m.WorkerB, 2, 1, 2, "active"}, InitialStateEmpty: true, LedgerComplete: true, RunComplete: true}
	if !rem {
		ev.Effects = append(ev.Effects, Effect{10, m.WorkerA, 1, "A-stale"})
	}
	return ev
}
func clone(e Evidence) Evidence {
	b, _ := json.Marshal(e)
	var x Evidence
	json.Unmarshal(b, &x)
	return x
}
func TestAdversarialEvidence(t *testing.T) {
	for _, rem := range []bool{false, true} {
		base := fixture(t, rem)
		if verifyEvidence(base).Verdict == "UNKNOWN" {
			t.Fatal("fixture")
		}
		if VerifyAgainstDB(context.Background(), nil, base).Verdict != "UNKNOWN" {
			t.Fatal("evidence-only authority")
		}
		mutations := map[string]func(*Evidence){"PFC": func(e *Evidence) { e.Manifest.PFC = "other" }, "invariant": func(e *Evidence) { e.Manifest.Invariant = "other" }, "environment": func(e *Evidence) { e.Manifest.Environment = "production" }, "build": func(e *Evidence) { e.Manifest.Build = "other" }, "resource": func(e *Evidence) { e.Manifest.ResourceID = "other" }, "worker A": func(e *Evidence) { e.Manifest.WorkerA = "other" }, "worker B": func(e *Evidence) { e.Manifest.WorkerB = "other" }, "same token": func(e *Evidence) { e.Manifest.TokenB = 1 }, "lower token": func(e *Evidence) { e.Manifest.TokenB = 0 }, "fake token": func(e *Evidence) { e.Manifest.TokenB = 3 }, "missing expiry": func(e *Evidence) { e.Events = append(e.Events[:3], e.Events[4:]...) }, "missing acquisition": func(e *Evidence) { e.Events = append(e.Events[:4], e.Events[5:]...) }, "acquisition early": func(e *Evidence) { e.Events[3], e.Events[4] = e.Events[4], e.Events[3] }, "B commit missing": func(e *Evidence) { e.Effects = nil }, "resume missing": func(e *Evidence) { e.Events = append(e.Events[:7], e.Events[8:]...) }, "resume early": func(e *Evidence) { e.Events[5], e.Events[7] = e.Events[7], e.Events[5] }, "attempt missing": func(e *Evidence) { e.Attempts = e.Attempts[:1] }, "attempt new token": func(e *Evidence) { e.Attempts[1].Token = 2 }, "extra event": func(e *Evidence) { e.Events = append(e.Events, e.Events[9]) }, "duplicated event": func(e *Evidence) { e.Events[8] = e.Events[7] }, "wrong sequence": func(e *Evidence) { e.Events[8].Seq = 11 }, "safety": func(e *Evidence) { e.Safety.MaxSteps = 11 }, "adapter": func(e *Evidence) { e.Manifest.AdapterVersion = "other" }, "verifier": func(e *Evidence) { e.Manifest.VerifierVersion = "other" }, "dirty": func(e *Evidence) { e.InitialStateEmpty = false }, "incomplete": func(e *Evidence) { e.RunComplete = false }, "ledger": func(e *Evidence) { e.LedgerComplete = false }, "wrong effect worker": func(e *Evidence) { e.Effects[0].WorkerID = "other" }, "wrong effect token": func(e *Evidence) { e.Effects[0].Token = 1 }, "effect early": func(e *Evidence) { e.Effects[0].Seq = 4 }, "history": func(e *Evidence) { e.History[2].Token = 3 }}
		for name, mutate := range mutations {
			t.Run(name, func(t *testing.T) {
				e := clone(base)
				mutate(&e)
				if verifyEvidence(e).Verdict != "UNKNOWN" {
					t.Fatal("tampering accepted")
				}
			})
		}
	}
}
func TestPostgresEndToEnd(t *testing.T) {
	dsn := os.Getenv("FAILUREMESH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set PostgreSQL test URL")
	}
	db, e := phase1.Open(dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	ctx := context.Background()
	if e = Init(ctx, db); e != nil {
		t.Fatal(e)
	}
	p := testPFC(t)
	for _, rem := range []bool{false, true} {
		var first Proof
		for repeat := 0; repeat < 2; repeat++ {
			if e = Reset(ctx, db); e != nil {
				t.Fatal(e)
			}
			if VerifyAgainstDB(ctx, db, fixture(t, rem)).Verdict != "UNKNOWN" {
				t.Fatal("fabricated proof")
			}
			a := ReferenceAC(rem)
			m, e := Compile(p, a, Evaluate(p, a))
			if e != nil {
				t.Fatal(e)
			}
			s := Plan(p, m)
			ev := Run(ctx, db, m, s, rem)
			proof := VerifyAgainstDB(ctx, db, ev)
			want := "EXPOSED"
			if rem {
				want = "PROVEN_RESILIENT"
			}
			if proof.Verdict != want {
				t.Fatalf("%s: %s %s", want, proof.Verdict, proof.Reason)
			}
			if repeat == 1 && !reflect.DeepEqual(first, proof) {
				t.Fatal("nondeterministic")
			}
			first = proof
			dirty := Run(ctx, db, m, s, rem)
			if dirty.InitialStateEmpty || VerifyAgainstDB(ctx, db, dirty).Verdict != "UNKNOWN" {
				t.Fatal("dirty reset accepted")
			}
			_, e = db.ExecContext(ctx, "UPDATE lease_resources SET current_token=3 WHERE run_id=$1", m.RunID)
			if e != nil {
				t.Fatal(e)
			}
			if VerifyAgainstDB(ctx, db, ev).Verdict != "UNKNOWN" {
				t.Fatal("mutated authority accepted")
			}
			if e = Reset(ctx, db); e != nil {
				t.Fatal(e)
			}
		}
	}
}

func TestPostgresAuthorityMutations(t *testing.T) {
	dsn := os.Getenv("FAILUREMESH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set PostgreSQL test URL")
	}
	db, e := phase1.Open(dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	ctx := context.Background()
	p := testPFC(t)
	cases := []struct {
		name string
		rem  bool
		sql  string
	}{
		{"vulnerable stale effect missing", false, "DELETE FROM protected_effects WHERE seq=10"},
		{"vulnerable wrong worker", false, "UPDATE protected_effects SET worker_id='other' WHERE seq=10"},
		{"vulnerable wrong token", false, "UPDATE protected_effects SET fencing_token=2 WHERE seq=10"},
		{"vulnerable effect before takeover", false, "UPDATE protected_effects SET seq=4 WHERE seq=10"},
		{"remediated stale effect exists", true, "INSERT INTO protected_effects VALUES('pfc3-run','resource-1',10,'worker-A',1,'A-stale')"},
		{"remediated rejection missing", true, "DELETE FROM protected_attempts WHERE seq=10"},
		{"ownership history missing", true, "DELETE FROM lease_history WHERE step=5"},
		{"trace missing", true, "DELETE FROM lease_trace WHERE seq=9"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if e := Init(ctx, db); e != nil {
				t.Fatal(e)
			}
			if e := Reset(ctx, db); e != nil {
				t.Fatal(e)
			}
			a := ReferenceAC(tc.rem)
			m, e := Compile(p, a, Evaluate(p, a))
			if e != nil {
				t.Fatal(e)
			}
			ev := Run(ctx, db, m, Plan(p, m), tc.rem)
			if VerifyAgainstDB(ctx, db, ev).Verdict == "UNKNOWN" {
				t.Fatal("setup")
			}
			if _, e = db.ExecContext(ctx, tc.sql); e != nil {
				t.Fatal(e)
			}
			if VerifyAgainstDB(ctx, db, ev).Verdict != "UNKNOWN" {
				t.Fatal("mutated database earned positive verdict")
			}
		})
	}
}
