package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/corpus"
)

func TestCorpusOversizedCLIRefusal(t *testing.T) {
	root := filepath.Join(t.TempDir(), "store")
	s := corpus.Store{Root: root}
	var policy, candidate corpus.Object
	wire := wireObjects(t)
	json.Unmarshal(wire[0], &policy)
	json.Unmarshal(wire[2], &candidate)
	p, e := s.Apply(corpus.Request{Operation: "policy", Action: "policy register", Object: policy})
	if e != nil {
		t.Fatal(e)
	}
	a, e := s.Apply(corpus.Request{Operation: "candidate", ExpectedHead: &p.Head, Action: "candidate add", Object: candidate})
	if e != nil {
		t.Fatal(e)
	}
	q := corpus.Request{Operation: "oversized-review", ExpectedHead: &a.Head, Action: "candidate add", Object: candidate}
	raw, _ := json.Marshal(q)
	raw = append(raw, bytes.Repeat([]byte(" "), corpus.MaxObjectBytes+1-len(raw))...)
	input := filepath.Join(t.TempDir(), "oversized.json")
	if e = os.WriteFile(input, raw, 0600); e != nil {
		t.Fatal(e)
	}
	out, code := cli(t, "corpus", "candidate", "add", "--store", root, "--input", input, "--attempt-id", "oversized-review")
	if code != 2 || !bytes.Contains(out, []byte("OBJECT_LIMIT")) {
		t.Fatalf("exit=%d %s", code, out)
	}
	var receipt corpus.Result
	if e = json.Unmarshal(bytes.SplitN(out, []byte("\n"), 2)[0], &receipt); e != nil || receipt.Sequence == 0 {
		t.Fatalf("oversized refusal omitted from audit ledger: %s", out)
	}
	head, e := s.Head()
	if e != nil || *head == a.Head || *head != receipt.Head {
		t.Fatal("audit head not advanced", e)
	}
}

func TestCorpusInputLimitParity(t *testing.T) {
	for _, size := range []int{corpus.MaxObjectBytes, corpus.MaxObjectBytes + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			setup := func() (corpus.Store, corpus.Object, corpus.Object, corpus.Result, corpus.Result) {
				s := corpus.Store{Root: filepath.Join(t.TempDir(), "store")}
				var p, c corpus.Object
				w := wireObjects(t)
				json.Unmarshal(w[0], &p)
				json.Unmarshal(w[2], &c)
				pr, e := s.Apply(corpus.Request{Operation: "policy", Action: "policy register", Object: p})
				if e != nil {
					t.Fatal(e)
				}
				cr, e := s.Apply(corpus.Request{Operation: "candidate", ExpectedHead: &pr.Head, Action: "candidate add", Object: c})
				if e != nil {
					t.Fatal(e)
				}
				return s, p, c, pr, cr
			}
			s, p, c, pr, cr := setup()
			library, _, _, _, _ := setup()
			raw := append([]byte("{invalid [private] oversized-private-marker"), bytes.Repeat([]byte(" "), size-len("{invalid [private] oversized-private-marker"))...)
			raw = raw[:size]
			input := filepath.Join(t.TempDir(), "input.json")
			if e := os.WriteFile(input, raw, 0600); e != nil {
				t.Fatal(e)
			}
			expected, err := library.ApplyBytes(raw, "oversized-review")
			if err == nil {
				t.Fatal("accepted malformed input")
			}
			reason := err.Error()
			if size > corpus.MaxObjectBytes && reason != "OBJECT_LIMIT" {
				t.Fatal(reason)
			}
			if size == corpus.MaxObjectBytes && reason == "OBJECT_LIMIT" {
				t.Fatal("limit inclusive boundary rejected")
			}
			call := func() corpus.Result {
				out, code := cli(t, "corpus", "candidate", "add", "--store", s.Root, "--input", input, "--attempt-id", "oversized-review")
				var r corpus.Result
				if code != 2 || !bytes.Contains(out, []byte(reason)) || json.Unmarshal(bytes.SplitN(out, []byte("\n"), 2)[0], &r) != nil {
					t.Errorf("exit=%d %s", code, out)
				}
				return r
			}
			r := call()
			if r != expected || call() != r {
				t.Fatal("CLI/library or retry mismatch", r, expected)
			}
			var wg sync.WaitGroup
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if call() != r {
						t.Error("concurrent retry differs")
					}
				}()
			}
			wg.Wait()
			head, e := s.Head()
			if e != nil || *head != r.Head || r.Sequence != 3 {
				t.Fatal("audit state", head, e)
			}
			inspected, e := s.Inspect(cr.Object)
			if e != nil || len(inspected.Revisions) != 1 {
				t.Fatal("domain changed", e)
			}
			if e = filepath.WalkDir(s.Root, func(path string, d os.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if !d.IsDir() {
					b, e := os.ReadFile(path)
					if e != nil {
						return e
					}
					for _, marker := range []string{"oversized-private-marker", "[private]", corpus.Hash(raw).Hex} {
						if bytes.Contains(b, []byte(marker)) {
							t.Errorf("payload retained: %s", path)
						}
					}
				}
				return nil
			}); e != nil {
				t.Fatal(e)
			}
			assignment := p
			assignment.Kind = "assignment"
			assignment.ID = "assignment"
			assignment.Policy = nil
			assignment.PolicyDigest = pr.Object
			assignment.Assignments = &[]corpus.Assignment{{Candidate: c.ID, Split: "DEVELOPMENT", Rationale: "synthetic test"}}
			ar, e := s.Apply(corpus.Request{Operation: "assignment", ExpectedHead: head, Action: "assign-split", Object: assignment})
			if e != nil {
				t.Fatal(e)
			}
			manifest := p
			manifest.Schema = corpus.ManifestSchema
			manifest.Kind = "manifest"
			manifest.ID = "manifest"
			manifest.Policy = nil
			manifest.PolicyDigest = pr.Object
			manifest.Manifest = &corpus.Manifest{Policy: corpus.Ref{ID: p.ID, Revision: 1, Digest: pr.Object}, Assignment: corpus.Ref{ID: assignment.ID, Revision: 1, Digest: ar.Object}, CheckTime: p.ReportedAt, Build: corpus.Hash([]byte("synthetic")), SourceCommit: "synthetic"}
			manifest, e = s.PrepareFreeze(manifest)
			if e != nil {
				t.Fatal(e)
			}
			found := false
			for _, n := range manifest.Manifest.Summary.Counts {
				if n.Name == "failed_attempts" && n.Count == 1 {
					found = true
				}
			}
			if !found {
				t.Fatal("freeze omitted refusal")
			}
			fr, e := s.Apply(corpus.Request{Operation: "freeze", ExpectedHead: &ar.Head, Action: "freeze", Object: manifest})
			if e != nil {
				t.Fatal(e)
			}
			if _, e = s.Verify(fr.Object, &fr.Head, p.ReportedAt, true); e != nil {
				t.Fatal(e)
			}
			if e = os.Remove(filepath.Join(s.Root, "events", fmt.Sprintf("%05d.json", r.Sequence))); e != nil {
				t.Fatal(e)
			}
			if _, e = s.Verify(fr.Object, &fr.Head, p.ReportedAt, true); e == nil {
				t.Fatal("deleted attempt not detected")
			}
		})
	}
}
