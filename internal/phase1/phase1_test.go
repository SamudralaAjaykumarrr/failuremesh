package phase1

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func pfc(t *testing.T) PFC {
	t.Helper()
	p, e := LoadPFC("../../contracts/timeout-after-external-commit.v1.json")
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestContractsAndMatching(t *testing.T) {
	p := pfc(t)
	a := ReferenceAC(false)
	if Evaluate(p, a).Verdict != "APPLICABLE" {
		t.Fatal("vulnerable AC")
	}
	if Evaluate(p, ReferenceAC(true)).Verdict != "APPLICABLE" {
		t.Fatal("control must remain under test")
	}
	a.Facts["external_effect"] = Fact{Value: Bool(false), Evidence: "observed absent", Source: "test", Quality: "direct"}
	if Evaluate(p, a).Verdict != "NOT_APPLICABLE" {
		t.Fatal("false prerequisite")
	}
	delete(a.Facts, "external_effect")
	if Evaluate(p, a).Verdict != "UNKNOWN" {
		t.Fatal("missing fact")
	}
	a = ReferenceAC(false)
	f := a.Facts["same_operation_retry"]
	f.Conflict = true
	a.Facts["same_operation_retry"] = f
	if Evaluate(p, a).Verdict != "UNKNOWN" {
		t.Fatal("conflict")
	}
	if _, e := LoadPFC("../../contracts/missing.json"); e == nil {
		t.Fatal("missing PFC accepted")
	}
}
func TestSafetyAndUnknown(t *testing.T) {
	p := pfc(t)
	a := ReferenceAC(false)
	m, e := Compile(p, a, Evaluate(p, a))
	if e != nil {
		t.Fatal(e)
	}
	s := Plan(p, m)
	if !s.Approved {
		t.Fatal(s)
	}
	bad := m
	bad.Environment = "production"
	if Plan(p, bad).Approved {
		t.Fatal("production approved")
	}
	if Run(context.Background(), nil, bad, s, false).Error != "safety plan rejected" {
		t.Fatal("unsafe run reached database")
	}
	if ValidateDSN("postgres://failuremesh@payments.example:5432/failuremesh") == nil {
		t.Fatal("remote database accepted")
	}
	if Verify(Evidence{Manifest: m, Safety: s}).Verdict != "UNKNOWN" {
		t.Fatal("empty evidence")
	}
	if _, e = Compile(p, a, Match{Verdict: "UNKNOWN"}); e == nil {
		t.Fatal("unknown compiled")
	}
}

func TestSyntheticConnectionConfig(t *testing.T) {
	for _, port := range []string{"5432", "55432"} {
		dsn := "postgres://failuremesh:synthetic-only@127.0.0.1:" + port + "/failuremesh?sslmode=disable"
		config, err := parseSyntheticConfig(dsn)
		if err != nil {
			t.Fatalf("authorized %s: %v", port, err)
		}
		if config.Host != "127.0.0.1" || config.Database != "failuremesh" || config.User != "failuremesh" || config.Password != "synthetic-only" || len(config.Fallbacks) != 0 {
			t.Fatalf("unexpected parsed config: host=%q database=%q fallbacks=%d", config.Host, config.Database, len(config.Fallbacks))
		}
	}
	base := "postgres://failuremesh:synthetic-only@127.0.0.1:5432/failuremesh?sslmode=disable"
	cases := map[string]string{
		"remote hostname":     "postgres://failuremesh:synthetic-only@payments.example:5432/failuremesh?sslmode=disable",
		"localhost name":      "postgres://failuremesh:synthetic-only@localhost:5432/failuremesh?sslmode=disable",
		"servicefile":         base + "&servicefile=/tmp/other.conf",
		"service":             base + "&service=other",
		"query host override": base + "&host=payments.example",
		"query port override": base + "&port=5432,5433",
		"alternate hosts":     "postgres://failuremesh:synthetic-only@127.0.0.1:5432,payments.example:5432/failuremesh?sslmode=disable",
		"passfile":            base + "&passfile=/tmp/other.pass",
		"sslkey":              base + "&sslkey=/tmp/key",
		"sslcert":             base + "&sslcert=/tmp/cert",
		"sslrootcert":         base + "&sslrootcert=/tmp/root",
		"runtime param":       base + "&search_path=other",
		"unapproved param":    base + "&future_option=value",
		"wrong database":      "postgres://failuremesh:synthetic-only@127.0.0.1:5432/other?sslmode=disable",
		"missing port":        "postgres://failuremesh:synthetic-only@127.0.0.1/failuremesh?sslmode=disable",
		"invalid port":        "postgres://failuremesh:synthetic-only@127.0.0.1:abc/failuremesh?sslmode=disable",
		"missing password":    "postgres://failuremesh@127.0.0.1:5432/failuremesh?sslmode=disable",
		"empty password":      "postgres://failuremesh:@127.0.0.1:5432/failuremesh?sslmode=disable",
		"wrong password":      "postgres://failuremesh:real-password@127.0.0.1:5432/failuremesh?sslmode=disable",
		"missing user":        "postgres://:synthetic-only@127.0.0.1:5432/failuremesh?sslmode=disable",
		"wrong user":          "postgres://other:synthetic-only@127.0.0.1:5432/failuremesh?sslmode=disable",
		"duplicate sslmode":   base + "&sslmode=require",
		"malformed query":     base + "&%zz=ignored",
		"keyword DSN":         "host=127.0.0.1 port=5432 dbname=failuremesh user=failuremesh sslmode=disable",
	}
	for name, dsn := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseSyntheticConfig(dsn); err == nil {
				t.Fatal("unsafe DSN accepted")
			}
		})
	}
	for _, key := range []string{"PGHOST", "PGSERVICE", "PGSERVICEFILE", "PGPASSFILE", "PGOPTIONS", "PGSSLCERT"} {
		t.Run("environment "+key, func(t *testing.T) {
			t.Setenv(key, "payments.example")
			if _, err := parseSyntheticConfig(base); err == nil {
				t.Fatal("PG environment setting accepted")
			}
		})
	}
}
func TestPostgresEndToEnd(t *testing.T) {
	dsn := os.Getenv("FAILUREMESH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set FAILUREMESH_TEST_DATABASE_URL for PostgreSQL integration")
	}
	db, e := Open(dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	ctx := context.Background()
	p := pfc(t)
	if e = Init(ctx, db); e != nil {
		t.Fatal(e)
	}
	for _, remediated := range []bool{false, true} {
		var first Proof
		for repeat := 0; repeat < 2; repeat++ {
			if e = Reset(ctx, db); e != nil {
				t.Fatal(e)
			}
			a := ReferenceAC(remediated)
			m, e := Compile(p, a, Evaluate(p, a))
			if e != nil {
				t.Fatal(e)
			}
			s := Plan(p, m)
			ev := Run(ctx, db, m, s, remediated)
			proof := Verify(ev)
			want := "EXPOSED"
			if remediated {
				want = "PROVEN_RESILIENT"
			}
			if proof.Verdict != want {
				t.Fatalf("%s: %s, %s", want, proof.Verdict, proof.Reason)
			}
			if len(ev.Effects) != 2 && !remediated {
				t.Fatal("missing duplicate")
			}
			if remediated && len(ev.Effects) != 1 {
				t.Fatal("remediation failed")
			}
			if repeat == 0 {
				first = proof
			} else if !reflect.DeepEqual(first, proof) {
				t.Fatal("nondeterministic proof")
			}
			missing := ev
			missing.LedgerComplete = false
			if Verify(missing).Verdict != "UNKNOWN" {
				t.Fatal("missing ledger accepted")
			}
			wrong := ev
			wrong.Events = append([]Event(nil), ev.Events...)
			for i := range wrong.Events {
				if wrong.Events[i].Kind == "response_lost_after_commit" {
					wrong.Events[i].Seq = 1
					break
				}
			}
			if Verify(wrong).Verdict != "UNKNOWN" {
				t.Fatal("conflicting ordering accepted")
			}
		}
	}
}

func fixture(t *testing.T, remediated bool) Evidence {
	t.Helper()
	p := pfc(t)
	a := ReferenceAC(remediated)
	m, err := Compile(p, a, Evaluate(p, a))
	if err != nil {
		t.Fatal(err)
	}
	e := Evidence{Manifest: m, Safety: Plan(p, m), LedgerComplete: true, RunComplete: true, InitialStateEmpty: true}
	e.add("caller_request", 1, 0)
	e.add("provider_commit", 1, 1)
	e.add("barrier_holds_success", 1, 1)
	e.add("response_lost_after_commit", 1, 1)
	e.add("caller_observed_loss", 1, 1)
	e.add("caller_retry", 2, 0)
	e.add("caller_request", 2, 0)
	secondID := int64(2)
	secondKind := "provider_commit"
	if remediated {
		secondID = 1
		secondKind = "provider_dedupe"
	}
	e.add(secondKind, 2, secondID)
	e.add("caller_success", 2, secondID)
	e.Effects = []Effect{{ID: 1, Attempt: 1, LogicalID: m.LogicalID}}
	if !remediated {
		e.Effects = append(e.Effects, Effect{ID: 2, Attempt: 2, LogicalID: m.LogicalID})
	}
	return e
}

func TestAdversarialEvidence(t *testing.T) {
	for _, remediated := range []bool{false, true} {
		base := fixture(t, remediated)
		want := "EXPOSED"
		if remediated {
			want = "PROVEN_RESILIENT"
		}
		if got := Verify(base).Verdict; got != want {
			t.Fatalf("valid fixture: %s", got)
		}
		mutations := map[string]func(*Evidence){
			"both requests attempt one": func(e *Evidence) { e.Events[6].Attempt = 1 },
			"retry wrong attempt":       func(e *Evidence) { e.Events[5].Attempt = 1 },
			"second provider before retry": func(e *Evidence) {
				e.Events[5], e.Events[7] = e.Events[7], e.Events[5]
				e.Events[5].Seq, e.Events[7].Seq = 6, 8
			},
			"second provider wrong attempt": func(e *Evidence) { e.Events[7].Attempt = 1 },
			"second provider wrong effect":  func(e *Evidence) { e.Events[7].EffectID = 77 },
			"success wrong effect":          func(e *Evidence) { e.Events[8].EffectID = 77 },
			"missing event":                 func(e *Evidence) { e.Events = append(e.Events[:3], e.Events[4:]...) },
			"duplicate event":               func(e *Evidence) { e.Events = append(e.Events, e.Events[3]) },
			"altered order": func(e *Evidence) {
				e.Events[3], e.Events[4] = e.Events[4], e.Events[3]
				e.Events[3].Seq, e.Events[4].Seq = 4, 5
			},
			"ledger attempt":    func(e *Evidence) { e.Effects[0].Attempt = 2 },
			"logical identity":  func(e *Evidence) { e.Events[7].LogicalID = "payment-Y" },
			"PFC scope":         func(e *Evidence) { e.Manifest.PFC = "pfc.other@1.0.0" },
			"manifest format":   func(e *Evidence) { e.Manifest.Version = "2" },
			"AC version":        func(e *Evidence) { e.Manifest.ACVersion = "2" },
			"AC digest missing": func(e *Evidence) { e.Manifest.ACDigest = "" },
			"AC digest wrong":   func(e *Evidence) { e.Manifest.ACDigest = "forged" },
			"invariant":         func(e *Evidence) { e.Manifest.Invariant = "something else" },
			"wrong build mode": func(e *Evidence) {
				if remediated {
					e.Manifest.Build = "reference-vulnerable-v1"
				} else {
					e.Manifest.Build = "reference-idempotent-v1"
				}
			},
			"adapter version":        func(e *Evidence) { e.Manifest.AdapterVersion = "other" },
			"verifier version":       func(e *Evidence) { e.Manifest.VerifierVersion = "other" },
			"retry timing":           func(e *Evidence) { e.Manifest.RetryTiming = "before observation" },
			"concurrency":            func(e *Evidence) { e.Manifest.Concurrency = 2 },
			"environment":            func(e *Evidence) { e.Manifest.Environment = "production" },
			"safety level":           func(e *Evidence) { e.Manifest.SafetyLevel = 5 },
			"run identity":           func(e *Evidence) { e.Manifest.RunID = "other-run" },
			"logical identity scope": func(e *Evidence) { e.Manifest.LogicalID = "other-payment" },
			"safety digest":          func(e *Evidence) { e.Safety.ManifestDigest = "wrong" },
			"safety attempts":        func(e *Evidence) { e.Safety.MaxAttempts = 3 },
			"safety duration":        func(e *Evidence) { e.Safety.MaxDurationSeconds = 100 },
			"observer drift":         func(e *Evidence) { e.Manifest.Observers[0] = "untrusted" },
			"reset drift":            func(e *Evidence) { e.Manifest.Reset = "none" },
		}
		for name, mutate := range mutations {
			t.Run(name+"/"+base.Manifest.Build, func(t *testing.T) {
				b, _ := json.Marshal(base)
				var e Evidence
				if err := json.Unmarshal(b, &e); err != nil {
					t.Fatal(err)
				}
				mutate(&e)
				if name != "safety digest" && name != "safety attempts" && name != "safety duration" {
					e.Safety.ManifestDigest = Digest(e.Manifest)
					e.Safety.Approved = true
				}
				if got := Verify(e).Verdict; got != "UNKNOWN" {
					t.Fatalf("forged bundle earned %s", got)
				}
			})
		}
	}
}

func TestPFCSemanticDrift(t *testing.T) {
	base := pfc(t)
	mutations := map[string]func(*PFC){
		"family":             func(p *PFC) { p.Family = "other" },
		"lifecycle":          func(p *PFC) { p.Lifecycle = "verified" },
		"verification level": func(p *PFC) { p.VerificationLevel = "L1 REPRODUCED" },
		"prerequisite":       func(p *PFC) { p.Prerequisites[0] = "other" },
		"causal sequence":    func(p *PFC) { p.CausalSequence[1] = "maybe committed" },
		"proof requirements": func(p *PFC) { p.ProofRequirements[0] = "optional ledger" },
		"control":            func(p *PFC) { p.ControlUnderTest = "none" },
		"forbidden outcome":  func(p *PFC) { p.ForbiddenOutcome = "retry" },
		"invariant":          func(p *PFC) { p.Invariant = "anything" },
		"fault":              func(p *PFC) { p.Experiment.Fault = "DROP_BEFORE_COMMIT" },
		"attempts":           func(p *PFC) { p.Experiment.Attempts = 3 },
		"concurrency":        func(p *PFC) { p.Experiment.Concurrency = 2 },
		"safety ceiling":     func(p *PFC) { p.SafetyCeiling = 5 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			b, _ := json.Marshal(base)
			var p PFC
			if err := json.Unmarshal(b, &p); err != nil {
				t.Fatal(err)
			}
			mutate(&p)
			data, _ := json.Marshal(p)
			path := t.TempDir() + "/pfc.json"
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadPFC(path); err == nil {
				t.Fatal("semantic drift accepted")
			}
		})
	}
}
