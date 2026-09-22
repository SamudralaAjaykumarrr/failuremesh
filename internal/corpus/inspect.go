package corpus

import (
	"encoding/json"
	"slices"
)

// Verify pins both the historical receipt and an explicit read head. Historical
// mode makes no current-use assertion. Current mode requires the actual head.
func (s Store) Verify(d Digest, head *Digest, at string, current bool) (Report, error) {
	out := Report{Integrity: "INVALID", SelectionTargetStatus: "UNKNOWN", CurrentUseStatus: "NOT_ASSESSED", AuditCapability: "REGISTER_ONLY", Freshness: "UNKNOWN", ChronologyTier: "SELF_REPORTED", Pretraining: "UNKNOWN", Notices: []Ref{}, Limitations: []string{"LOCAL_ADMINISTRATOR_TRUSTED", "UNRECORDED_ACCESS_UNDETECTABLE", "SOURCE_TRUTH_NOT_ASSESSED"}}
	if !validDigest(d) || !stamp(at) {
		return out, fail("INVALID_VERIFY_INPUT", 2)
	}
	f, e := s.lock(false)
	if e != nil {
		return out, e
	}
	defer f.Close()
	st, e := s.scan()
	if e != nil {
		return out, e
	}
	if len(st.orphans) > 0 {
		return out, fail("ORPHAN_PUBLICATION", 2)
	}
	if current && !same(head, st.head) {
		return out, fail("STALE_READ_HEAD", 4)
	}
	snapshot := emptyState()
	found := false
	var manifest Object
	for _, ev := range st.events {
		o := st.objects[ev.Output.Digest]
		if ev.Output.Digest == d {
			if o.Manifest == nil || ev.Action != "freeze" {
				return out, fail("NOT_COMMITTED_MANIFEST", 2)
			}
			manifest = o
			if e := s.checkManifestSnapshot(snapshot, o); e != nil {
				return out, e
			}
			found = true
			break
		}
		snapshot.objects[ev.Output.Digest] = o
		snapshot.order = append(snapshot.order, ev.Output.Digest)
		snapshot.heads[o.ID] = ev.Output
		snapshot.events = append(snapshot.events, ev)
		b, _ := json.Marshal(ev)
		h := Hash(b)
		snapshot.head = &h
	}
	if head == nil {
		return out, fail("READ_HEAD_REQUIRED", 2)
	}
	headSequence := uint32(0)
	for _, ev := range st.events {
		b, _ := json.Marshal(ev)
		if Hash(b) == *head {
			headSequence = ev.Sequence
		}
	}
	if found && headSequence <= manifest.Manifest.LedgerSequence {
		return out, fail("READ_HEAD_BEFORE_RECEIPT", 2)
	}
	if !found {
		return out, fail("NOT_COMMITTED_MANIFEST", 2)
	}
	out.Integrity = "INTACT"
	out.Historical = manifest.Manifest.Summary
	out.SelectionTargetStatus = out.Historical.TargetStatus
	out.Head = head
	if !current {
		out.CurrentUseStatus = "HISTORICAL_ONLY"
		return out, nil
	}
	out.CurrentUseStatus = "METADATA_REGISTER_ONLY"
	withdrawn, withdrawErr := withdrawals(st, manifest.Manifest.Inventory)
	if withdrawErr != nil {
		return out, withdrawErr
	}
	if len(withdrawn) > 0 {
		out.CurrentUseStatus = "BLOCKED"
		out.Notices = append(out.Notices, withdrawn...)
		out.Limitations = append(out.Limitations, "WITHDRAWN_COMPONENT")
	}

	out.Freshness = "ELIGIBLE_UNREVEALED"
	summary, e := s.summarize(st, manifest.PolicyDigest, at, false, true)
	if e != nil {
		out.CurrentUseStatus = "BLOCKED"
		out.Limitations = append(out.Limitations, e.Error())
		out.Freshness = "UNKNOWN"
	} else {
		out.Current = &summary
		for _, code := range []string{"COMPONENT_SPLIT_CONFLICT", "SPLIT_HISTORY_CONFLICT", "FRESH_SPLIT_REFUSED", "EXCLUDED_MEMBERSHIP", "STALE_ASSIGNMENT_STATE"} {
			if slices.Contains(summary.Limitations, code) {
				out.CurrentUseStatus = "BLOCKED"
				out.Limitations = append(out.Limitations, code)
			}
		}
	}
	// Report exposure even when it is precisely the reason current split checks fail.
	if e != nil && e.Error() == "FRESH_SPLIT_REFUSED" {
		out.Freshness = "EXPOSED_OR_UNKNOWN"
	}
	out.AuditCapability = "LOCAL_AUDIT_AVAILABLE"
	auditRefs := slices.Clone(manifest.Manifest.Inventory)
	for _, r := range manifest.Manifest.Inventory {
		if latest, ok := st.heads[r.ID]; ok && latest.Digest != r.Digest {
			auditRefs = append(auditRefs, latest)
		}
	}
	for _, r := range auditRefs {
		o := st.objects[r.Digest]
		if o.Candidate == nil {
			continue
		}
		tier, err := s.audit(o, at, false)
		if err != nil {
			out.CurrentUseStatus = "BLOCKED"
			out.Limitations = append(out.Limitations, err.Error())
		}
		if tier == "CURRENT_USE_BLOCKED" {
			out.CurrentUseStatus = "BLOCKED"
		}
		if tier == "CONDITIONAL_AUDIT" && out.AuditCapability != "REGISTER_ONLY" {
			out.AuditCapability = "CONDITIONAL_AUDIT"
		}
		if tier == "REGISTER_ONLY" {
			out.AuditCapability = "REGISTER_ONLY"
		}
	}
	for _, ev := range st.events {
		if ev.Sequence <= manifest.Manifest.LedgerSequence+1 {
			continue
		}
		o := st.objects[ev.Output.Digest]
		if o.Notice != nil && o.Notice.Manifest == d {
			if !slices.Contains(out.Notices, ev.Output) {
				out.Notices = append(out.Notices, ev.Output)
			}
			out.CurrentUseStatus = "BLOCKED"
			if o.Kind == "retirement" {
				out.CurrentUseStatus = "RETIRED"
			}
		}
		if o.Exposure != nil || o.Relations != nil || o.Candidate != nil {
			out.Notices = append(out.Notices, ev.Output)
		}
	}
	if out.Current != nil {
		for _, m := range out.Current.Members {
			if m.Freshness == "EXPOSED" {
				out.Freshness = "EXPOSED"
			} else if m.Freshness == "UNKNOWN" && out.Freshness != "EXPOSED" {
				out.Freshness = "UNKNOWN"
			}
		}
	}
	return out, nil
}

type Inspection struct {
	Object      Object   `json:"object"`
	Revisions   []Ref    `json:"revisions"`
	Events      []Event  `json:"events"`
	Limitations []string `json:"limitations"`
}

func (s Store) Inspect(d Digest) (Inspection, error) {
	f, e := s.lock(false)
	if e != nil {
		return Inspection{}, e
	}
	defer f.Close()
	st, e := s.scan()
	if e != nil {
		return Inspection{}, e
	}
	o, ok := st.objects[d]
	if !ok {
		return Inspection{}, fail("OBJECT_NOT_REGISTERED", 2)
	}
	out := Inspection{Object: o, Revisions: []Ref{}, Events: []Event{}, Limitations: []string{"METADATA_ONLY", "SELF_REPORTED_CHRONOLOGY", "NO_SOURCE_BODY_EXPANSION"}}
	for _, ev := range st.events {
		v := st.objects[ev.Output.Digest]
		if v.ID == o.ID {
			out.Revisions = append(out.Revisions, ev.Output)
		}
		if v.ID == o.ID || v.Exposure != nil || v.Notice != nil || v.Relations != nil {
			out.Events = append(out.Events, ev)
		}
	}
	return out, nil
}

// RebuildIndex publishes a disposable snapshot keyed by the authoritative head.
// No reader consults it for membership, counts, decisions, or verification.
func (s Store) RebuildIndex() (*Digest, error) {
	f, e := s.lock(true)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	st, e := s.scan()
	if e != nil {
		return nil, e
	}
	if len(st.orphans) > 0 {
		return nil, fail("ORPHAN_PUBLICATION", 2)
	}
	name := "empty"
	if st.head != nil {
		name = st.head.Hex
	}
	b, e := json.Marshal(struct {
		Head      *Digest `json:"head"`
		Inventory []Ref   `json:"inventory"`
	}{st.head, inventory(st)})
	if e != nil {
		return nil, fail("INDEX_ENCODING", 2)
	}
	if e = s.publish("index/"+name+".json", b); e != nil {
		return nil, e
	}
	return st.head, nil
}
