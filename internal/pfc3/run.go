package pfc3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"reflect"
	"time"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

//go:embed schema.sql
var schema embed.FS

func Init(ctx context.Context, db *sql.DB) error {
	b, _ := schema.ReadFile("schema.sql")
	_, e := db.ExecContext(ctx, string(b))
	return e
}
func Reset(ctx context.Context, db *sql.DB) error {
	_, e := db.ExecContext(ctx, "TRUNCATE protected_attempts, protected_effects, lease_trace, lease_history, lease_resources")
	return e
}

type Manifest struct {
	Version         string `json:"version"`
	PFC             string `json:"pfc"`
	ACVersion       string `json:"ac_version"`
	ACDigest        string `json:"ac_digest"`
	Build           string `json:"build"`
	Environment     string `json:"environment"`
	Fault           string `json:"fault"`
	Attempts        int    `json:"attempts"`
	Concurrency     int    `json:"concurrency"`
	Invariant       string `json:"invariant"`
	ResourceID      string `json:"resource_id"`
	WorkerA         string `json:"worker_a"`
	WorkerB         string `json:"worker_b"`
	TokenA          int64  `json:"token_a"`
	TokenB          int64  `json:"token_b"`
	AdapterVersion  string `json:"adapter_version"`
	VerifierVersion string `json:"verifier_version"`
	RunID           string `json:"run_id"`
	SafetyLevel     int    `json:"safety_level"`
}
type SafetyPlan struct {
	Approved           bool   `json:"approved"`
	Reason             string `json:"reason"`
	ManifestDigest     string `json:"manifest_digest"`
	MaxSteps           int    `json:"max_steps"`
	MaxDurationSeconds int    `json:"max_duration_seconds"`
}

func validManifest(m Manifest) bool {
	return m.Version == "1" && m.PFC == Identity && m.ACVersion == "1" && m.ACDigest == phase1.Digest(ReferenceAC(m.Build == remediatedBuild)) && (m.Build == vulnerableBuild || m.Build == remediatedBuild) && m.Environment == "synthetic-reference" && m.Fault == Fault && m.Attempts == 2 && m.Concurrency == 2 && m.Invariant == Invariant && m.ResourceID == "resource-1" && m.WorkerA == "worker-A" && m.WorkerB == "worker-B" && m.TokenA == 1 && m.TokenB == 2 && m.AdapterVersion == "logical-lease-pg-v1" && m.VerifierVersion == "pfc3-v1" && m.RunID == "pfc3-run" && m.SafetyLevel == 1
}
func Plan(p PFC, m Manifest) SafetyPlan {
	s := SafetyPlan{ManifestDigest: phase1.Digest(m), MaxSteps: 10, MaxDurationSeconds: 10}
	if Valid(p) && validManifest(m) {
		s.Approved = true
		s.Reason = "bounded synthetic logical lease expiry"
	} else {
		s.Reason = "outside authorized synthetic experiment"
	}
	return s
}
func validPlan(m Manifest, s SafetyPlan) bool {
	return validManifest(m) && s.Approved && s.Reason == "bounded synthetic logical lease expiry" && s.ManifestDigest == phase1.Digest(m) && s.MaxSteps == 10 && s.MaxDurationSeconds == 10
}

type Event struct {
	Seq        int    `json:"seq"`
	Kind       string `json:"kind"`
	ResourceID string `json:"resource_id"`
	WorkerID   string `json:"worker_id"`
	Token      int64  `json:"token"`
}
type History struct {
	Step     int    `json:"step"`
	WorkerID string `json:"worker_id"`
	Token    int64  `json:"token"`
	Action   string `json:"action"`
	Tick     int    `json:"tick"`
}
type Effect struct {
	Seq      int    `json:"seq"`
	WorkerID string `json:"worker_id"`
	Token    int64  `json:"token"`
	Value    string `json:"value"`
}
type Attempt struct {
	Seq      int    `json:"seq"`
	WorkerID string `json:"worker_id"`
	Token    int64  `json:"token"`
	Accepted bool   `json:"accepted"`
}
type Resource struct {
	Owner   string `json:"owner"`
	Token   int64  `json:"token"`
	Tick    int    `json:"tick"`
	Expires int    `json:"expires"`
	State   string `json:"state"`
}
type Evidence struct {
	Manifest          Manifest   `json:"manifest"`
	Safety            SafetyPlan `json:"safety"`
	Events            []Event    `json:"events"`
	History           []History  `json:"history"`
	Effects           []Effect   `json:"effects"`
	Attempts          []Attempt  `json:"attempts"`
	Resource          Resource   `json:"resource"`
	InitialStateEmpty bool       `json:"initial_state_empty"`
	LedgerComplete    bool       `json:"ledger_complete"`
	RunComplete       bool       `json:"run_complete"`
	Error             string     `json:"error,omitempty"`
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

func expected(m Manifest) []Event {
	k := []string{"worker_a_lease_acquired", "worker_a_work_started", "worker_a_stalled", "worker_a_lease_expired", "worker_b_lease_acquired", "worker_b_work_started", "worker_b_effect_committed", "worker_a_resumed", "worker_a_stale_write_attempted", "worker_a_stale_effect_committed"}
	if m.Build == remediatedBuild {
		k[9] = "worker_a_stale_write_rejected"
	}
	out := make([]Event, 10)
	for i := range out {
		w, t := m.WorkerA, m.TokenA
		if i == 4 || i == 5 || i == 6 {
			w, t = m.WorkerB, m.TokenB
		}
		out[i] = Event{i + 1, k[i], m.ResourceID, w, t}
	}
	return out
}
func record(ctx context.Context, db *sql.DB, m Manifest, e Event) error {
	_, err := db.ExecContext(ctx, "INSERT INTO lease_trace(run_id,seq,resource_id,worker_id,fencing_token,kind) VALUES($1,$2,$3,$4,$5,$6)", m.RunID, e.Seq, e.ResourceID, e.WorkerID, e.Token, e.Kind)
	return err
}
func acquireA(ctx context.Context, db *sql.DB, m Manifest) error {
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(ctx, "INSERT INTO lease_resources VALUES($1,$2,$3,1,0,1,'active')", m.ResourceID, m.RunID, m.WorkerA); e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, "INSERT INTO lease_history VALUES($1,$2,1,$3,1,'acquired',0)", m.RunID, m.ResourceID, m.WorkerA); e != nil {
		return e
	}
	return tx.Commit()
}
func expire(ctx context.Context, db *sql.DB, m Manifest) error {
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	r, e := tx.ExecContext(ctx, "UPDATE lease_resources SET logical_tick=1,lease_state='expired' WHERE resource_id=$1 AND run_id=$2 AND current_owner=$3 AND current_token=1 AND expires_at_tick=1 AND lease_state='active'", m.ResourceID, m.RunID, m.WorkerA)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return errors.New("expiry precondition failed")
	}
	if _, e = tx.ExecContext(ctx, "INSERT INTO lease_history VALUES($1,$2,4,$3,1,'expired',1)", m.RunID, m.ResourceID, m.WorkerA); e != nil {
		return e
	}
	return tx.Commit()
}
func acquireB(ctx context.Context, db *sql.DB, m Manifest) error {
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var token int64
	e = tx.QueryRowContext(ctx, "UPDATE lease_resources SET current_owner=$3,current_token=current_token+1,lease_state='active',expires_at_tick=2 WHERE resource_id=$1 AND run_id=$2 AND lease_state='expired' AND logical_tick>=expires_at_tick RETURNING current_token", m.ResourceID, m.RunID, m.WorkerB).Scan(&token)
	if e != nil {
		return e
	}
	if token != m.TokenB {
		return errors.New("unexpected fencing generation")
	}
	if _, e = tx.ExecContext(ctx, "INSERT INTO lease_history VALUES($1,$2,5,$3,$4,'acquired',1)", m.RunID, m.ResourceID, m.WorkerB, token); e != nil {
		return e
	}
	return tx.Commit()
}

// protectedWrite locks the authoritative resource row through effect insertion.
// Vulnerable mode deliberately ignores the token while still recording the attempt.
func protectedWrite(ctx context.Context, db *sql.DB, m Manifest, seq int, worker string, token int64, remediated bool) (bool, error) {
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return false, e
	}
	defer tx.Rollback()
	var current int64
	var state string
	e = tx.QueryRowContext(ctx, "SELECT current_token,lease_state FROM lease_resources WHERE resource_id=$1 AND run_id=$2 FOR UPDATE", m.ResourceID, m.RunID).Scan(&current, &state)
	if e != nil {
		return false, e
	}
	accepted := !remediated || (current == token && state == "active")
	if _, e = tx.ExecContext(ctx, "INSERT INTO protected_attempts VALUES($1,$2,$3,$4,$5,$6)", m.RunID, m.ResourceID, seq, worker, token, accepted); e != nil {
		return false, e
	}
	if accepted {
		value := "B-legitimate"
		if worker == m.WorkerA {
			value = "A-stale"
		}
		if _, e = tx.ExecContext(ctx, "INSERT INTO protected_effects VALUES($1,$2,$3,$4,$5,$6)", m.RunID, m.ResourceID, seq, worker, token, value); e != nil {
			return false, e
		}
	}
	return accepted, tx.Commit()
}
func snapshot(ctx context.Context, db *sql.DB, m Manifest) (Evidence, error) {
	var ev Evidence
	rows, e := db.QueryContext(ctx, "SELECT seq,kind,resource_id,worker_id,fencing_token FROM lease_trace WHERE run_id=$1 ORDER BY seq", m.RunID)
	if e != nil {
		return ev, e
	}
	for rows.Next() {
		var x Event
		if e = rows.Scan(&x.Seq, &x.Kind, &x.ResourceID, &x.WorkerID, &x.Token); e != nil {
			rows.Close()
			return ev, e
		}
		ev.Events = append(ev.Events, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return ev, e
	}
	rows, e = db.QueryContext(ctx, "SELECT step,worker_id,fencing_token,action,logical_tick FROM lease_history WHERE run_id=$1 ORDER BY step", m.RunID)
	if e != nil {
		return ev, e
	}
	for rows.Next() {
		var x History
		if e = rows.Scan(&x.Step, &x.WorkerID, &x.Token, &x.Action, &x.Tick); e != nil {
			rows.Close()
			return ev, e
		}
		ev.History = append(ev.History, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return ev, e
	}
	rows, e = db.QueryContext(ctx, "SELECT seq,worker_id,fencing_token,effect_value FROM protected_effects WHERE run_id=$1 ORDER BY seq", m.RunID)
	if e != nil {
		return ev, e
	}
	for rows.Next() {
		var x Effect
		if e = rows.Scan(&x.Seq, &x.WorkerID, &x.Token, &x.Value); e != nil {
			rows.Close()
			return ev, e
		}
		ev.Effects = append(ev.Effects, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return ev, e
	}
	rows, e = db.QueryContext(ctx, "SELECT seq,worker_id,fencing_token,accepted FROM protected_attempts WHERE run_id=$1 ORDER BY seq", m.RunID)
	if e != nil {
		return ev, e
	}
	for rows.Next() {
		var x Attempt
		if e = rows.Scan(&x.Seq, &x.WorkerID, &x.Token, &x.Accepted); e != nil {
			rows.Close()
			return ev, e
		}
		ev.Attempts = append(ev.Attempts, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return ev, e
	}
	e = db.QueryRowContext(ctx, "SELECT current_owner,current_token,logical_tick,expires_at_tick,lease_state FROM lease_resources WHERE resource_id=$1 AND run_id=$2", m.ResourceID, m.RunID).Scan(&ev.Resource.Owner, &ev.Resource.Token, &ev.Resource.Tick, &ev.Resource.Expires, &ev.Resource.State)
	return ev, e
}
func Run(ctx context.Context, db *sql.DB, m Manifest, s SafetyPlan, remediated bool) Evidence {
	ev := Evidence{Manifest: m, Safety: s}
	if !validPlan(m, s) || remediated != (m.Build == remediatedBuild) {
		ev.Error = "safety plan or build rejected"
		return ev
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.MaxDurationSeconds)*time.Second)
	defer cancel()
	if e := Init(ctx, db); e != nil {
		ev.Error = e.Error()
		return ev
	}
	var count int
	e := db.QueryRowContext(ctx, "SELECT (SELECT count(*) FROM lease_resources)+(SELECT count(*) FROM lease_history)+(SELECT count(*) FROM lease_trace)+(SELECT count(*) FROM protected_effects)+(SELECT count(*) FROM protected_attempts)").Scan(&count)
	if e != nil || count != 0 {
		ev.Error = "clean reset precondition failed"
		return ev
	}
	ev.InitialStateEmpty = true
	steps := expected(m)
	for i, x := range steps {
		switch i {
		case 0:
			e = acquireA(ctx, db, m)
		case 3:
			e = expire(ctx, db, m)
		case 4:
			e = acquireB(ctx, db, m)
		case 6:
			_, e = protectedWrite(ctx, db, m, 7, m.WorkerB, m.TokenB, remediated)
		case 9:
			var accepted bool
			accepted, e = protectedWrite(ctx, db, m, 10, m.WorkerA, m.TokenA, remediated)
			if e == nil && accepted == remediated {
				e = errors.New("stale write outcome mismatched build")
			}
		}
		if e != nil {
			ev.Error = e.Error()
			return ev
		}
		if e = record(ctx, db, m, x); e != nil {
			ev.Error = e.Error()
			return ev
		}
	}
	snap, e := snapshot(ctx, db, m)
	if e != nil {
		ev.Error = e.Error()
		return ev
	}
	ev.Events, ev.History, ev.Effects, ev.Attempts, ev.Resource = snap.Events, snap.History, snap.Effects, snap.Attempts, snap.Resource
	ev.LedgerComplete = true
	ev.RunComplete = true
	return ev
}
func verifyEvidence(ev Evidence) Proof {
	p := Proof{FormatVersion: "1", Manifest: ev.Manifest, Safety: ev.Safety, Evidence: ev, EvidenceDigest: phase1.Digest(ev), Verdict: "UNKNOWN", Reason: "incomplete or inconsistent evidence", Replay: "reset PFC #3 tables and replay identical logical interleaving"}
	m := ev.Manifest
	if ev.Error != "" {
		p.Reason = ev.Error
		return p
	}
	if !validPlan(m, ev.Safety) || !ev.InitialStateEmpty || !ev.LedgerComplete || !ev.RunComplete || !reflect.DeepEqual(ev.Events, expected(m)) || !reflect.DeepEqual(ev.History, []History{{1, m.WorkerA, 1, "acquired", 0}, {4, m.WorkerA, 1, "expired", 1}, {5, m.WorkerB, 2, "acquired", 1}}) || ev.Resource != (Resource{m.WorkerB, 2, 1, 2, "active"}) {
		return p
	}
	attempts := []Attempt{{7, m.WorkerB, 2, true}, {10, m.WorkerA, 1, true}}
	effects := []Effect{{7, m.WorkerB, 2, "B-legitimate"}, {10, m.WorkerA, 1, "A-stale"}}
	if m.Build == remediatedBuild {
		attempts[1].Accepted = false
		effects = effects[:1]
	}
	if !reflect.DeepEqual(ev.Attempts, attempts) || !reflect.DeepEqual(ev.Effects, effects) {
		return p
	}
	if m.Build == vulnerableBuild {
		p.Verdict = "EXPOSED"
		p.Reason = "PostgreSQL records token-1 stale effect after token-2 takeover"
	} else {
		p.Verdict = "PROVEN_RESILIENT"
		p.InvariantHolds = true
		p.Reason = "token-1 attempt rejected after token-2 takeover in bounded synthetic run"
	}
	return p
}
func VerifyAgainstDB(ctx context.Context, db *sql.DB, ev Evidence) Proof {
	p := verifyEvidence(ev)
	unknown := func(reason string) Proof {
		p.Verdict = "UNKNOWN"
		p.InvariantHolds = false
		p.Reason = reason
		return p
	}
	if p.Verdict == "UNKNOWN" {
		return p
	}
	if db == nil {
		return unknown("authoritative PostgreSQL unavailable")
	}
	actual, e := snapshot(ctx, db, ev.Manifest)
	if e != nil || !reflect.DeepEqual(actual.Events, ev.Events) || !reflect.DeepEqual(actual.History, ev.History) || !reflect.DeepEqual(actual.Effects, ev.Effects) || !reflect.DeepEqual(actual.Attempts, ev.Attempts) || actual.Resource != ev.Resource {
		return unknown("authoritative PostgreSQL state missing or changed")
	}
	var counts [5]int
	e = db.QueryRowContext(ctx, "SELECT (SELECT count(*) FROM lease_resources),(SELECT count(*) FROM lease_history),(SELECT count(*) FROM lease_trace),(SELECT count(*) FROM protected_effects),(SELECT count(*) FROM protected_attempts)").Scan(&counts[0], &counts[1], &counts[2], &counts[3], &counts[4])
	if e != nil || counts != [5]int{1, 3, 10, len(ev.Effects), 2} {
		return unknown("authoritative PostgreSQL ledger incomplete")
	}
	return p
}
