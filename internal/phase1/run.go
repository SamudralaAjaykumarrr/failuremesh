package phase1

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed schema.sql
var schema embed.FS

func Open(dsn string) (*sql.DB, error) {
	config, e := parseSyntheticConfig(dsn)
	if e != nil {
		return nil, e
	}
	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(4)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e = db.PingContext(ctx); e != nil {
		db.Close()
		return nil, e
	}
	return db, nil
}
func Init(ctx context.Context, db *sql.DB) error {
	b, _ := schema.ReadFile("schema.sql")
	_, e := db.ExecContext(ctx, string(b))
	return e
}
func Reset(ctx context.Context, db *sql.DB) error {
	_, e := db.ExecContext(ctx, "TRUNCATE provider_idempotency, provider_effects, business_operations RESTART IDENTITY")
	return e
}

type Manifest struct {
	Version         string   `json:"version"`
	PFC             string   `json:"pfc"`
	ACVersion       string   `json:"ac_version"`
	ACDigest        string   `json:"ac_digest"`
	Build           string   `json:"build"`
	Environment     string   `json:"environment"`
	Fault           string   `json:"fault"`
	Attempts        int      `json:"attempts"`
	RetryTiming     string   `json:"retry_timing"`
	Concurrency     int      `json:"concurrency"`
	Invariant       string   `json:"invariant"`
	Observers       []string `json:"observers"`
	Reset           string   `json:"reset"`
	AdapterVersion  string   `json:"adapter_version"`
	VerifierVersion string   `json:"verifier_version"`
	RunID           string   `json:"run_id"`
	LogicalID       string   `json:"logical_id"`
	SafetyLevel     int      `json:"safety_level"`
}
type SafetyPlan struct {
	Approved           bool   `json:"approved"`
	Reason             string `json:"reason"`
	ManifestDigest     string `json:"manifest_digest"`
	MaxAttempts        int    `json:"max_attempts"`
	MaxDurationSeconds int    `json:"max_duration_seconds"`
}

var phase1Observers = []string{"caller attempts", "simulator barrier", "authoritative provider ledger"}

func validManifest(m Manifest) bool {
	remediated := m.Build == "reference-idempotent-v1"
	return m.Version == "1" && m.PFC == pfcIdentity && m.ACVersion == "1" && m.ACDigest != "" &&
		(m.Build == "reference-vulnerable-v1" || m.Build == "reference-idempotent-v1") &&
		m.ACDigest == Digest(ReferenceAC(remediated)) &&
		m.Environment == "synthetic-reference" && m.Fault == phase1Fault && m.Attempts == 2 &&
		m.RetryTiming == "immediate after observed loss" && m.Concurrency == 1 &&
		m.Invariant == phase1Invariant && slices.Equal(m.Observers, phase1Observers) &&
		m.Reset == "TRUNCATE reference tables RESTART IDENTITY" &&
		m.AdapterVersion == "reference-pg-v1" && m.VerifierVersion == "phase1-v1" &&
		m.RunID == "phase1-run" && m.LogicalID == "payment-X" && m.SafetyLevel == 1
}

func validPlan(m Manifest, s SafetyPlan) bool {
	return validManifest(m) && s.Approved && s.Reason == "bounded synthetic reference run" &&
		s.ManifestDigest == Digest(m) && s.MaxAttempts == 2 && s.MaxDurationSeconds == 10
}

func Compile(p PFC, a AC, m Match) (Manifest, error) {
	if !validPFC(p) || m.Verdict != "APPLICABLE" || m.PFC != pfcIdentity || m.AC != Digest(a) {
		return Manifest{}, errors.New("applicability does not authorize compilation")
	}
	return Manifest{Version: "1", PFC: p.ID + "@" + p.Version, ACVersion: a.Version, ACDigest: Digest(a), Build: a.Build, Environment: a.Environment, Fault: p.Experiment.Fault, Attempts: p.Experiment.Attempts, RetryTiming: "immediate after observed loss", Concurrency: p.Experiment.Concurrency, Invariant: p.Invariant, Observers: []string{"caller attempts", "simulator barrier", "authoritative provider ledger"}, Reset: "TRUNCATE reference tables RESTART IDENTITY", AdapterVersion: "reference-pg-v1", VerifierVersion: "phase1-v1", RunID: "phase1-run", LogicalID: "payment-X", SafetyLevel: 1}, nil
}
func Plan(p PFC, m Manifest) SafetyPlan {
	s := SafetyPlan{ManifestDigest: Digest(m), MaxAttempts: 2, MaxDurationSeconds: 10}
	if !validPFC(p) || !validManifest(m) {
		s.Reason = "outside authorized synthetic level-1 experiment"
		return s
	}
	s.Approved = true
	s.Reason = "bounded synthetic reference run"
	return s
}

type Event struct {
	Seq       int    `json:"seq"`
	Kind      string `json:"kind"`
	Attempt   int    `json:"attempt,omitempty"`
	EffectID  int64  `json:"effect_id,omitempty"`
	LogicalID string `json:"logical_id"`
}
type Effect struct {
	ID        int64  `json:"id"`
	Attempt   int    `json:"attempt"`
	LogicalID string `json:"logical_id"`
}
type Evidence struct {
	Manifest          Manifest   `json:"manifest"`
	Safety            SafetyPlan `json:"safety"`
	Events            []Event    `json:"events"`
	Effects           []Effect   `json:"effects"`
	LedgerComplete    bool       `json:"ledger_complete"`
	RunComplete       bool       `json:"run_complete"`
	InitialStateEmpty bool       `json:"initial_state_empty"`
	Error             string     `json:"error,omitempty"`
}

func (e *Evidence) add(kind string, attempt int, id int64) {
	e.Events = append(e.Events, Event{Seq: len(e.Events) + 1, Kind: kind, Attempt: attempt, EffectID: id, LogicalID: e.Manifest.LogicalID})
}
func ledger(ctx context.Context, db *sql.DB, m Manifest) ([]Effect, error) {
	rows, e := db.QueryContext(ctx, "SELECT effect_id, attempt, logical_id FROM provider_effects WHERE run_id=$1 AND logical_id=$2 ORDER BY effect_id", m.RunID, m.LogicalID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Effect
	for rows.Next() {
		var x Effect
		if e = rows.Scan(&x.ID, &x.Attempt, &x.LogicalID); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func provider(ctx context.Context, db *sql.DB, m Manifest, attempt int, remediated bool) (int64, bool, error) {
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return 0, false, e
	}
	defer tx.Rollback()
	if remediated {
		var id int64
		e = tx.QueryRowContext(ctx, "SELECT effect_id FROM provider_idempotency WHERE run_id=$1 AND logical_id=$2 FOR UPDATE", m.RunID, m.LogicalID).Scan(&id)
		if e == nil {
			return id, false, tx.Commit()
		}
		if e != sql.ErrNoRows {
			return 0, false, e
		}
	}
	var id int64
	e = tx.QueryRowContext(ctx, "INSERT INTO provider_effects(run_id,logical_id,attempt) VALUES($1,$2,$3) RETURNING effect_id", m.RunID, m.LogicalID, attempt).Scan(&id)
	if e != nil {
		return 0, false, e
	}
	if remediated {
		_, e = tx.ExecContext(ctx, "INSERT INTO provider_idempotency(run_id,logical_id,effect_id) VALUES($1,$2,$3)", m.RunID, m.LogicalID, id)
		if e != nil {
			return 0, false, e
		}
	}
	return id, true, tx.Commit()
}
func Run(ctx context.Context, db *sql.DB, m Manifest, s SafetyPlan, remediated bool) Evidence {
	ev := Evidence{Manifest: m, Safety: s}
	if !validPlan(m, s) {
		ev.Error = "safety plan rejected"
		return ev
	}
	if (remediated && m.Build != "reference-idempotent-v1") || (!remediated && m.Build != "reference-vulnerable-v1") {
		ev.Error = "build/control mismatch"
		return ev
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.MaxDurationSeconds)*time.Second)
	defer cancel()
	if e := Init(ctx, db); e != nil {
		ev.Error = e.Error()
		return ev
	}
	var count int
	e := db.QueryRowContext(ctx, "SELECT count(*) FROM provider_effects WHERE run_id=$1 OR logical_id=$2", m.RunID, m.LogicalID).Scan(&count)
	if e != nil || count != 0 {
		ev.Error = "reset precondition failed"
		return ev
	}
	ev.InitialStateEmpty = true
	_, e = db.ExecContext(ctx, "INSERT INTO business_operations(run_id,logical_id,status) VALUES($1,$2,'pending')", m.RunID, m.LogicalID)
	if e != nil {
		ev.Error = e.Error()
		return ev
	}
	for attempt := 1; attempt <= m.Attempts; attempt++ {
		ev.add("caller_request", attempt, 0)
		id, created, e := provider(ctx, db, m, attempt, remediated)
		if e != nil {
			ev.Error = e.Error()
			return ev
		}
		if created {
			ev.add("provider_commit", attempt, id)
		} else {
			ev.add("provider_dedupe", attempt, id)
		}
		if attempt == 1 {
			ev.add("barrier_holds_success", attempt, id)
			ev.add("response_lost_after_commit", attempt, id)
			ev.add("caller_observed_loss", attempt, id)
			ev.add("caller_retry", attempt+1, 0)
			continue
		}
		ev.add("caller_success", attempt, id)
	}
	_, e = db.ExecContext(ctx, "UPDATE business_operations SET status='completed' WHERE run_id=$1 AND logical_id=$2", m.RunID, m.LogicalID)
	if e != nil {
		ev.Error = e.Error()
		return ev
	}
	ev.Effects, e = ledger(ctx, db, m)
	if e != nil {
		ev.Error = e.Error()
		return ev
	}
	var total int
	e = db.QueryRowContext(ctx, "SELECT count(*) FROM provider_effects WHERE run_id=$1", m.RunID).Scan(&total)
	if e != nil {
		ev.Error = e.Error()
		return ev
	}
	ev.LedgerComplete = total == len(ev.Effects)
	ev.RunComplete = true
	return ev
}
func Verify(ev Evidence) Proof {
	p := Proof{FormatVersion: "1", Manifest: ev.Manifest, Safety: ev.Safety, Evidence: ev, Verdict: "UNKNOWN", Reason: "incomplete or inconsistent evidence", InvariantHolds: false, Replay: "reset reference database; run same mode and manifest"}
	p.EvidenceDigest = Digest(ev)
	if ev.Error != "" {
		p.Reason = ev.Error
		return p
	}
	m := ev.Manifest
	if !validPlan(m, ev.Safety) || !ev.RunComplete || !ev.InitialStateEmpty || !ev.LedgerComplete {
		return p
	}
	if len(ev.Events) != 9 {
		p.Reason = "exact nine-event causal trace required"
		return p
	}
	a := ev.Events[1].EffectID
	if a <= 0 {
		p.Reason = "first committed effect missing"
		return p
	}
	mode := m.Build
	secondKind := "provider_commit"
	b := ev.Events[7].EffectID
	if mode == "reference-idempotent-v1" {
		secondKind = "provider_dedupe"
		b = a
	}
	expected := []Event{
		{Kind: "caller_request", Attempt: 1},
		{Kind: "provider_commit", Attempt: 1, EffectID: a},
		{Kind: "barrier_holds_success", Attempt: 1, EffectID: a},
		{Kind: "response_lost_after_commit", Attempt: 1, EffectID: a},
		{Kind: "caller_observed_loss", Attempt: 1, EffectID: a},
		{Kind: "caller_retry", Attempt: 2},
		{Kind: "caller_request", Attempt: 2},
		{Kind: secondKind, Attempt: 2, EffectID: b},
		{Kind: "caller_success", Attempt: 2, EffectID: b},
	}
	for i := range expected {
		expected[i].Seq = i + 1
		expected[i].LogicalID = m.LogicalID
		if ev.Events[i] != expected[i] {
			p.Reason = "causal event, attempt, effect, identity, or order mismatch"
			return p
		}
	}
	if mode == "reference-vulnerable-v1" {
		if b <= 0 || a == b || len(ev.Effects) != 2 ||
			ev.Effects[0] != (Effect{ID: a, Attempt: 1, LogicalID: m.LogicalID}) ||
			ev.Effects[1] != (Effect{ID: b, Attempt: 2, LogicalID: m.LogicalID}) {
			p.Reason = "vulnerable ledger does not match both ordered commits"
			return p
		}
		p.Verdict = "EXPOSED"
		p.Reason = "two committed provider effects for one logical operation"
		return p
	}
	if ev.Events[7].EffectID != a || len(ev.Effects) != 1 ||
		ev.Effects[0] != (Effect{ID: a, Attempt: 1, LogicalID: m.LogicalID}) {
		p.Reason = "remediated ledger or dedupe does not match first commit"
		return p
	}
	p.InvariantHolds = true
	p.Verdict = "PROVEN_RESILIENT"
	p.Reason = "one committed effect across the complete two-attempt experiment"
	return p
}

type Proof struct {
	FormatVersion  string     `json:"format_version"`
	Manifest       Manifest   `json:"manifest"`
	Safety         SafetyPlan `json:"safety"`
	Evidence       Evidence   `json:"evidence"`
	EvidenceDigest string     `json:"evidence_digest"`
	InvariantHolds bool       `json:"invariant_holds"`
	Verdict        string     `json:"verdict"`
	Reason         string     `json:"reason"`
	Replay         string     `json:"replay"`
}

func ValidateDSN(dsn string) error {
	_, e := parseSyntheticConfig(dsn)
	return e
}

func parseSyntheticConfig(dsn string) (*pgx.ConnConfig, error) {
	// pgx merges PG* settings before building the config. Reject them before
	// parsing so a service file or credential path is never consulted.
	for _, setting := range os.Environ() {
		name, _, _ := strings.Cut(setting, "=")
		if strings.HasPrefix(name, "PG") {
			return nil, errors.New("PG environment settings are not allowed for the synthetic database")
		}
	}
	u, e := url.Parse(dsn)
	if e != nil || u.Scheme != "postgres" || u.Path != "/failuremesh" || u.User == nil || u.Fragment != "" {
		return nil, errors.New("synthetic PostgreSQL URL required")
	}
	password, present := u.User.Password()
	if u.User.Username() != "failuremesh" || !present || password != "synthetic-only" {
		return nil, errors.New("explicit synthetic PostgreSQL fixture credentials required")
	}
	if u.Hostname() != "127.0.0.1" && u.Hostname() != "::1" {
		return nil, errors.New("only literal loopback PostgreSQL hosts are allowed")
	}
	if u.Port() == "" || strings.Contains(u.Host, ",") {
		return nil, errors.New("one explicit PostgreSQL port and host are required")
	}
	query, e := url.ParseQuery(u.RawQuery)
	if e != nil {
		return nil, errors.New("invalid synthetic PostgreSQL URL query")
	}
	if len(query) != 1 || len(query["sslmode"]) != 1 || query.Get("sslmode") != "disable" {
		return nil, errors.New("only sslmode=disable is allowed in the synthetic URL query")
	}
	// pgx opens the default ~/.pgpass during parsing even with a supplied
	// password. Override only that internal default; the user URL cannot carry
	// passfile because its query was restricted to sslmode above.
	config, e := pgx.ParseConfigWithOptions(dsn+"&passfile=", pgx.ParseConfigOptions{ParseConfigOptions: pgconn.ParseConfigOptions{
		ConnStringAllowedKeys: []string{"host", "port", "database", "user", "password", "sslmode", "passfile"},
	}})
	if e != nil {
		return nil, e
	}
	if config.Database != "failuremesh" || config.User != "failuremesh" || config.Password != "synthetic-only" ||
		(config.Host != "127.0.0.1" && config.Host != "::1") ||
		config.Port == 0 || len(config.Fallbacks) != 0 || config.TLSConfig != nil ||
		len(config.RuntimeParams) != 0 {
		return nil, errors.New("parsed PostgreSQL config is outside the synthetic loopback boundary")
	}
	return config, nil
}
