package evidence

import (
	"encoding/json"
	"slices"
	"time"
)

// syntheticObservation is a deliberately narrow typed fixture, not an adapter,
// imported runtime trace, or assertion about the upstream system. Attempts are
// counted only for the exact operation identity. Negative evidence requires a
// complete bounded observer with at least one initial attempt.
type syntheticObservation struct {
	SchemaVersion string    `json:"schema_version"`
	Subject       string    `json:"subject"`
	Scope         Scope     `json:"scope"`
	Freshness     Freshness `json:"freshness"`
	Observer      string    `json:"observer"`
	Completeness  string    `json:"completeness"`
	Attempts      []string  `json:"attempts"`
}

type ruleResult struct {
	value   *bool
	reasons []string
	stale   bool
	unknown bool
}
type predicateRule struct {
	id       string
	eligible func(AtomicClaim) []string
	evaluate func(AtomicClaim, SourceArtifact, Location, []byte, time.Time) ruleResult
}

// Rules are code-owned and pinned to Policy. Input records cannot register code,
// change a rule or nominate an LLM/human/source as the authority.
func lookupRule(predicate string) (predicateRule, bool) {
	switch predicate {
	case "synthetic.same_operation_retry":
		return predicateRule{"synthetic-retry-v1", retryEligibility, retryRule}, true
	default:
		return predicateRule{}, false
	}
}

// Eligibility belongs to the proposition, before any related evidence can
// establish support. The same rule is used for targets and supporting claims.
func retryEligibility(c AtomicClaim) []string {
	var reasons []string
	if c.Value.Type != "boolean" && c.Value.Type != "unknown" {
		reasons = append(reasons, "predicate_requires_boolean_value")
	}
	if c.Sensitivity != "synthetic" || c.Scope.Environment != "synthetic-reference" || c.EpistemicBasis != "experiment_observed" {
		reasons = append(reasons, "rule_requires_synthetic_experiment_scope")
	}
	if c.EpistemicBasis == "derived" || len(c.Derivation.Parents) > 0 || c.Derivation.Rule != "" {
		reasons = append(reasons, "derivation_rule_not_implemented")
	}
	return reasons
}
func retryRule(c AtomicClaim, a SourceArtifact, l Location, b []byte, at time.Time) ruleResult {
	out := ruleResult{reasons: retryEligibility(c)}
	if len(out.reasons) > 0 {
		return out
	}
	reject := func(s string) { out.reasons = append(out.reasons, s) }
	if a.SourceType != "synthetic_observation" || a.Sensitivity != "synthetic" || l.Sensitivity != "synthetic" {
		reject("rule_requires_synthetic_experiment_scope")
	}
	if a.Retrieval.IdentityClass != "tool" || a.Retrieval.Collector != "fixture-author" || a.Retrieval.Version != "1" || a.Retrieval.Model != "" {
		reject("not_fixture_observer_provenance")
	}
	if a.Retrieval.Status != "supplied" || len(a.Retrieval.Errors) > 0 {
		reject("retrieval_error")
	}
	if a.Availability == "retracted" || a.Availability == "unknown" {
		reject("source_unusable")
	}
	q := a.Quality
	if q.Attribution != "fixture-author" || q.Directness != "structured_fixture" || q.ScopeFit != "exact" || q.VersionFit != "exact" || q.Integrity != "exact_bytes" || q.Completeness != "bounded_observer" || q.Independence != "synthetic_not_independent" {
		reject("quality_dimensions_not_admissible_for_rule")
	}
	var obs syntheticObservation
	if strictDecode(b, &obs) != nil {
		reject("unsupported_semantic_mapping")
		out.unknown = true
		return out
	}
	if obs.SchemaVersion != "synthetic-retry-v1" || !scopeValid(obs.Scope) || !freshnessValid(obs.Freshness, obs.Scope) || obs.Observer != "synthetic-complete-attempts-v1" {
		reject("invalid_structured_observation")
		return out
	}
	if obs.Scope != c.Scope || obs.Subject != c.Subject {
		reject("observation_scope_mismatch")
	}
	if obs.Freshness != c.Freshness {
		reject("observation_freshness_mismatch")
	}
	if !fresh(obs.Freshness, at) {
		reject("outside_validity_interval")
		out.stale = true
	}
	if obs.Completeness != "complete_bounded_operation" {
		reject("observer_completeness_missing")
		out.unknown = true
	}
	n := 0
	for _, operation := range obs.Attempts {
		if operation == "" {
			reject("missing_operation_identity")
			out.unknown = true
		}
		if operation == obs.Scope.Operation {
			n++
		}
	}
	if n == 0 {
		reject("initial_attempt_missing")
		out.unknown = true
	}
	if len(out.reasons) == 0 || (len(out.reasons) == 1 && out.stale) {
		v := n >= 2
		out.value = &v
	}
	return out
}

// Assess derives authority from the complete supplied input set, not candidate
// status/confidence/value. Invalid structure wins; then current material conflict,
// otherwise sufficient stale support, then admissibility/unknown/unsupported.
func Assess(r Record, parents ...[]byte) Assessment {
	b, marshalErr := json.Marshal(r)
	out := Assessment{SchemaVersion: Schema, PolicyVersion: Policy, ClaimID: r.TargetClaim, EvaluationTime: r.EvaluationTime, InputDigest: Hash(b), State: Unsupported, SupportStatus: "invalid", ContradictionStatus: "none"}
	if marshalErr != nil {
		out.InputDigest = Digest{}
		out.Reasons = []string{"input_not_encodable"}
		return out
	}
	if err := Validate(r, parents...); err != nil {
		out.Reasons = []string{err.Error()}
		return out
	}
	var target AtomicClaim
	for _, c := range r.Claims {
		if c.ID == r.TargetClaim {
			target = c
		}
	}
	out.ClaimRevision = target.Revision
	targetRule, supported := lookupRule(target.Predicate)
	if !supported {
		out.Reasons = []string{"unsupported_predicate"}
	} else {
		out.Reasons = targetRule.eligible(target)
	}
	if len(out.Reasons) > 0 {
		out.Decisions = []Decision{{ClaimID: target.ID, Rule: targetRule.id, Reasons: slices.Clone(out.Reasons)}}
		return out
	}
	out.SupportStatus = "absent"
	at, _ := stamp(r.EvaluationTime)
	arts := map[string]SourceArtifact{}
	locations := map[string]Location{}
	for _, a := range r.Artifacts {
		arts[a.ArtifactID] = a
		for _, l := range a.Locations {
			locations[l.ID] = l
		}
	}
	positive, negative, stale, unresolved := false, false, false, false
	unresolvedChallenge := false
	// Sort a copy; input order and caller memory remain intact. Trace ordering is stable.
	claims := slices.Clone(r.Claims)
	slices.SortFunc(claims, func(a, b AtomicClaim) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	for _, c := range claims {
		ids := slices.Clone(c.EvidenceRefs)
		slices.Sort(ids)
		if len(ids) == 0 {
			ids = []string{""}
		}
		for _, id := range ids {
			d := Decision{ClaimID: c.ID, EvidenceID: id}
			same := c.Predicate == target.Predicate && c.Subject == target.Subject && c.Scope == target.Scope && c.EpistemicBasis == target.EpistemicBasis
			if !same {
				d.Reasons = append(d.Reasons, "outside_target_scope_or_basis")
			}
			rule, ok := lookupRule(c.Predicate)
			d.Rule = rule.id
			if !ok {
				d.Reasons = append(d.Reasons, "unsupported_predicate")
			}
			if id == "" {
				d.Reasons = append(d.Reasons, "missing_evidence")
				if same && c.Value.Type == "unknown" {
					unresolved = true
				}
			}
			if !safeClass(c.Sensitivity) || c.Extraction.Context == "quarantined" {
				d.Reasons = append(d.Reasons, "claim_quarantined")
			}
			if id != "" {
				l := locations[id]
				a := arts[l.ArtifactID]
				if !safeClass(a.Sensitivity) || !safeClass(l.Sensitivity) {
					d.Reasons = append(d.Reasons, "source_or_location_quarantined")
				}
				retrieved, _ := stamp(a.RetrievedAt)
				checked, _ := stamp(a.CheckedAt)
				if at.Before(retrieved) || at.Before(checked) {
					d.Reasons = append(d.Reasons, "source_observation_after_evaluation")
				}
				if a.Rights.RetainUntil != "" {
					end, _ := stamp(a.Rights.RetainUntil)
					if !at.Before(end) {
						d.Reasons = append(d.Reasons, "retention_expired")
					}
				}
				var raw []byte
				switch l.Access {
				case "body":
					raw = a.Body[l.Selector.Start:l.Selector.End]
				case "excerpt":
					raw = l.Excerpt
					if !a.Retained && (l.Selector.Start != 0 || l.Selector.End != a.ByteLength || Hash(raw) != a.ContentDigest) {
						d.Reasons = append(d.Reasons, "excerpt_binding_not_replayable")
					}
				default:
					d.Reasons = append(d.Reasons, "bytes_unavailable")
				}
				if ok && raw != nil {
					rr := rule.evaluate(c, a, l, raw, at)
					d.Reasons = append(d.Reasons, rr.reasons...)
					d.Value = rr.value
					// A stale result is sufficient only if every other structural/admissibility check passed.
					if rr.value != nil && len(d.Reasons) == 1 && rr.stale {
						stale = true
					}
					if same && rr.unknown && len(d.Reasons) == len(rr.reasons) {
						unresolved = true
						if a.Retrieval.IdentityClass == "tool" && a.SourceType == "synthetic_observation" && fresh(c.Freshness, at) {
							unresolvedChallenge = true
						}
					}
					if rr.value != nil && len(d.Reasons) == 0 {
						d.Accepted = true
						if *rr.value {
							positive = true
						} else {
							negative = true
						}
						if c.Value.Type == "boolean" && *c.Value.Boolean != *rr.value {
							d.Reasons = append(d.Reasons, "candidate_value_disagrees_with_observation")
						}
						d.Reasons = append(d.Reasons, "structured_fixture_support")
					}
				}
			}
			out.Decisions = append(out.Decisions, d)
		}
	}
	switch {
	case !safeClass(target.Sensitivity) || target.Extraction.Context == "quarantined":
		out.State = Unsupported
		out.Reasons = []string{"target_quarantined"}
	case positive && negative:
		out.State = Conflicting
		out.SupportStatus = "present"
		out.ContradictionStatus = "open"
		out.Reasons = []string{"material_current_same_scope_conflict"}
	case unresolvedChallenge:
		out.State = Unknown
		out.Reasons = []string{"unresolved_evidence_interpretation"}
	case positive:
		out.State = SupportedTrue
		out.SupportStatus = "present"
		out.Reasons = []string{"current_scoped_positive_evidence"}
	case negative:
		out.State = SupportedFalse
		out.SupportStatus = "present"
		out.Reasons = []string{"current_scoped_negative_evidence_with_completeness"}
	case stale:
		out.State = Stale
		out.Reasons = []string{"otherwise_sufficient_support_outside_validity"}
	case unresolved:
		out.State = Unknown
		out.Reasons = []string{"missing_value_or_unresolved_interpretation"}
	default:
		out.State = Unsupported
		out.Reasons = []string{"no_admissible_support"}
	}
	return out
}
