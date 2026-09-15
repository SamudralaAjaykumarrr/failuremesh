package githubmay4

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

func fixture(t *testing.T) (context.Context, *sql.DB, Manifest, PlanResult) {
	t.Helper()
	dsn := os.Getenv("FAILUREMESH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("real PostgreSQL integration URL not set")
	}
	db, e := phase1.Open(dsn)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	if e = Init(ctx, db); e != nil {
		t.Fatal(e)
	}
	if e = Reset(ctx, db); e != nil {
		t.Fatal(e)
	}
	a := ReferenceAC()
	m, e := Compile(a, Evaluate(a), SourceDigest)
	if e != nil {
		t.Fatal(e)
	}
	return ctx, db, m, Plan(m)
}
func positive(t *testing.T, ctx context.Context, db *sql.DB, m Manifest, p PlanResult) Evidence {
	t.Helper()
	ev := Run(ctx, db, m, p)
	if ev.Error != "" {
		t.Fatal(ev.Error)
	}
	r := VerifyAgainstDB(ctx, db, ev)
	if r.Comparison != "MATCH" || r.Recommendation != "PASS" {
		t.Fatalf("%+v", r)
	}
	return ev
}
func TestApplicability(t *testing.T) {
	a := ReferenceAC()
	if Evaluate(a).Verdict != "APPLICABLE" || len(a.QuantitativeUnknowns) == 0 {
		t.Fatal("reference AC")
	}
	for _, k := range causal {
		for _, mode := range []string{"missing", "stale", "conflict", "insufficient", "wrong source", "wrong quality", "wrong scope", "absent"} {
			t.Run(k+"/"+mode, func(t *testing.T) {
				b := ReferenceAC()
				f := b.Facts[k]
				switch mode {
				case "missing":
					delete(b.Facts, k)
				case "stale":
					f.Stale = true
				case "conflict":
					f.Conflict = true
				case "insufficient":
					f.Evidence = ""
				case "wrong source":
					f.Source = "unsupported"
				case "wrong quality":
					if f.Quality == "direct" {
						f.Quality = "reported"
					} else {
						f.Quality = "direct"
					}
				case "wrong scope":
					f.Scope = "exact GitHub telemetry"
				case "absent":
					f.Value = phase1.Bool(false)
				}
				if mode != "missing" {
					b.Facts[k] = f
				}
				want := "UNKNOWN"
				if mode == "absent" {
					want = "NOT_APPLICABLE"
				}
				if got := Evaluate(b).Verdict; got != want {
					t.Fatalf("%s != %s", got, want)
				}
			})
		}
	}
	a.QuantitativeUnknowns = append(a.QuantitativeUnknowns, "exact_pool_count_still_unknown")
	if Evaluate(a).Verdict != "APPLICABLE" {
		t.Fatal("noncausal quantitative unknown blocked service-level applicability")
	}
}
func TestProvenanceAndUnknowns(t *testing.T) {
	a := ReferenceAC()
	local := map[string]bool{"requests_wait_on_capacity": true, "timeout_state_observable": true, "capacity_state_observable": true, "local_dependency_graph": true}
	for k, f := range a.Facts {
		if local[k] {
			if f.Quality != "direct" || f.Source != localSource || f.Sensitivity != "synthetic" {
				t.Fatalf("local capability mislabeled: %s: %+v", k, f)
			}
		} else if f.Quality != "reported" || f.Source != Source || f.Sensitivity != "public" {
			t.Fatalf("reported prerequisite mislabeled: %s: %+v", k, f)
		}
	}
	for _, k := range []string{"exact_connection_pool_size", "exact_occupancy", "exact_request_deadlines", "exact_queue_lengths", "complete_per_request_topology", "complete_traces", "exact_retry_policy"} {
		found := false
		for _, unknown := range a.QuantitativeUnknowns {
			if unknown == k {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing UNKNOWN: %s", k)
		}
		if _, ok := a.Facts[k]; ok {
			t.Fatalf("unknown converted to fact: %s", k)
		}
	}
}

func TestRunRepeatabilityAndDirty(t *testing.T) {
	ctx, db, m, p := fixture(t)
	ev := positive(t, ctx, db, m, p)
	if ev.State.History[3].Available != 0 || ev.State.History[6].Available != 2 {
		t.Fatal("capacity ledger")
	}
	if got := Run(ctx, db, m, p).Error; !strings.Contains(got, "dirty initial state") {
		t.Fatal(got)
	}
	if e := Reset(ctx, db); e != nil {
		t.Fatal(e)
	}
	again := positive(t, ctx, db, m, p)
	if phase1.Digest(ev.State) != phase1.Digest(again.State) {
		t.Fatal("non-deterministic state")
	}
}
func TestFabricationAndChangedDB(t *testing.T) {
	ctx, db, m, p := fixture(t)
	fake := Evidence{Manifest: m, Plan: p, InitialStateEmpty: true, RunComplete: true, State: State{Complete: true}}
	if VerifyAgainstDB(ctx, db, fake).Comparison == "MATCH" {
		t.Fatal("fabrication accepted")
	}
	ev := positive(t, ctx, db, m, p)
	if _, e := db.ExecContext(ctx, "UPDATE historical_github_capacity_history SET available=1 WHERE tick=3"); e != nil {
		t.Fatal(e)
	}
	if VerifyAgainstDB(ctx, db, ev).Comparison == "MATCH" {
		t.Fatal("changed DB accepted")
	}
}
func TestAdversarialMutations(t *testing.T) {
	cases := map[string]string{
		"missing dependency":        "DELETE FROM historical_github_dependencies WHERE from_actor='pull-request-service'",
		"independent second domain": "UPDATE historical_github_requests SET domain='other-db' WHERE id='secondary-1'",
		"migration absent":          "DELETE FROM historical_github_workloads WHERE id='migration-M'",
		"production absent":         "DELETE FROM historical_github_workloads WHERE id='normal-production'",
		"no overlap":                "UPDATE historical_github_workloads SET start_tick=6 WHERE id='normal-production'",
		"not finite":                "UPDATE historical_github_capacity SET units=100",
		"never saturated":           "UPDATE historical_github_capacity_history SET available=1 WHERE tick=3",
		"early arrival":             "UPDATE historical_github_requests SET arrival_tick=1 WHERE id='pull-1'",
		"wrong dependency":          "UPDATE historical_github_dependencies SET to_actor='other-db' WHERE from_actor='secondary-service'",
		"no wait":                   "UPDATE historical_github_requests SET wait_tick=4 WHERE id='pull-1'",
		"deadline not crossed":      "UPDATE historical_github_requests SET deadline_tick=8 WHERE id='pull-1'",
		"one service":               "DELETE FROM historical_github_requests WHERE id='secondary-1'",
		"dependent relation absent": "DELETE FROM historical_github_dependencies WHERE from_actor='dependent-service'",
		"fake timeout":              "UPDATE historical_github_capacity_history SET available=1, occupied=3 WHERE tick=3",
		"fake saturation":           "DELETE FROM historical_github_requests WHERE id='pull-1'",
		"early migration release":   "UPDATE historical_github_workloads SET release_tick=4 WHERE id='migration-M'",
		"recovery fails":            "UPDATE historical_github_requests SET outcome='TIMEOUT' WHERE id='recovery-1'",
		"bad arithmetic":            "UPDATE historical_github_capacity_history SET occupied=3 WHERE tick=3",
		"impossible occupancy":      "UPDATE historical_github_capacity_history SET occupied=5 WHERE tick=3",
		"missing history":           "DELETE FROM historical_github_capacity_history WHERE tick=3",
		"missing trace":             "DELETE FROM historical_github_trace WHERE step=6",
		"extra request":             "INSERT INTO historical_github_requests VALUES('extra','pull-request-service','primary-db',3,4,3,5,'TIMEOUT')",
		"extra service":             "INSERT INTO historical_github_dependencies VALUES('extra-service','primary-db','primary-db')",
		"incomplete":                "UPDATE historical_github_runs SET complete=false",
	}
	for name, q := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, db, m, p := fixture(t)
			ev := positive(t, ctx, db, m, p)
			if strings.Contains(q, "other-db") {
				if _, e := db.ExecContext(ctx, "INSERT INTO historical_github_capacity VALUES('other-db',4)"); e != nil {
					t.Fatal(e)
				}
			}
			if _, e := db.ExecContext(ctx, q); e != nil {
				t.Fatal(e)
			}
			if VerifyAgainstDB(ctx, db, ev).Comparison == "MATCH" {
				t.Fatal("changed state accepted")
			}
			if VerifyAgainstDB(ctx, db, CaptureCurrent(ctx, db, m, p)).Comparison == "MATCH" {
				t.Fatal("fresh corrupted state accepted")
			}
		})
	}
}
func TestSourceAndManifest(t *testing.T) {
	b, e := os.ReadFile("../../../docs/validation/historical-reproductions/002-github-may4-cascade/source-record.md")
	if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != SourceDigest {
		t.Fatal("source digest changed")
	}
	a := ReferenceAC()
	if _, e := Compile(a, Evaluate(a), "changed"); e == nil {
		t.Fatal("source change accepted")
	}
	ctx, db, m, p := fixture(t)
	ev := positive(t, ctx, db, m, p)
	ev.Manifest.SourceDigest = "changed"
	if VerifyAgainstDB(ctx, db, ev).Comparison == "MATCH" {
		t.Fatal("manifest change accepted")
	}
}
