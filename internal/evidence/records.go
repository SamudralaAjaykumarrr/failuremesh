package evidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

const maxBytes = 4 << 20

// Hash identifies exact bytes only. It proves neither truth, authenticity,
// chronology, independent origin nor permission to use those bytes.
func Hash(b []byte) Digest { h := sha256.Sum256(b); return Digest{"sha256", hex.EncodeToString(h[:])} }

// strictDecode accepts only this implementation's pinned JSON representation.
// Remarshal comparison also rejects duplicate keys, casing aliases, alternate
// numbers/escapes, trailing data, invalid UTF-8 and ambiguous omitted fields.
func strictDecode(b []byte, v any) error {
	if len(b) > maxBytes || !utf8.Valid(b) {
		return errors.New("invalid encoding or size")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return errors.New("invalid JSON structure")
	}
	encoded, err := json.Marshal(v)
	if err != nil || !bytes.Equal(b, encoded) {
		return errors.New("noncanonical or ambiguous JSON")
	}
	return nil
}
func Encode(r Record, parents ...[]byte) ([]byte, Digest, error) {
	if err := Validate(r, parents...); err != nil {
		return nil, Digest{}, err
	}
	b, err := json.Marshal(r)
	if err != nil || len(b) > maxBytes {
		return nil, Digest{}, errors.New("record encoding failed or too large")
	}
	return b, Hash(b), nil
}
func Decode(b []byte, expected Digest, parents ...[]byte) (Record, error) {
	var r Record
	if !validDigest(expected) || Hash(b) != expected {
		return r, errors.New("record digest mismatch")
	}
	if err := strictDecode(b, &r); err != nil {
		return Record{}, err
	}
	if err := Validate(r, parents...); err != nil {
		return Record{}, err
	}
	return r, nil
}

// Save writes a local content-addressed input plus deterministic assessment.
// The supplied root must be a trusted local directory. Atomic link publication
// refuses replacement; existing exact bytes are idempotent, corruption is an error.
// Administrative filesystem access can still rewrite history: no WORM claim.
func Save(root string, r Record, now string, parents ...[]byte) (Digest, Digest, error) {
	if err := retentionAllowed(r, now, parents...); err != nil {
		return Digest{}, Digest{}, err
	}
	b, d, err := Encode(r, parents...)
	if err != nil {
		return Digest{}, Digest{}, err
	}
	assessment := Assess(r, parents...)
	ab, err := json.Marshal(assessment)
	if err != nil {
		return Digest{}, Digest{}, err
	}
	ad := Hash(ab)
	// Serialize the read/check/publish boundary across cooperating processes.
	// Per-object exclusive publication alone cannot prevent complementary captures
	// from both passing a check against an earlier store state.
	if err := os.MkdirAll(root, 0700); err != nil {
		return Digest{}, Digest{}, err
	}
	lock, err := os.OpenFile(filepath.Join(root, ".save.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return Digest{}, Digest{}, err
	}
	defer lock.Close() // Closing releases the OS lock, including on a failed save.
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return Digest{}, Digest{}, err
	}
	persisted, err := persistedArtifacts(root)
	if err != nil {
		return Digest{}, Digest{}, err
	}
	if err := validateContentCoverage(append(persisted, r.Artifacts...), parents...); err != nil {
		return Digest{}, Digest{}, err
	}
	// Bind logical identities to immutable bytes across saves in this local store.
	for _, a := range r.Artifacts {
		raw, _ := json.Marshal(a)
		revision, _ := json.Marshal(struct {
			Source   string
			Revision int
		}{a.SourceID, a.Revision})
		if err := publish(root, artifactBindingName(a.ArtifactID), raw); err != nil {
			return Digest{}, Digest{}, err
		}
		if err := publish(root, "source-revision-"+Hash(revision).Hex+".json", raw); err != nil {
			return Digest{}, Digest{}, err
		}
	}
	for _, c := range r.Claims {
		raw, _ := json.Marshal(c)
		identity, _ := json.Marshal(struct {
			Claim    string
			Revision int
		}{c.ID, c.Revision})
		if err := publish(root, "claim-revision-"+Hash(identity).Hex+".json", raw); err != nil {
			return Digest{}, Digest{}, err
		}
	}
	if err := writeObject(root, d, b); err != nil {
		return Digest{}, Digest{}, err
	}
	if err := writeObject(root, ad, ab); err != nil {
		return d, Digest{}, err
	}
	return d, ad, nil
}
func artifactBindingName(id string) string {
	identity, _ := json.Marshal([]string{id})
	return "artifact-" + Hash(identity).Hex + ".json"
}

// Every Save publishes an artifact binding before its source-revision binding
// and record. Thus even orphaned bindings from interrupted saves retain coverage.
// Unfinished temporary publications fail closed instead of hiding retained bytes.
func persistedArtifacts(root string) ([]SourceArtifact, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var artifacts []SourceArtifact
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".evidence-") {
			return nil, errors.New("unfinished evidence publication requires local review")
		}
		if !strings.HasPrefix(name, "artifact-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return nil, errors.New("invalid persisted artifact binding")
		}
		raw, err := readLimited(filepath.Join(root, name))
		if err != nil {
			return nil, err
		}
		var a SourceArtifact
		if err := strictDecode(raw, &a); err != nil {
			return nil, err
		}
		if artifactBindingName(a.ArtifactID) != name {
			return nil, errors.New("artifact binding identity mismatch")
		}
		if err := validateArtifact(a); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, a)
	}
	return artifacts, nil
}

func writeObject(root string, d Digest, b []byte) error {
	return publish(root, "sha256-"+d.Hex+".json", b)
}
func publish(root, name string, b []byte) error {
	if err := os.MkdirAll(root, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(root, ".evidence-")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	path := filepath.Join(root, name)
	if err = os.Link(temp, path); err != nil {
		info, statErr := os.Lstat(path)
		if statErr != nil || !info.Mode().IsRegular() {
			return errors.New("object publication refused")
		}
		old, readErr := readLimited(path)
		if readErr != nil || !bytes.Equal(old, b) {
			return errors.New("existing object differs")
		}
	}
	return nil
}
func readLimited(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil || len(b) > maxBytes {
		return nil, errors.New("object read failed or too large")
	}
	return b, nil
}
func Read(root string, d Digest, parents ...[]byte) (Record, error) {
	if !validDigest(d) {
		return Record{}, errors.New("invalid object identity")
	}
	path := filepath.Join(root, "sha256-"+d.Hex+".json")
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return Record{}, errors.New("object is not a regular file")
	}
	b, err := readLimited(path)
	if err != nil {
		return Record{}, err
	}
	return Decode(b, d, parents...)
}
func retentionAllowed(r Record, now string, parents ...[]byte) error {
	t, ok := stamp(now)
	if !ok {
		return errors.New("invalid policy check time")
	}
	if err := Validate(r, parents...); err != nil {
		return err
	}
	for _, a := range r.Artifacts {
		if a.Rights.RetainUntil != "" {
			until, _ := stamp(a.Rights.RetainUntil)
			if !t.Before(until) {
				return errors.New("retention deadline passed")
			}
		}
	}
	return nil
}

// Export is all-or-nothing; no silent redaction or inherited derivative proof.
// Explicit approval binds the exact record digest and destination. This is a
// policy decision, not a legal entitlement. It does not send data anywhere.
type ExportApproval struct {
	Record      Digest
	Destination string
	Approved    bool
}

func Export(r Record, approval ExportApproval, now string, parents ...[]byte) ([]byte, error) {
	if err := retentionAllowed(r, now, parents...); err != nil {
		return nil, err
	}
	b, d, err := Encode(r, parents...)
	if err != nil {
		return nil, err
	}
	if !approval.Approved || approval.Record != d || approval.Destination == "" {
		return nil, errors.New("exact record export approval required")
	}
	// Scan metadata separately: byte payloads are base64 in the JSON encoding.
	if !safeClass(Classify(b, "public")) {
		return nil, errors.New("metadata quarantined")
	}
	for _, a := range r.Artifacts {
		if !safeClass(a.Sensitivity) || a.Rights.Export != "allow" || !safeClass(Classify(a.Body, a.Sensitivity)) {
			return nil, errors.New("source export denied")
		}
		for _, l := range a.Locations {
			if !safeClass(l.Sensitivity) || !safeClass(Classify(l.Excerpt, l.Sensitivity)) {
				return nil, errors.New("location export denied")
			}
		}
	}
	for _, c := range r.Claims {
		if !safeClass(c.Sensitivity) || c.Extraction.Context == "quarantined" {
			return nil, errors.New("claim export denied")
		}
	}
	return b, nil
}

// Revalidation uses a caller-recorded evaluation time; never the ambient clock.
// Changing it produces a new input root and assessment, preserving historical bytes.
func Revalidate(r Record, at time.Time) Record {
	r.EvaluationTime = at.UTC().Format(time.RFC3339Nano)
	return r
}
