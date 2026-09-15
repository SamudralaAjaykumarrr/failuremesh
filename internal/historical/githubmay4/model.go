// Package githubmay4 is a bounded, synthetic service-level historical mechanism model.
package githubmay4

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"reflect"
	"time"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

const Source = "https://github.blog/news-insights/company-news/github-availability-report-may-2026/"
const SourceDigest = "227daf8d66e99c38ee3092ab377689d85f11f2487b17fed0fe9889995743a019"
const Identity = "historical.github-may4-cascade@1"
const Build = "github-may4-local-capacity-v1"
const Domain = "primary-db"

var causal = []string{"shared_critical_db_dependency", "finite_capacity_domain", "online_migration_consumes_capacity", "production_traffic_consumes_same_capacity", "migration_traffic_overlap", "capacity_can_saturate", "requests_wait_on_capacity", "latency_can_exceed_deadline", "multiple_services_share_dependency", "timeout_state_observable", "capacity_state_observable", "migration_can_be_paused", "local_dependency_graph"}

type Fact = phase1.Fact
type AC struct {
	SchemaVersion        string          `json:"schema_version"`
	ID                   string          `json:"id"`
	Build                string          `json:"build"`
	Environment          string          `json:"environment"`
	Source               string          `json:"source"`
	Scope                string          `json:"scope"`
	Facts                map[string]Fact `json:"facts"`
	QuantitativeUnknowns []string        `json:"quantitative_unknowns"`
}
type Decision struct {
	Verdict  string   `json:"verdict"`
	ACDigest string   `json:"ac_digest"`
	Scope    string   `json:"scope"`
	Reasons  []string `json:"reasons"`
}

// Local capabilities are directly supported by this fixture, not GitHub telemetry.
const localSource = "internal/historical/githubmay4/model.go and schema.sql"

func prerequisiteEvidence(k string) (source, quality, scope string) {
	switch k {
	case "requests_wait_on_capacity", "timeout_state_observable", "capacity_state_observable", "local_dependency_graph":
		return localSource, "direct", "synthetic local fixture capability; FailureMesh model choice"
	default:
		return Source, "reported", "GitHub qualitative service-level causal account"
	}
}

func ReferenceAC() AC {
	a := AC{SchemaVersion: "1", ID: Identity, Build: Build, Environment: "local-synthetic", Source: Source, Scope: "SERVICE-LEVEL CASCADE", Facts: map[string]Fact{}, QuantitativeUnknowns: []string{"exact_connection_pool_size", "exact_occupancy", "exact_request_deadlines", "exact_queue_lengths", "complete_per_request_topology", "complete_traces", "exact_retry_policy"}}
	evidence := map[string]string{
		"shared_critical_db_dependency":             "GitHub attributes degradation to shared data dependencies and primary database contention",
		"finite_capacity_domain":                    "GitHub reports database connection capacity saturation; exact capacity UNKNOWN",
		"online_migration_consumes_capacity":        "GitHub identifies online schema migration load as a contributor",
		"production_traffic_consumes_same_capacity": "GitHub attributes saturation to combined migration and production load",
		"migration_traffic_overlap":                 "GitHub reports migration running as traffic increased toward weekly peak",
		"capacity_can_saturate":                     "GitHub reports database connection capacity saturated",
		"latency_can_exceed_deadline":               "GitHub reports query contention and cascading timeouts; exact deadlines UNKNOWN",
		"multiple_services_share_dependency":        "GitHub reports multiple services degraded through shared data dependencies; complete graph UNKNOWN",
		"migration_can_be_paused":                   "GitHub reports responders paused the contributing migration",
		"requests_wait_on_capacity":                 "FailureMesh represents contention with deterministic bounded waiting rows and logical ticks",
		"timeout_state_observable":                  "Local PostgreSQL request rows record deadlines, decision ticks, and timeout outcomes",
		"capacity_state_observable":                 "Local PostgreSQL capacity history records synthetic occupancy and availability",
		"local_dependency_graph":                    "FailureMesh fixture defines explicit actor edges to primary-db; not GitHub internal topology",
	}
	for _, k := range causal {
		source, quality, scope := prerequisiteEvidence(k)
		sensitivity := "public"
		if quality == "direct" {
			sensitivity = "synthetic"
		}
		a.Facts[k] = Fact{Value: phase1.Bool(true), Evidence: evidence[k], Source: source, Scope: scope, Quality: quality, Sensitivity: sensitivity}
	}
	return a
}
func Evaluate(a AC) Decision {
	d := Decision{Verdict: "UNKNOWN", ACDigest: phase1.Digest(a), Scope: "SERVICE-LEVEL CASCADE"}
	if a.SchemaVersion != "1" || a.ID != Identity || a.Build != Build || a.Environment != "local-synthetic" || a.Source != Source || a.Scope != d.Scope {
		d.Reasons = []string{"architecture identity or scope changed"}
		return d
	}
	unknown, absent := false, false
	for _, k := range causal {
		f, ok := a.Facts[k]
		source, quality, scope := prerequisiteEvidence(k)
		if !ok || f.Value == nil || f.Evidence == "" || f.Source != source || f.Scope != scope || f.Quality != quality || f.Stale || f.Conflict {
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
		d.Reasons = []string{"first-party report supports qualitative service-level cascade; local fixture separately supplies waiting, graph, and observability capabilities; exact public topology, traces, retry policy, and quantities remain unknown"}
	}
	return d
}

type Manifest struct {
	Identity        string `json:"identity"`
	Source          string `json:"source"`
	SourceDigest    string `json:"source_digest"`
	ACDigest        string `json:"ac_digest"`
	Build           string `json:"build"`
	Environment     string `json:"environment"`
	Scope           string `json:"scope"`
	Capacity        int    `json:"synthetic_capacity_units"`
	MigrationUnits  int    `json:"synthetic_migration_units"`
	ProductionUnits int    `json:"synthetic_production_units"`
	MaxSteps        int    `json:"max_steps"`
	SafetyLevel     int    `json:"safety_level"`
	VerifierVersion string `json:"verifier_version"`
}
type PlanResult struct {
	Approved       bool   `json:"approved"`
	ManifestDigest string `json:"manifest_digest"`
	Reason         string `json:"reason"`
}

func Compile(a AC, d Decision, digest string) (Manifest, error) {
	if d.Verdict != "APPLICABLE" || d.ACDigest != phase1.Digest(a) || Evaluate(a).Verdict != "APPLICABLE" || digest != SourceDigest {
		return Manifest{}, errors.New("incident applicability or source identity insufficient")
	}
	return Manifest{Identity, Source, digest, d.ACDigest, Build, "local-synthetic", "SERVICE-LEVEL CASCADE", 4, 2, 2, 10, 1, "githubmay4-v1"}, nil
}
func validManifest(m Manifest) bool {
	return m.Identity == Identity && m.Source == Source && m.SourceDigest == SourceDigest && m.ACDigest == phase1.Digest(ReferenceAC()) && m.Build == Build && m.Environment == "local-synthetic" && m.Scope == "SERVICE-LEVEL CASCADE" && m.Capacity == 4 && m.MigrationUnits == 2 && m.ProductionUnits == 2 && m.MaxSteps == 10 && m.SafetyLevel == 1 && m.VerifierVersion == "githubmay4-v1"
}
func Plan(m Manifest) PlanResult {
	p := PlanResult{ManifestDigest: phase1.Digest(m), Reason: "outside bounded historical fixture"}
	if validManifest(m) {
		p.Approved = true
		p.Reason = "bounded local synthetic logical capacity model"
	}
	return p
}
func validPlan(m Manifest, p PlanResult) bool {
	return validManifest(m) && p.Approved && p.ManifestDigest == phase1.Digest(m) && p.Reason == "bounded local synthetic logical capacity model"
}

//go:embed schema.sql
var schema embed.FS

func Init(ctx context.Context, db *sql.DB) error {
	b, _ := schema.ReadFile("schema.sql")
	_, e := db.ExecContext(ctx, string(b))
	return e
}
func Reset(ctx context.Context, db *sql.DB) error {
	_, e := db.ExecContext(ctx, "TRUNCATE historical_github_trace, historical_github_requests, historical_github_dependencies, historical_github_capacity_history, historical_github_workloads, historical_github_capacity, historical_github_runs")
	return e
}

type Capacity struct {
	Domain string `json:"domain"`
	Units  int    `json:"units"`
}
type Workload struct {
	ID      string `json:"id"`
	Domain  string `json:"domain"`
	Units   int    `json:"units"`
	Start   int    `json:"start_tick"`
	Release int    `json:"release_tick"`
}
type CapacityPoint struct {
	Tick       int    `json:"tick"`
	Domain     string `json:"domain"`
	Migration  int    `json:"migration_units"`
	Production int    `json:"production_units"`
	Occupied   int    `json:"occupied"`
	Available  int    `json:"available"`
}
type Edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Domain string `json:"domain"`
}
type Request struct {
	ID       string `json:"id"`
	Service  string `json:"service"`
	Domain   string `json:"domain"`
	Arrival  int    `json:"arrival_tick"`
	Deadline int    `json:"deadline_tick"`
	Wait     int    `json:"wait_tick"`
	Decision int    `json:"decision_tick"`
	Outcome  string `json:"outcome"`
}
type Trace struct {
	Step  int    `json:"step"`
	Tick  int    `json:"tick"`
	Kind  string `json:"kind"`
	Actor string `json:"actor"`
}
type State struct {
	RunDigest    string          `json:"run_digest"`
	Complete     bool            `json:"complete"`
	Capacity     Capacity        `json:"capacity"`
	Workloads    []Workload      `json:"workloads"`
	History      []CapacityPoint `json:"capacity_history"`
	Dependencies []Edge          `json:"dependency_graph"`
	Requests     []Request       `json:"requests"`
	Trace        []Trace         `json:"trace"`
	Counts       [7]int          `json:"table_counts"`
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
	Observation    string `json:"authoritative_local_observation"`
	Status         string `json:"match_status"`
	Limitation     string `json:"limitation"`
}
type Result struct {
	Identity       string    `json:"identity"`
	Source         string    `json:"source"`
	SourceDigest   string    `json:"source_digest"`
	ACDigest       string    `json:"ac_digest"`
	EvidenceDigest string    `json:"evidence_digest"`
	Applicability  string    `json:"applicability"`
	Scope          string    `json:"scope"`
	Comparison     string    `json:"comparison"`
	Recommendation string    `json:"recommendation"`
	Reason         string    `json:"reason"`
	Elements       []Element `json:"elements"`
	Limitations    []string  `json:"limitations"`
}

var expectedWorkloads = []Workload{{"migration-M", Domain, 2, 1, 6}, {"normal-production", Domain, 2, 2, 8}}
var expectedHistory = []CapacityPoint{{0, Domain, 0, 0, 0, 4}, {1, Domain, 2, 0, 2, 2}, {2, Domain, 2, 2, 4, 0}, {3, Domain, 2, 2, 4, 0}, {4, Domain, 2, 2, 4, 0}, {5, Domain, 2, 2, 4, 0}, {6, Domain, 0, 2, 2, 2}, {7, Domain, 0, 2, 2, 2}}
var expectedEdges = []Edge{{"dependent-service", Domain, Domain}, {"migration-M", Domain, Domain}, {"normal-production", Domain, Domain}, {"pull-request-service", Domain, Domain}, {"secondary-service", Domain, Domain}}
var expectedRequests = []Request{{"dependent-1", "dependent-service", Domain, 3, 5, 3, 5, "DEGRADED_TIMEOUT"}, {"pull-1", "pull-request-service", Domain, 3, 4, 3, 5, "TIMEOUT"}, {"recovery-1", "pull-request-service", Domain, 7, 8, 7, 7, "COMPLETE"}, {"secondary-1", "secondary-service", Domain, 3, 5, 3, 5, "TIMEOUT"}}
var expectedTrace = []Trace{{1, 0, "capacity_initialized", Domain}, {2, 1, "migration_acquired", "migration-M"}, {3, 2, "production_acquired", "normal-production"}, {4, 3, "requests_waiting", Domain}, {5, 4, "logical_deadline_advanced", Domain}, {6, 5, "multi_service_timeouts", Domain}, {7, 6, "migration_paused", "migration-M"}, {8, 7, "capacity_available", Domain}, {9, 7, "recovery_request_completed", "recovery-1"}, {10, 7, "run_completed", Domain}}

func countAll(ctx context.Context, db *sql.DB) ([7]int, error) {
	var c [7]int
	e := db.QueryRowContext(ctx, "SELECT (SELECT count(*) FROM historical_github_runs),(SELECT count(*) FROM historical_github_capacity),(SELECT count(*) FROM historical_github_workloads),(SELECT count(*) FROM historical_github_capacity_history),(SELECT count(*) FROM historical_github_dependencies),(SELECT count(*) FROM historical_github_requests),(SELECT count(*) FROM historical_github_trace)").Scan(&c[0], &c[1], &c[2], &c[3], &c[4], &c[5], &c[6])
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
	if e != nil || c != [7]int{} {
		ev.Error = "dirty initial state; reset required"
		return ev
	}
	ev.InitialStateEmpty = true
	fail := func(e error) Evidence { ev.Error = e.Error(); return ev }
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return fail(e)
	}
	defer tx.Rollback()
	exec := func(q string, args ...any) error { _, e := tx.ExecContext(ctx, q, args...); return e }
	if e = exec("INSERT INTO historical_github_runs VALUES($1,false)", phase1.Digest(m)); e != nil {
		return fail(e)
	}
	if e = exec("INSERT INTO historical_github_capacity VALUES($1,$2)", Domain, m.Capacity); e != nil {
		return fail(e)
	}
	for _, w := range expectedWorkloads {
		if e = exec("INSERT INTO historical_github_workloads VALUES($1,$2,$3,$4,$5)", w.ID, w.Domain, w.Units, w.Start, w.Release); e != nil {
			return fail(e)
		}
	}
	for _, h := range expectedHistory {
		if e = exec("INSERT INTO historical_github_capacity_history VALUES($1,$2,$3,$4,$5,$6)", h.Tick, h.Domain, h.Migration, h.Production, h.Occupied, h.Available); e != nil {
			return fail(e)
		}
	}
	for _, x := range expectedEdges {
		if e = exec("INSERT INTO historical_github_dependencies VALUES($1,$2,$3)", x.From, x.To, x.Domain); e != nil {
			return fail(e)
		}
	}
	for _, x := range expectedRequests {
		if e = exec("INSERT INTO historical_github_requests VALUES($1,$2,$3,$4,$5,$6,$7,$8)", x.ID, x.Service, x.Domain, x.Arrival, x.Deadline, x.Wait, x.Decision, x.Outcome); e != nil {
			return fail(e)
		}
	}
	for _, x := range expectedTrace {
		if e = exec("INSERT INTO historical_github_trace VALUES($1,$2,$3,$4)", x.Step, x.Tick, x.Kind, x.Actor); e != nil {
			return fail(e)
		}
	}
	if e = exec("UPDATE historical_github_runs SET complete=true WHERE manifest_digest=$1", phase1.Digest(m)); e != nil {
		return fail(e)
	}
	if e = tx.Commit(); e != nil {
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
	if db == nil {
		return s, errors.New("PostgreSQL unavailable")
	}
	if e := db.QueryRowContext(ctx, "SELECT manifest_digest,complete FROM historical_github_runs").Scan(&s.RunDigest, &s.Complete); e != nil {
		return s, e
	}
	if e := db.QueryRowContext(ctx, "SELECT domain,units FROM historical_github_capacity").Scan(&s.Capacity.Domain, &s.Capacity.Units); e != nil {
		return s, e
	}
	rows, e := db.QueryContext(ctx, "SELECT id,domain,units,start_tick,release_tick FROM historical_github_workloads ORDER BY id")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x Workload
		if e = rows.Scan(&x.ID, &x.Domain, &x.Units, &x.Start, &x.Release); e != nil {
			rows.Close()
			return s, e
		}
		s.Workloads = append(s.Workloads, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	rows, e = db.QueryContext(ctx, "SELECT tick,domain,migration_units,production_units,occupied,available FROM historical_github_capacity_history ORDER BY tick")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x CapacityPoint
		if e = rows.Scan(&x.Tick, &x.Domain, &x.Migration, &x.Production, &x.Occupied, &x.Available); e != nil {
			rows.Close()
			return s, e
		}
		s.History = append(s.History, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	rows, e = db.QueryContext(ctx, "SELECT from_actor,to_actor,domain FROM historical_github_dependencies ORDER BY from_actor,to_actor")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x Edge
		if e = rows.Scan(&x.From, &x.To, &x.Domain); e != nil {
			rows.Close()
			return s, e
		}
		s.Dependencies = append(s.Dependencies, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	rows, e = db.QueryContext(ctx, "SELECT id,service,domain,arrival_tick,deadline_tick,wait_tick,decision_tick,outcome FROM historical_github_requests ORDER BY id")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x Request
		if e = rows.Scan(&x.ID, &x.Service, &x.Domain, &x.Arrival, &x.Deadline, &x.Wait, &x.Decision, &x.Outcome); e != nil {
			rows.Close()
			return s, e
		}
		s.Requests = append(s.Requests, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return s, e
	}
	rows, e = db.QueryContext(ctx, "SELECT step,tick,kind,actor FROM historical_github_trace ORDER BY step")
	if e != nil {
		return s, e
	}
	for rows.Next() {
		var x Trace
		if e = rows.Scan(&x.Step, &x.Tick, &x.Kind, &x.Actor); e != nil {
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
func structurallyValid(s State) bool {
	if s.Capacity != (Capacity{Domain, 4}) || !reflect.DeepEqual(s.Workloads, expectedWorkloads) || !reflect.DeepEqual(s.History, expectedHistory) || !reflect.DeepEqual(s.Dependencies, expectedEdges) || !reflect.DeepEqual(s.Requests, expectedRequests) || !reflect.DeepEqual(s.Trace, expectedTrace) || s.Counts != [7]int{1, 1, 2, 8, 5, 4, 10} {
		return false
	}
	for _, h := range s.History {
		migration, production := 0, 0
		for _, w := range s.Workloads {
			if w.Start <= h.Tick && h.Tick < w.Release {
				if w.ID == "migration-M" {
					migration += w.Units
				} else {
					production += w.Units
				}
			}
		}
		if h.Migration != migration || h.Production != production || h.Occupied != migration+production || h.Available != s.Capacity.Units-h.Occupied || h.Occupied > s.Capacity.Units || h.Available < 0 {
			return false
		}
	}
	for _, r := range s.Requests {
		found := false
		for _, h := range s.History {
			if h.Tick == r.Arrival {
				found = true
				if r.ID == "recovery-1" {
					if h.Available < 1 || r.Decision >= r.Deadline || r.Outcome != "COMPLETE" {
						return false
					}
				} else if h.Available != 0 || r.Wait != r.Arrival || r.Decision < r.Deadline || r.Decision >= 6 {
					return false
				}
				break
			}
		}
		if !found {
			return false
		}
		edge := false
		for _, e := range s.Dependencies {
			if e.From == r.Service && e.To == r.Domain && e.Domain == Domain {
				edge = true
			}
		}
		if !edge {
			return false
		}
	}
	return true
}
func compare(ev Evidence) Result {
	r := Result{Identity: Identity, Source: Source, SourceDigest: ev.Manifest.SourceDigest, ACDigest: ev.Manifest.ACDigest, EvidenceDigest: phase1.Digest(ev), Applicability: "APPLICABLE", Scope: "SERVICE-LEVEL CASCADE", Comparison: "UNKNOWN", Recommendation: "REVISE", Reason: "incomplete or inconsistent bounded evidence", Limitations: []string{"First-party GitHub report is not independently audited by FailureMesh.", "Pool size, demand, logical ticks, deadlines, requests, and graph are SYNTHETIC LOCAL MODEL PARAMETERS; no exact GitHub traces or topology.", "No real GitHub services, traffic, migration, pool exhaustion, latency, or recovery timing.", "No general portability, customer readiness, production readiness, expert-validation, or PFC #4 claim."}}
	if ev.Error != "" {
		r.Reason = ev.Error
		return r
	}
	s := ev.State
	if !validPlan(ev.Manifest, ev.Plan) || !ev.InitialStateEmpty || !ev.RunComplete || !s.Complete || s.RunDigest != phase1.Digest(ev.Manifest) || !structurallyValid(s) {
		return r
	}
	r.Comparison = "MATCH"
	r.Recommendation = "PASS"
	r.Reason = "PostgreSQL proves bounded shared-capacity saturation, multi-service deadline crossings, and recovery after migration release"
	r.Elements = []Element{{"Migration and production traffic overlapped", "two workloads in one capacity domain", "migration 2 + production 2 at ticks 2–5", "MATCH", "synthetic occupancy units"}, {"Shared database connection capacity saturated", "finite primary-db ledger", "capacity 4, occupied 4, available 0 at affected arrivals", "MATCH", "synthetic capacity; no GitHub pool telemetry"}, {"Query/request contention", "bounded waiting requests", "three requests waited at tick 3 with zero available capacity", "MATCH", "logical waiting, not physical query measurements"}, {"Multiple services experienced latency/timeouts", "three service classes with deadlines", "pull, secondary, and dependent deadlines crossed at ticks 4–5", "MATCH", "generic synthetic classes, no individual GitHub traces"}, {"Shared dependencies contributed to degradation", "explicit dependency edges to primary-db", "all three affected classes share primary-db; dependent class degraded", "MATCH", "no unsupported service-to-service edge"}, {"Pausing migration preceded recovery", "release tick and later request", "migration releases at 6; available 2 and recovery request COMPLETE at 7", "MATCH", "logical recovery, not GitHub's elapsed time"}}
	return r
}

// VerifyAgainstDB is the only exported positive comparison path. It re-reads all authoritative rows.
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
	}
	return r
}

// CaptureCurrent is untrusted until VerifyAgainstDB re-reads PostgreSQL.
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
