// Package evidence implements the Phase 2A local, structural evidence substrate.
// Its synthetic rule validates fixture semantics, not real-world observations.
package evidence

const (
	Schema   = "evidence-v1"
	Policy   = "structural-v1"
	Encoding = "go-json-v1"
)

type Digest struct {
	Algorithm string `json:"algorithm"`
	Hex       string `json:"hex"`
}

type Rights struct {
	License      string `json:"license"`
	Basis        string `json:"basis"`
	Review       string `json:"review"`
	Metadata     string `json:"metadata"`
	Digest       string `json:"digest"`
	Excerpt      string `json:"excerpt"`
	WholeBody    string `json:"whole_body"`
	Export       string `json:"export"`
	ExcerptScope string `json:"excerpt_scope"`
	RetainUntil  string `json:"retain_until"`
}

type Retrieval struct {
	Method           string   `json:"method"`
	Collector        string   `json:"collector"`
	Version          string   `json:"version"`
	IdentityClass    string   `json:"identity_class"`
	Model            string   `json:"model"`
	Redirects        []string `json:"redirects"`
	Status           string   `json:"status"`
	ContentType      string   `json:"content_type"`
	UpstreamRevision string   `json:"upstream_revision"`
	Errors           []string `json:"errors"`
}

type Quality struct {
	Attribution  string `json:"attribution"`
	Directness   string `json:"directness"`
	ScopeFit     string `json:"scope_fit"`
	VersionFit   string `json:"version_fit"`
	Integrity    string `json:"integrity"`
	Completeness string `json:"completeness"`
	Independence string `json:"independence"`
}

type Relation struct {
	Kind       string `json:"kind"`
	ArtifactID string `json:"artifact_id"`
	Basis      string `json:"basis"`
}

type Selector struct {
	Kind           string `json:"kind"`
	Representation string `json:"representation"`
	Start          int64  `json:"start"`
	End            int64  `json:"end"`
	Label          string `json:"label"`
}

type Location struct {
	ID             string   `json:"id"`
	ArtifactID     string   `json:"artifact_id"`
	ArtifactDigest Digest   `json:"artifact_digest"`
	Selector       Selector `json:"selector"`
	ExcerptDigest  *Digest  `json:"excerpt_digest"`
	Excerpt        []byte   `json:"excerpt"`
	Sensitivity    string   `json:"sensitivity"`
	Access         string   `json:"access"`
	Limitation     string   `json:"limitation"`
}

type SourceArtifact struct {
	SchemaVersion             string     `json:"schema_version"`
	SourceID                  string     `json:"source_id"`
	ArtifactID                string     `json:"artifact_id"`
	Revision                  int        `json:"revision"`
	Supersedes                string     `json:"supersedes"`
	CanonicalURL              string     `json:"canonical_url"`
	RetrievedURL              string     `json:"retrieved_url"`
	Publisher                 string     `json:"publisher"`
	PublisherIdentityBasis    string     `json:"publisher_identity_basis"`
	PublicationTimestamp      string     `json:"publication_timestamp"`
	PublicationTimestampBasis string     `json:"publication_timestamp_basis"`
	RetrievedAt               string     `json:"retrieved_at"`
	SourceType                string     `json:"source_type"`
	QualityClass              string     `json:"quality_class"`
	QualityPolicyVersion      string     `json:"quality_policy_version"`
	Quality                   Quality    `json:"quality"`
	ContentDigest             Digest     `json:"content_digest"`
	Representation            string     `json:"representation"`
	ByteLength                int64      `json:"byte_length"`
	Retrieval                 Retrieval  `json:"retrieval_provenance"`
	Rights                    Rights     `json:"rights"`
	Locations                 []Location `json:"locations"`
	Retained                  bool       `json:"retained"`
	Body                      []byte     `json:"body"`
	Mutability                string     `json:"mutability"`
	Availability              string     `json:"availability"`
	CheckedAt                 string     `json:"checked_at"`
	Relations                 []Relation `json:"relations"`
	Sensitivity               string     `json:"sensitivity"`
}

type Scope struct {
	Build          string `json:"build"`
	Configuration  string `json:"configuration"`
	Environment    string `json:"environment"`
	Run            string `json:"run"`
	Operation      string `json:"operation"`
	IdentityDomain string `json:"identity_domain"`
	Quantifier     string `json:"quantifier"`
}

type Value struct {
	Type          string  `json:"type"`
	Boolean       *bool   `json:"boolean"`
	Text          *string `json:"text"`
	UnknownReason string  `json:"unknown_reason"`
}

type SourceRef struct {
	SourceID   string `json:"source_id"`
	ArtifactID string `json:"artifact_id"`
	Revision   int    `json:"revision"`
	Digest     Digest `json:"digest"`
}

type Extraction struct {
	Method         string   `json:"method"`
	Tool           string   `json:"tool"`
	Version        string   `json:"version"`
	InputDigests   []Digest `json:"input_digests"`
	Transformation string   `json:"transformation"`
	Model          string   `json:"model"`
	PromptDigest   *Digest  `json:"prompt_digest"`
	ProposerClass  string   `json:"proposer_class"`
	Context        string   `json:"context"`
}

type Confidence struct {
	Score  float64 `json:"score"`
	Method string  `json:"method"`
}

type Freshness struct {
	ObservedAt       string `json:"observed_at"`
	ValidFrom        string `json:"valid_from"`
	ValidUntil       string `json:"valid_until"`
	PinnedBuild      string `json:"pinned_build"`
	RevalidationRule string `json:"revalidation_rule"`
}

type Derivation struct {
	Parents       []string `json:"parents"`
	ParentDigests []Digest `json:"parent_digests"`
	Rule          string   `json:"rule"`
	Version       string   `json:"version"`
	CausalRefs    []string `json:"causal_refs"`
}

type AtomicClaim struct {
	SchemaVersion      string      `json:"schema_version"`
	ID                 string      `json:"claim_id"`
	Revision           int         `json:"revision"`
	Predicate          string      `json:"predicate"`
	Value              Value       `json:"value"`
	Subject            string      `json:"subject"`
	Scope              Scope       `json:"scope"`
	EpistemicBasis     string      `json:"epistemic_basis"`
	SourceRefs         []SourceRef `json:"source_refs"`
	EvidenceRefs       []string    `json:"evidence_refs"`
	Extraction         Extraction  `json:"extraction"`
	AdvisoryConfidence *Confidence `json:"advisory_confidence"`
	Freshness          Freshness   `json:"freshness"`
	Sensitivity        string      `json:"sensitivity"`
	Boundary           string      `json:"boundary"`
	Derivation         Derivation  `json:"derivation"`
	// Candidate metadata is never an input to authoritative state selection.
	SupportStatus       string `json:"support_status"`
	ContradictionStatus string `json:"contradiction_status"`
}

type State string

const (
	SupportedTrue  State = "SUPPORTED_TRUE"
	SupportedFalse State = "SUPPORTED_FALSE"
	Unknown        State = "UNKNOWN"
	Conflicting    State = "CONFLICTING"
	Stale          State = "STALE"
	Unsupported    State = "UNSUPPORTED"
)

func (s State) Projection() string {
	switch s {
	case SupportedTrue:
		return "TRUE"
	case SupportedFalse:
		return "FALSE"
	default:
		return "UNKNOWN"
	}
}

type Decision struct {
	ClaimID    string   `json:"claim_id"`
	EvidenceID string   `json:"evidence_id"`
	Rule       string   `json:"rule"`
	Value      *bool    `json:"value"`
	Accepted   bool     `json:"accepted"`
	Reasons    []string `json:"reasons"`
}

type Assessment struct {
	SchemaVersion       string     `json:"schema_version"`
	PolicyVersion       string     `json:"policy_version"`
	ClaimID             string     `json:"claim_id"`
	ClaimRevision       int        `json:"claim_revision"`
	EvaluationTime      string     `json:"evaluation_time"`
	InputDigest         Digest     `json:"input_digest"`
	State               State      `json:"state"`
	SupportStatus       string     `json:"support_status"`
	ContradictionStatus string     `json:"contradiction_status"`
	Reasons             []string   `json:"reasons"`
	Decisions           []Decision `json:"decisions"`
	Supersedes          *Digest    `json:"supersedes"`
	Invalidations       []string   `json:"invalidations"`
}

// Record is a frozen local input set. Assessments are separate immutable objects.
// All challenges included in the set are inspected; completeness beyond it is not claimed.
type Record struct {
	SchemaVersion  string           `json:"schema_version"`
	Encoding       string           `json:"encoding"`
	PolicyVersion  string           `json:"policy_version"`
	EvaluationTime string           `json:"evaluation_time"`
	TargetClaim    string           `json:"target_claim"`
	Artifacts      []SourceArtifact `json:"artifacts"`
	Claims         []AtomicClaim    `json:"claims"`
}
