// Package corpus is a local metadata register, never a causal evidence authority.
package corpus

import "github.com/SamudralaAjaykumarrr/failuremesh/internal/evidence"

const (
	Schema            = "corpus-v1"
	ManifestSchema    = "corpus-manifest-v1"
	Encoding          = "corpus-go-json-v1"
	MaxObjectBytes    = 4 << 20
	MaxManifestBytes  = 4 << 20
	MaxSources        = 64
	MaxRelations      = 4096
	MaxExposures      = 8192
	MaxClosureObjects = 16384
	MaxClosureBytes   = 64 << 20
	MaxStoreFiles     = 65536
	MaxStoreBytes     = 256 << 20
	MaxOrdinal        = 65535
)

type Digest = evidence.Digest

var families = []string{"durable-effect replay/identity", "ownership/lease/fencing", "shared-capacity/deadline propagation", "retry feedback amplification", "cache/coordination contention"}
var categories = []string{"target_pfc", "matcher_rules", "expected_result", "prior_case_outcome", "architecture_packet", "causal_family_label", "previous_evaluation_result"}

type Identity struct {
	Actor   string `json:"actor"`
	Class   string `json:"class"`
	Role    string `json:"role"`
	Tool    string `json:"tool"`
	Version string `json:"version"`
	Model   string `json:"model"`
	Context string `json:"context"`
}
type Fact struct {
	State  string   `json:"state"`
	Value  *string  `json:"value"`
	Reason string   `json:"reason"`
	Basis  []string `json:"basis"`
}
type Date struct {
	Value     Fact   `json:"value"`
	Precision string `json:"precision"`
	Timezone  string `json:"timezone"`
}
type Ref struct {
	ID       string `json:"id"`
	Revision uint32 `json:"revision"`
	Digest   Digest `json:"digest"`
}
type Citation struct {
	ID             string          `json:"id"`
	Locator        string          `json:"locator"`
	PublisherBasis string          `json:"publisher_basis"`
	SourceType     string          `json:"source_type"`
	Role           string          `json:"role"`
	Status         string          `json:"status"`
	Limitation     string          `json:"limitation"`
	Rights         evidence.Rights `json:"rights"`
	Availability   string          `json:"availability"`
	CheckedAt      string          `json:"checked_at"`
}
type Captured struct {
	CitationID       string             `json:"citation_id"`
	StoreID          string             `json:"store_id"`
	Record           Digest             `json:"record"`
	Source           evidence.SourceRef `json:"source"`
	Representation   string             `json:"representation"`
	Locations        []string           `json:"locations"`
	ParentDependency string             `json:"parent_dependency"`
}
type Discovery struct {
	Method              string   `json:"method"`
	Discoverer          Identity `json:"discoverer"`
	ReportedTime        string   `json:"reported_time"`
	Lead                Fact     `json:"lead"`
	Scope               string   `json:"scope"`
	Round               string   `json:"round"`
	RetrievalPermission string   `json:"retrieval_permission"`
	RetrievalBasis      string   `json:"retrieval_basis"`
}
type Attribution struct {
	Group          Fact     `json:"group"`
	Names          []string `json:"names"`
	OwnershipBasis string   `json:"ownership_basis"`
}
type Family struct {
	Vocabulary   Digest   `json:"vocabulary"`
	State        string   `json:"state"`
	Primary      Fact     `json:"primary"`
	Alternatives []string `json:"alternatives"`
	Mechanism    string   `json:"mechanism"`
	Scope        string   `json:"scope"`
	Rationale    string   `json:"rationale"`
}
type Assessment struct {
	Dimension  string   `json:"dimension"`
	State      string   `json:"state"`
	Rationale  string   `json:"rationale"`
	References []string `json:"references"`
	Assessor   string   `json:"assessor"`
}
type Decision struct {
	State        string   `json:"state"`
	Reasons      []string `json:"reasons"`
	Rationale    string   `json:"rationale"`
	Policy       Ref      `json:"policy"`
	Actor        string   `json:"actor"`
	Round        string   `json:"round"`
	Alternatives string   `json:"alternatives"`
}
type Candidate struct {
	Discovery           Discovery    `json:"discovery"`
	Sources             []Citation   `json:"source_references"`
	Captured            []Captured   `json:"captured_sources"`
	Publisher           Attribution  `json:"publisher"`
	Project             Fact         `json:"project"`
	ProjectLineage      Fact         `json:"project_lineage"`
	IncidentTime        Date         `json:"incident_time"`
	PublicationTime     Date         `json:"publication_time"`
	System              Fact         `json:"system"`
	Stack               Fact         `json:"stack"`
	ArchitectureRefs    []string     `json:"architecture_refs"`
	Family              Family       `json:"family"`
	ChallengeCategories []string     `json:"challenge_categories"`
	OriginClass         string       `json:"origin_class"`
	Assessments         []Assessment `json:"selection_assessments"`
	Origin              Fact         `json:"origin_cluster"`
	Episode             Fact         `json:"causal_episode"`
	Independence        Fact         `json:"independence"`
	ExposureRefs        []Digest     `json:"exposure_refs"`
	Pretraining         string       `json:"model_pretraining"`
	Decision            Decision     `json:"decision"`
	Split               string       `json:"split_snapshot"`
	Assignment          *Digest      `json:"assignment_snapshot"`
	FreezeState         string       `json:"freeze_state"`
	Parents             []Ref        `json:"synthetic_parents"`
	Transformation      string       `json:"transformation"`
}
type Budget struct {
	Family         string `json:"family"`
	Candidates     uint32 `json:"candidates"`
	CuratorMinutes uint32 `json:"curator_minutes"`
}
type Target struct {
	Name  string `json:"name"`
	Count uint32 `json:"count"`
}
type Policy struct {
	Round               string   `json:"round"`
	Vocabulary          Digest   `json:"vocabulary"`
	Scope               string   `json:"scope"`
	Channels            []string `json:"manual_channels"`
	Start               string   `json:"start"`
	Cutoff              string   `json:"cutoff"`
	Budgets             []Budget `json:"budgets"`
	Eligibility         []string `json:"eligibility"`
	Challenges          []string `json:"challenges"`
	Targets             []Target `json:"targets"`
	PublisherMultiplier uint32   `json:"publisher_multiplier"`
	DiversityException  string   `json:"diversity_exception"`
	SplitMethod         string   `json:"split_method"`
	Seed                string   `json:"seed"`
	RightsRequirements  []string `json:"rights_requirements"`
	ExposureExclusions  []string `json:"exposure_exclusions"`
	StopRules           []string `json:"stop_rules"`
}
type Relation struct {
	ID        string `json:"id"`
	From      Ref    `json:"from"`
	To        Ref    `json:"to"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Basis     string `json:"basis"`
	Rationale string `json:"rationale"`
}
type Assignment struct {
	Candidate string `json:"candidate"`
	Split     string `json:"split"`
	Rationale string `json:"rationale"`
}
type Exposure struct {
	Candidates      []string `json:"candidates"`
	Round           string   `json:"round"`
	Category        string   `json:"category"`
	State           string   `json:"state"`
	AvailableAccess string   `json:"available_access"`
	Basis           string   `json:"basis"`
	Restrictions    string   `json:"restrictions"`
	OccurredAt      Fact     `json:"occurred_at"`
	RecordedAt      string   `json:"recorded_at"`
	Artifacts       []Digest `json:"artifacts"`
}
type Notice struct {
	Manifest   Digest   `json:"manifest"`
	Candidates []string `json:"candidates"`
	Reason     string   `json:"reason"`
}
type Member struct {
	Candidate      string   `json:"candidate"`
	Origin         string   `json:"origin"`
	Component      string   `json:"component"`
	Representative bool     `json:"representative"`
	Natural        bool     `json:"natural"`
	Publisher      string   `json:"publisher"`
	Project        string   `json:"project"`
	Stack          string   `json:"stack"`
	Family         string   `json:"family"`
	Decision       string   `json:"decision"`
	Split          string   `json:"split"`
	Exposure       []Access `json:"exposure"`
	Freshness      string   `json:"freshness"`
	Audit          string   `json:"audit"`
}
type Access struct {
	Category string `json:"category"`
	State    string `json:"state"`
}
type ContextAccess struct {
	Component string   `json:"component"`
	Context   string   `json:"actor_context_model"`
	Access    []Access `json:"access"`
}
type Summary struct {
	Contexts     []ContextAccess `json:"contexts"`
	Members      []Member        `json:"members"`
	Counts       []Target        `json:"counts"`
	Deficits     []Target        `json:"deficits"`
	Limitations  []string        `json:"limitations"`
	TargetStatus string          `json:"target_status"`
}
type Manifest struct {
	Inventory      []Ref    `json:"inventory"`
	Closure        []Digest `json:"closure"`
	LedgerHead     *Digest  `json:"ledger_head"`
	LedgerSequence uint32   `json:"ledger_sequence"`
	Policy         Ref      `json:"policy"`
	Assignment     Ref      `json:"assignment"`
	CheckTime      string   `json:"check_time"`
	Build          Digest   `json:"build"`
	SourceCommit   string   `json:"source_commit"`
	DirtyTree      *Digest  `json:"dirty_tree"`
	ParentManifest *Digest  `json:"parent_manifest"`
	Summary        Summary  `json:"summary"`
	Method         string   `json:"method"`
	Seed           string   `json:"seed"`
	Chronology     string   `json:"chronology"`
	Pretraining    string   `json:"pretraining"`
}

// Object is an explicit tagged union. Exactly one payload matches object_kind.
// Policy objects bind the code-owned vocabulary as their policy digest, avoiding
// self-reference. All other objects bind an already registered policy object.
type Object struct {
	Schema        string        `json:"schema_version"`
	Encoding      string        `json:"encoding"`
	Kind          string        `json:"object_kind"`
	ID            string        `json:"logical_id"`
	Revision      uint32        `json:"revision"`
	Parent        *Digest       `json:"parent_revision_digest"`
	Producer      Identity      `json:"producer"`
	PolicyDigest  Digest        `json:"policy_digest"`
	ReportedAt    string        `json:"reported_at"`
	Sensitivity   string        `json:"sensitivity"`
	Reasons       []string      `json:"reasons"`
	Rationale     string        `json:"rationale"`
	Policy        *Policy       `json:"policy"`
	Candidate     *Candidate    `json:"candidate"`
	Relations     *[]Relation   `json:"relations"`
	Assignments   *[]Assignment `json:"assignments"`
	Exposure      *Exposure     `json:"exposure"`
	AttemptReason *string       `json:"attempt_reason"`
	Manifest      *Manifest     `json:"manifest"`
	Notice        *Notice       `json:"notice"`
}
type Event struct {
	Schema     string   `json:"schema_version"`
	Encoding   string   `json:"encoding"`
	Sequence   uint32   `json:"sequence"`
	Previous   *Digest  `json:"previous_digest"`
	Operation  string   `json:"operation_id"`
	Request    Digest   `json:"request_digest"`
	Action     string   `json:"action"`
	Output     Ref      `json:"output"`
	Producer   Identity `json:"producer"`
	ReportedAt string   `json:"reported_at"`
}
type Request struct {
	Operation    string  `json:"operation_id"`
	ExpectedHead *Digest `json:"expected_head"`
	Action       string  `json:"action"`
	Object       Object  `json:"object"`
}
type Result struct {
	Object   Digest `json:"object"`
	Head     Digest `json:"head"`
	Sequence uint32 `json:"sequence"`
	Reason   string `json:"reason"`
}
type Report struct {
	Integrity             string   `json:"integrity"`
	SelectionTargetStatus string   `json:"selection_target_status"`
	CurrentUseStatus      string   `json:"current_use_status"`
	AuditCapability       string   `json:"audit_capability"`
	Freshness             string   `json:"freshness"`
	ChronologyTier        string   `json:"chronology_tier"`
	Pretraining           string   `json:"pretraining"`
	Historical            Summary  `json:"historical"`
	Current               *Summary `json:"current"`
	Notices               []Ref    `json:"notices"`
	Head                  *Digest  `json:"head"`
	Limitations           []string `json:"limitations"`
}

// Vocabulary accessors return copies; callers cannot change the pinned profile.
func FamilyVocabulary() []string   { return append([]string(nil), families...) }
func ExposureCategories() []string { return append([]string(nil), categories...) }
