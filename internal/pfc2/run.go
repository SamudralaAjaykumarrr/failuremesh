package pfc2

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"reflect"
	"slices"
	"time"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

//go:embed schema.sql
var schema embed.FS

func Init(ctx context.Context, db *sql.DB) error {
	b, _ := schema.ReadFile("schema.sql")
	_, err := db.ExecContext(ctx, string(b))
	return err
}
func Reset(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "TRUNCATE consumer_idempotency, consumer_effects RESTART IDENTITY")
	return err
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
	Concurrency     int      `json:"concurrency"`
	Invariant       string   `json:"invariant"`
	Observers       []string `json:"observers"`
	Reset           string   `json:"reset"`
	AdapterVersion  string   `json:"adapter_version"`
	VerifierVersion string   `json:"verifier_version"`
	RunID           string   `json:"run_id"`
	MessageID       string   `json:"message_id"`
	SafetyLevel     int      `json:"safety_level"`
}
type SafetyPlan struct {
	Approved           bool   `json:"approved"`
	Reason             string `json:"reason"`
	ManifestDigest     string `json:"manifest_digest"`
	MaxDeliveries      int    `json:"max_deliveries"`
	MaxDurationSeconds int    `json:"max_duration_seconds"`
}

var observers = []string{"synthetic delivery trace", "authoritative PostgreSQL consumer ledger"}

func validManifest(m Manifest) bool {
	remediated := m.Build == "queue-idempotent-v1"
	return m.Version == "1" && m.PFC == Identity && m.ACVersion == "1" && m.ACDigest == phase1.Digest(ReferenceAC(remediated)) &&
		(m.Build == "queue-vulnerable-v1" || remediated) && m.Environment == "synthetic-reference" && m.Fault == Fault && m.Attempts == 2 && m.Concurrency == 1 && m.Invariant == Invariant && slices.Equal(m.Observers, observers) && m.Reset == "TRUNCATE consumer reference tables RESTART IDENTITY" && m.AdapterVersion == "synthetic-queue-pg-v1" && m.VerifierVersion == "pfc2-v1" && m.RunID == "pfc2-run" && m.MessageID == "message-X" && m.SafetyLevel == 1
}
func validPlan(m Manifest, s SafetyPlan) bool {
	return validManifest(m) && s.Approved && s.Reason == "bounded synthetic duplicate delivery" && s.ManifestDigest == phase1.Digest(m) && s.MaxDeliveries == 2 && s.MaxDurationSeconds == 10
}
func Compile(p PFC, a AC, match Match) (Manifest, error) {
	if !Valid(p) || match.Verdict != "APPLICABLE" || match.PFC != Identity || match.AC != phase1.Digest(a) || Evaluate(p, a).Verdict != "APPLICABLE" || !reflect.DeepEqual(a, ReferenceAC(a.Build == "queue-idempotent-v1")) {
		return Manifest{}, errors.New("applicability does not authorize compilation")
	}
	return Manifest{Version: "1", PFC: Identity, ACVersion: a.Version, ACDigest: phase1.Digest(a), Build: a.Build, Environment: a.Environment, Fault: Fault, Attempts: 2, Concurrency: 1, Invariant: Invariant, Observers: slices.Clone(observers), Reset: "TRUNCATE consumer reference tables RESTART IDENTITY", AdapterVersion: "synthetic-queue-pg-v1", VerifierVersion: "pfc2-v1", RunID: "pfc2-run", MessageID: "message-X", SafetyLevel: 1}, nil
}
func Plan(p PFC, m Manifest) SafetyPlan {
	s := SafetyPlan{ManifestDigest: phase1.Digest(m), MaxDeliveries: 2, MaxDurationSeconds: 10}
	if Valid(p) && validManifest(m) {
		s.Approved = true
		s.Reason = "bounded synthetic duplicate delivery"
	} else {
		s.Reason = "outside authorized synthetic level-1 experiment"
	}
	return s
}

type Event struct {
	Seq        int    `json:"seq"`
	Kind       string `json:"kind"`
	Attempt    int    `json:"attempt"`
	MessageID  string `json:"message_id"`
	DeliveryID string `json:"delivery_id"`
	EffectID   int64  `json:"effect_id"`
}
type Effect struct {
	ID        int64  `json:"id"`
	Attempt   int    `json:"attempt"`
	MessageID string `json:"message_id"`
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
	delivery := ""
	if attempt > 0 {
		if attempt == 1 {
			delivery = "delivery-1"
		} else {
			delivery = "delivery-2"
		}
	}
	e.Events = append(e.Events, Event{Seq: len(e.Events) + 1, Kind: kind, Attempt: attempt, MessageID: e.Manifest.MessageID, DeliveryID: delivery, EffectID: id})
}

func consume(ctx context.Context, db *sql.DB, m Manifest, attempt int, remediated bool) (int64, bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()
	if remediated {
		var prior int64
		err = tx.QueryRowContext(ctx, "SELECT effect_id FROM consumer_idempotency WHERE run_id=$1 AND message_id=$2 FOR UPDATE", m.RunID, m.MessageID).Scan(&prior)
		if err == nil {
			return prior, false, tx.Commit()
		}
		if err != sql.ErrNoRows {
			return 0, false, err
		}
	}
	var id int64
	err = tx.QueryRowContext(ctx, "INSERT INTO consumer_effects(run_id,message_id,delivery_attempt) VALUES($1,$2,$3) RETURNING effect_id", m.RunID, m.MessageID, attempt).Scan(&id)
	if err != nil {
		return 0, false, err
	}
	if remediated {
		_, err = tx.ExecContext(ctx, "INSERT INTO consumer_idempotency(run_id,message_id,effect_id) VALUES($1,$2,$3)", m.RunID, m.MessageID, id)
		if err != nil {
			return 0, false, err
		}
	}
	return id, true, tx.Commit()
}
func ledger(ctx context.Context, db *sql.DB, m Manifest) ([]Effect, error) {
	rows, err := db.QueryContext(ctx, "SELECT effect_id,delivery_attempt,message_id FROM consumer_effects WHERE run_id=$1 ORDER BY effect_id", m.RunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Effect
	for rows.Next() {
		var x Effect
		if err = rows.Scan(&x.ID, &x.Attempt, &x.MessageID); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func Run(ctx context.Context, db *sql.DB, m Manifest, s SafetyPlan, remediated bool) Evidence {
	ev := Evidence{Manifest: m, Safety: s}
	if !validPlan(m, s) || remediated != (m.Build == "queue-idempotent-v1") {
		ev.Error = "safety plan or build rejected"
		return ev
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.MaxDurationSeconds)*time.Second)
	defer cancel()
	if err := Init(ctx, db); err != nil {
		ev.Error = err.Error()
		return ev
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT (SELECT count(*) FROM consumer_effects)+(SELECT count(*) FROM consumer_idempotency)").Scan(&count); err != nil || count != 0 {
		ev.Error = "clean reset precondition failed"
		return ev
	}
	ev.InitialStateEmpty = true
	ev.add("message_published", 0, 0)
	for attempt := 1; attempt <= 2; attempt++ {
		if attempt == 2 {
			ev.add("duplicate_injected", 0, 0)
		}
		ev.add("delivery_started", attempt, 0)
		ev.add("consumer_received", attempt, 0)
		id, created, err := consume(ctx, db, m, attempt, remediated)
		if err != nil {
			ev.Error = err.Error()
			return ev
		}
		if created {
			ev.add("consumer_committed", attempt, id)
		} else {
			ev.add("consumer_deduped", attempt, id)
		}
		ev.add("delivery_acknowledged", attempt, id)
	}
	var err error
	ev.Effects, err = ledger(ctx, db, m)
	if err != nil {
		ev.Error = err.Error()
		return ev
	}
	var total int
	err = db.QueryRowContext(ctx, "SELECT count(*) FROM consumer_effects").Scan(&total)
	if err != nil {
		ev.Error = err.Error()
		return ev
	}
	ev.LedgerComplete = total == len(ev.Effects)
	ev.RunComplete = true
	return ev
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

// verifyEvidence checks structure only. Its candidate verdict is private to this
// package; only VerifyAgainstDB may return a positive execution proof to callers.
func verifyEvidence(ev Evidence) Proof {
	p := Proof{FormatVersion: "1", Manifest: ev.Manifest, Safety: ev.Safety, Evidence: ev, EvidenceDigest: phase1.Digest(ev), Verdict: "UNKNOWN", Reason: "incomplete or inconsistent evidence", Replay: "reset consumer reference tables; replay the same serial message and mode"}
	if ev.Error != "" {
		p.Reason = ev.Error
		return p
	}
	m := ev.Manifest
	if !validPlan(m, ev.Safety) || !ev.RunComplete || !ev.InitialStateEmpty || !ev.LedgerComplete || len(ev.Events) != 10 {
		return p
	}
	a := ev.Events[3].EffectID
	if a <= 0 {
		return p
	}
	b := ev.Events[8].EffectID
	secondKind := "consumer_committed"
	if m.Build == "queue-idempotent-v1" {
		secondKind = "consumer_deduped"
		b = a
	}
	expected := []Event{{Kind: "message_published"}, {Kind: "delivery_started", Attempt: 1}, {Kind: "consumer_received", Attempt: 1}, {Kind: "consumer_committed", Attempt: 1, EffectID: a}, {Kind: "delivery_acknowledged", Attempt: 1, EffectID: a}, {Kind: "duplicate_injected"}, {Kind: "delivery_started", Attempt: 2}, {Kind: "consumer_received", Attempt: 2}, {Kind: secondKind, Attempt: 2, EffectID: b}, {Kind: "delivery_acknowledged", Attempt: 2, EffectID: b}}
	for i := range expected {
		expected[i].Seq = i + 1
		expected[i].MessageID = m.MessageID
		if expected[i].Attempt > 0 {
			if expected[i].Attempt == 1 {
				expected[i].DeliveryID = "delivery-1"
			} else {
				expected[i].DeliveryID = "delivery-2"
			}
		}
		if ev.Events[i] != expected[i] {
			return p
		}
	}
	if m.Build == "queue-vulnerable-v1" {
		if b <= 0 || a == b || len(ev.Effects) != 2 || ev.Effects[0] != (Effect{ID: a, Attempt: 1, MessageID: m.MessageID}) || ev.Effects[1] != (Effect{ID: b, Attempt: 2, MessageID: m.MessageID}) {
			return p
		}
		p.Verdict = "EXPOSED"
		p.Reason = "two distinct committed consumer effects for the same logical message"
		return p
	}
	if len(ev.Effects) != 1 || ev.Effects[0] != (Effect{ID: a, Attempt: 1, MessageID: m.MessageID}) {
		return p
	}
	p.Verdict = "PROVEN_RESILIENT"
	p.InvariantHolds = true
	p.Reason = "one committed effect across the complete two-delivery serial experiment"
	return p
}

// VerifyAgainstDB re-reads the durable effects and consumer control state before
// issuing a CLI verdict. The local evidence digest is not an attestation.
func VerifyAgainstDB(ctx context.Context, db *sql.DB, ev Evidence) Proof {
	p := verifyEvidence(ev)
	if p.Verdict == "UNKNOWN" {
		return p
	}
	unknown := func(reason string) Proof {
		p.Verdict = "UNKNOWN"
		p.InvariantHolds = false
		p.Reason = reason
		return p
	}
	if db == nil {
		return unknown("authoritative PostgreSQL connection unavailable")
	}
	actual, err := ledger(ctx, db, ev.Manifest)
	if err != nil || !reflect.DeepEqual(actual, ev.Effects) {
		return unknown("authoritative consumer ledger unavailable or changed")
	}
	var effectCount, controlCount int
	if err = db.QueryRowContext(ctx, "SELECT count(*) FROM consumer_effects").Scan(&effectCount); err != nil || effectCount != len(actual) {
		return unknown("authoritative consumer ledger incomplete")
	}
	if err = db.QueryRowContext(ctx, "SELECT count(*) FROM consumer_idempotency").Scan(&controlCount); err != nil {
		return unknown("durable idempotency state unavailable")
	}
	if ev.Manifest.Build == "queue-vulnerable-v1" {
		if controlCount != 0 {
			return unknown("unexpected durable consumer control state")
		}
		return p
	}
	var prior int64
	err = db.QueryRowContext(ctx, "SELECT effect_id FROM consumer_idempotency WHERE run_id=$1 AND message_id=$2", ev.Manifest.RunID, ev.Manifest.MessageID).Scan(&prior)
	if err != nil || controlCount != 1 || prior != actual[0].ID {
		return unknown("durable consumer dedupe does not bind first effect")
	}
	return p
}
