package corpus

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/evidence"
)

type Error struct {
	Code string
	Exit int
}

func (e *Error) Error() string         { return e.Code }
func fail(code string, exit int) error { return &Error{code, exit} }
func ExitCode(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Exit
	}
	return 4
}
func Hash(b []byte) Digest     { return evidence.Hash(b) }
func VocabularyDigest() Digest { b, _ := json.Marshal(families); return Hash(b) }
func validDigest(d Digest) bool {
	b, e := hex.DecodeString(d.Hex)
	return e == nil && len(b) == 32 && d.Algorithm == "sha256" && strings.ToLower(d.Hex) == d.Hex
}
func same(a, b *Digest) bool { return reflect.DeepEqual(a, b) }
func stamp(s string) bool {
	t, e := time.Parse(time.RFC3339Nano, s)
	return e == nil && t.UTC().Format(time.RFC3339Nano) == s
}
func before(a, b string) bool {
	x, _ := time.Parse(time.RFC3339Nano, a)
	y, _ := time.Parse(time.RFC3339Nano, b)
	return x.Before(y)
}
func one(s string, v ...string) bool { return slices.Contains(v, s) }
func present(v ...string) bool {
	for _, s := range v {
		if strings.TrimSpace(s) == "" || len(s) > 4096 {
			return false
		}
	}
	return true
}
func id(s string) bool {
	return utf8.ValidString(s) && present(s) && len(s) <= 128 && !strings.ContainsRune(s, 0)
}
func fact(f Fact) bool {
	return present(f.Reason) && ((f.State == "UNKNOWN" && f.Value == nil) || (f.State == "KNOWN" && f.Value != nil && present(*f.Value)))
}
func value(f Fact) string {
	if f.State == "KNOWN" && f.Value != nil {
		return *f.Value
	}
	return "UNKNOWN"
}
func identity(i Identity) bool {
	return !strings.ContainsRune(i.Context+i.Model, 0) && id(i.Actor) && present(i.Role, i.Tool, i.Version, i.Context) && one(i.Class, "human", "tool", "ai", "unknown") && (i.Class != "ai" || present(i.Model))
}
func ref(r Ref) bool {
	return id(r.ID) && r.Revision > 0 && r.Revision <= MaxOrdinal && validDigest(r.Digest)
}
func sorted[T any](v []T, key func(T) string) error {
	slices.SortFunc(v, func(a, b T) int { return strings.Compare(key(a), key(b)) })
	for i := 1; i < len(v); i++ {
		if key(v[i-1]) == key(v[i]) {
			return fail("DUPLICATE_SET_MEMBER", 2)
		}
	}
	return nil
}
func stringsSet(v []string) error { return sorted(v, func(s string) string { return s }) }
func digestsSet(v []Digest) error {
	for _, d := range v {
		if !validDigest(d) {
			return fail("INVALID_DIGEST", 2)
		}
	}
	return sorted(v, func(d Digest) string { return d.Hex })
}
func normalize(o *Object) error {
	var errs []error
	errs = append(errs, stringsSet(o.Reasons))
	if p := o.Policy; p != nil {
		errs = append(errs, stringsSet(p.Channels), stringsSet(p.Eligibility), stringsSet(p.Challenges), stringsSet(p.RightsRequirements), stringsSet(p.ExposureExclusions), stringsSet(p.StopRules), sorted(p.Budgets, func(b Budget) string { return b.Family }), sorted(p.Targets, func(t Target) string { return t.Name }))
	}
	if c := o.Candidate; c != nil {
		errs = append(errs, sorted(c.Sources, func(v Citation) string { return v.ID }), sorted(c.Captured, func(v Captured) string { return v.CitationID + "\x00" + v.Record.Hex }), stringsSet(c.Publisher.Names), stringsSet(c.ArchitectureRefs), stringsSet(c.ChallengeCategories), stringsSet(c.Family.Alternatives), stringsSet(c.Decision.Reasons), digestsSet(c.ExposureRefs), sorted(c.Parents, func(v Ref) string { return v.ID }), sorted(c.Assessments, func(v Assessment) string { return v.Dimension }))
		for i := range c.Captured {
			errs = append(errs, stringsSet(c.Captured[i].Locations))
		}
		for i := range c.Assessments {
			errs = append(errs, stringsSet(c.Assessments[i].References))
		}
	}
	if o.Relations != nil {
		errs = append(errs, sorted(*o.Relations, func(v Relation) string { return v.ID }))
	}
	if o.Assignments != nil {
		errs = append(errs, sorted(*o.Assignments, func(v Assignment) string { return v.Candidate }))
	}
	if e := o.Exposure; e != nil {
		errs = append(errs, stringsSet(e.Candidates), digestsSet(e.Artifacts))
	}
	if o.Notice != nil {
		errs = append(errs, stringsSet(o.Notice.Candidates))
	}
	// Sort every Fact's basis, including facts nested in dates and hypotheses.
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		if v.Kind() == reflect.Pointer {
			if !v.IsNil() {
				walk(v.Elem())
			}
			return
		}
		if v.Type() == reflect.TypeOf(Fact{}) {
			f := v.Addr().Interface().(*Fact)
			if f.Basis == nil {
				f.Basis = []string{}
			}
			errs = append(errs, stringsSet(f.Basis))
			return
		}
		switch v.Kind() {
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				walk(v.Field(i))
			}
		case reflect.Slice:
			if v.IsNil() {
				v.Set(reflect.MakeSlice(v.Type(), 0, 0))
			}
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i))
			}
		}
	}
	walk(reflect.ValueOf(o).Elem())
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
func strict(b []byte, v any) error {
	if len(b) > MaxObjectBytes {
		return fail("OBJECT_LIMIT", 2)
	}
	if !utf8.Valid(b) {
		return fail("INVALID_UTF8", 2)
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return fail("INVALID_JSON", 2)
	}
	raw, e := json.Marshal(v)
	if e != nil || !bytes.Equal(raw, b) {
		return fail("NONCANONICAL_JSON", 2)
	}
	return nil
}
func Encode(o Object) ([]byte, Digest, error) {
	// Copy before sorting so callers cannot mutate store state through slice aliases.
	raw, e := json.Marshal(o)
	if e != nil || len(raw) > MaxObjectBytes {
		return nil, Digest{}, fail("OBJECT_LIMIT", 2)
	}
	var c Object
	if json.Unmarshal(raw, &c) != nil || !reflect.DeepEqual(c, o) {
		return nil, Digest{}, fail("NON_ROUND_TRIPPABLE", 2)
	}
	if e = normalize(&c); e != nil {
		return nil, Digest{}, e
	}
	if e = validate(c); e != nil {
		return nil, Digest{}, e
	}
	b, e := json.Marshal(c)
	if e != nil || len(b) > MaxObjectBytes {
		return nil, Digest{}, fail("OBJECT_LIMIT", 2)
	}
	return b, Hash(b), nil
}
func Decode(b []byte, d Digest) (Object, error) {
	var o Object
	if !validDigest(d) || Hash(b) != d {
		return o, fail("DIGEST_MISMATCH", 2)
	}
	if e := strict(b, &o); e != nil {
		return Object{}, e
	}
	raw, _, e := Encode(o)
	if e != nil {
		return Object{}, e
	}
	if !bytes.Equal(b, raw) {
		return Object{}, fail("UNSORTED_SET", 2)
	}
	return o, nil
}
func DecodeRequest(b []byte) (Request, error) {
	var q Request
	if e := strict(b, &q); e != nil {
		return q, e
	}
	return q, nil
}
func validRights(r evidence.Rights) bool {
	if !present(r.License, r.Basis, r.Review) {
		return false
	}
	for _, s := range []string{r.Metadata, r.Digest, r.Excerpt, r.WholeBody, r.Export} {
		if !one(s, "allow", "deny", "unknown") {
			return false
		}
	}
	return (r.RetainUntil == "" || stamp(r.RetainUntil)) && (r.Excerpt != "allow" || present(r.ExcerptScope))
}
func validate(o Object) error {
	if o.Encoding != Encoding || (o.Kind == "manifest" && o.Schema != ManifestSchema) || (o.Kind != "manifest" && o.Schema != Schema) {
		return fail("UNSUPPORTED_VERSION", 2)
	}
	if !id(o.ID) || o.Revision < 1 || o.Revision > MaxOrdinal || !identity(o.Producer) || !validDigest(o.PolicyDigest) || !stamp(o.ReportedAt) || !present(o.Rationale) || len(o.Reasons) == 0 {
		return fail("INVALID_ENVELOPE", 2)
	}
	if (o.Revision == 1) != (o.Parent == nil) || (o.Parent != nil && !validDigest(*o.Parent)) {
		return fail("INVALID_PARENT", 2)
	}
	if !one(o.Sensitivity, "public", "synthetic") {
		return fail("PRIVACY_QUARANTINE", 3)
	}
	var bounded func(reflect.Value) bool
	bounded = func(v reflect.Value) bool {
		switch v.Kind() {
		case reflect.Pointer:
			return v.IsNil() || bounded(v.Elem())
		case reflect.String:
			return len(v.String()) <= 4096
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if !bounded(v.Field(i)) {
					return false
				}
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				if !bounded(v.Index(i)) {
					return false
				}
			}
		}
		return true
	}
	if !bounded(reflect.ValueOf(o)) {
		return fail("METADATA_FIELD_LIMIT", 2)
	}
	raw, _ := json.Marshal(o)
	if evidence.Classify(raw, o.Sensitivity) != o.Sensitivity {
		return fail("PRIVACY_QUARANTINE", 3)
	}
	n := 0
	for _, yes := range []bool{o.Policy != nil, o.Candidate != nil, o.Relations != nil, o.Assignments != nil, o.Exposure != nil, o.AttemptReason != nil, o.Manifest != nil, o.Notice != nil} {
		if yes {
			n++
		}
	}
	if n != 1 {
		return fail("INVALID_PAYLOAD", 2)
	}
	switch o.Kind {
	case "policy":
		p := o.Policy
		if p == nil {
			return fail("INVALID_PAYLOAD", 2)
		}
		if p.Vocabulary != VocabularyDigest() || o.PolicyDigest != VocabularyDigest() || !id(p.Round) || !present(p.Scope) || !stamp(p.Start) || !stamp(p.Cutoff) || !before(p.Start, p.Cutoff) || len(p.Budgets) != 5 || p.PublisherMultiplier != 3 || p.SplitMethod != "explicit-table-v1" || p.Seed != "NOT_USED" {
			return fail("INVALID_POLICY", 3)
		}
		for _, b := range p.Budgets {
			if !slices.Contains(families, b.Family) || b.Candidates == 0 || b.Candidates > MaxOrdinal || b.CuratorMinutes == 0 || b.CuratorMinutes > MaxOrdinal {
				return fail("INVALID_BUDGET", 3)
			}
		}
		for _, v := range [][]string{p.Channels, p.Eligibility, p.Challenges, p.RightsRequirements, p.ExposureExclusions, p.StopRules} {
			if len(v) == 0 {
				return fail("INCOMPLETE_POLICY", 3)
			}
		}
		if len(p.Targets) == 0 {
			return fail("INCOMPLETE_TARGETS", 3)
		}
		for _, t := range p.Targets {
			if !present(t.Name) || t.Count > MaxOrdinal {
				return fail("INVALID_TARGET", 2)
			}
		}
	case "candidate":
		c := o.Candidate
		if c == nil {
			return fail("INVALID_PAYLOAD", 2)
		}
		if len(c.Sources) > MaxSources || len(c.Captured) > MaxSources {
			return fail("SOURCE_LIMIT", 2)
		}
		if c.Discovery.Method != "manual-local" || !identity(c.Discovery.Discoverer) || !stamp(c.Discovery.ReportedTime) || !fact(c.Discovery.Lead) || !present(c.Discovery.Scope, c.Discovery.Round, c.Discovery.RetrievalBasis) || !one(c.Discovery.RetrievalPermission, "allow", "deny", "unknown") {
			return fail("INVALID_DISCOVERY", 2)
		}
		for _, f := range []Fact{c.Publisher.Group, c.Project, c.ProjectLineage, c.IncidentTime.Value, c.PublicationTime.Value, c.System, c.Stack, c.Family.Primary, c.Origin, c.Episode, c.Independence} {
			if !fact(f) {
				return fail("UNKNOWN_REQUIRES_REASON", 2)
			}
		}
		for _, d := range []Date{c.IncidentTime, c.PublicationTime} {
			if !validDate(d) {
				return fail("INVALID_DATE", 2)
			}
		}
		if !present(c.Publisher.OwnershipBasis) || len(c.Publisher.Names) == 0 || c.Pretraining != "UNKNOWN" || !one(c.OriginClass, "natural_incident", "synthetic_descendant", "synthetic_original", "unknown") {
			return fail("INVALID_CLASSIFICATION", 2)
		}
		if c.OriginClass == "natural_incident" && o.Sensitivity != "public" {
			return fail("SYNTHETIC_NATURAL_REFUSED", 3)
		}
		if c.OriginClass == "synthetic_descendant" && (len(c.Parents) == 0 || !present(c.Transformation)) {
			return fail("SYNTHETIC_PARENT_REQUIRED", 3)
		}
		for _, p := range c.Parents {
			if !ref(p) {
				return fail("INVALID_REFERENCE", 2)
			}
		}
		f := c.Family
		if f.State == "EVIDENCE_SUPPORTED" {
			return fail("FAMILY_AUTHORITY_REFUSED", 3)
		}
		if f.Vocabulary != VocabularyDigest() || !one(f.State, "UNASSIGNED", "CANDIDATE", "DISPUTED") || !present(f.Rationale, f.Scope, f.Mechanism) {
			return fail("INVALID_FAMILY", 2)
		}
		if f.State == "UNASSIGNED" && f.Primary.State != "UNKNOWN" {
			return fail("INVALID_FAMILY", 2)
		}
		if f.State != "UNASSIGNED" && !slices.Contains(families, value(f.Primary)) {
			return fail("INVALID_FAMILY", 2)
		}
		for _, s := range f.Alternatives {
			if !slices.Contains(families, s) {
				return fail("INVALID_FAMILY", 2)
			}
		}
		if len(c.Assessments) != 4 {
			return fail("INCOMPLETE_ASSESSMENTS", 2)
		}
		for _, a := range c.Assessments {
			if !one(a.Dimension, "causal_clarity", "architecture_context", "reproduction", "directness") || !one(a.State, "adequate", "limited", "unavailable", "unknown") || !present(a.Rationale) || !id(a.Assessor) {
				return fail("INVALID_ASSESSMENT", 2)
			}
		}
		d := c.Decision
		if !one(d.State, "PENDING", "INCLUDED", "DEFERRED", "EXCLUDED") || !ref(d.Policy) || !id(d.Actor) || !present(d.Round, d.Rationale, d.Alternatives) || len(d.Reasons) == 0 {
			return fail("INVALID_DECISION", 3)
		}
		if !one(c.Split, "UNASSIGNED", "DEVELOPMENT", "VALIDATION", "HOLDOUT", "NOT_SELECTED") || !one(c.FreezeState, "DRAFT", "SELECTED") {
			return fail("INVALID_LIFECYCLE", 3)
		}
		for _, s := range c.Sources {
			if !citationURL(s.Locator) || !id(s.ID) || !present(s.Locator, s.PublisherBasis, s.SourceType, s.Limitation) || !one(s.Role, "origin", "report", "secondary", "architecture") || s.Status != "citation-only" || !validRights(s.Rights) || !stamp(s.CheckedAt) || !one(s.Availability, "available", "unavailable", "deleted", "retracted", "unknown") {
				return fail("INVALID_CITATION", 2)
			}
			if s.Rights.Metadata != "allow" || s.Rights.RetainUntil != "" && !before(o.ReportedAt, s.Rights.RetainUntil) {
				return fail("RIGHTS_BLOCKED", 3)
			}
		}
		for _, s := range c.Captured {
			if !id(s.StoreID) || !validDigest(s.Record) || !validDigest(s.Source.Digest) || s.Source.Revision < 1 || s.Source.Revision > MaxOrdinal || !present(s.Source.SourceID, s.Source.ArtifactID, s.Representation, s.ParentDependency) {
				return fail("INVALID_CAPTURE_REFERENCE", 2)
			}
			found := false
			for _, r := range c.Sources {
				if r.ID == s.CitationID {
					found = true
					if r.Rights.Digest != "allow" {
						return fail("RIGHTS_BLOCKED", 3)
					}
				}
			}
			if !found {
				return fail("DANGLING_CITATION", 2)
			}
		}
	case "cluster":
		if o.Relations == nil {
			return fail("INVALID_PAYLOAD", 2)
		}
		if len(*o.Relations) > MaxRelations {
			return fail("RELATION_LIMIT", 2)
		}
		for _, r := range *o.Relations {
			if !id(r.ID) || !ref(r.From) || !ref(r.To) || r.From.ID == r.To.ID || !one(r.Type, "SAME_INCIDENT", "MIRROR_OF", "COPIED_SUMMARY_OF", "FOLLOWUP_OF", "ISSUE_REPORT_OF", "QUOTES_ORIGIN", "REVISION_OF", "SYNTHETIC_DESCENDANT_OF", "SAME_CAUSAL_EPISODE", "SHARED_PROJECT_LINEAGE") || !one(r.Status, "ESTABLISHED", "POSSIBLE", "REJECTED") || !present(r.Basis, r.Rationale) {
				return fail("INVALID_RELATION", 2)
			}
		}
	case "assignment":
		if o.Assignments == nil {
			return fail("INVALID_PAYLOAD", 2)
		}
		for _, a := range *o.Assignments {
			if !id(a.Candidate) || !one(a.Split, "DEVELOPMENT", "VALIDATION", "HOLDOUT", "NOT_SELECTED") || !present(a.Rationale) {
				return fail("INVALID_ASSIGNMENT", 3)
			}
		}
	case "exposure":
		e := o.Exposure
		if e == nil {
			return fail("INVALID_PAYLOAD", 2)
		}
		if len(e.Candidates) == 0 || !id(e.Round) || !slices.Contains(categories, e.Category) || !one(e.State, "YES", "NO_WITH_BASIS", "UNKNOWN") || !one(e.AvailableAccess, "YES", "NO_WITH_BASIS", "UNKNOWN") || !present(e.Basis, e.Restrictions) || !fact(e.OccurredAt) || !stamp(e.RecordedAt) {
			return fail("INVALID_EXPOSURE", 3)
		}
	case "attempt":
		if o.AttemptReason == nil || !safeReason(*o.AttemptReason) {
			return fail("INVALID_ATTEMPT", 2)
		}
	case "manifest":
		if o.Manifest == nil {
			return fail("INVALID_PAYLOAD", 2)
		}
		m := o.Manifest
		if !stamp(m.CheckTime) || !ref(m.Policy) || !ref(m.Assignment) || !validDigest(m.Build) || !present(m.SourceCommit) || m.Method != "explicit-table-v1" || m.Seed != "NOT_USED" || m.Chronology != "SELF_REPORTED" || m.Pretraining != "UNKNOWN" {
			return fail("INVALID_MANIFEST", 2)
		}
	case "retirement", "withdrawal":
		if o.Notice == nil || !validDigest(o.Notice.Manifest) || !present(o.Notice.Reason) {
			return fail("INVALID_NOTICE", 2)
		}
	default:
		return fail("UNSUPPORTED_KIND", 2)
	}
	return nil
}

func citationURL(s string) bool {
	u, e := url.Parse(s)
	return e == nil && one(u.Scheme, "https", "http") && u.Hostname() != "" && u.User == nil
}

func validDate(d Date) bool {
	if !fact(d.Value) || !present(d.Timezone) {
		return false
	}
	if d.Value.State == "UNKNOWN" {
		return d.Precision == "unknown"
	}
	parse := func(v string) (time.Time, error) {
		if len(v) == 10 {
			return time.Parse("2006-01-02", v)
		}
		return time.Parse(time.RFC3339Nano, v)
	}
	switch d.Precision {
	case "date":
		_, e := time.Parse("2006-01-02", value(d.Value))
		return e == nil
	case "timestamp":
		_, e := time.Parse(time.RFC3339Nano, value(d.Value))
		return e == nil
	case "interval":
		v := strings.Split(value(d.Value), "/")
		if len(v) != 2 {
			return false
		}
		a, e := parse(v[0])
		b, f := parse(v[1])
		return e == nil && f == nil && !b.Before(a)
	}
	return false
}

func safeReason(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, r := range s {
		if r != '_' && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func evidenceUnsafeOperation(s string) bool {
	return strings.ContainsAny(s, "/\\ \n\r\t") || strings.Contains(s, "=") || strings.Contains(s, "[") || evidence.Classify([]byte(s), "synthetic") != "synthetic"
}
