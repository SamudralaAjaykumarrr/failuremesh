package evidence

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"time"
)

func one(v string, allowed ...string) bool { return slices.Contains(allowed, v) }
func present(v ...string) bool {
	for _, s := range v {
		if strings.TrimSpace(s) == "" {
			return false
		}
	}
	return true
}
func stamp(s string) (time.Time, bool) {
	t, e := time.Parse(time.RFC3339Nano, s)
	return t, e == nil && t.UTC().Format(time.RFC3339Nano) == s
}
func validDigest(d Digest) bool {
	b, e := hex.DecodeString(d.Hex)
	return d.Algorithm == "sha256" && e == nil && len(b) == 32 && strings.ToLower(d.Hex) == d.Hex
}
func safeClass(s string) bool { return one(s, "public", "synthetic") }
func classification(s string) bool {
	return one(s, "public", "synthetic", "private", "secret", "unknown")
}

// Classify recognizes explicit fixture/private markers only; it is not a general
// secret detector or a substitute for a source classification policy.
func Classify(b []byte, declared string) string {
	s := strings.ToLower(string(b))
	for _, marker := range []string{"[secret]", "private_key", "api_key=", "password=", "-----begin private key"} {
		if strings.Contains(s, marker) {
			return "secret"
		}
	}
	if strings.Contains(s, "[private]") {
		return "private"
	}
	if !classification(declared) {
		return "unknown"
	}
	return declared
}
func safeURL(s string) bool {
	u, e := url.Parse(s)
	return e == nil && one(u.Scheme, "https", "http") && u.Host != "" && u.User == nil && u.RawQuery == "" && u.Fragment == ""
}
func validRights(r Rights) bool {
	for _, p := range []string{r.Metadata, r.Digest, r.Excerpt, r.WholeBody, r.Export} {
		if !one(p, "allow", "deny", "unknown") {
			return false
		}
	}
	if !present(r.License, r.Basis, r.Review) {
		return false
	}
	if r.RetainUntil != "" {
		if _, ok := stamp(r.RetainUntil); !ok {
			return false
		}
	}
	return r.Excerpt != "allow" || r.ExcerptScope != ""
}
func scopeValid(s Scope) bool {
	return present(s.Build, s.Configuration, s.Environment, s.Run, s.Operation, s.IdentityDomain) && s.Quantifier == "one_bounded_operation"
}
func freshnessValid(f Freshness, s Scope) bool {
	observed, ok := stamp(f.ObservedAt)
	if !ok {
		return false
	}
	from, ok := stamp(f.ValidFrom)
	if !ok || from.Before(observed) || f.PinnedBuild != s.Build {
		return false
	}
	if f.RevalidationRule == "pinned-build-v1" {
		return f.ValidUntil == ""
	}
	until, ok := stamp(f.ValidUntil)
	return f.RevalidationRule == "explicit-interval-v1" && ok && until.After(from)
}
func fresh(f Freshness, at time.Time) bool {
	from, _ := stamp(f.ValidFrom)
	if at.Before(from) {
		return false
	}
	if f.ValidUntil != "" {
		until, _ := stamp(f.ValidUntil)
		return at.Before(until)
	}
	return true
}

// Capture binds exactly supplied bytes without normalization. Retention is opt-in;
// classification is preserved even when the bytes are discarded. No acquisition occurs.
func Capture(a SourceArtifact, raw []byte, retain bool) (SourceArtifact, error) {
	a.ContentDigest = Hash(raw)
	a.ByteLength = int64(len(raw))
	a.Representation = "supplied-exact-bytes"
	a.Sensitivity = Classify(raw, a.Sensitivity)
	metadata, _ := json.Marshal(a)
	a.Sensitivity = Classify(metadata, a.Sensitivity)
	a.Retained = retain
	a.Body = nil
	if retain {
		a.Body = bytes.Clone(raw)
	}
	if err := validateArtifact(a); err != nil {
		return SourceArtifact{}, err
	}
	if err := validateContentCoverage([]SourceArtifact{a}, raw); err != nil {
		return SourceArtifact{}, err
	}
	return a, nil
}

func validateArtifact(a SourceArtifact) error {
	fail := errors.New("invalid source artifact or retention policy")
	metadata, err := json.Marshal(a)
	if err != nil || Classify(metadata, a.Sensitivity) != a.Sensitivity {
		return fail
	}
	if a.SchemaVersion != Schema || !present(a.SourceID, a.ArtifactID, a.Publisher, a.PublisherIdentityBasis, a.PublicationTimestampBasis, a.SourceType, a.QualityClass) || a.Revision < 1 || a.QualityPolicyVersion != Policy || !validDigest(a.ContentDigest) || a.Representation != "supplied-exact-bytes" || a.ByteLength < 0 || a.ByteLength > maxBytes {
		return fail
	}
	if !safeURL(a.CanonicalURL) || !safeURL(a.RetrievedURL) {
		return fail
	}
	retrieved, ok := stamp(a.RetrievedAt)
	if !ok {
		return fail
	}
	checked, ok := stamp(a.CheckedAt)
	if !ok || checked.Before(retrieved) {
		return fail
	}
	if a.PublicationTimestamp != "" {
		if _, ok := stamp(a.PublicationTimestamp); !ok {
			return fail
		}
	}
	if !one(a.Mutability, "mutable", "revision-pinned", "unknown") || !one(a.Availability, "available", "unavailable", "deleted", "retracted", "unknown") || !classification(a.Sensitivity) {
		return fail
	}
	q := a.Quality
	if !present(q.Attribution, q.Directness, q.ScopeFit, q.VersionFit, q.Integrity, q.Completeness, q.Independence) {
		return fail
	}
	p := a.Retrieval
	if p.Method != "manual-local" || !present(p.Collector, p.Version, p.Status, p.ContentType) || !one(p.IdentityClass, "human", "tool", "ai", "unknown") || (p.IdentityClass == "ai" && p.Model == "") || (p.Model != "" && p.IdentityClass != "ai") {
		return fail
	}
	for _, u := range p.Redirects {
		if !safeURL(u) {
			return fail
		}
	}
	if !validRights(a.Rights) || a.Rights.Metadata != "allow" || a.Rights.Digest != "allow" {
		return fail
	}
	if a.Retained {
		if a.Rights.WholeBody != "allow" || int64(len(a.Body)) != a.ByteLength || Hash(a.Body) != a.ContentDigest || Classify(a.Body, a.Sensitivity) != a.Sensitivity {
			return fail
		}
	} else if a.Body != nil {
		return fail
	}
	seen := map[string]bool{}
	for _, l := range a.Locations {
		if l.ID == "" || seen[l.ID] || l.ArtifactID != a.ArtifactID || l.ArtifactDigest != a.ContentDigest || !classification(l.Sensitivity) {
			return fail
		}
		seen[l.ID] = true
		s := l.Selector
		// Only byte selectors resolve locally in v1. Other selectors are citation-only.
		if s.Representation != a.Representation || !one(s.Kind, "byte", "line", "page", "section") {
			return fail
		}
		if s.Kind == "byte" {
			if s.Start < 0 || s.End <= s.Start || s.End > a.ByteLength {
				return fail
			}
		} else if !present(s.Label) || l.Access != "unavailable" {
			return fail
		}
		switch l.Access {
		case "body":
			if !a.Retained || s.Kind != "byte" {
				return fail
			}
		case "excerpt":
			if s.Kind != "byte" || l.Excerpt == nil {
				return fail
			}
		case "unavailable":
			if l.Limitation == "" || l.Excerpt != nil {
				return fail
			}
		default:
			return fail
		}
		if l.Excerpt != nil {
			if l.ExcerptDigest == nil || !validDigest(*l.ExcerptDigest) || Hash(l.Excerpt) != *l.ExcerptDigest || a.Rights.Excerpt != "allow" || int64(len(l.Excerpt)) != s.End-s.Start || Classify(l.Excerpt, l.Sensitivity) != l.Sensitivity {
				return fail
			}
			if a.Retained && !bytes.Equal(a.Body[s.Start:s.End], l.Excerpt) {
				return fail
			}
		} else if l.ExcerptDigest != nil && (!validDigest(*l.ExcerptDigest) || l.Access != "unavailable") {
			return fail
		}
	}
	return nil
}

// All supplied captures of the same exact bytes share one coverage boundary,
// regardless of logical identity or revision. Called only after artifact range
// validation. Metadata-only captures contribute no permission or byte coverage.
func validateContentCoverage(artifacts []SourceArtifact, parents ...[]byte) error {
	// Parent bytes are transient verification inputs, never part of the record.
	// Index by their computed digest, not a caller-supplied identity or trust flag.
	bodies := map[Digest][]byte{}
	for _, body := range parents {
		if len(body) > maxBytes {
			return errors.New("parent body too large")
		}
		bodies[Hash(body)] = body
	}
	for _, a := range artifacts {
		if a.Retained {
			bodies[Hash(a.Body)] = a.Body
		}
		for _, l := range a.Locations {
			// An excerpt that hashes to the whole parent is itself a complete parent.
			// Its coordinates and whole-body rights still must be checked below.
			if l.Excerpt != nil && Hash(l.Excerpt) == a.ContentDigest {
				bodies[a.ContentDigest] = l.Excerpt
			}
		}
	}

	type coverage struct {
		length         int64
		representation string
		ranges         []Selector
		wholeAllowed   bool
	}
	groups := map[Digest]*coverage{}
	var order []Digest
	for _, a := range artifacts {
		parent, verified := bodies[a.ContentDigest]
		// Without a verified parent, this length is metadata only: no retained
		// range may enter coverage below. It cannot establish safe partial coverage.
		length := a.ByteLength
		if verified {
			length = int64(len(parent))
			if a.ByteLength != length {
				return errors.New("parent_length_mismatch")
			}
		}
		g, ok := groups[a.ContentDigest]
		if !ok {
			g = &coverage{length: length, representation: a.Representation, wholeAllowed: true}
			groups[a.ContentDigest] = g
			order = append(order, a.ContentDigest)
		} else if g.length != length || g.representation != a.Representation {
			return errors.New("inconsistent content identity metadata")
		}
		contributes := false
		if a.Retained && length > 0 {
			g.ranges = append(g.ranges, Selector{Start: 0, End: length})
			contributes = true
		}
		for _, l := range a.Locations {
			if l.Excerpt != nil {
				if !verified {
					return errors.New("excerpt_parent_unavailable")
				}
				s := l.Selector
				if s.Kind != "byte" || s.Start < 0 || s.End <= s.Start || s.End > length {
					return errors.New("excerpt_parent_range_invalid")
				}
				if !bytes.Equal(l.Excerpt, parent[s.Start:s.End]) {
					return errors.New("excerpt_parent_slice_mismatch")
				}
				g.ranges = append(g.ranges, s)
				contributes = true
			}
		}
		// No capture's allow can silently override another contributor's deny or
		// unknown policy. Whole-body permission must cover every contributing capture.
		if contributes && a.Rights.WholeBody != "allow" {
			g.wholeAllowed = false
		}
	}
	for _, d := range order {
		g := groups[d]
		slices.SortFunc(g.ranges, func(a, b Selector) int {
			if a.Start < b.Start {
				return -1
			}
			if a.Start > b.Start {
				return 1
			}
			return 0
		})
		end := int64(0)
		for _, s := range g.ranges {
			if s.Start > end {
				break
			}
			if s.End > end {
				end = s.End
			}
		}
		if g.length > 0 && end == g.length && !g.wholeAllowed {
			return errors.New("cumulative_whole_body_permission_required")
		}
	}
	return nil
}

func claimKey(c AtomicClaim) string { return c.ID } // One immutable revision per logical claim per input set.
func validateClaim(c AtomicClaim) error {
	fail := errors.New("invalid claim structure")
	metadata, err := json.Marshal(c)
	if err != nil || Classify(metadata, c.Sensitivity) != c.Sensitivity {
		return fail
	}
	if c.SchemaVersion != Schema || !present(c.ID, c.Predicate, c.Subject) || c.Revision < 1 || !scopeValid(c.Scope) || !freshnessValid(c.Freshness, c.Scope) || !classification(c.Sensitivity) || c.Boundary != "local" || !one(c.EpistemicBasis, "reported", "documented_design", "code_observed", "experiment_observed", "derived") {
		return fail
	}
	v := c.Value
	switch v.Type {
	case "boolean":
		if v.Boolean == nil || v.Text != nil || v.UnknownReason != "" {
			return fail
		}
	case "string":
		if v.Text == nil || v.Boolean != nil || v.UnknownReason != "" {
			return fail
		}
	case "unknown":
		if v.Boolean != nil || v.Text != nil || v.UnknownReason == "" {
			return fail
		}
	default:
		return fail
	}
	e := c.Extraction
	if !present(e.Method, e.Tool, e.Version, e.Transformation) || !one(e.ProposerClass, "human", "tool", "ai", "unknown") || !one(e.Context, "public-only", "synthetic-only", "quarantined") {
		return fail
	}
	if e.ProposerClass == "ai" || e.Method == "ai" || e.Model != "" {
		if e.ProposerClass != "ai" || e.Model == "" || e.PromptDigest == nil || !validDigest(*e.PromptDigest) {
			return fail
		}
	}
	if e.PromptDigest != nil && !validDigest(*e.PromptDigest) {
		return fail
	}
	if c.AdvisoryConfidence != nil {
		v := c.AdvisoryConfidence
		if v.Method == "" || math.IsNaN(v.Score) || math.IsInf(v.Score, 0) || v.Score < 0 || v.Score > 1 {
			return fail
		}
	}
	if !one(c.SupportStatus, "candidate", "absent", "present", "invalid") || !one(c.ContradictionStatus, "none", "open", "resolved") {
		return fail
	}
	return nil
}

// Validate rejects ambiguous identities, invalid/dangling references and cycles.
// Expiry is an assessment/export concern: a frozen historical record can be replayed.
// Optional parents supply exact bytes for excerpt verification only. Callers must
// supply them again on replay when neither the record nor a complete excerpt
// contains the parent. No caller-declared binding flag is trusted.
func Validate(r Record, parents ...[]byte) error {
	// Go replaces invalid UTF-8 strings during marshaling. Reject that mutation,
	// and bound programmatically constructed records as well as parsed inputs.
	raw, err := json.Marshal(r)
	if err != nil || len(raw) > maxBytes {
		return errors.New("record encoding failed or too large")
	}
	var roundTrip Record
	if json.Unmarshal(raw, &roundTrip) != nil || !reflect.DeepEqual(r, roundTrip) {
		return errors.New("non-round-trippable record")
	}

	if r.SchemaVersion != Schema || r.Encoding != Encoding || r.PolicyVersion != Policy {
		return errors.New("unsupported record version")
	}
	if _, ok := stamp(r.EvaluationTime); !ok {
		return errors.New("invalid evaluation time")
	}
	arts := map[string]SourceArtifact{}
	locs := map[string]string{}
	revisions := map[string]map[int]bool{}
	for _, a := range r.Artifacts {
		if err := validateArtifact(a); err != nil {
			return err
		}
		if _, ok := arts[a.ArtifactID]; ok {
			return errors.New("duplicate artifact identity")
		}
		if revisions[a.SourceID] == nil {
			revisions[a.SourceID] = map[int]bool{}
		}
		if revisions[a.SourceID][a.Revision] {
			return errors.New("duplicate source revision")
		}
		revisions[a.SourceID][a.Revision] = true
		arts[a.ArtifactID] = a
		for _, l := range a.Locations {
			if _, ok := locs[l.ID]; ok {
				return errors.New("duplicate location identity")
			}
			locs[l.ID] = a.ArtifactID
		}
	}
	if err := validateContentCoverage(r.Artifacts, parents...); err != nil {
		return err
	}
	for _, a := range r.Artifacts {
		if a.Supersedes != "" {
			p, ok := arts[a.Supersedes]
			if !ok || p.SourceID != a.SourceID || p.Revision >= a.Revision {
				return errors.New("invalid supersession")
			}
		}
		for _, rel := range a.Relations {
			_, ok := arts[rel.ArtifactID]
			if !ok || rel.ArtifactID == a.ArtifactID || !one(rel.Kind, "origin", "mirror", "archive", "revision") || rel.Basis == "" {
				return errors.New("invalid artifact relation")
			}
		}
	}
	claims := map[string]AtomicClaim{}
	for _, c := range r.Claims {
		if err := validateClaim(c); err != nil {
			return err
		}
		if _, ok := claims[claimKey(c)]; ok {
			return errors.New("duplicate claim identity")
		}
		claims[claimKey(c)] = c
		refs := map[string]bool{}
		inputs := map[Digest]bool{}
		for _, ref := range c.SourceRefs {
			a, ok := arts[ref.ArtifactID]
			if !ok || refs[ref.ArtifactID] || ref.SourceID != a.SourceID || ref.Revision != a.Revision || ref.Digest != a.ContentDigest {
				return errors.New("invalid source reference")
			}
			refs[ref.ArtifactID] = true
			inputs[ref.Digest] = true
		}
		seen := map[string]bool{}
		for _, id := range c.EvidenceRefs {
			aid, ok := locs[id]
			if !ok || !refs[aid] || seen[id] {
				return errors.New("invalid evidence reference")
			}
			seen[id] = true
		}
		digests := map[Digest]bool{}
		for _, d := range c.Extraction.InputDigests {
			if !validDigest(d) || !inputs[d] || digests[d] {
				return errors.New("invalid extraction input")
			}
			digests[d] = true
		}
		if len(digests) != len(inputs) {
			return errors.New("missing extraction input")
		}
	}
	if _, ok := claims[r.TargetClaim]; !ok {
		return errors.New("unresolvable target")
	}
	visiting := map[string]bool{}
	done := map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] {
			return errors.New("cyclic derivation")
		}
		if done[id] {
			return nil
		}
		visiting[id] = true
		c := claims[id]
		d := c.Derivation
		if len(d.Parents) != len(d.ParentDigests) || (len(d.Parents) > 0 && !present(d.Rule, d.Version)) {
			return errors.New("invalid derivation")
		}
		seen := map[string]bool{}
		for i, p := range d.Parents {
			parent, ok := claims[p]
			if !ok || seen[p] {
				return errors.New("invalid parent")
			}
			seen[p] = true
			b, _ := json.Marshal(parent)
			if Hash(b) != d.ParentDigests[i] {
				return errors.New("parent digest mismatch")
			}
			if err := visit(p); err != nil {
				return err
			}
		}
		visiting[id] = false
		done[id] = true
		return nil
	}
	for _, c := range r.Claims {
		if err := visit(c.ID); err != nil {
			return err
		}
	}
	return nil
}
