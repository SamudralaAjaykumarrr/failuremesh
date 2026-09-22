package corpus

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
)

func (s Store) transition(st *state, o Object, action string) error {
	if o.Kind != "policy" && !(o.Kind == "attempt" && o.PolicyDigest == VocabularyDigest()) {
		p, ok := st.objects[o.PolicyDigest]
		if !ok || p.Policy == nil {
			return fail("POLICY_NOT_REGISTERED", 3)
		}
	}
	if action == "candidate exclude" && o.Candidate.Decision.State != "EXCLUDED" {
		return fail("EXCLUSION_REQUIRED", 3)
	}
	switch o.Kind {
	case "policy":
		if o.Parent != nil {
			old := st.objects[*o.Parent]
			if old.Policy.Round != o.Policy.Round {
				return fail("ROUND_ID_IMMUTABLE", 3)
			}
		}
		for _, d := range st.order {
			old := st.objects[d]
			if old.Exposure != nil && (old.Exposure.State == "YES" || old.Exposure.AvailableAccess == "YES") && one(old.Exposure.Category, "expected_result", "prior_case_outcome", "previous_evaluation_result") && o.Policy.DiversityException != "" {
				return fail("POST_RESULT_WAIVER", 3)
			}
		}
	case "candidate":
		c := o.Candidate
		p := st.objects[o.PolicyDigest]
		if c.Discovery.Round != p.Policy.Round || c.Decision.Round != p.Policy.Round || c.Decision.Policy.Digest != o.PolicyDigest {
			return fail("POLICY_BINDING_MISMATCH", 2)
		}
		if _, e := st.resolve(c.Decision.Policy); e != nil {
			return e
		}
		if before(o.ReportedAt, p.Policy.Start) || (o.Revision == 1 && before(p.Policy.Cutoff, o.ReportedAt)) {
			return fail("OUTSIDE_POLICY_WINDOW", 3)
		}
		if c.Split != "UNASSIGNED" {
			if c.Assignment == nil {
				return fail("ASSIGNMENT_SNAPSHOT_MISSING", 2)
			}
			a, ok := st.objects[*c.Assignment]
			if !ok || a.Assignments == nil {
				return fail("DANGLING_ASSIGNMENT", 2)
			}
			matched := false
			for _, row := range *a.Assignments {
				if row.Candidate == o.ID && row.Split == c.Split {
					matched = true
				}
			}
			if !matched {
				return fail("ASSIGNMENT_SNAPSHOT_MISMATCH", 2)
			}
		}
		for _, d := range c.ExposureRefs {
			v, ok := st.objects[d]
			if !ok || v.Exposure == nil {
				return fail("DANGLING_EXPOSURE", 2)
			}
		}
		for _, r := range c.Parents {
			v, e := st.resolve(r)
			if e != nil || v.Candidate == nil {
				return fail("DANGLING_SYNTHETIC_PARENT", 2)
			}
			if r.ID == o.ID {
				return fail("ANCESTRY_CYCLE", 2)
			}
		}
		if o.Parent != nil {
			old := st.objects[*o.Parent].Candidate
			if old == nil {
				return fail("INVALID_PARENT", 2)
			}
			if old.Discovery.Round != c.Discovery.Round {
				return fail("ROUND_ID_IMMUTABLE", 3)
			}
			if old.OriginClass != c.OriginClass {
				return fail("ORIGIN_CLASS_IMMUTABLE", 3)
			}
			for _, r := range old.Parents {
				if !slices.Contains(c.Parents, r) {
					return fail("ANCESTRY_REMOVAL_REFUSED", 3)
				}
			}
		}
		if _, e := graphs(st.with(o)); e != nil {
			return e
		}
		if _, e := s.audit(o, o.ReportedAt, true); e != nil {
			return e
		}
		total := uint32(0)
		for _, r := range st.heads {
			v := st.objects[r.Digest]
			if v.Candidate != nil && v.Candidate.Discovery.Round == p.Policy.Round {
				total++
			}
		}
		var budget uint32
		for _, b := range p.Policy.Budgets {
			budget += b.Candidates
		}
		if o.Revision == 1 && total >= budget {
			return fail("CANDIDATE_BUDGET_EXHAUSTED", 3)
		}
	case "cluster":
		count := len(*o.Relations)
		for _, d := range st.order {
			v := st.objects[d]
			if v.Relations != nil {
				count += len(*v.Relations)
			}
		}
		if count > MaxRelations {
			return fail("RELATION_LIMIT", 2)
		}
		for _, r := range *o.Relations {
			for _, x := range []Ref{r.From, r.To} {
				v, e := st.resolve(x)
				if e != nil || v.Candidate == nil {
					return fail("DANGLING_RELATION", 2)
				}
			}
		}
		tmp := st.with(o)
		if _, e := graphs(tmp); e != nil {
			return e
		}
	case "assignment":
		expected := 0
		for _, r := range st.heads {
			v := st.objects[r.Digest]
			if v.Candidate != nil && v.PolicyDigest == o.PolicyDigest {
				expected++
			}
		}
		for _, row := range *o.Assignments {
			r, ok := st.heads[row.Candidate]
			candidate := st.objects[r.Digest]
			if !ok || candidate.Candidate == nil {
				return fail("DANGLING_ASSIGNMENT", 2)
			}
			if candidate.PolicyDigest != o.PolicyDigest {
				return fail("ASSIGNMENT_POLICY_MISMATCH", 3)
			}
		}
		if len(*o.Assignments) != expected {
			return fail("ASSIGNMENT_INCOMPLETE", 3)
		}
		tmp := st.with(o)
		_, e := s.summarize(tmp, o.PolicyDigest, o.ReportedAt, false)
		return e
	case "exposure":
		n := 0
		for _, v := range st.objects {
			if v.Exposure != nil {
				n++
			}
		}
		if n >= MaxExposures {
			return fail("EXPOSURE_LIMIT", 2)
		}
		for _, id := range o.Exposure.Candidates {
			r, ok := st.heads[id]
			if !ok || st.objects[r.Digest].Candidate == nil {
				return fail("DANGLING_EXPOSURE_TARGET", 2)
			}
		}
	case "manifest":
		return s.checkManifest(st, o)
	case "retirement", "withdrawal":
		v, ok := st.objects[o.Notice.Manifest]
		if !ok || v.Manifest == nil {
			return fail("DANGLING_MANIFEST", 2)
		}
		for _, id := range o.Notice.Candidates {
			r, ok := st.heads[id]
			if !ok || st.objects[r.Digest].Candidate == nil {
				return fail("DANGLING_NOTICE_TARGET", 2)
			}
		}
	}
	return nil
}
func (st *state) with(o Object) *state {
	out := emptyState()
	for k, v := range st.heads {
		out.heads[k] = v
	}
	for k, v := range st.objects {
		out.objects[k] = v
	}
	out.order = slices.Clone(st.order)
	out.events = st.events
	out.head = st.head
	b, _ := json.Marshal(o)
	d := Hash(b)
	out.objects[d] = o
	out.order = append(out.order, d)
	out.heads[o.ID] = Ref{o.ID, o.Revision, d}
	return out
}

type graph struct {
	origin, allocation    map[string]string
	unresolved, synthetic map[string]bool
	candidates            map[string]Object
	ids                   []string
}

func root(m map[string]string, id string) string {
	for m[id] != id {
		id = m[id]
	}
	return id
}
func join(m map[string]string, a, b string) {
	a = root(m, a)
	b = root(m, b)
	if a < b {
		m[b] = a
	} else {
		m[a] = b
	}
}
func graphs(st *state) (graph, error) {
	g := graph{map[string]string{}, map[string]string{}, map[string]bool{}, map[string]bool{}, map[string]Object{}, nil}
	for id, r := range st.heads {
		o := st.objects[r.Digest]
		if o.Candidate != nil {
			g.ids = append(g.ids, id)
			g.candidates[id] = o
			g.origin[id] = id
			g.allocation[id] = id
		}
	}
	slices.Sort(g.ids)
	origins, episodes, projects, locators := map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}
	ancestry := map[string][]string{}
	for _, d := range st.order {
		o := st.objects[d]
		if c := o.Candidate; c != nil {
			// All historical identity assignments remain conservatively connected. A
			// correction cannot erase exposure or split dependencies from older revisions.
			for i, f := range []Fact{c.Origin, c.Episode, c.ProjectLineage} {
				if f.State != "KNOWN" {
					continue
				}
				m := origins
				if i == 1 {
					m = episodes
				}
				if i == 2 {
					m = projects
				}
				key := value(f)
				if prev, ok := m[key]; ok {
					join(g.allocation, o.ID, prev)
					if i < 2 {
						join(g.origin, o.ID, prev)
					}
				} else {
					m[key] = o.ID
				}
			}
			for _, c := range c.Sources {
				if prev, ok := locators[c.Locator]; ok {
					join(g.origin, o.ID, prev)
					join(g.allocation, o.ID, prev)
				} else {
					locators[c.Locator] = o.ID
				}
			}
			for _, p := range c.Parents {
				if _, ok := g.candidates[p.ID]; !ok {
					return g, fail("DANGLING_SYNTHETIC_PARENT", 2)
				}
				join(g.allocation, o.ID, p.ID)
				join(g.origin, o.ID, p.ID)
				ancestry[o.ID] = append(ancestry[o.ID], p.ID)
				g.synthetic[o.ID] = true
			}
			if c.OriginClass != "natural_incident" {
				g.synthetic[o.ID] = true
			}
		}
		if o.Relations != nil {
			for _, r := range *o.Relations {
				if _, e := st.resolve(r.From); e != nil {
					return g, e
				}
				if _, e := st.resolve(r.To); e != nil {
					return g, e
				}
				if r.Status == "REJECTED" {
					continue
				}
				// QUOTES_ORIGIN is a citation edge; it must not fuse unrelated episodes.
				if r.Type == "QUOTES_ORIGIN" {
					join(g.allocation, r.From.ID, r.To.ID)
					continue
				}
				join(g.allocation, r.From.ID, r.To.ID)
				if r.Status == "POSSIBLE" {
					g.unresolved[r.From.ID] = true
					g.unresolved[r.To.ID] = true
					continue
				}
				if r.Type != "SHARED_PROJECT_LINEAGE" {
					join(g.origin, r.From.ID, r.To.ID)
				}
				if one(r.Type, "MIRROR_OF", "COPIED_SUMMARY_OF", "FOLLOWUP_OF", "REVISION_OF", "SYNTHETIC_DESCENDANT_OF") {
					ancestry[r.From.ID] = append(ancestry[r.From.ID], r.To.ID)
				}
				if r.Type == "SYNTHETIC_DESCENDANT_OF" {
					g.synthetic[r.From.ID] = true
				}
			}
		}
	}
	visiting, done := map[string]bool{}, map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] {
			return fail("ANCESTRY_CYCLE", 2)
		}
		if done[id] {
			return nil
		}
		visiting[id] = true
		for _, p := range ancestry[id] {
			if e := visit(p); e != nil {
				return e
			}
		}
		visiting[id] = false
		done[id] = true
		return nil
	}
	for _, id := range g.ids {
		if e := visit(id); e != nil {
			return g, e
		}
		c := g.candidates[id].Candidate
		if c.Origin.State != "KNOWN" || c.Episode.State != "KNOWN" || c.Independence.State != "KNOWN" {
			g.unresolved[id] = true
		}
	}
	// Synthetic ancestry propagates transitively, even if a descendant was mislabeled.
	for changed := true; changed; {
		changed = false
		for id, parents := range ancestry {
			for _, p := range parents {
				if g.synthetic[p] && !g.synthetic[id] {
					g.synthetic[id] = true
					changed = true
				}
			}
		}
	}
	for _, id := range g.ids {
		g.origin[id] = root(g.origin, id)
		g.allocation[id] = root(g.allocation, id)
	}
	unresolvedOrigins := map[string]bool{}
	for id := range g.unresolved {
		unresolvedOrigins[g.origin[id]] = true
	}
	for _, id := range g.ids {
		if unresolvedOrigins[g.origin[id]] {
			g.unresolved[id] = true
		}
	}
	return g, nil
}
func rank(s string) int {
	switch s {
	case "YES":
		return 2
	case "UNKNOWN":
		return 1
	case "NO_WITH_BASIS":
		return 0
	}
	return 1
}
func union(a, b string) string {
	if rank(a) >= rank(b) {
		return a
	}
	return b
}
func (s Store) audit(o Object, at string, require bool) (string, error) {
	c := *o.Candidate
	tier := "REGISTER_ONLY"
	for _, r := range c.Sources {
		if r.Rights.Metadata != "allow" || (r.Rights.RetainUntil != "" && !before(at, r.Rights.RetainUntil)) {
			if require {
				return tier, fail("RIGHTS_BLOCKED", 3)
			}
			return "CURRENT_USE_BLOCKED", nil
		}
		if one(r.Availability, "deleted", "retracted") {
			if require {
				return "CURRENT_USE_BLOCKED", fail("SOURCE_WITHDRAWN", 3)
			}
			return "CURRENT_USE_BLOCKED", nil
		}
	}
	if len(c.Captured) == 0 {
		return tier, nil
	}
	tier = "LOCAL_AUDIT_AVAILABLE"
	if s.EvidenceRoot == "" || secureDir(s.EvidenceRoot, false) != nil {
		if require {
			return tier, fail("EVIDENCE_STORE_UNAVAILABLE", 2)
		}
		return "CONDITIONAL_AUDIT", nil
	}
	cSources := c.Sources
	originClass := c.OriginClass
	for _, c := range c.Captured {
		if c.StoreID != s.EvidenceStoreID {
			if require {
				return tier, fail("EVIDENCE_STORE_MISMATCH", 2)
			}
			return "CONDITIONAL_AUDIT", nil
		}
		r, e := s.readEvidence(c.Record)
		if e != nil {
			if require {
				return tier, e
			}
			return "CONDITIONAL_AUDIT", nil
		}
		found := false
		for _, a := range r.Artifacts {
			if a.ArtifactID == c.Source.ArtifactID && a.SourceID == c.Source.SourceID && a.Revision == c.Source.Revision && a.ContentDigest == c.Source.Digest && a.Representation == c.Representation {
				found = true
				// Phase 2A classification is a constraint, never proof of incident truth.
				synthetic := a.Sensitivity == "synthetic" || a.QualityClass == "synthetic"
				if !one(a.Sensitivity, "public", "synthetic") || synthetic && (o.Sensitivity != "synthetic" || originClass == "natural_incident") {
					return tier, fail("CAPTURE_CLASSIFICATION_MISMATCH", 2)
				}
				for _, citation := range cSources {
					if citation.ID == c.CitationID && citation.Rights != a.Rights {
						return tier, fail("CAPTURE_RIGHTS_MISMATCH", 2)
					}
				}
				if a.Rights.Metadata != "allow" || a.Rights.Digest != "allow" || (a.Rights.RetainUntil != "" && !before(at, a.Rights.RetainUntil)) || one(a.Availability, "deleted", "retracted") {
					if require {
						return tier, fail("RIGHTS_BLOCKED", 3)
					}
					return "CURRENT_USE_BLOCKED", nil
				}
				for _, id := range c.Locations {
					ok := false
					for _, l := range a.Locations {
						if l.ID == id {
							if !one(l.Sensitivity, "public", "synthetic") || l.Sensitivity == "synthetic" && (o.Sensitivity != "synthetic" || originClass == "natural_incident") {
								return tier, fail("CAPTURE_CLASSIFICATION_MISMATCH", 2)
							}
							ok = true
						}
					}
					if !ok {
						return tier, fail("CAPTURE_LOCATION_MISMATCH", 2)
					}
				}
				if !a.Retained {
					tier = "CONDITIONAL_AUDIT"
				}
			}
		}
		if !found {
			return tier, fail("CAPTURE_IDENTITY_MISMATCH", 2)
		}
	}
	return tier, nil
}
func (s Store) summarize(st *state, policyDigest Digest, at string, freeze bool, currentView ...bool) (Summary, error) {
	out := Summary{Contexts: []ContextAccess{}, Members: []Member{}, Counts: []Target{}, Deficits: []Target{}, Limitations: []string{"MODEL_PRETRAINING_UNKNOWN", "SELF_REPORTED_CHRONOLOGY", "ATTRIBUTION_NOT_SEMANTIC_EVIDENCE"}, TargetStatus: "TARGETS_MET"}
	relaxed := len(currentView) > 0 && currentView[0]
	po, ok := st.objects[policyDigest]
	if !ok || po.Policy == nil {
		return out, fail("POLICY_NOT_REGISTERED", 3)
	}
	p := po.Policy
	g, e := graphs(st)
	if e != nil {
		return out, e
	}
	assignments := map[string]string{}
	history := map[string]int{}
	splitRank := map[string]int{"DEVELOPMENT": 1, "VALIDATION": 2, "HOLDOUT": 3}
	// Ordered history permits retirement into development, never promotion back to
	// prospective material. Excluded rows retain historical component use.
	for _, d := range st.order {
		if a := st.objects[d].Assignments; a != nil {
			table := map[string]string{}
			if st.objects[d].PolicyDigest == policyDigest {
				assignments = map[string]string{}
			}
			for _, r := range *a {
				if _, ok := g.candidates[r.Candidate]; !ok {
					return out, fail("DANGLING_ASSIGNMENT", 2)
				}
				comp := g.allocation[r.Candidate]
				if r.Split != "NOT_SELECTED" {
					if prior, ok := table[comp]; ok && prior != r.Split {
						if !relaxed {
							return out, fail("COMPONENT_SPLIT_CONFLICT", 3)
						}
						out.Limitations = append(out.Limitations, "COMPONENT_SPLIT_CONFLICT")
					}
					table[comp] = r.Split
					if prior, ok := history[comp]; ok && splitRank[r.Split] > prior {
						if !relaxed {
							return out, fail("SPLIT_HISTORY_CONFLICT", 3)
						}
						out.Limitations = append(out.Limitations, "SPLIT_HISTORY_CONFLICT")
					}
				}
				if st.objects[d].PolicyDigest == policyDigest {
					assignments[r.Candidate] = r.Split
				}
			}
			for comp, split := range table {
				history[comp] = splitRank[split]
			}
		}
	}
	assignmentStale := false
	hasAssignment := false
	for _, d := range st.order {
		o := st.objects[d]
		if o.Assignments != nil && o.PolicyDigest == policyDigest {
			hasAssignment = true
			assignmentStale = false
			continue
		}
		if hasAssignment && (o.Candidate != nil || o.Relations != nil || o.Exposure != nil) {
			assignmentStale = true
		}
	}
	if relaxed && assignmentStale {
		out.Limitations = append(out.Limitations, "STALE_ASSIGNMENT_STATE")
	}
	// Every actor/context involved in intake, decisions, assignment and clustering
	// needs explicit category declarations; missing context is UNKNOWN.
	contexts := map[string]map[string]bool{}
	ctxKey := func(i Identity) string { return i.Actor + "\x00" + i.Context + "\x00" + i.Model }
	addContext := func(id string, i Identity) {
		if contexts[id] == nil {
			contexts[id] = map[string]bool{}
		}
		contexts[id][ctxKey(i)] = true
	}
	for _, id := range g.ids {
		o := g.candidates[id]
		if o.PolicyDigest == policyDigest {
			addContext(id, po.Producer)
			if st.freezeProducer != nil {
				addContext(id, *st.freezeProducer)
			}
		}
	}
	accesses := map[string]map[string]map[string]string{}
	for _, d := range st.order {
		o := st.objects[d]
		if o.Candidate != nil {
			addContext(o.ID, o.Producer)
			addContext(o.ID, o.Candidate.Discovery.Discoverer)
			// Actor-only selection fields inherit the producing context/model; a
			// different discoverer keeps its independently declared identity too.
			selector := o.Producer
			selector.Actor = o.Candidate.Decision.Actor
			addContext(o.ID, selector)
			for _, assessment := range o.Candidate.Assessments {
				selector.Actor = assessment.Assessor
				addContext(o.ID, selector)
			}
		}
		if o.Manifest != nil {
			for _, r := range o.Manifest.Inventory {
				if st.objects[r.Digest].Candidate != nil {
					addContext(r.ID, o.Producer)
				}
			}
		}
		if o.Assignments != nil {
			for _, a := range *o.Assignments {
				addContext(a.Candidate, o.Producer)
			}
		}
		if o.Relations != nil {
			for _, r := range *o.Relations {
				addContext(r.From.ID, o.Producer)
				addContext(r.To.ID, o.Producer)
			}
		}
		if ex := o.Exposure; ex != nil {
			for _, id := range ex.Candidates {
				if _, ok := g.candidates[id]; !ok {
					return out, fail("DANGLING_EXPOSURE_TARGET", 2)
				}
				addContext(id, o.Producer)
				comp := id
				if accesses[comp] == nil {
					accesses[comp] = map[string]map[string]string{}
				}
				key := ctxKey(o.Producer)
				if accesses[comp][key] == nil {
					accesses[comp][key] = map[string]string{}
				}
				v := union(ex.State, ex.AvailableAccess)
				old, ok := accesses[comp][key][ex.Category]
				if ok {
					v = union(old, v)
				}
				accesses[comp][key][ex.Category] = v
			}
		}
	}
	// Missing declarations are UNKNOWN per candidate BEFORE component union.
	componentAccess := map[string]map[string]map[string]string{}
	for id, ctxs := range contexts {
		comp := g.allocation[id]
		if componentAccess[comp] == nil {
			componentAccess[comp] = map[string]map[string]string{}
		}
		for ctx := range ctxs {
			if componentAccess[comp][ctx] == nil {
				componentAccess[comp][ctx] = map[string]string{}
			}
			for _, cat := range categories {
				v, ok := accesses[id][ctx][cat]
				if !ok {
					v = "UNKNOWN"
				}
				if old, ok := componentAccess[comp][ctx][cat]; ok {
					v = union(old, v)
				}
				componentAccess[comp][ctx][cat] = v
			}
		}
	}
	for comp, ctxs := range componentAccess {
		for ctx, cats := range ctxs {
			row := ContextAccess{Component: comp, Context: ctx, Access: []Access{}}
			for _, cat := range categories {
				row.Access = append(row.Access, Access{cat, cats[cat]})
			}
			out.Contexts = append(out.Contexts, row)
		}
	}
	_ = sorted(out.Contexts, func(v ContextAccess) string { return v.Component + "\x00" + v.Context })
	publishers := publisherGroups(st, g)
	usedOrigin := map[string]bool{}
	componentSplit := map[string]string{}
	counts := map[string]uint32{}
	publisher := map[string]uint32{}
	unknownPublisher := false
	for _, id := range g.ids {
		o := g.candidates[id]
		c := o.Candidate
		if c.Discovery.Round != p.Round {
			continue
		}
		if freeze && c.Decision.State == "INCLUDED" && c.FreezeState != "SELECTED" {
			return out, fail("CANDIDATE_NOT_SELECTED", 3)
		}
		if freeze && c.Decision.State == "PENDING" {
			return out, fail("PENDING_INVENTORY", 3)
		}
		split, has := assignments[id]
		if !has {
			if freeze {
				return out, fail("ASSIGNMENT_INCOMPLETE", 3)
			}
			split = "UNASSIGNED"
		}
		if c.Decision.State != "INCLUDED" && split != "NOT_SELECTED" && split != "UNASSIGNED" {
			if !relaxed {
				return out, fail("EXCLUDED_MEMBERSHIP", 3)
			}
			out.Limitations = append(out.Limitations, "EXCLUDED_MEMBERSHIP")
		}
		if c.Decision.State == "INCLUDED" && one(split, "NOT_SELECTED", "UNASSIGNED") && freeze {
			return out, fail("INCLUDED_UNASSIGNED", 3)
		}
		comp := g.allocation[id]
		if prior, ok := history[comp]; ok && splitRank[split] > prior {
			if !relaxed {
				return out, fail("SPLIT_HISTORY_CONFLICT", 3)
			}
			out.Limitations = append(out.Limitations, "SPLIT_HISTORY_CONFLICT")
		}
		if split != "NOT_SELECTED" && split != "UNASSIGNED" {
			if prev, ok := componentSplit[comp]; ok && prev != split {
				if !relaxed {
					return out, fail("COMPONENT_SPLIT_CONFLICT", 3)
				}
				out.Limitations = append(out.Limitations, "COMPONENT_SPLIT_CONFLICT")
			}
			componentSplit[comp] = split
		}
		access := []Access{}
		fresh := "ELIGIBLE_UNREVEALED"
		for _, category := range categories {
			v := "NO_WITH_BASIS"
			for ctx := range componentAccess[comp] {
				declared, ok := componentAccess[comp][ctx][category]
				if !ok {
					declared = "UNKNOWN"
				}
				v = union(v, declared)
			}
			access = append(access, Access{category, v})
			if !one(category, "causal_family_label", "architecture_packet") {
				if v == "YES" {
					fresh = "EXPOSED"
				} else if v == "UNKNOWN" && fresh != "EXPOSED" {
					fresh = "UNKNOWN"
				}
			}
		}
		if (c.ProjectLineage.State != "KNOWN" || relaxed && assignmentStale) && fresh != "EXPOSED" {
			fresh = "UNKNOWN"
		}
		if one(split, "VALIDATION", "HOLDOUT") && fresh != "ELIGIBLE_UNREVEALED" {
			if !relaxed {
				return out, fail("FRESH_SPLIT_REFUSED", 3)
			}
			out.Limitations = append(out.Limitations, "FRESH_SPLIT_REFUSED")
		}
		// The corpus capsule has metadata only. Captured bytes are an external,
		// conditional audit path; current audit is recomputed separately by Verify.
		audit := "REGISTER_ONLY"
		if len(c.Captured) > 0 {
			audit = "CONDITIONAL_AUDIT"
		}
		natural := !g.synthetic[id] && !g.unresolved[id] && o.Sensitivity == "public" && c.Decision.State == "INCLUDED"
		representative := natural && !usedOrigin[g.origin[id]]
		if representative {
			usedOrigin[g.origin[id]] = true
		}
		fam := "UNASSIGNED"
		if c.Family.State == "CANDIDATE" {
			fam = value(c.Family.Primary)
		}
		m := Member{id, g.origin[id], comp, representative, natural, publishers[id], value(c.ProjectLineage), value(c.Stack), fam, c.Decision.State, split, access, fresh, audit}
		out.Members = append(out.Members, m)
		counts["submissions"]++
		counts["decision:"+c.Decision.State]++
		for _, r := range c.Decision.Reasons {
			counts["reason:"+r]++
		}
		if c.Decision.State == "INCLUDED" {
			for _, challenge := range c.ChallengeCategories {
				counts["challenge:"+challenge]++
			}
		}
		counts["audit:"+audit]++
		if g.synthetic[id] {
			counts["synthetic"]++
		}
		if g.unresolved[id] {
			counts["unresolved"]++
		}
		if representative {
			counts["natural"]++
			counts["family:"+fam]++
			counts["split:"+split]++
			counts["family-split:"+fam+":"+split]++
			counts["project:"+m.Project]++
			counts["stack:"+m.Stack]++
			publisher[m.Publisher]++
			if m.Publisher == "UNKNOWN" {
				unknownPublisher = true
			}
		}
		if natural && !representative {
			counts["duplicates"]++
		}
	}
	for _, d := range st.order {
		if st.objects[d].AttemptReason != nil {
			counts["failed_attempts"]++
		}
	}
	counts["origins"] = uint32(len(usedOrigin))
	counts["publisher_groups"] = uint32(len(publisher))
	for name, n := range publisher {
		counts["publisher:"+name] = n
	}
	diversity := unknownPublisher
	for _, n := range publisher {
		if 3*n > counts["natural"] {
			diversity = true
		}
	}
	if p.DiversityException != "" {
		diversity = true
	}
	if diversity {
		if freeze && p.DiversityException == "" {
			return out, fail("PUBLISHER_CONCENTRATION", 3)
		}
		out.Limitations = append(out.Limitations, "DIVERSITY_SHORTFALL")
		out.TargetStatus = "SHORTFALL"
	}
	// The approved research target remains visible even if a policy supplies a
	// smaller budget. A small freeze is allowed; reducing targets cannot hide gaps.
	targets := map[string]uint32{"natural": 25, "split:DEVELOPMENT": 15, "split:VALIDATION": 5, "split:HOLDOUT": 5}
	for _, family := range families {
		targets["family:"+family] = 5
		for split, n := range map[string]uint32{"DEVELOPMENT": 3, "VALIDATION": 1, "HOLDOUT": 1} {
			targets["family-split:"+family+":"+split] = n
		}
	}
	for _, challenge := range p.Challenges {
		targets["challenge:"+challenge] = 1
	}
	for _, target := range p.Targets {
		if target.Count > targets[target.Name] {
			targets[target.Name] = target.Count
		}
	}
	for name, n := range targets {
		if counts[name] < n {
			out.Deficits = append(out.Deficits, Target{name, n - counts[name]})
			out.TargetStatus = "SHORTFALL"
		}
	}
	for name, n := range counts {
		out.Counts = append(out.Counts, Target{name, n})
	}
	slices.SortFunc(out.Counts, func(a, b Target) int { return strings.Compare(a.Name, b.Name) })
	slices.SortFunc(out.Deficits, func(a, b Target) int { return strings.Compare(a.Name, b.Name) })
	slices.Sort(out.Limitations)
	out.Limitations = slices.Compact(out.Limitations)
	return out, nil
}
func inventory(st *state) []Ref {
	out := []Ref{}
	for _, r := range st.heads {
		o := st.objects[r.Digest]
		if o.Candidate != nil || o.AttemptReason != nil {
			out = append(out, r)
		}
	}
	slices.SortFunc(out, func(a, b Ref) int { return strings.Compare(a.ID, b.ID) })
	return out
}
func closure(st *state) []Digest {
	out := slices.Clone(st.order)
	slices.SortFunc(out, func(a, b Digest) int { return strings.Compare(a.Hex, b.Hex) })
	return out
}
func (s Store) checkManifest(st *state, o Object) error {
	if refs, e := withdrawals(st, o.Manifest.Inventory); e != nil {
		return e
	} else if len(refs) > 0 {
		return fail("WITHDRAWN_COMPONENT", 3)
	}
	if e := s.checkManifestSnapshot(st, o); e != nil {
		return e
	}
	for _, r := range inventory(st) {
		v := st.objects[r.Digest]
		if v.Candidate != nil {
			if _, e := s.audit(v, o.Manifest.CheckTime, true); e != nil {
				return e
			}
		}
	}
	return nil
}
func (s Store) checkManifestSnapshot(st *state, o Object) error {
	m := o.Manifest
	if !same(m.LedgerHead, st.head) || m.LedgerSequence != uint32(len(st.events)) {
		return fail("MANIFEST_HEAD_MISMATCH", 2)
	}
	if !reflect.DeepEqual(m.Inventory, inventory(st)) {
		return fail("INVENTORY_MISMATCH", 2)
	}
	if !reflect.DeepEqual(m.Closure, closure(st)) {
		return fail("CLOSURE_MISMATCH", 2)
	}
	p, e := st.resolve(m.Policy)
	if e != nil || p.Policy == nil || m.Policy.Digest != o.PolicyDigest {
		return fail("POLICY_BINDING_MISMATCH", 2)
	}
	a, e := st.resolve(m.Assignment)
	if e != nil || a.Assignments == nil {
		return fail("ASSIGNMENT_BINDING_MISMATCH", 2)
	}
	latest := Ref{}
	for _, d := range st.order {
		v := st.objects[d]
		if v.Assignments != nil && v.PolicyDigest == o.PolicyDigest {
			latest = Ref{v.ID, v.Revision, d}
		}
	}
	if a.PolicyDigest != o.PolicyDigest {
		return fail("ASSIGNMENT_POLICY_MISMATCH", 3)
	}
	if latest != m.Assignment {
		return fail("STALE_ASSIGNMENT", 3)
	}
	// The immutable assignment receipt binds the exact input ledger prefix.
	// Any later candidate/dependency/exposure change requires a new table. This
	// conservatively binds aliases and cross-round dependencies without a second
	// writable copy of candidate/graph/exposure identities in each row.
	assigned := false
	for _, d := range st.order {
		if d == m.Assignment.Digest {
			assigned = true
			continue
		}
		v := st.objects[d]
		if assigned && (v.Candidate != nil || v.Relations != nil || v.Exposure != nil) {
			return fail("STALE_ASSIGNMENT_STATE", 3)
		}
	}
	expectedRows := map[string]bool{}
	for _, r := range inventory(st) {
		v := st.objects[r.Digest]
		if v.Candidate != nil && v.Candidate.Discovery.Round == p.Policy.Round {
			expectedRows[r.ID] = true
		}
	}
	if len(*a.Assignments) != len(expectedRows) {
		return fail("ASSIGNMENT_INCOMPLETE", 3)
	}
	for _, row := range *a.Assignments {
		if !expectedRows[row.Candidate] {
			return fail("ASSIGNMENT_INVENTORY_MISMATCH", 2)
		}
	}
	if st.heads[p.ID].Digest != m.Policy.Digest {
		return fail("STALE_POLICY", 3)
	}
	if o.Revision == 1 && m.ParentManifest != nil || o.Revision > 1 && !same(m.ParentManifest, o.Parent) {
		return fail("MANIFEST_PARENT_MISMATCH", 2)
	}
	for _, r := range inventory(st) {
		v := st.objects[r.Digest]
		if v.Candidate != nil {
			if v.Candidate.Decision.State == "PENDING" {
				return fail("PENDING_INVENTORY", 3)
			}
		}
	}
	summaryState := *st
	summaryState.freezeProducer = &o.Producer
	summary, e := s.summarize(&summaryState, o.PolicyDigest, m.CheckTime, true)
	if e != nil {
		return e
	}
	if !reflect.DeepEqual(summary, m.Summary) {
		return fail("FALSE_MANIFEST_SUMMARY", 2)
	}
	return nil
}

// PrepareFreeze computes the reviewable manifest. Apply rechecks it under lock;
// this method never commits or masks a concurrent expected-head change.
func (s Store) PrepareFreeze(o Object) (Object, error) {
	f, e := s.lock(false)
	if e != nil {
		return Object{}, e
	}
	defer f.Close()
	st, e := s.scan()
	if e != nil {
		return Object{}, e
	}
	if len(st.orphans) > 0 {
		return Object{}, fail("ORPHAN_PUBLICATION", 2)
	}
	if o.Manifest == nil {
		return Object{}, fail("INVALID_MANIFEST", 2)
	}
	m := *o.Manifest
	m.Inventory = inventory(st)
	m.Closure = closure(st)
	m.LedgerHead = st.head
	m.LedgerSequence = uint32(len(st.events))
	m.Method = "explicit-table-v1"
	m.Seed = "NOT_USED"
	m.Chronology = "SELF_REPORTED"
	m.Pretraining = "UNKNOWN"
	summaryState := *st
	summaryState.freezeProducer = &o.Producer
	m.Summary, e = s.summarize(&summaryState, o.PolicyDigest, m.CheckTime, true)
	if e != nil {
		return Object{}, e
	}
	o.Manifest = &m
	return o, nil
}

// Publisher aliases are metadata assertions, not verified corporate ownership.
// ASCII/Unicode lowercase and surrounding whitespace removal are pinned here;
// URL identity bytes are never normalized. Historical aliases cannot be erased.
func publisherGroups(st *state, g graph) map[string]string {
	groups := map[string]string{}
	names := map[string]string{}
	knownGroup := map[string]string{}
	for _, id := range g.ids {
		groups[id] = id
	}
	for _, d := range st.order {
		o := st.objects[d]
		if c := o.Candidate; c != nil && c.Publisher.Group.State == "KNOWN" {
			all := append(slices.Clone(c.Publisher.Names), value(c.Publisher.Group))
			for _, name := range all {
				key := strings.ToLower(strings.TrimSpace(name))
				if old, ok := names[key]; ok {
					join(groups, o.ID, old)
				} else {
					names[key] = o.ID
				}
			}
		}
	}
	for _, id := range g.ids {
		c := g.candidates[id].Candidate
		if c.Publisher.Group.State == "KNOWN" {
			r := root(groups, id)
			name := strings.ToLower(strings.TrimSpace(value(c.Publisher.Group)))
			if old, ok := knownGroup[r]; !ok || name < old {
				knownGroup[r] = name
			}
		}
	}
	out := map[string]string{}
	for _, id := range g.ids {
		out[id] = "UNKNOWN"
		if g.candidates[id].Candidate.Publisher.Group.State == "KNOWN" {
			out[id] = knownGroup[root(groups, id)]
		}
	}
	return out
}

// references is a historical structural check: unlike current audit it needs no
// external source bytes, and can run on every recovery/inspection replay.
func references(st *state, o Object) error {
	if o.Kind != "policy" && !(o.Kind == "attempt" && o.PolicyDigest == VocabularyDigest()) {
		p, ok := st.objects[o.PolicyDigest]
		if !ok || p.Policy == nil {
			return fail("DANGLING_POLICY", 2)
		}
	}
	if o.Assignments != nil {
		for _, row := range *o.Assignments {
			r, ok := st.heads[row.Candidate]
			v := st.objects[r.Digest]
			if !ok || v.Candidate == nil {
				return fail("DANGLING_ASSIGNMENT", 2)
			}
			if v.PolicyDigest != o.PolicyDigest {
				return fail("ASSIGNMENT_POLICY_MISMATCH", 3)
			}
		}
	}
	refs := []Ref{}
	if c := o.Candidate; c != nil {
		refs = append(refs, c.Decision.Policy)
		refs = append(refs, c.Parents...)
		if c.Decision.Policy.Digest != o.PolicyDigest {
			return fail("POLICY_BINDING_MISMATCH", 2)
		}
		for _, d := range c.ExposureRefs {
			v, ok := st.objects[d]
			if !ok || v.Exposure == nil {
				return fail("DANGLING_EXPOSURE", 2)
			}
		}
		if c.Assignment != nil {
			v, ok := st.objects[*c.Assignment]
			if !ok || v.Assignments == nil {
				return fail("DANGLING_ASSIGNMENT", 2)
			}
		}
	}
	if o.Relations != nil {
		for _, r := range *o.Relations {
			refs = append(refs, r.From, r.To)
		}
	}
	if o.Notice != nil {
		v, ok := st.objects[o.Notice.Manifest]
		if !ok || v.Manifest == nil {
			return fail("DANGLING_MANIFEST", 2)
		}
	}
	for _, r := range refs {
		if _, e := st.resolve(r); e != nil {
			return e
		}
	}
	return nil
}

// Withdrawal follows currently known dependency components, across manifest
// versions and aliases. Excluded inventory remains inspectable without being
// treated as selected material. There is no implicit withdrawal reset operation.
func withdrawals(st *state, inventory []Ref) ([]Ref, error) {
	g, e := graphs(st)
	if e != nil {
		return nil, e
	}
	selected := map[string]bool{}
	for _, r := range inventory {
		o, e := st.resolve(r)
		if e != nil {
			return nil, e
		}
		if o.Candidate != nil && o.Candidate.Decision.State == "INCLUDED" {
			selected[g.allocation[o.ID]] = true
		}
	}
	out := []Ref{}
	for _, ev := range st.events {
		o := st.objects[ev.Output.Digest]
		if o.Kind != "withdrawal" {
			continue
		}
		ids := slices.Clone(o.Notice.Candidates)
		if len(ids) == 0 {
			manifest := st.objects[o.Notice.Manifest]
			for _, r := range manifest.Manifest.Inventory {
				if st.objects[r.Digest].Candidate != nil {
					ids = append(ids, r.ID)
				}
			}
		}
		for _, id := range ids {
			if selected[g.allocation[id]] {
				out = append(out, ev.Output)
				break
			}
		}
	}
	return out, nil
}
