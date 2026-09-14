// Package iceberg16282 is an incident-specific, synthetic historical mechanism model.
package iceberg16282

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"reflect"
	"time"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

const Source = "https://github.com/apache/iceberg/issues/16282"
const SourceDigest = "b0e3a444070ad031df689f3a8caa33851c8e319a94a4fe6e99c8420f747299d8"
const Identity = "historical.iceberg-16282@1"
const Build = "iceberg-16282-reference-v1"
const EventID = "data-written-X"
const FileID = "file-X"
const Offset = 7 // Synthetic logical offset; the public source does not report one.

var causal = []string{"durable_table_commit", "separate_control_consumption", "control_event_replay", "commit_before_offset_advance", "shutdown_between_commit_and_offset", "recovery_from_committed_offset", "append_only", "no_equality_deletes", "non_idempotent_registration", "file_identity_observable", "snapshot_membership_observable", "control_offset_observable"}

type Fact = phase1.Fact
type AC struct {
	SchemaVersion string          `json:"schema_version"`
	ID            string          `json:"id"`
	Build         string          `json:"build"`
	Environment   string          `json:"environment"`
	Source        string          `json:"source"`
	Facts         map[string]Fact `json:"facts"`
	CatalogType   struct {
		Value    string   `json:"value"`
		Conflict []string `json:"conflict"`
	} `json:"catalog_type"`
}
type Decision struct {
	Verdict  string   `json:"verdict"`
	ACDigest string   `json:"ac_digest"`
	Reasons  []string `json:"reasons"`
}

func ReferenceAC() AC {
	a := AC{SchemaVersion: "1", ID: Identity, Build: Build, Environment: "local-synthetic", Source: Source, Facts: map[string]Fact{}}
	for _, k := range causal {
		a.Facts[k] = Fact{Value: phase1.Bool(true), Evidence: "public issue mechanism account mapped to deterministic local PostgreSQL fixture", Source: Source, Scope: "reported append-only coordinator recovery mechanism; local fixture", Quality: "reported", Sensitivity: "public/synthetic"}
	}
	a.CatalogType.Value = "UNKNOWN"
	a.CatalogType.Conflict = []string{"REST catalog reported in summary", "JDBC catalog reported in evidence section"}
	return a
}
func Evaluate(a AC) Decision {
	d := Decision{Verdict: "UNKNOWN", ACDigest: phase1.Digest(a)}
	if a.SchemaVersion != "1" || a.ID != Identity || a.Build != Build || a.Environment != "local-synthetic" || a.Source != Source {
		d.Reasons = append(d.Reasons, "architecture identity or source changed")
		return d
	}
	unknown, absent := false, false
	for _, k := range causal {
		f, ok := a.Facts[k]
		if !ok || f.Value == nil || f.Evidence == "" || f.Source != Source || f.Scope == "" || f.Quality != "reported" || f.Stale || f.Conflict {
			unknown = true
			d.Reasons = append(d.Reasons, k+": missing, stale, conflicting, or insufficient")
		} else if !*f.Value {
			absent = true
			d.Reasons = append(d.Reasons, k+": evidenced absence")
		}
	}
	if absent {
		d.Verdict = "NOT_APPLICABLE"
	} else if !unknown {
		d.Verdict = "APPLICABLE"
	}
	if d.Verdict == "APPLICABLE" {
		d.Reasons = []string{"causal prerequisites supported for bounded local model; catalog conflict is non-causal"}
	}
	return d
}

type Manifest struct {
	Identity        string    `json:"identity"`
	Source          string    `json:"source"`
	SourceDigest    string    `json:"source_digest"`
	ACDigest        string    `json:"ac_digest"`
	Build           string    `json:"build"`
	Environment     string    `json:"environment"`
	EventID         string    `json:"event_id"`
	FileID          string    `json:"file_id"`
	EventOffset     int       `json:"event_offset"`
	Snapshots       [2]string `json:"snapshots"`
	SafetyLevel     int       `json:"safety_level"`
	MaxSteps        int       `json:"max_steps"`
	VerifierVersion string    `json:"verifier_version"`
}
type PlanResult struct {
	Approved       bool   `json:"approved"`
	ManifestDigest string `json:"manifest_digest"`
	Reason         string `json:"reason"`
}

func Compile(a AC, d Decision, sourceDigest string) (Manifest, error) {
	if d.Verdict != "APPLICABLE" || d.ACDigest != phase1.Digest(a) || Evaluate(a).Verdict != "APPLICABLE" || sourceDigest != SourceDigest {
		return Manifest{}, errors.New("incident applicability or source identity insufficient")
	}
	return Manifest{Identity: Identity, Source: Source, SourceDigest: sourceDigest, ACDigest: d.ACDigest, Build: Build, Environment: "local-synthetic", EventID: EventID, FileID: FileID, EventOffset: Offset, Snapshots: [2]string{"snapshot-S1", "snapshot-S2"}, SafetyLevel: 1, MaxSteps: 11, VerifierVersion: "iceberg16282-v1"}, nil
}
func Plan(m Manifest) PlanResult {
	p := PlanResult{ManifestDigest: phase1.Digest(m), Reason: "outside bounded historical fixture"}
	if validManifest(m) {
		p.Approved = true
		p.Reason = "bounded local synthetic PostgreSQL model"
	}
	return p
}
func validManifest(m Manifest) bool {
	return m.Identity == Identity && m.Source == Source && m.SourceDigest == SourceDigest && m.ACDigest == phase1.Digest(ReferenceAC()) && m.Build == Build && m.Environment == "local-synthetic" && m.EventID == EventID && m.FileID == FileID && m.EventOffset == Offset && m.Snapshots == [2]string{"snapshot-S1", "snapshot-S2"} && m.SafetyLevel == 1 && m.MaxSteps == 11 && m.VerifierVersion == "iceberg16282-v1"
}
func validPlan(m Manifest, p PlanResult) bool {
	return validManifest(m) && p.Approved && p.ManifestDigest == phase1.Digest(m) && p.Reason == "bounded local synthetic PostgreSQL model"
}

//go:embed schema.sql
var schema embed.FS

func Init(ctx context.Context, db *sql.DB) error {
	b, _ := schema.ReadFile("schema.sql")
	_, e := db.ExecContext(ctx, string(b))
	return e
}
func Reset(ctx context.Context, db *sql.DB) error {
	_, e := db.ExecContext(ctx, "TRUNCATE historical_iceberg_trace, historical_iceberg_registrations, historical_iceberg_snapshots, historical_iceberg_consumptions, historical_iceberg_offset_history, historical_iceberg_offsets, historical_iceberg_events, historical_iceberg_runs")
	return e
}

type Trace struct {
	Step            int    `json:"step"`
	Kind            string `json:"kind"`
	Coordinator     string `json:"coordinator"`
	EventID         string `json:"event_id"`
	FileID          string `json:"file_id"`
	Snapshot        string `json:"snapshot"`
	CommittedOffset int    `json:"committed_offset"`
}
type Consumption struct {
	Step        int    `json:"step"`
	Coordinator string `json:"coordinator"`
	EventID     string `json:"event_id"`
	Offset      int    `json:"offset"`
	FileID      string `json:"file_id"`
}
type Registration struct {
	Snapshot    string `json:"snapshot"`
	FileID      string `json:"file_id"`
	EventID     string `json:"event_id"`
	Coordinator string `json:"coordinator"`
	Step        int    `json:"step"`
}
type Snapshot struct {
	ID          string `json:"id"`
	Coordinator string `json:"coordinator"`
	Step        int    `json:"step"`
}
type OffsetObservation struct {
	Step            int `json:"step"`
	CommittedOffset int `json:"committed_offset"`
}
type State struct {
	RunDigest       string              `json:"run_digest"`
	Complete        bool                `json:"complete"`
	EventCount      int                 `json:"event_count"`
	EventID         string              `json:"event_id"`
	EventFile       string              `json:"event_file"`
	EventOffset     int                 `json:"event_offset"`
	CommittedOffset int                 `json:"committed_offset"`
	OffsetHistory   []OffsetObservation `json:"offset_history"`
	Consumptions    []Consumption       `json:"consumptions"`
	Snapshots       []Snapshot          `json:"snapshots"`
	Registrations   []Registration      `json:"registrations"`
	Trace           []Trace             `json:"trace"`
	Counts          [8]int              `json:"counts"`
}
type Evidence struct {
	Manifest          Manifest   `json:"manifest"`
	Plan              PlanResult `json:"plan"`
	InitialStateEmpty bool       `json:"initial_state_empty"`
	RunComplete       bool       `json:"run_complete"`
	State             State      `json:"state"`
	Error             string     `json:"error,omitempty"`
}
type Element struct {
	Claim          string `json:"public_source_claim"`
	Representation string `json:"failuremesh_representation"`
	Observation    string `json:"reproduction_observation"`
	Status         string `json:"match_status"`
	Limitation     string `json:"limitation"`
}
type Result struct {
	Identity       string    `json:"identity"`
	Source         string    `json:"source"`
	SourceDigest   string    `json:"source_digest"`
	ACDigest       string    `json:"ac_digest"`
	EvidenceDigest string    `json:"evidence_digest"`
	Comparison     string    `json:"comparison"`
	Recommendation string    `json:"recommendation"`
	Reason         string    `json:"reason"`
	Elements       []Element `json:"elements"`
	Limitations    []string  `json:"limitations"`
}

var kinds = []string{"control_event_written", "coordinator_c1_consumes", "coordinator_c1_table_commit", "table_commit_confirmed", "control_offset_not_advanced", "coordinator_c1_shutdown", "coordinator_c2_started", "coordinator_c2_resumes_before_event", "coordinator_c2_reconsumes_same_event", "coordinator_c2_table_commit", "duplicate_file_registration_observed"}

func expectedTrace() []Trace {
	out := make([]Trace, 11)
	for i, k := range kinds {
		c := "coordinator-C1"
		if i >= 6 {
			c = "coordinator-C2"
		}
		s := ""
		if i == 2 || i == 3 {
			s = "snapshot-S1"
		}
		if i == 9 {
			s = "snapshot-S2"
		}
		out[i] = Trace{i + 1, k, c, EventID, FileID, s, Offset - 1}
	}
	return out
}

func countAll(ctx context.Context, db *sql.DB) ([8]int, error) {
	var c [8]int
	e := db.QueryRowContext(ctx, "SELECT (SELECT count(*) FROM historical_iceberg_runs),(SELECT count(*) FROM historical_iceberg_events),(SELECT count(*) FROM historical_iceberg_offsets),(SELECT count(*) FROM historical_iceberg_consumptions),(SELECT count(*) FROM historical_iceberg_snapshots),(SELECT count(*) FROM historical_iceberg_registrations),(SELECT count(*) FROM historical_iceberg_trace),(SELECT count(*) FROM historical_iceberg_offset_history)").Scan(&c[0], &c[1], &c[2], &c[3], &c[4], &c[5], &c[6], &c[7])
	return c, e
}
func Run(ctx context.Context, db *sql.DB, m Manifest, p PlanResult) Evidence {
	ev := Evidence{Manifest: m, Plan: p}
	if db == nil || !validPlan(m, p) {
		ev.Error = "database or safety plan unavailable"
		return ev
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if e := Init(ctx, db); e != nil {
		ev.Error = e.Error()
		return ev
	}
	c, e := countAll(ctx, db)
	if e != nil || c != [8]int{} {
		ev.Error = "dirty initial state; reset required"
		return ev
	}
	ev.InitialStateEmpty = true
	fail := func(e error) Evidence { ev.Error = e.Error(); return ev }
	exec := func(q string, args ...any) error { _, e := db.ExecContext(ctx, q, args...); return e }
	if e = exec("INSERT INTO historical_iceberg_runs VALUES($1,false)", phase1.Digest(m)); e != nil {
		return fail(e)
	}
	if e = exec("INSERT INTO historical_iceberg_events VALUES($1,$2,$3)", EventID, FileID, Offset); e != nil {
		return fail(e)
	}
	if e = exec("INSERT INTO historical_iceberg_offsets VALUES(1,$1)", Offset-1); e != nil {
		return fail(e)
	}
	for _, t := range expectedTrace() {
		// Each logical step commits independently. In particular, S1 is durable
		// before the offset-gap and coordinator-shutdown steps are recorded.
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fail(err)
		}
		stepExec := func(q string, args ...any) error { _, err := tx.ExecContext(ctx, q, args...); return err }
		switch t.Step {
		case 2, 9:
			e = stepExec("INSERT INTO historical_iceberg_consumptions VALUES($1,$2,$3,$4,$5)", t.Step, t.Coordinator, EventID, Offset, FileID)
		case 3, 10:
			e = stepExec("INSERT INTO historical_iceberg_snapshots VALUES($1,$2,$3)", t.Snapshot, t.Coordinator, t.Step)
			if e == nil {
				e = stepExec("INSERT INTO historical_iceberg_registrations VALUES($1,$2,$3,$4,$5)", t.Snapshot, FileID, EventID, t.Coordinator, t.Step)
			}
		}
		if e != nil {
			tx.Rollback()
			return fail(e)
		}
		if t.Step == 5 || t.Step == 6 || t.Step == 8 {
			e = stepExec("INSERT INTO historical_iceberg_offset_history SELECT $1,committed_offset FROM historical_iceberg_offsets WHERE singleton=1", t.Step)
			if e != nil {
				tx.Rollback()
				return fail(e)
			}
		}
		if e = stepExec("INSERT INTO historical_iceberg_trace VALUES($1,$2,$3,$4,$5,$6,$7)", t.Step, t.Kind, t.Coordinator, t.EventID, t.FileID, t.Snapshot, t.CommittedOffset); e != nil {
			tx.Rollback()
			return fail(e)
		}
		if e = tx.Commit(); e != nil {
			return fail(e)
		}
	}
	if e = exec("UPDATE historical_iceberg_runs SET complete=true WHERE manifest_digest=$1", phase1.Digest(m)); e != nil {
		return fail(e)
	}
	ev.RunComplete = true
	ev.State, e = readState(ctx, db)
	if e != nil {
		ev.Error = e.Error()
	}
	return ev
}

func readState(ctx context.Context, db *sql.DB) (State, error) {
	var s State
	var e error
	if db == nil {
		return s, errors.New("PostgreSQL unavailable")
	}
	e = db.QueryRowContext(ctx, "SELECT manifest_digest,complete FROM historical_iceberg_runs").Scan(&s.RunDigest, &s.Complete)
	if e != nil {
		return s, e
	}
	e = db.QueryRowContext(ctx, "SELECT count(*),coalesce(min(event_id),''),coalesce(min(file_id),''),coalesce(min(event_offset),0) FROM historical_iceberg_events").Scan(&s.EventCount, &s.EventID, &s.EventFile, &s.EventOffset)
	if e != nil {
		return s, e
	}
	e = db.QueryRowContext(ctx, "SELECT committed_offset FROM historical_iceberg_offsets WHERE singleton=1").Scan(&s.CommittedOffset)
	if e != nil {
		return s, e
	}
	rows, e := db.QueryContext(ctx, "SELECT step,committed_offset FROM historical_iceberg_offset_history ORDER BY step")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x OffsetObservation
		if e = rows.Scan(&x.Step, &x.CommittedOffset); e != nil {
			rows.Close()
			return s, e
		}
		s.OffsetHistory = append(s.OffsetHistory, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	rows, e = db.QueryContext(ctx, "SELECT step,coordinator,event_id,event_offset,file_id FROM historical_iceberg_consumptions ORDER BY step")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x Consumption
		if e = rows.Scan(&x.Step, &x.Coordinator, &x.EventID, &x.Offset, &x.FileID); e != nil {
			rows.Close()
			return s, e
		}
		s.Consumptions = append(s.Consumptions, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	rows, e = db.QueryContext(ctx, "SELECT snapshot_id,coordinator,step FROM historical_iceberg_snapshots ORDER BY step")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x Snapshot
		if e = rows.Scan(&x.ID, &x.Coordinator, &x.Step); e != nil {
			rows.Close()
			return s, e
		}
		s.Snapshots = append(s.Snapshots, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	rows, e = db.QueryContext(ctx, "SELECT snapshot_id,file_id,event_id,coordinator,step FROM historical_iceberg_registrations ORDER BY step")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x Registration
		if e = rows.Scan(&x.Snapshot, &x.FileID, &x.EventID, &x.Coordinator, &x.Step); e != nil {
			rows.Close()
			return s, e
		}
		s.Registrations = append(s.Registrations, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	rows, e = db.QueryContext(ctx, "SELECT step,kind,coordinator,event_id,file_id,snapshot_id,committed_offset FROM historical_iceberg_trace ORDER BY step")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x Trace
		if e = rows.Scan(&x.Step, &x.Kind, &x.Coordinator, &x.EventID, &x.FileID, &x.Snapshot, &x.CommittedOffset); e != nil {
			rows.Close()
			return s, e
		}
		s.Trace = append(s.Trace, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	s.Counts, e = countAll(ctx, db)
	return s, e
}

func compare(ev Evidence) Result {
	r := Result{Identity: Identity, Source: Source, SourceDigest: ev.Manifest.SourceDigest, ACDigest: ev.Manifest.ACDigest, EvidenceDigest: phase1.Digest(ev), Comparison: "UNKNOWN", Recommendation: "REVISE", Reason: "incomplete or inconsistent bounded evidence", Limitations: []string{"No exact Apache Iceberg 1.10.1 runtime, Kafka broker, GCS, or REST/JDBC catalog execution or conclusion.", "No real coordinator scheduling or reproduction of the reported approximately 20-minute race.", "No proof of every upstream causal detail, upstream fix, general family-10 portability, customer applicability, production readiness, or market validation.", "This is an incident-specific model, not a fourth generalized executable PFC."}}
	s := ev.State
	if ev.Error != "" {
		r.Reason = ev.Error
		return r
	}
	if !validPlan(ev.Manifest, ev.Plan) || !ev.InitialStateEmpty || !ev.RunComplete || !s.Complete || s.RunDigest != phase1.Digest(ev.Manifest) || s.EventCount != 1 || s.EventID != EventID || s.EventFile != FileID || s.EventOffset != Offset || s.CommittedOffset != Offset-1 || s.Counts != [8]int{1, 1, 1, 2, 2, 2, 11, 3} || !reflect.DeepEqual(s.OffsetHistory, []OffsetObservation{{5, Offset - 1}, {6, Offset - 1}, {8, Offset - 1}}) || !reflect.DeepEqual(s.Consumptions, []Consumption{{2, "coordinator-C1", EventID, Offset, FileID}, {9, "coordinator-C2", EventID, Offset, FileID}}) || !reflect.DeepEqual(s.Snapshots, []Snapshot{{"snapshot-S1", "coordinator-C1", 3}, {"snapshot-S2", "coordinator-C2", 10}}) || !reflect.DeepEqual(s.Registrations, []Registration{{"snapshot-S1", FileID, EventID, "coordinator-C1", 3}, {"snapshot-S2", FileID, EventID, "coordinator-C2", 10}}) || !reflect.DeepEqual(s.Trace, expectedTrace()) {
		return r
	}
	r.Comparison = "MATCH"
	r.Recommendation = "PASS"
	r.Reason = "same durable DATA_WRITTEN event replayed after an unchanged control offset; same file registered in two distinct PostgreSQL snapshots"
	r.Elements = []Element{
		{"One DATA_WRITTEN for file X", "one durable event and stable identity", "one event row; both consumption rows reference it", "MATCH", "synthetic event and offset"},
		{"C1 commits X into S1", "first snapshot and registration", "S1 registration at step 3", "MATCH", "reference table model"},
		{"offset not advanced before shutdown", "durable committed offset below event offset", "offset remains 6 through shutdown trace and final row", "MATCH", "no Kafka broker"},
		{"C2 resumes and reconsumes same event", "recovery and replay rows", "second consumption uses same event ID, offset, and file", "MATCH", "sequential logical scheduling"},
		{"same file appears in S1 and S2", "two authoritative registrations", "one file-X in each distinct snapshot", "MATCH", "row-count symptom not reproduced"},
	}
	return r
}

// VerifyAgainstDB is the sole exported positive comparison path. Captured evidence is
// only a claim until the full authoritative PostgreSQL state is re-read.
func VerifyAgainstDB(ctx context.Context, db *sql.DB, ev Evidence) Result {
	r := compare(ev)
	if r.Comparison != "MATCH" {
		return r
	}
	actual, e := readState(ctx, db)
	if e != nil || !reflect.DeepEqual(actual, ev.State) {
		r.Comparison = "UNKNOWN"
		r.Recommendation = "REVISE"
		r.Reason = "authoritative PostgreSQL state missing or changed"
		r.Elements = nil
		return r
	}
	return r
}

// CaptureCurrent returns an untrusted snapshot for a prior completed local run.
// Only VerifyAgainstDB can turn it into a positive result.
func CaptureCurrent(ctx context.Context, db *sql.DB, m Manifest, p PlanResult) Evidence {
	ev := Evidence{Manifest: m, Plan: p}
	s, e := readState(ctx, db)
	if e != nil {
		ev.Error = e.Error()
		return ev
	}
	ev.State = s
	ev.InitialStateEmpty = s.Complete
	ev.RunComplete = s.Complete
	return ev
}
