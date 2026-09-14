package pfc3

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

const Identity = "pfc.stale-worker-after-lease-expiry@1.0.0"
const Invariant = "an older fencing token must never commit a protected effect after a newer token becomes authoritative"
const Fault = "DELAY_WORKER_PAST_LEASE_EXPIRY"
const vulnerableBuild = "lease-vulnerable-v1"
const remediatedBuild = "lease-fenced-v1"

var prerequisites = []string{"time_bounded_lease", "ownership_expires", "work_outlives_lease", "reacquisition_after_expiry", "stale_worker_resumes", "durable_protected_effect", "ownership_token_observable", "protected_effect_observable"}
var sequence = []string{"worker A acquires token 1 and starts work", "worker A stalls past deterministic lease expiry", "worker B acquires newer token 2", "worker B commits protected effect", "worker A resumes and attempts protected write with stale token 1"}
var proofNeeds = []string{"authoritative PostgreSQL lease and token history", "complete ordered worker and expiry trace", "authoritative protected-effect ledger", "durable stale-write attempt and rejection when fenced"}

type PFC = phase1.PFC
type AC = phase1.AC
type Match = phase1.Match

func Load(path string) (PFC, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return PFC{}, e
	}
	var p PFC
	if e = json.Unmarshal(b, &p); e != nil {
		return PFC{}, e
	}
	if !Valid(p) {
		return PFC{}, errors.New("unsupported family-3 PFC")
	}
	return p, nil
}
func Valid(p PFC) bool {
	return p.ID == "pfc.stale-worker-after-lease-expiry" && p.Version == "1.0.0" && p.Family == "stale worker after lease expiry" && p.Lifecycle == "candidate" && p.VerificationLevel == "L0 CANDIDATE" && p.Provenance == "Gate A family 3 mechanism mapping; synthetic execution is not public incident evidence" && slices.Equal(p.Prerequisites, prerequisites) && p.ControlUnderTest == "protected-resource fencing token enforcement" && slices.Equal(p.CausalSequence, sequence) && p.ForbiddenOutcome == "older token commits protected effect after newer token becomes authoritative" && p.Invariant == Invariant && p.Experiment.Fault == Fault && p.Experiment.Attempts == 2 && p.Experiment.Concurrency == 2 && slices.Equal(p.ProofRequirements, proofNeeds) && p.SafetyCeiling == 1
}
func ReferenceAC(remediated bool) AC {
	a := AC{SchemaVersion: "1", ID: "reference-lease", Version: "1", Build: vulnerableBuild, Environment: "synthetic-reference", Runtime: "Go", DatabaseBoundary: "PostgreSQL lease and protected resource", ProviderBoundary: "deterministic logical expiry", RetryPolicy: "stalled worker A resumes after worker B commit", ProviderGuarantee: "one current owner and monotonically increasing durable token", LogicalIdentity: "run ID and resource-1", Idempotency: "absent", Reconciliation: "absent", Facts: map[string]phase1.Fact{}}
	if remediated {
		a.Build = remediatedBuild
	}
	for _, k := range append(slices.Clone(prerequisites), "fencing_enforced") {
		v := true
		if k == "fencing_enforced" {
			v = remediated
		}
		a.Facts[k] = phase1.Fact{Value: phase1.Bool(v), Evidence: "PostgreSQL reference schema and adapter", Source: "reference implementation", Scope: "one resource and two logical workers", Quality: "direct", Sensitivity: "synthetic"}
	}
	return a
}
func Evaluate(p PFC, a AC) Match {
	m := Match{Verdict: "UNKNOWN", PFC: p.ID + "@" + p.Version, AC: phase1.Digest(a)}
	if !Valid(p) {
		return m
	}
	m = phase1.Evaluate(p, a)
	for _, k := range append(slices.Clone(prerequisites), "fencing_enforced") {
		f, ok := a.Facts[k]
		if !ok || f.Value == nil || f.Evidence == "" || f.Source == "" || f.Scope != "one resource and two logical workers" || f.Quality != "direct" || f.Sensitivity != "synthetic" || f.Stale || f.Conflict {
			m.Verdict = "UNKNOWN"
		}
	}
	if a.ID != "reference-lease" || a.Version != "1" || a.Runtime != "Go" || a.Environment != "synthetic-reference" || a.DatabaseBoundary != "PostgreSQL lease and protected resource" || a.ProviderBoundary != "deterministic logical expiry" || a.RetryPolicy != "stalled worker A resumes after worker B commit" || a.ProviderGuarantee != "one current owner and monotonically increasing durable token" || a.LogicalIdentity != "run ID and resource-1" || (a.Build != vulnerableBuild && a.Build != remediatedBuild) {
		m.Verdict = "UNKNOWN"
	}
	return m
}
func Compile(p PFC, a AC, match Match) (Manifest, error) {
	if !Valid(p) || match.Verdict != "APPLICABLE" || match.PFC != Identity || match.AC != phase1.Digest(a) || Evaluate(p, a).Verdict != "APPLICABLE" || !reflect.DeepEqual(a, ReferenceAC(a.Build == remediatedBuild)) {
		return Manifest{}, errors.New("applicability does not authorize compilation")
	}
	return Manifest{Version: "1", PFC: Identity, ACVersion: "1", ACDigest: phase1.Digest(a), Build: a.Build, Environment: a.Environment, Fault: Fault, Attempts: 2, Concurrency: 2, Invariant: Invariant, ResourceID: "resource-1", WorkerA: "worker-A", WorkerB: "worker-B", TokenA: 1, TokenB: 2, AdapterVersion: "logical-lease-pg-v1", VerifierVersion: "pfc3-v1", RunID: "pfc3-run", SafetyLevel: 1}, nil
}
