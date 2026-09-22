package corpus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
)

// Store paths are configuration, not persisted content. Evidence data stays in
// one separately governed store. Parent bytes are transient and never published.
type Store struct {
	Root            string
	EvidenceRoot    string
	EvidenceStoreID string
	Parents         [][]byte
	// fault is a package-private deterministic crash-injection seam used by tests.
	fault func(string) error
}
type state struct {
	events         []Event
	heads          map[string]Ref
	objects        map[Digest]Object
	order          []Digest
	head           *Digest
	bytes          int
	storeFiles     int
	storeBytes     int64
	orphans        []string
	pending        *Request
	freezeProducer *Identity
}

func emptyState() *state         { return &state{heads: map[string]Ref{}, objects: map[Digest]Object{}} }
func objectPath(d Digest) string { return "objects/sha256/" + d.Hex + ".json" }
func bindingPath(r Ref) string {
	b, _ := json.Marshal(struct {
		ID       string `json:"id"`
		Revision uint32 `json:"revision"`
	}{r.ID, r.Revision})
	return "bindings/" + Hash(b).Hex + ".json"
}
func eventPath(n uint32) string { return fmt.Sprintf("events/%05d.json", n) }
func regularRead(path string, limit int) ([]byte, error) {
	f, e := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if e != nil {
		return nil, fail("IO_READ", 4)
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() {
		return nil, fail("NONREGULAR_FILE", 2)
	}
	if info.Size() > int64(limit) {
		return nil, fail("OBJECT_LIMIT", 2)
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if e != nil {
		return nil, fail("IO_READ", 4)
	}
	if len(b) > limit {
		return nil, fail("OBJECT_LIMIT", 2)
	}
	return b, nil
}

// secureDir checks every ancestor; concurrent hostile root replacement is outside
// the documented trusted-administrator/cooperating-process boundary.
func secureDir(path string, create bool) error {
	abs, e := filepath.Abs(path)
	if e != nil {
		return fail("INVALID_ROOT", 4)
	}
	cur := "/"
	for _, part := range strings.Split(strings.TrimPrefix(abs, "/"), "/") {
		if part == "" {
			continue
		}
		cur = filepath.Join(cur, part)
		info, e := os.Lstat(cur)
		if os.IsNotExist(e) && create {
			if e = os.Mkdir(cur, 0700); e != nil && !os.IsExist(e) {
				return fail("IO_MKDIR", 4)
			}
			if e = syncDir(filepath.Dir(cur)); e != nil {
				return e
			}
			info, e = os.Lstat(cur)
		}
		if e != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fail("UNTRUSTED_PATH", 2)
		}
	}
	return nil
}
func syncDir(path string) error {
	f, e := os.Open(path)
	if e != nil {
		return fail("IO_DIRECTORY", 4)
	}
	defer f.Close()
	if f.Sync() != nil {
		return fail("IO_DIRECTORY_SYNC", 4)
	}
	return nil
}
func (s Store) lock(create bool) (*os.File, error) {
	if runtime.GOOS != "linux" {
		return nil, fail("UNSUPPORTED_FILESYSTEM", 2)
	}
	if e := secureDir(s.Root, create); e != nil {
		return nil, e
	}
	info, e := os.Stat(s.Root)
	if e != nil || info.Mode().Perm()&0077 != 0 {
		return nil, fail("ROOT_PERMISSIONS", 3)
	}
	var fs syscall.Statfs_t
	if syscall.Statfs(s.Root, &fs) != nil {
		return nil, fail("IO_STATFS", 4)
	}
	switch uint64(fs.Type) {
	case 0xef53, 0x58465342, 0x9123683e, 0x01021994, 0x794c7630:
	default:
		return nil, fail("UNSUPPORTED_FILESYSTEM", 2)
	}
	for _, p := range []string{"objects", "objects/sha256", "events", "bindings", "index"} {
		if e := secureDir(filepath.Join(s.Root, p), create); e != nil {
			return nil, e
		}
	}
	f, e := os.OpenFile(filepath.Join(s.Root, ".writer.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0600)
	if e != nil {
		return nil, fail("LOCK_IO", 4)
	}
	info, e = f.Stat()
	if e != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, fail("INVALID_LOCK", 2)
	}
	if syscall.Flock(int(f.Fd()), syscall.LOCK_EX) != nil {
		f.Close()
		return nil, fail("LOCK_FAILED", 4)
	}
	return f, nil
}
func (s Store) publish(name string, b []byte) error {
	path := filepath.Join(s.Root, name)
	if _, e := os.Lstat(path); e == nil {
		old, e := regularRead(path, MaxObjectBytes)
		if e != nil {
			return e
		}
		if !bytes.Equal(old, b) {
			return fail("IMMUTABLE_CONFLICT", 2)
		}
		return syncDir(filepath.Dir(path))
	} else if !os.IsNotExist(e) {
		return fail("IO_STAT", 4)
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".pending-")
	if e != nil {
		return fail("IO_TEMP", 4)
	}
	temp := f.Name()
	// Leave failed temporary publications visible to recovery. Never hide a failed
	// durability attempt with best-effort cleanup or infer that it committed.
	if _, e = f.Write(b); e != nil {
		f.Close()
		return fail("IO_WRITE", 4)
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return fail("IO_SYNC", 4)
	}
	if f.Close() != nil {
		return fail("IO_CLOSE", 4)
	}
	if e = os.Link(temp, path); e != nil {
		return fail("IO_PUBLICATION", 4)
	}
	if e = syncDir(filepath.Dir(path)); e != nil {
		return e
	}
	if e = os.Remove(temp); e != nil {
		return fail("IO_TEMP_REMOVE", 4)
	}
	return syncDir(filepath.Dir(path))
}
func (s Store) scan() (*state, error) {
	st := emptyState()
	files := 0
	total := int64(0)
	e := filepath.WalkDir(s.Root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return fail("IO_WALK", 4)
		}
		rel, _ := filepath.Rel(s.Root, path)
		if rel == "index" {
			if !d.IsDir() || d.Type()&os.ModeSymlink != 0 {
				return fail("UNTRUSTED_PATH", 2)
			}
			return filepath.SkipDir
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fail("SYMLINK_REFUSED", 2)
		}
		if d.IsDir() {
			if !one(rel, ".", "objects", "objects/sha256", "events", "bindings") {
				return fail("UNEXPECTED_DIRECTORY", 2)
			}
			return nil
		}
		info, e := d.Info()
		if e != nil || !info.Mode().IsRegular() {
			return fail("NONREGULAR_FILE", 2)
		}
		files++
		total += info.Size()
		if files > MaxStoreFiles || total > MaxStoreBytes {
			return fail("STORE_LIMIT", 2)
		}
		if strings.HasPrefix(d.Name(), ".pending-") {
			return fail("UNFINISHED_PUBLICATION", 2)
		}
		if rel != ".writer.lock" && rel != "pending-refusal.json" && !strings.HasPrefix(rel, "objects/sha256/") && !strings.HasPrefix(rel, "events/") && !strings.HasPrefix(rel, "bindings/") {
			return fail("UNEXPECTED_FILE", 2)
		}
		return nil
	})
	if e != nil {
		return nil, e
	}
	st.storeFiles = files
	st.storeBytes = total
	entries, e := os.ReadDir(filepath.Join(s.Root, "events"))
	if e != nil {
		return nil, fail("IO_EVENTS", 4)
	}
	ops := map[string]bool{}
	for i, en := range entries {
		n := uint32(i + 1)
		if n > MaxOrdinal || en.Name() != filepath.Base(eventPath(n)) {
			return nil, fail("MISSING_EVENT", 2)
		}
		b, e := regularRead(filepath.Join(s.Root, eventPath(n)), MaxObjectBytes)
		if e != nil {
			return nil, e
		}
		var ev Event
		if e = strict(b, &ev); e != nil {
			return nil, e
		}
		if ev.Schema != Schema || ev.Encoding != Encoding || ev.Sequence != n || !same(ev.Previous, st.head) || ops[ev.Operation] || !id(ev.Operation) || !ref(ev.Output) || !validDigest(ev.Request) || !identity(ev.Producer) || !stamp(ev.ReportedAt) {
			return nil, fail("EVENT_CHAIN_MISMATCH", 2)
		}
		ops[ev.Operation] = true
		raw, e := regularRead(filepath.Join(s.Root, objectPath(ev.Output.Digest)), MaxObjectBytes)
		if e != nil {
			return nil, fail("COMMITTED_CLOSURE_MISSING", 2)
		}
		o, e := Decode(raw, ev.Output.Digest)
		if e != nil {
			return nil, e
		}
		if o.ID != ev.Output.ID || o.Revision != ev.Output.Revision {
			return nil, fail("BINDING_MISMATCH", 2)
		}
		bind, e := regularRead(filepath.Join(s.Root, bindingPath(ev.Output)), MaxObjectBytes)
		if e != nil {
			return nil, fail("MISSING_BINDING", 2)
		}
		want, _ := json.Marshal(ev.Output)
		if !bytes.Equal(bind, want) {
			return nil, fail("BINDING_MISMATCH", 2)
		}
		if e = st.lineage(o); e != nil {
			return nil, e
		}
		if !actionKind(ev.Action, o) {
			return nil, fail("EVENT_ACTION_MISMATCH", 2)
		}
		requestBytes, _ := json.Marshal(Request{ev.Operation, ev.Previous, ev.Action, o})
		if Hash(requestBytes) != ev.Request || ev.Producer != o.Producer || ev.ReportedAt != o.ReportedAt {
			return nil, fail("EVENT_REQUEST_MISMATCH", 2)
		}
		if e := references(st, o); e != nil {
			return nil, e
		}
		if o.Manifest != nil {
			if e := s.checkManifestSnapshot(st, o); e != nil {
				return nil, e
			}
		}
		st.objects[ev.Output.Digest] = o
		st.order = append(st.order, ev.Output.Digest)
		st.heads[o.ID] = ev.Output
		st.bytes += len(raw) + len(bind) + len(b)
		if len(st.objects) > MaxClosureObjects || st.bytes > MaxClosureBytes {
			return nil, fail("CLOSURE_LIMIT", 2)
		}
		st.events = append(st.events, ev)
		d := Hash(b)
		st.head = &d
	}
	// Every published object must be recognizable, even if not committed.
	expected := map[string]bool{}
	for _, ev := range st.events {
		expected[objectPath(ev.Output.Digest)] = true
		expected[bindingPath(ev.Output)] = true
	}
	for _, dir := range []string{"objects/sha256", "bindings"} {
		entries, e := os.ReadDir(filepath.Join(s.Root, dir))
		if e != nil {
			return nil, fail("IO_DIRECTORY", 4)
		}
		for _, en := range entries {
			rel := dir + "/" + en.Name()
			raw, e := regularRead(filepath.Join(s.Root, rel), MaxObjectBytes)
			if e != nil {
				return nil, e
			}
			if dir == "objects/sha256" {
				d := Hash(raw)
				if rel != objectPath(d) {
					return nil, fail("CORRUPT_OBJECT", 2)
				}
				if _, e := Decode(raw, d); e != nil {
					return nil, e
				}
			} else {
				var r Ref
				if e := strict(raw, &r); e != nil {
					return nil, e
				}
				if !ref(r) || rel != bindingPath(r) {
					return nil, fail("BINDING_MISMATCH", 2)
				}
			}
			if !expected[rel] {
				st.orphans = append(st.orphans, rel)
			}
		}
	}
	if _, err := os.Lstat(filepath.Join(s.Root, "pending-refusal.json")); err == nil {
		b, err := regularRead(filepath.Join(s.Root, "pending-refusal.json"), MaxObjectBytes)
		if err != nil {
			return nil, err
		}
		var q Request
		if strict(b, &q) != nil || !id(q.Operation) || evidenceUnsafeOperation(q.Operation) || q.Action != "attempt" || q.Object.AttemptReason == nil {
			return nil, fail("INVALID_PENDING_REFUSAL", 2)
		}
		if _, _, err := Encode(q.Object); err != nil {
			return nil, err
		}
		committed := false
		for _, ev := range st.events {
			if ev.Operation == q.Operation {
				qb, _ := json.Marshal(q)
				if Hash(qb) != ev.Request {
					return nil, fail("INVALID_PENDING_REFUSAL", 2)
				}
				committed = true
			}
		}
		if !committed && !same(q.ExpectedHead, st.head) {
			return nil, fail("INVALID_PENDING_REFUSAL", 2)
		}
		st.pending = &q
		st.orphans = append(st.orphans, "pending-refusal.json")
	} else if !os.IsNotExist(err) {
		return nil, fail("IO_READ", 4)
	}
	slices.Sort(st.orphans)
	return st, nil
}
func (st *state) lineage(o Object) error {
	old, exists := st.heads[o.ID]
	if !exists {
		if o.Revision != 1 || o.Parent != nil {
			return fail("DANGLING_PARENT", 2)
		}
	} else {
		if o.Revision != old.Revision+1 || o.Parent == nil || *o.Parent != old.Digest {
			return fail("REVISION_CONFLICT", 3)
		}
		if st.objects[old.Digest].Kind != o.Kind {
			return fail("IDENTITY_KIND_CONFLICT", 3)
		}
	}
	return nil
}
func actionKind(a string, o Object) bool {
	switch a {
	case "policy register":
		return o.Kind == "policy"
	case "candidate add":
		return o.Kind == "candidate" && o.Revision == 1
	case "candidate revise", "candidate exclude", "assign-family":
		return o.Kind == "candidate" && o.Revision > 1
	case "cluster":
		return o.Kind == "cluster"
	case "assign-split":
		return o.Kind == "assignment"
	case "exposure append":
		return o.Kind == "exposure"
	case "freeze":
		return o.Kind == "manifest"
	case "attempt":
		return o.Kind == "attempt"
	case "retire":
		return o.Kind == "retirement"
	case "withdraw":
		return o.Kind == "withdrawal"
	}
	return false
}
func (st *state) resolve(r Ref) (Object, error) {
	o, ok := st.objects[r.Digest]
	if !ok || o.ID != r.ID || o.Revision != r.Revision {
		return Object{}, fail("DANGLING_REFERENCE", 2)
	}
	return o, nil
}
func (s Store) Head() (*Digest, error) {
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
		return st.head, fail("ORPHAN_PUBLICATION", 2)
	}
	return st.head, nil
}

type Recovery struct {
	Head    *Digest  `json:"head"`
	Orphans []string `json:"orphan_unregistered"`
}

func (s Store) Recover() (Recovery, error) {
	f, e := s.lock(false)
	if e != nil {
		return Recovery{}, e
	}
	defer f.Close()
	st, e := s.scan()
	if e != nil {
		return Recovery{}, e
	}
	return Recovery{st.head, st.orphans}, nil
}
func (s Store) Apply(q Request) (Result, error) { return s.apply(q, nil) }

// RejectOversizedInput accounts for a submission refused by a bounded file
// reader before decoding. Only the safe administrative identity crosses this
// boundary; no oversized bytes or payload digest are needed or retained.
func (s Store) RejectOversizedInput(operation string) (Result, error) {
	return s.apply(Request{Operation: operation}, fail("OBJECT_LIMIT", 2))
}

// ApplyBytes retains only a safe administrative refusal when wire decoding fails.
// operation is an optional separately supplied opaque ID for an unreadable
// envelope. DecodeRequest itself remains a pure decoder, not a submission API.
func (s Store) ApplyBytes(b []byte, operation string) (Result, error) {
	q, e := DecodeRequest(b)
	if e == nil {
		if operation != "" && operation != q.Operation {
			return s.apply(Request{Operation: operation}, fail("ATTEMPT_ID_MISMATCH", 2))
		}
		return s.Apply(q)
	}
	if operation != "" {
		q.Operation = operation
	}
	return s.apply(Request{Operation: q.Operation}, e)
}

// ApplyCommand adds a command binding to the same submission path used by ApplyBytes.
func (s Store) ApplyCommand(action string, b []byte, operation string) (Result, error) {
	q, e := DecodeRequest(b)
	if e != nil {
		return s.ApplyBytes(b, operation)
	}
	if operation != "" && operation != q.Operation {
		return s.ApplyBytes(b, operation)
	}
	if q.Action != action {
		return s.apply(q, fail("ACTION_MISMATCH", 2))
	}
	return s.Apply(q)
}
func (s Store) apply(q Request, inputErr error) (Result, error) {
	// The operation ID is an independently supplied safe administrative identity.
	// Forbidden payloads, including their digests, never enter refusal receipts.
	f, e := s.lock(true)
	if e != nil {
		return Result{}, e
	}
	defer f.Close()
	st, e := s.scan()
	if e != nil {
		return Result{}, e
	}
	if !id(q.Operation) || evidenceUnsafeOperation(q.Operation) {
		// An unreadable/unsafe envelope still gets an opaque identity derived
		// only from the safe local prefix, never from rejected input bytes.
		seed, _ := json.Marshal(struct {
			Head     *Digest
			Sequence int
		}{st.head, len(st.events) + 1})
		q = Request{Operation: "anonymous-" + Hash(seed).Hex}
		if inputErr == nil {
			inputErr = fail("UNSAFE_ATTEMPT_ID", 3)
		}
	}
	if st.pending != nil && st.pending.Object.AttemptReason != nil && *st.pending.Object.AttemptReason == "OPERATION_CONFLICT" && st.pending.Operation == "conflict-"+Hash([]byte(q.Operation)).Hex {
		q = *st.pending
	}
	if st.pending != nil && st.pending.Operation != q.Operation {
		return Result{}, fail("PENDING_REFUSAL", 2)
	}
resolveOperation:
	for _, ev := range st.events {
		if ev.Operation != q.Operation {
			continue
		}
		prior := st.objects[ev.Output.Digest]
		eb, _ := json.Marshal(ev)
		if prior.AttemptReason != nil {
			if st.pending != nil {
				if e = s.clearPending(); e != nil {
					return Result{}, e
				}
			}
			return Result{ev.Output.Digest, Hash(eb), ev.Sequence, *prior.AttemptReason}, fail(*prior.AttemptReason, attemptExit(prior))
		}
		b, d, err := Encode(q.Object)
		if err == nil {
			q.Object, err = Decode(b, d)
		}
		qb, _ := json.Marshal(q)
		if err != nil || inputErr != nil || ev.Request != Hash(qb) {
			// Reusing a successful operation for different bytes preserves that
			// success and records one idempotent protocol-refusal identity. No
			// rejected content or content digest participates in this identity.
			q = Request{Operation: "conflict-" + Hash([]byte(q.Operation)).Hex, Object: Object{PolicyDigest: prior.PolicyDigest}}
			inputErr = fail("OPERATION_CONFLICT", 3)
			goto resolveOperation
		}
		return Result{ev.Output.Digest, Hash(eb), ev.Sequence, ""}, nil
	}
	pendingRetry := st.pending != nil
	if pendingRetry {
		q = *st.pending
	}
	b, d, validationErr := Encode(q.Object)
	if inputErr != nil && !pendingRetry {
		validationErr = inputErr
	}
	o := q.Object
	if validationErr == nil {
		o, validationErr = Decode(b, d)
	}
	q.Object = o
	if validationErr == nil && !pendingRetry && strings.HasPrefix(o.ID, "attempt-") {
		validationErr = fail("RESERVED_ATTEMPT_ID", 3)
	}
	if validationErr == nil && !actionKind(q.Action, o) {
		validationErr = fail("ACTION_KIND_MISMATCH", 2)
	}
	if validationErr == nil && !same(q.ExpectedHead, st.head) {
		validationErr = fail("STALE_HEAD", 4)
	}
	if validationErr == nil {
		validationErr = st.lineage(o)
	}
	if validationErr == nil {
		validationErr = s.transition(st, o, q.Action)
	}
	if validationErr == nil && !pendingRetry {
		qb, _ := json.Marshal(q)
		r := Ref{o.ID, o.Revision, d}
		rb, _ := json.Marshal(r)
		ev := Event{Schema, Encoding, uint32(len(st.events) + 1), st.head, q.Operation, Hash(qb), q.Action, r, o.Producer, o.ReportedAt}
		eb, _ := json.Marshal(ev)
		validationErr = publicationLimits(st, len(b)+len(rb)+len(eb))
	}
	refusal := ""
	refusalExit := 0
	if pendingRetry {
		refusal = *q.Object.AttemptReason
		refusalExit = attemptExit(q.Object)
	}
	if validationErr != nil {
		refusal, refusalExit = validationErr.Error(), ExitCode(validationErr)
		// Bind only a registered policy and its safe administrative timestamp.
		// Unknown-policy submissions use the vocabulary anchor, not caller bytes.
		p, ok := st.objects[q.Object.PolicyDigest]
		q = opaqueAttempt(q, refusal)
		q.Object.PolicyDigest = VocabularyDigest()
		q.Object.ReportedAt = "1970-01-01T00:00:00Z"
		if ok && p.Policy != nil {
			q.Object.PolicyDigest = hashObject(p)
			q.Object.ReportedAt = p.ReportedAt
		}
		q.Object.Reasons = append(q.Object.Reasons, fmt.Sprintf("EXIT_%d", refusalExit))
		q.ExpectedHead = st.head
		b, d, e = Encode(q.Object)
		if e != nil {
			return Result{}, e
		}
		o, e = Decode(b, d)
		if e != nil {
			return Result{}, e
		}
		q.Object = o
		if e = st.lineage(o); e != nil {
			return Result{}, e
		}
	}
	qb, _ := json.Marshal(q)
	qd := Hash(qb)

	r := Ref{o.ID, o.Revision, d}
	rb, _ := json.Marshal(r)
	n := uint32(len(st.events) + 1)
	if n > MaxOrdinal {
		return Result{}, fail("EVENT_LIMIT", 2)
	}
	ev := Event{Schema, Encoding, n, st.head, q.Operation, qd, q.Action, r, o.Producer, o.ReportedAt}
	eb, _ := json.Marshal(ev)
	// An explicit identical retry may complete its own orphan publications only
	// after all checks rerun. Nothing is recovered merely by opening the store.
	for _, p := range st.orphans {
		if p != objectPath(d) && p != bindingPath(r) && !(refusal != "" && p == "pending-refusal.json") {
			return Result{}, fail("ORPHAN_PUBLICATION", 2)
		}
	}
	if refusal != "" {
		// A safe pending refusal makes failed publication visible to recovery;
		// only an explicit retry of this opaque operation may complete it.
		if st.pending == nil && (st.storeFiles+2 > MaxStoreFiles || st.storeBytes+int64(2*len(qb)) > MaxStoreBytes) {
			return Result{}, fail("ATTEMPT_DURABILITY_FAILED", 4)
		}
		if e = s.publish("pending-refusal.json", qb); e != nil {
			return Result{}, e
		}
		if st.pending == nil {
			st.storeFiles++
			st.storeBytes += int64(len(qb))
		}
		if s.fault != nil {
			if e = s.fault("after_refusal_intent"); e != nil {
				return Result{}, e
			}
		}
	}

	if e = publicationLimits(st, len(b)+len(rb)+len(eb)); e != nil {
		return Result{}, e
	}
	if e = s.publish(objectPath(d), b); e != nil {
		return Result{}, e
	}
	if s.fault != nil {
		if e = s.fault("after_object"); e != nil {
			return Result{}, e
		}
	}
	if e = s.publish(bindingPath(r), rb); e != nil {
		return Result{}, e
	}
	if s.fault != nil {
		if e = s.fault("before_receipt"); e != nil {
			return Result{}, e
		}
	}
	if e = s.publish(eventPath(n), eb); e != nil {
		return Result{}, e
	}
	if refusal != "" {
		if s.fault != nil {
			if e = s.fault("after_refusal_receipt"); e != nil {
				return Result{}, e
			}
		}
		if e = s.clearPending(); e != nil {
			return Result{}, e
		}
	}
	result := Result{d, Hash(eb), n, refusal}
	if refusal != "" {
		return result, fail(refusal, refusalExit)
	}
	return result, nil
}

func opaqueAttempt(q Request, reason string) Request {
	q.Object = Object{Schema: Schema, Encoding: Encoding, Kind: "attempt", ID: "attempt-" + Hash([]byte(q.Operation)).Hex, Revision: 1, Producer: Identity{"local-intake", "tool", "register", "failuremesh-corpus", "1", "", "opaque-attempt"}, PolicyDigest: q.Object.PolicyDigest, ReportedAt: q.Object.ReportedAt, Sensitivity: "synthetic", Reasons: []string{reason}, Rationale: "Rejected submission; payload not retained", AttemptReason: &reason}
	q.Action = "attempt"
	return q
}

// ReadInput applies the same bounded regular-file/no-symlink boundary to explicit
// CLI metadata and transient parent files. It performs no retrieval.
func ReadInput(path string) ([]byte, error) {
	if e := secureDir(filepath.Dir(path), false); e != nil {
		return nil, e
	}
	return regularRead(path, MaxObjectBytes)
}

func publicationLimits(st *state, additionalBytes int) error {
	if len(st.objects)+1 > MaxClosureObjects || st.bytes+additionalBytes > MaxClosureBytes {
		return fail("CLOSURE_LIMIT", 2)
	}
	if st.storeFiles+4 > MaxStoreFiles || st.storeBytes+int64(2*additionalBytes) > MaxStoreBytes {
		return fail("STORE_LIMIT", 2)
	}
	return nil
}

// hashObject is internal deterministic encoding of an already validated object.
func hashObject(o Object) Digest { b, _ := json.Marshal(o); return Hash(b) }
func attemptExit(o Object) int {
	for _, code := range o.Reasons {
		switch code {
		case "EXIT_2":
			return 2
		case "EXIT_4":
			return 4
		}
	}
	return 3
}

func (s Store) clearPending() error {
	if e := os.Remove(filepath.Join(s.Root, "pending-refusal.json")); e != nil {
		return fail("IO_PENDING_REMOVE", 4)
	}
	return syncDir(s.Root)
}
