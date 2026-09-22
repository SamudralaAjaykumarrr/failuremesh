package corpus

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/evidence"
)

// readEvidence revalidates a real existing Phase 2A record and the one-store
// retained-artifact coverage boundary. It never creates an AtomicClaim, calls
// Assess, copies source bytes into this store, or invokes an export path.
func (s Store) readEvidence(d Digest) (evidence.Record, error) {
	totalParents := 0
	for _, b := range s.Parents {
		totalParents += len(b)
		if len(b) > MaxObjectBytes || totalParents > MaxClosureBytes {
			return evidence.Record{}, fail("PARENT_LIMIT", 2)
		}
	}
	f, e := os.OpenFile(filepath.Join(s.EvidenceRoot, ".save.lock"), os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if e != nil {
		return evidence.Record{}, fail("EVIDENCE_LOCK_UNAVAILABLE", 2)
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() {
		return evidence.Record{}, fail("EVIDENCE_LOCK_INVALID", 2)
	}
	if syscall.Flock(int(f.Fd()), syscall.LOCK_SH) != nil {
		return evidence.Record{}, fail("EVIDENCE_LOCK_FAILED", 4)
	}
	raw, e := regularRead(filepath.Join(s.EvidenceRoot, "sha256-"+d.Hex+".json"), MaxObjectBytes)
	if e != nil {
		return evidence.Record{}, fail("CAPTURE_UNAVAILABLE", 2)
	}
	r, e := evidence.Decode(raw, d, s.Parents...)
	if e != nil {
		return evidence.Record{}, fail("PHASE2A_VALIDATION_REFUSED", 2)
	}
	entries, e := os.ReadDir(s.EvidenceRoot)
	if e != nil {
		return evidence.Record{}, fail("EVIDENCE_STORE_UNAVAILABLE", 2)
	}
	if len(entries) > MaxStoreFiles {
		return evidence.Record{}, fail("EVIDENCE_STORE_LIMIT", 2)
	}
	artifacts := map[string]evidence.SourceArtifact{}
	total := 0
	for _, en := range entries {
		if strings.HasPrefix(en.Name(), ".evidence-") {
			return evidence.Record{}, fail("PHASE2A_UNFINISHED_PUBLICATION", 2)
		}
		if !strings.HasPrefix(en.Name(), "artifact-") {
			continue
		}
		b, e := regularRead(filepath.Join(s.EvidenceRoot, en.Name()), MaxObjectBytes)
		if e != nil {
			return evidence.Record{}, e
		}
		total += len(b)
		if total > MaxClosureBytes || len(artifacts) >= MaxClosureObjects {
			return evidence.Record{}, fail("EVIDENCE_STORE_LIMIT", 2)
		}
		var a evidence.SourceArtifact
		if e := strict(b, &a); e != nil {
			return evidence.Record{}, fail("PHASE2A_BINDING_INVALID", 2)
		}
		key, _ := json.Marshal([]string{a.ArtifactID})
		if en.Name() != "artifact-"+Hash(key).Hex+".json" {
			return evidence.Record{}, fail("PHASE2A_BINDING_INVALID", 2)
		}
		revision, _ := json.Marshal(struct {
			Source   string
			Revision int
		}{a.SourceID, a.Revision})
		bound, e := regularRead(filepath.Join(s.EvidenceRoot, "source-revision-"+Hash(revision).Hex+".json"), MaxObjectBytes)
		if e != nil || !bytes.Equal(bound, b) {
			return evidence.Record{}, fail("PHASE2A_BINDING_INVALID", 2)
		}
		artifacts[a.ArtifactID] = a
	}
	for _, a := range r.Artifacts {
		old, ok := artifacts[a.ArtifactID]
		if !ok {
			return evidence.Record{}, fail("PHASE2A_BINDING_MISSING", 2)
		}
		x, _ := json.Marshal(a)
		y, _ := json.Marshal(old)
		if !bytes.Equal(x, y) {
			return evidence.Record{}, fail("PHASE2A_BINDING_CONFLICT", 2)
		}
	}
	combined := r
	combined.Artifacts = nil
	keys := make([]string, 0, len(artifacts))
	for k := range artifacts {
		keys = append(keys, k)
	}
	_ = stringsSet(keys)
	for _, k := range keys {
		combined.Artifacts = append(combined.Artifacts, artifacts[k])
	}
	if evidence.Validate(combined, s.Parents...) != nil {
		return evidence.Record{}, fail("PHASE2A_STORE_COVERAGE_REFUSED", 2)
	}
	return r, nil
}
