package pfc2

import (
	"encoding/json"
	"errors"
	"os"
	"slices"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

const Identity = "pfc.duplicate-queue-delivery@1.0.0"
const Invariant = "at most one committed consumer effect per logical message"
const Fault = "DUPLICATE_OPERATION_MESSAGE"

var prerequisites = []string{"queue_delivery", "duplicate_possible", "stable_message_identity", "durable_consumer_effect", "same_message_redelivery", "effect_ledger_observable"}
var sequence = []string{"one logical message published", "delivery one reaches consumer and commits effect", "delivery one acknowledged", "duplicate operation injects the same logical message", "delivery two reaches consumer and may commit another effect"}
var proofNeeds = []string{"complete authoritative PostgreSQL effect ledger", "same logical message in both deliveries", "ordered duplicate injection after first acknowledgement", "both delivery attempts reach consumer", "dedupe references first committed effect when remediated"}

type PFC = phase1.PFC
type AC = phase1.AC
type Match = phase1.Match

func Load(path string) (PFC, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return PFC{}, err
	}
	var p PFC
	if err = json.Unmarshal(b, &p); err != nil {
		return PFC{}, err
	}
	if !Valid(p) {
		return PFC{}, errors.New("unsupported or incomplete family-2 PFC")
	}
	return p, nil
}

func Valid(p PFC) bool {
	return p.ID == "pfc.duplicate-queue-delivery" && p.Version == "1.0.0" && p.Family == "duplicate queue delivery" &&
		p.Lifecycle == "candidate" && p.VerificationLevel == "L0 CANDIDATE" &&
		p.Provenance == "Gate A family 2 mechanism mapping; synthetic execution is separate from public incident evidence" &&
		slices.Equal(p.Prerequisites, prerequisites) &&
		p.ControlUnderTest == "durable consumer idempotency keyed by run and logical message ID" &&
		slices.Equal(p.CausalSequence, sequence) && p.ForbiddenOutcome == "more than one committed consumer effect for one logical message" &&
		p.Invariant == Invariant && p.Experiment.Fault == Fault && p.Experiment.Attempts == 2 && p.Experiment.Concurrency == 1 &&
		slices.Equal(p.ProofRequirements, proofNeeds) && p.SafetyCeiling == 1
}

func ReferenceAC(remediated bool) AC {
	a := AC{SchemaVersion: "1", ID: "reference-queue", Version: "1", Build: "queue-vulnerable-v1", Environment: "synthetic-reference", Runtime: "Go", DatabaseBoundary: "consumer effects committed in PostgreSQL", ProviderBoundary: "deterministic in-process synthetic queue", RetryPolicy: "one serial duplicate delivery after first acknowledgement", ProviderGuarantee: "same logical message may be delivered twice", LogicalIdentity: "run ID plus stable logical message ID", Idempotency: "absent", Reconciliation: "absent", Facts: map[string]phase1.Fact{}}
	if remediated {
		a.Build = "queue-idempotent-v1"
		a.Idempotency = "durable consumer key (run ID, logical message ID)"
	}
	for _, k := range prerequisites {
		a.Facts[k] = phase1.Fact{Value: phase1.Bool(true), Evidence: "synthetic queue adapter and PostgreSQL schema", Source: "reference implementation", Scope: "one serial logical message", Quality: "direct", Sensitivity: "synthetic"}
	}
	a.Facts["durable_consumer_idempotency"] = phase1.Fact{Value: phase1.Bool(remediated), Evidence: "consumer transaction and PostgreSQL idempotency table", Source: "reference implementation", Scope: "one serial logical message", Quality: "direct", Sensitivity: "synthetic"}
	return a
}

func Evaluate(p PFC, a AC) Match {
	m := Match{Verdict: "UNKNOWN", PFC: p.ID + "@" + p.Version, AC: phase1.Digest(a)}
	if !Valid(p) {
		return m
	}
	m = phase1.Evaluate(p, a)
	for _, k := range append(slices.Clone(prerequisites), "durable_consumer_idempotency") {
		f, ok := a.Facts[k]
		if !ok || f.Value == nil || f.Evidence == "" || f.Source == "" || f.Scope != "one serial logical message" || f.Quality != "direct" || f.Sensitivity != "synthetic" || f.Stale || f.Conflict {
			m.Verdict = "UNKNOWN"
		}
	}
	if a.ID != "reference-queue" || a.Version != "1" || a.Runtime != "Go" || a.ProviderBoundary != "deterministic in-process synthetic queue" || a.DatabaseBoundary != "consumer effects committed in PostgreSQL" || a.RetryPolicy != "one serial duplicate delivery after first acknowledgement" || a.ProviderGuarantee != "same logical message may be delivered twice" || a.LogicalIdentity != "run ID plus stable logical message ID" || (a.Build != "queue-vulnerable-v1" && a.Build != "queue-idempotent-v1") {
		m.Verdict = "UNKNOWN"
	}
	return m
}
