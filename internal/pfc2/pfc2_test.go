package pfc2

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
	p, e := Load("../../contracts/duplicate-queue-delivery.v1.json")
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func fixture(t *testing.T, remediated bool) Evidence {
	t.Helper()
	p := testPFC(t)
	a := ReferenceAC(remediated)
	m, e := Compile(p, a, Evaluate(p, a))
	if e != nil {
		t.Fatal(e)
	}
	ev := Evidence{Manifest: m, Safety: Plan(p, m), LedgerComplete: true, RunComplete: true, InitialStateEmpty: true}
	ev.add("message_published", 0, 0)
	ev.add("delivery_started", 1, 0)
	ev.add("consumer_received", 1, 0)
	ev.add("consumer_committed", 1, 1)
	ev.add("delivery_acknowledged", 1, 1)
	ev.add("duplicate_injected", 0, 0)
	ev.add("delivery_started", 2, 0)
	ev.add("consumer_received", 2, 0)
	if remediated {
		ev.add("consumer_deduped", 2, 1)
		ev.add("delivery_acknowledged", 2, 1)
	} else {
		ev.add("consumer_committed", 2, 2)
		ev.add("delivery_acknowledged", 2, 2)
	}
	ev.Effects = []Effect{{1, 1, m.MessageID}}
	if !remediated {
		ev.Effects = append(ev.Effects, Effect{2, 2, m.MessageID})
	}
	return ev
}

func TestApplicability(t *testing.T) {
	p := testPFC(t)
	for _, remediated := range []bool{false, true} {
		a := ReferenceAC(remediated)
		if got := Evaluate(p, a).Verdict; got != "APPLICABLE" {
			t.Fatal(got)
		}
		f := a.Facts["queue_delivery"]
		f.Value = phase1.Bool(false)
		a.Facts["queue_delivery"] = f
		if got := Evaluate(p, a).Verdict; got != "NOT_APPLICABLE" {
			t.Fatal(got)
		}
		a = ReferenceAC(remediated)
		f = a.Facts["stable_message_identity"]
		f.Stale = true
		a.Facts["stable_message_identity"] = f
		if got := Evaluate(p, a).Verdict; got != "UNKNOWN" {
			t.Fatal(got)
		}
		a = ReferenceAC(remediated)
		delete(a.Facts, "duplicate_possible")
		if got := Evaluate(p, a).Verdict; got != "UNKNOWN" {
			t.Fatal(got)
		}
		a = ReferenceAC(remediated)
		f = a.Facts["effect_ledger_observable"]
		f.Conflict = true
		a.Facts["effect_ledger_observable"] = f
		if got := Evaluate(p, a).Verdict; got != "UNKNOWN" {
			t.Fatal(got)
		}
		a = ReferenceAC(remediated)
		f = a.Facts["queue_delivery"]
		f.Value = phase1.Bool(false)
		a.Facts["queue_delivery"] = f
		delete(a.Facts, "stable_message_identity")
		if got := Evaluate(p, a).Verdict; got != "UNKNOWN" {
			t.Fatal("missing fact with false fact must remain UNKNOWN", got)
		}
		a = ReferenceAC(remediated)
		f = a.Facts["durable_consumer_idempotency"]
		f.Scope = ""
		a.Facts["durable_consumer_idempotency"] = f
		if got := Evaluate(p, a).Verdict; got != "UNKNOWN" {
			t.Fatal("insufficient control evidence accepted", got)
		}
	}
}

func TestSafetyPlanRejectsDrift(t *testing.T) {
	base := fixture(t, false)
	base.Manifest.Environment = "production"
	base.Safety.ManifestDigest = phase1.Digest(base.Manifest)
	if Plan(testPFC(t), base.Manifest).Approved || verifyEvidence(base).Verdict != "UNKNOWN" {
		t.Fatal("unsafe environment accepted")
	}
	if ev := Run(context.Background(), nil, base.Manifest, base.Safety, false); ev.Error == "" || ev.RunComplete {
		t.Fatal("unsafe plan reached database")
	}
}

func TestContractSemanticDrift(t *testing.T) {
	base := testPFC(t)
	mutations := map[string]func(*PFC){"identity": func(p *PFC) { p.ID = "other" }, "family": func(p *PFC) { p.Family = "other" }, "level": func(p *PFC) { p.VerificationLevel = "L1 REPRODUCED" }, "provenance": func(p *PFC) { p.Provenance = "other" }, "prerequisite": func(p *PFC) { p.Prerequisites[0] = "other" }, "control": func(p *PFC) { p.ControlUnderTest = "none" }, "sequence": func(p *PFC) { p.CausalSequence[0] = "other" }, "outcome": func(p *PFC) { p.ForbiddenOutcome = "other" }, "invariant": func(p *PFC) { p.Invariant = "other" }, "fault": func(p *PFC) { p.Experiment.Fault = "other" }, "attempts": func(p *PFC) { p.Experiment.Attempts = 3 }, "concurrency": func(p *PFC) { p.Experiment.Concurrency = 2 }, "proof": func(p *PFC) { p.ProofRequirements[0] = "other" }, "safety": func(p *PFC) { p.SafetyCeiling = 5 }}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			b, _ := json.Marshal(base)
			var p PFC
			json.Unmarshal(b, &p)
			mutate(&p)
			if Valid(p) {
				t.Fatal("same-version semantic drift accepted")
			}
		})
	}
}

func TestVerifierAdversarial(t *testing.T) {
	for _, remediated := range []bool{false, true} {
		base := fixture(t, remediated)
		want := "EXPOSED"
		if remediated {
			want = "PROVEN_RESILIENT"
		}
		if got := verifyEvidence(base).Verdict; got != want {
			t.Fatalf("valid fixture %s", got)
		}
		if got := VerifyAgainstDB(context.Background(), nil, base).Verdict; got != "UNKNOWN" {
			t.Fatalf("evidence without PostgreSQL authority earned %s", got)
		}
		mutations := map[string]func(*Evidence){
			"message id": func(e *Evidence) { e.Events[7].MessageID = "message-Y" }, "missing duplicate": func(e *Evidence) { e.Events = append(e.Events[:5], e.Events[6:]...) }, "wrong duplicate order": func(e *Evidence) { e.Events[4], e.Events[5] = e.Events[5], e.Events[4] }, "missing delivery two": func(e *Evidence) { e.Events = append(e.Events[:6], e.Events[7:]...) }, "attempt one twice": func(e *Evidence) { e.Events[6].Attempt = 1 }, "delivery identity": func(e *Evidence) { e.Events[6].DeliveryID = "delivery-1" }, "effect id": func(e *Evidence) { e.Events[3].EffectID = 77 }, "ledger attempt": func(e *Evidence) { e.Effects[0].Attempt = 2 }, "second effect before fault": func(e *Evidence) { e.Events[5], e.Events[8] = e.Events[8], e.Events[5] }, "dedupe wrong effect": func(e *Evidence) { e.Events[8].EffectID = 77 }, "extra event": func(e *Evidence) { e.Events = append(e.Events, e.Events[9]) }, "missing event": func(e *Evidence) { e.Events = e.Events[:9] }, "duplicated event": func(e *Evidence) { e.Events[7] = e.Events[6] }, "PFC": func(e *Evidence) { e.Manifest.PFC = "other" }, "invariant": func(e *Evidence) { e.Manifest.Invariant = "other" }, "build": func(e *Evidence) {
				if remediated {
					e.Manifest.Build = "queue-vulnerable-v1"
				} else {
					e.Manifest.Build = "queue-idempotent-v1"
				}
			}, "environment": func(e *Evidence) { e.Manifest.Environment = "production" }, "safety bounds": func(e *Evidence) { e.Safety.MaxDeliveries = 3 }, "adapter": func(e *Evidence) { e.Manifest.AdapterVersion = "other" }, "verifier": func(e *Evidence) { e.Manifest.VerifierVersion = "other" }, "ledger incomplete": func(e *Evidence) { e.LedgerComplete = false }, "ledger row missing": func(e *Evidence) { e.Effects = e.Effects[:0] }, "run incomplete": func(e *Evidence) { e.RunComplete = false }, "dirty initial state": func(e *Evidence) { e.InitialStateEmpty = false }, "fault": func(e *Evidence) { e.Manifest.Fault = "other" }, "concurrency": func(e *Evidence) { e.Manifest.Concurrency = 2 }, "safety plan": func(e *Evidence) { e.Safety.ManifestDigest = "other" },
		}
		for name, mutate := range mutations {
			t.Run(name+base.Manifest.Build, func(t *testing.T) {
				b, _ := json.Marshal(base)
				var e Evidence
				json.Unmarshal(b, &e)
				mutate(&e)
				if name != "safety plan" {
					e.Safety.ManifestDigest = phase1.Digest(e.Manifest)
				}
				if got := verifyEvidence(e).Verdict; got != "UNKNOWN" {
					t.Fatalf("tampered evidence earned %s", got)
				}
			})
		}
	}
}

func TestPostgresEndToEnd(t *testing.T) {
	dsn := os.Getenv("FAILUREMESH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set FAILUREMESH_TEST_DATABASE_URL for PostgreSQL integration")
	}
	db, err := phase1.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err = Init(ctx, db); err != nil {
		t.Fatal(err)
	}
	p := testPFC(t)
	for _, remediated := range []bool{false, true} {
		var first Proof
		for repeat := 0; repeat < 2; repeat++ {
			if err = Reset(ctx, db); err != nil {
				t.Fatal(err)
			}
			fabricated := fixture(t, remediated)
			if got := VerifyAgainstDB(ctx, db, fabricated).Verdict; got != "UNKNOWN" {
				t.Fatalf("fabricated evidence without PostgreSQL effects earned %s", got)
			}
			a := ReferenceAC(remediated)
			m, err := Compile(p, a, Evaluate(p, a))
			if err != nil {
				t.Fatal(err)
			}
			s := Plan(p, m)
			ev := Run(ctx, db, m, s, remediated)
			proof := VerifyAgainstDB(ctx, db, ev)
			want := "EXPOSED"
			if remediated {
				want = "PROVEN_RESILIENT"
			}
			if proof.Verdict != want {
				t.Fatalf("%s: %s %s", want, proof.Verdict, proof.Reason)
			}
			if len(ev.Effects) != 2 && !remediated || len(ev.Effects) != 1 && remediated {
				t.Fatal("wrong durable ledger")
			}
			if ev.Events[5].Kind != "duplicate_injected" || ev.Events[7].MessageID != m.MessageID {
				t.Fatal("duplicate fault not delivered")
			}
			if repeat == 0 {
				first = proof
			} else if !reflect.DeepEqual(first, proof) {
				t.Fatal("nondeterministic proof")
			}
			if remediated {
				if _, err := db.ExecContext(ctx, "DELETE FROM consumer_idempotency WHERE run_id=$1 AND message_id=$2", m.RunID, m.MessageID); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, err := db.ExecContext(ctx, "UPDATE consumer_effects SET delivery_attempt=1 WHERE run_id=$1 AND delivery_attempt=2", m.RunID); err != nil {
					t.Fatal(err)
				}
			}
			if got := VerifyAgainstDB(ctx, db, ev).Verdict; got != "UNKNOWN" {
				t.Fatalf("changed PostgreSQL state earned %s", got)
			}
			dirty := Run(ctx, db, m, s, remediated)
			if dirty.InitialStateEmpty || verifyEvidence(dirty).Verdict != "UNKNOWN" {
				t.Fatal("dirty reset accepted")
			}
		}
	}
}
