package phase1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"slices"
)

type PFC struct {
	ID                string     `json:"id"`
	Version           string     `json:"version"`
	Family            string     `json:"family"`
	Lifecycle         string     `json:"lifecycle"`
	VerificationLevel string     `json:"verification_level"`
	Provenance        string     `json:"provenance"`
	Prerequisites     []string   `json:"prerequisites"`
	ControlUnderTest  string     `json:"control_under_test"`
	CausalSequence    []string   `json:"causal_sequence"`
	ForbiddenOutcome  string     `json:"forbidden_outcome"`
	Invariant         string     `json:"invariant"`
	Experiment        Experiment `json:"experiment"`
	ProofRequirements []string   `json:"proof_requirements"`
	SafetyCeiling     int        `json:"safety_ceiling"`
}
type Experiment struct {
	Fault       string `json:"fault"`
	Attempts    int    `json:"attempts"`
	Concurrency int    `json:"concurrency"`
}

var required = []string{"external_effect", "post_commit_uncertainty", "post_commit_fault_placeable", "same_operation_retry", "effect_ledger_observable"}
var canonicalSequence = []string{"caller requests one logical payment", "provider durably commits effect", "response is lost before caller success", "caller retries same logical payment", "provider may commit another effect"}
var canonicalProof = []string{"authoritative complete ledger", "ordered commit-before-loss", "caller-observed loss before retry", "same logical operation identity", "all attempts and effects observed"}

const pfcIdentity = "pfc.timeout-after-external-commit@1.0.0"
const phase1Invariant = "at most one committed provider effect per logical payment operation"
const phase1Fault = "DROP_RESPONSE_AFTER_REMOTE_COMMIT"

func LoadPFC(path string) (PFC, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return PFC{}, e
	}
	var p PFC
	if e = json.Unmarshal(b, &p); e != nil {
		return p, e
	}
	if !validPFC(p) {
		return p, errors.New("unsupported or incomplete family-1 PFC")
	}
	return p, nil
}

func validPFC(p PFC) bool {
	return !(p.ID != "pfc.timeout-after-external-commit" || p.Version != "1.0.0" ||
		p.Lifecycle != "candidate" || p.VerificationLevel != "L0 CANDIDATE" ||
		p.Family != "timeout after external commit" ||
		p.ControlUnderTest != "durable provider idempotency keyed by logical operation ID" ||
		p.ForbiddenOutcome != "more than one committed provider effect for one logical payment operation" ||
		p.Invariant != phase1Invariant || p.Experiment.Fault != phase1Fault ||
		p.Experiment.Attempts != 2 || p.Experiment.Concurrency != 1 || p.SafetyCeiling != 1 ||
		p.Provenance == "" || !slices.Equal(p.Prerequisites, required) ||
		!slices.Equal(p.CausalSequence, canonicalSequence) || !slices.Equal(p.ProofRequirements, canonicalProof))
}

type Fact struct {
	Value       *bool  `json:"value"`
	Evidence    string `json:"evidence"`
	Source      string `json:"source"`
	Scope       string `json:"scope"`
	Quality     string `json:"quality"`
	Sensitivity string `json:"sensitivity"`
	Conflict    bool   `json:"conflict,omitempty"`
	Stale       bool   `json:"stale,omitempty"`
}
type AC struct {
	SchemaVersion     string          `json:"schema_version"`
	ID                string          `json:"id"`
	Version           string          `json:"version"`
	Build             string          `json:"build"`
	Environment       string          `json:"environment"`
	Runtime           string          `json:"runtime"`
	DatabaseBoundary  string          `json:"database_boundary"`
	ProviderBoundary  string          `json:"provider_boundary"`
	RetryPolicy       string          `json:"retry_policy"`
	ProviderGuarantee string          `json:"provider_guarantee"`
	LogicalIdentity   string          `json:"logical_identity"`
	Idempotency       string          `json:"idempotency"`
	Reconciliation    string          `json:"reconciliation"`
	Facts             map[string]Fact `json:"facts"`
}

func Bool(v bool) *bool { return &v }
func ReferenceAC(remediated bool) AC {
	a := AC{SchemaVersion: "1", ID: "reference-payment", Version: "1", Build: "reference-vulnerable-v1", Environment: "synthetic-reference", Runtime: "Go", DatabaseBoundary: "business state in PostgreSQL; external provider commit is independent", ProviderBoundary: "separate simulator transaction commits provider effect", RetryPolicy: "one immediate retry of same logical ID after lost response", ProviderGuarantee: "append-only effect ledger; optional durable idempotency key", LogicalIdentity: "run ID plus logical payment ID", Idempotency: "absent", Reconciliation: "absent", Facts: map[string]Fact{}}
	if remediated {
		a.Build = "reference-idempotent-v1"
		a.Idempotency = "provider durable key (run ID, logical ID)"
	}
	for _, k := range required {
		a.Facts[k] = Fact{Value: Bool(true), Evidence: "reference adapter and PostgreSQL schema", Source: "reference implementation", Scope: "payment operation", Quality: "direct", Sensitivity: "synthetic"}
	}
	return a
}
func Digest(v any) string {
	b, _ := json.Marshal(v)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

type FactDecision struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Evidence string `json:"evidence,omitempty"`
	Reason   string `json:"reason"`
}
type Match struct {
	Verdict string         `json:"verdict"`
	PFC     string         `json:"pfc"`
	AC      string         `json:"ac"`
	Facts   []FactDecision `json:"facts"`
}

func Evaluate(p PFC, a AC) Match {
	m := Match{Verdict: "UNKNOWN", PFC: p.ID + "@" + p.Version, AC: Digest(a)}
	falseFound := false
	unknown := false
	for _, k := range p.Prerequisites {
		f, ok := a.Facts[k]
		d := FactDecision{Name: k, Evidence: f.Evidence}
		switch {
		case !ok || f.Value == nil || f.Evidence == "" || f.Source == "" || f.Stale || f.Conflict || f.Quality != "direct":
			d.State = "UNKNOWN"
			d.Reason = "missing, stale, conflicting, or unsupported fact"
			unknown = true
		case !*f.Value:
			d.State = "FALSE"
			d.Reason = "necessary causal prerequisite absent"
			falseFound = true
		default:
			d.State = "TRUE"
			d.Reason = "evidenced prerequisite"
		}
		m.Facts = append(m.Facts, d)
	}
	if a.SchemaVersion != "1" || a.Build == "" || a.Environment == "" || a.ProviderBoundary == "" || a.DatabaseBoundary == "" || a.LogicalIdentity == "" {
		unknown = true
	}
	if falseFound {
		m.Verdict = "NOT_APPLICABLE"
	} else if !unknown {
		m.Verdict = "APPLICABLE"
	}
	return m
}
