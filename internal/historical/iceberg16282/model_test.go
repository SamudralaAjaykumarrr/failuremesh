package iceberg16282

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
	if got := Evaluate(a).Verdict; got != "APPLICABLE" {
		t.Fatal(got)
	}
	if a.CatalogType.Value != "UNKNOWN" || len(a.CatalogType.Conflict) != 2 {
		t.Fatal("catalog conflict erased")
	}
	for _, mode := range []string{"missing", "stale", "conflict", "insufficient", "absent"} {
		t.Run(mode, func(t *testing.T) {
			b := ReferenceAC()
			f := b.Facts["commit_before_offset_advance"]
			switch mode {
			case "missing":
				delete(b.Facts, "commit_before_offset_advance")
			case "stale":
				f.Stale = true
			case "conflict":
				f.Conflict = true
			case "insufficient":
				f.Evidence = ""
			case "absent":
				f.Value = phase1.Bool(false)
			}
			if mode != "missing" {
				b.Facts["commit_before_offset_advance"] = f
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
	a.CatalogType.Conflict = append(a.CatalogType.Conflict, "still unresolved")
	if Evaluate(a).Verdict != "APPLICABLE" {
		t.Fatal("non-causal conflict changed match")
	}
}
func TestRunAndRepeatability(t *testing.T) {
	ctx, db, m, p := fixture(t)
	ev := positive(t, ctx, db, m, p)
	if ev.State.EventID != EventID || ev.State.EventFile != FileID || ev.State.CommittedOffset >= ev.State.EventOffset || ev.State.Consumptions[0].EventID != ev.State.Consumptions[1].EventID || ev.State.Registrations[0].FileID != ev.State.Registrations[1].FileID || ev.State.Snapshots[0] == ev.State.Snapshots[1] {
		t.Fatal("causal identity invariant")
	}
	dirty := Run(ctx, db, m, p)
	if !strings.Contains(dirty.Error, "dirty initial state") {
		t.Fatalf("dirty run: %+v", dirty)
	}
	if e := Reset(ctx, db); e != nil {
		t.Fatal(e)
	}
	again := positive(t, ctx, db, m, p)
	if phase1.Digest(ev.State) != phase1.Digest(again.State) {
		t.Fatal("non-deterministic durable state")
	}
}
func TestFabricatedAndChangedEvidence(t *testing.T) {
	ctx, db, m, p := fixture(t)
	fabricated := Evidence{Manifest: m, Plan: p, InitialStateEmpty: true, RunComplete: true, State: State{Complete: true}}
	if VerifyAgainstDB(ctx, db, fabricated).Comparison == "MATCH" {
		t.Fatal("fabrication accepted")
	}
	ev := positive(t, ctx, db, m, p)
	if _, e := db.ExecContext(ctx, "UPDATE historical_iceberg_offsets SET committed_offset=7"); e != nil {
		t.Fatal(e)
	}
	if VerifyAgainstDB(ctx, db, ev).Comparison == "MATCH" {
		t.Fatal("changed database accepted")
	}
}
func TestAdversarialDBMutations(t *testing.T) {
	cases := map[string]string{
		"different event on replay":    "INSERT INTO historical_iceberg_events VALUES('other','file-X',8); UPDATE historical_iceberg_consumptions SET event_id='other' WHERE step=9",
		"different file on replay":     "UPDATE historical_iceberg_consumptions SET file_id='file-Y' WHERE step=9",
		"S1 missing":                   "DELETE FROM historical_iceberg_registrations WHERE snapshot_id='snapshot-S1'",
		"S2 missing":                   "DELETE FROM historical_iceberg_registrations WHERE snapshot_id='snapshot-S2'",
		"same snapshot":                "UPDATE historical_iceberg_trace SET snapshot_id='snapshot-S1' WHERE step=10",
		"offset advanced":              "UPDATE historical_iceberg_offset_history SET committed_offset=7 WHERE step=6",
		"C2 after event":               "UPDATE historical_iceberg_offset_history SET committed_offset=7 WHERE step=8",
		"C2 no reconsume":              "DELETE FROM historical_iceberg_consumptions WHERE step=9",
		"same file not in both":        "UPDATE historical_iceberg_registrations SET file_id='file-Y' WHERE step=10",
		"duplicate without replay":     "DELETE FROM historical_iceberg_consumptions WHERE step=9",
		"replay without second commit": "DELETE FROM historical_iceberg_registrations WHERE step=10",
		"trace incomplete":             "DELETE FROM historical_iceberg_trace WHERE step=8",
		"run incomplete":               "UPDATE historical_iceberg_runs SET complete=false",
		"extra registration":           "INSERT INTO historical_iceberg_snapshots VALUES('snapshot-S3','coordinator-C2',12); INSERT INTO historical_iceberg_registrations VALUES('snapshot-S3','file-X','data-written-X','coordinator-C2',12)",
	}
	for name, query := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, db, m, p := fixture(t)
			ev := positive(t, ctx, db, m, p)
			for _, q := range strings.Split(query, "; ") {
				if _, e := db.ExecContext(ctx, q); e != nil {
					t.Fatal(e)
				}
			}
			if got := VerifyAgainstDB(ctx, db, ev).Comparison; got == "MATCH" {
				t.Fatal("mutation accepted")
			}
			altered := CaptureCurrent(ctx, db, m, p)
			if got := VerifyAgainstDB(ctx, db, altered).Comparison; got == "MATCH" {
				t.Fatal("fresh mutated evidence accepted")
			}
		})
	}
}
func TestSourceAndManifestIdentity(t *testing.T) {
	b, err := os.ReadFile("../../../docs/validation/historical-reproductions/001-iceberg-16282/source-record.md")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != SourceDigest {
		t.Fatal("source record missing or changed")
	}
	a := ReferenceAC()
	d := Evaluate(a)
	if _, e := Compile(a, d, "changed"); e == nil {
		t.Fatal("source change accepted")
	}
	ctx, db, m, p := fixture(t)
	ev := positive(t, ctx, db, m, p)
	ev.Manifest.SourceDigest = "changed"
	if VerifyAgainstDB(ctx, db, ev).Comparison == "MATCH" {
		t.Fatal("manifest source changed accepted")
	}
}
