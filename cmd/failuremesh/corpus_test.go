package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/corpus"
)

func TestCorpusCLIHelper(t *testing.T) {
	if os.Getenv("FAILUREMESH_CORPUS_TEST_HELPER") != "1" {
		return
	}
	for i, s := range os.Args {
		if s == "--" {
			os.Args = append([]string{"failuremesh"}, os.Args[i+1:]...)
			main()
			os.Exit(0)
		}
	}
	t.Fatal("missing helper args")
}
func cli(t *testing.T, args ...string) ([]byte, int) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCorpusCLIHelper$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "FAILUREMESH_CORPUS_TEST_HELPER=1")
	b, e := cmd.CombinedOutput()
	if e == nil {
		return b, 0
	}
	if exit, ok := e.(*exec.ExitError); ok {
		return b, exit.ExitCode()
	}
	t.Fatal(e)
	return nil, -1
}
func wireObjects(t *testing.T) []json.RawMessage {
	t.Helper()
	b, e := os.ReadFile("../../internal/corpus/testdata/wire.golden.json")
	if e != nil {
		t.Fatal(e)
	}
	var v []json.RawMessage
	if json.Unmarshal(b, &v) != nil {
		t.Fatal("golden")
	}
	return v
}
func TestCorpusCLIEndToEnd(t *testing.T) {
	root := filepath.Join(t.TempDir(), "store")
	var head, manifest corpus.Digest
	for i, v := range wireObjects(t) {
		if i%2 == 0 {
			continue
		}
		var ev corpus.Event
		json.Unmarshal(v, &ev)
		var object corpus.Object
		json.Unmarshal(wireObjects(t)[i-1], &object)
		q := corpus.Request{Operation: ev.Operation, ExpectedHead: ev.Previous, Action: ev.Action, Object: object}
		raw, _ := json.Marshal(q)
		input := filepath.Join(t.TempDir(), "input.json")
		if os.WriteFile(input, raw, 0600) != nil {
			t.Fatal("input")
		}
		args := append([]string{"corpus"}, strings.Fields(ev.Action)...)
		args = append(args, "--store", root, "--input", input)
		if ev.Action == "freeze" {
			prepared, code := cli(t, append(args, "--prepare")...)
			if code != 0 {
				t.Fatalf("prepare code %d: %s", code, prepared)
			}
			if _, e := corpus.DecodeRequest(prepared); e != nil {
				t.Fatal("prepare output cannot be used as input", e)
			}
			os.WriteFile(input, prepared, 0600)
		}
		out, code := cli(t, args...)
		if code != 0 {
			t.Fatalf("%s code=%d: %s", ev.Action, code, out)
		}
		var result corpus.Result
		if json.Unmarshal(bytes.TrimSuffix(out, []byte("PASS\n")), &result) != nil {
			t.Fatalf("result %s", out)
		}
		head = result.Head
		if ev.Action == "freeze" {
			manifest = result.Object
		}
	}
	for _, action := range []string{"verify", "inspect"} {
		args := []string{"corpus", action, "--store", root, "--digest", manifest.Hex}
		if action == "verify" {
			args = append(args, "--head", head.Hex, "--at", "2026-09-17T12:00:00Z")
		}
		out, code := cli(t, args...)
		if code != 0 {
			t.Fatalf("%s code=%d %s", action, code, out)
		}
		if action == "verify" && (!bytes.Contains(out, []byte(`"integrity":"INTACT"`)) || !bytes.Contains(out, []byte(`"selection_target_status":"SHORTFALL"`)) || !bytes.Contains(out, []byte(`"current_use_status":"RETIRED"`))) {
			t.Fatalf("bad status %s", out)
		}
	}
	if out, code := cli(t, "corpus", "index", "rebuild", "--store", root); code != 0 {
		t.Fatal(code, string(out))
	}
}
func TestCorpusCLIExitContract(t *testing.T) {
	root := filepath.Join(t.TempDir(), "store")
	input := filepath.Join(t.TempDir(), "input.json")
	os.WriteFile(input, []byte(`{"unknown":1}`), 0600)
	for _, tt := range []struct {
		args []string
		code int
	}{{[]string{"corpus", "unknown", "--store", root}, 2}, {[]string{"corpus", "candidate", "add", "--store", root, "--input", input}, 2}, {[]string{"corpus", "candidate", "add", "--store", root, "--input", input + "missing"}, 4}} {
		if out, code := cli(t, tt.args...); code != tt.code {
			t.Fatalf("want %d got %d: %s", tt.code, code, out)
		}
	}
	raw := wireObjects(t)
	var o corpus.Object
	json.Unmarshal(raw[0], &o)
	head, e := (corpus.Store{Root: root}).Head()
	if e != nil {
		t.Fatal(e)
	}
	q := corpus.Request{Operation: "policy-cli", ExpectedHead: head, Action: "policy register", Object: o}
	b, _ := json.Marshal(q)
	os.WriteFile(input, b, 0600)
	if out, code := cli(t, "corpus", "policy", "register", "--store", root, "--input", input); code != 0 {
		t.Fatal(code, string(out))
	}
	q.Operation = "stale"
	b, _ = json.Marshal(q)
	os.WriteFile(input, b, 0600)
	if out, code := cli(t, "corpus", "policy", "register", "--store", root, "--input", input); code != 4 || !bytes.Contains(out, []byte("STALE_HEAD")) {
		t.Fatal(code, string(out))
	}
	// Unsupported family authority is a policy refusal (3), before mutation.
	json.Unmarshal(raw[2], &o)
	o.Candidate.Family.State = "EVIDENCE_SUPPORTED"
	q.Object = o
	q.Operation = "authority-refusal"
	q.Action = "candidate add"
	b, _ = json.Marshal(q)
	os.WriteFile(input, b, 0600)
	if out, code := cli(t, "corpus", "candidate", "add", "--store", root, "--input", input); code != 3 || !bytes.Contains(out, []byte("FAMILY_AUTHORITY_REFUSED")) {
		t.Fatal(code, string(out))
	}
}

func TestCorpusConcurrentProcesses(t *testing.T) {
	root := filepath.Join(t.TempDir(), "store")
	wire := wireObjects(t)
	var p corpus.Object
	json.Unmarshal(wire[0], &p)
	s := corpus.Store{Root: root}
	r, e := s.Apply(corpus.Request{Operation: "policy", Action: "policy register", Object: p})
	if e != nil {
		t.Fatal(e)
	}
	inputs := []string{}
	for _, name := range []string{"writer-a", "writer-b"} {
		var o corpus.Object
		json.Unmarshal(wire[2], &o)
		o.ID = name
		q := corpus.Request{Operation: name, ExpectedHead: &r.Head, Action: "candidate add", Object: o}
		b, _ := json.Marshal(q)
		path := filepath.Join(t.TempDir(), name+".json")
		if os.WriteFile(path, b, 0600) != nil {
			t.Fatal("input")
		}
		inputs = append(inputs, path)
	}
	type result struct {
		out  []byte
		code int
	}
	done := make(chan result, 2)
	for _, input := range inputs {
		go func(input string) {
			out, code := cli(t, "corpus", "candidate", "add", "--store", root, "--input", input)
			done <- result{out, code}
		}(input)
	}
	counts := map[int]int{}
	for i := 0; i < 2; i++ {
		v := <-done
		counts[v.code]++
		if v.code != 0 && !(v.code == 4 && bytes.Contains(v.out, []byte("STALE_HEAD"))) {
			t.Fatalf("unexpected concurrent result %d %s", v.code, v.out)
		}
	}
	if counts[0] != 1 || counts[4] != 1 {
		t.Fatalf("multiple writers committed %+v", counts)
	}
}

// Package submission and subprocess CLI must produce identical refusal bytes and
// receipts against the same local prefix, including malformed classifications.
func TestCorpusReviewBlockerParity(t *testing.T) {
	for _, kind := range []string{"classification", "decision actor", "late merge", "stale assignment", "cross policy", "malformed wire", "action mismatch"} {
		t.Run(kind, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "store")
			s := corpus.Store{Root: root}
			var head *corpus.Digest
			n := 0
			apply := func(action string, o corpus.Object) corpus.Result {
				t.Helper()
				n++
				r, e := s.Apply(corpus.Request{Operation: fmt.Sprintf("setup-%d", n), ExpectedHead: head, Action: action, Object: o})
				if e != nil {
					t.Fatal(e)
				}
				head = &r.Head
				return r
			}
			wire := wireObjects(t)
			var p, a corpus.Object
			json.Unmarshal(wire[0], &p)
			pr := apply("policy register", p)
			pref := corpus.Ref{ID: p.ID, Revision: 1, Digest: pr.Object}
			json.Unmarshal(wire[2], &a)
			a.PolicyDigest = pr.Object
			a.Candidate.Decision.Policy = pref
			if kind == "decision actor" {
				a.Candidate.Decision.Actor = "alternate-selector"
			}
			if kind != "classification" && kind != "malformed wire" {
				apply("candidate add", a)
			}
			negative := func(id string) {
				for _, category := range corpus.ExposureCategories() {
					o := p
					o.Kind = "exposure"
					o.Policy = nil
					o.PolicyDigest = pr.Object
					o.ID = fmt.Sprintf("exposure-%d", n)
					o.Exposure = &corpus.Exposure{Candidates: []string{id}, Round: p.Policy.Round, Category: category, State: "NO_WITH_BASIS", AvailableAccess: "NO_WITH_BASIS", Basis: "synthetic restrictions", Restrictions: "authored fixture", OccurredAt: corpus.Fact{State: "UNKNOWN", Reason: "not asserted", Basis: []string{}}, RecordedAt: p.ReportedAt, Artifacts: []corpus.Digest{}}
					apply("exposure append", o)
				}
			}
			q := corpus.Request{Operation: "attack", Action: "candidate add", Object: a}
			switch kind {
			case "classification":
				q.Object.Candidate.OriginClass = "invalid"
			case "decision actor", "late merge", "stale assignment", "cross policy":
				negative(a.ID)
				rows := []corpus.Assignment{{Candidate: a.ID, Split: "HOLDOUT", Rationale: "review"}}
				if kind == "late merge" {
					var b corpus.Object
					json.Unmarshal(wire[4], &b)
					br := apply("candidate add", b)
					o := p
					o.ID = "merge"
					o.Kind = "cluster"
					o.Policy = nil
					o.PolicyDigest = pr.Object
					// The original a identity is retained in the setup candidate receipt.
					aBytes, ad, e := corpus.Encode(a)
					_ = aBytes
					if e != nil {
						t.Fatal(e)
					}
					relations := []corpus.Relation{{ID: "link", From: corpus.Ref{ID: b.ID, Revision: 1, Digest: br.Object}, To: corpus.Ref{ID: a.ID, Revision: 1, Digest: ad}, Type: "SHARED_PROJECT_LINEAGE", Status: "ESTABLISHED", Basis: "synthetic", Rationale: "review"}}
					o.Relations = &relations
					apply("cluster", o)
					rows = append(rows, corpus.Assignment{Candidate: b.ID, Split: "HOLDOUT", Rationale: "review"})
				}
				o := p
				o.Kind = "assignment"
				o.ID = "assignment"
				o.Policy = nil
				o.PolicyDigest = pr.Object
				o.Assignments = &rows
				q.Action = "assign-split"
				q.Object = o
				if kind == "stale assignment" {
					ar := apply("assign-split", o)
					_, ad, _ := corpus.Encode(a)
					a.Revision = 2
					a.Parent = &ad
					a.Candidate.Sources[0].Locator = "https://example.invalid/revised"
					apply("candidate revise", a)
					m := p
					m.Schema = corpus.ManifestSchema
					m.Kind = "manifest"
					m.ID = "manifest"
					m.Policy = nil
					m.PolicyDigest = pr.Object
					m.Manifest = &corpus.Manifest{Policy: pref, Assignment: corpus.Ref{ID: o.ID, Revision: 1, Digest: ar.Object}, CheckTime: p.ReportedAt, Build: corpus.Hash([]byte("review")), SourceCommit: "synthetic"}
					prepared, e := s.PrepareFreeze(m)
					if e != nil {
						t.Fatal(e)
					}
					q.Action = "freeze"
					q.Object = prepared
				}
				if kind == "cross policy" {
					apply("assign-split", o)
					p.ID = "p2"
					p.Policy.Round = "round2"
					p2 := apply("policy register", p)
					q.Object.ID = "p2-assignment"
					q.Object.PolicyDigest = p2.Object
					(*q.Object.Assignments)[0].Split = "DEVELOPMENT"
				}
			}
			q.ExpectedHead = head
			b, _ := json.Marshal(q)
			if kind == "malformed wire" {
				b = bytes.Replace(b, []byte(`"candidate":`), []byte(`"body":"[private] blocked-text","candidate":`), 1)
			}
			// Copy only the already accepted synthetic prefix, before either submission.
			mirror := filepath.Join(t.TempDir(), "store")
			err := filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
				if e != nil {
					return e
				}
				rel, _ := filepath.Rel(root, path)
				dest := filepath.Join(mirror, rel)
				if d.IsDir() {
					return os.MkdirAll(dest, 0700)
				}
				raw, e := os.ReadFile(path)
				if e != nil {
					return e
				}
				return os.WriteFile(dest, raw, 0600)
			})
			if err != nil {
				t.Fatal(err)
			}
			command := q.Action
			if kind == "action mismatch" {
				command = "candidate revise"
			}
			expected, err := (corpus.Store{Root: mirror}).ApplyCommand(command, b, "")
			if err == nil || expected.Sequence == 0 {
				t.Fatal("package did not reject audibly", kind, err)
			}
			input := filepath.Join(t.TempDir(), "input.json")
			os.WriteFile(input, b, 0600)
			args := append([]string{"corpus"}, strings.Fields(command)...)
			args = append(args, "--store", root, "--input", input)
			out, code := cli(t, args...)
			if code != corpus.ExitCode(err) || !bytes.Contains(out, []byte(err.Error())) {
				t.Fatalf("CLI differs: %d %s / %v", code, out, err)
			}
			var got corpus.Result
			if json.Unmarshal(bytes.SplitN(out, []byte("\n"), 2)[0], &got) != nil || got != expected {
				t.Fatalf("receipt differs: %s", out)
			}
		})
	}
}
