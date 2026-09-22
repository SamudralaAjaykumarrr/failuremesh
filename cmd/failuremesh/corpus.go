package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/corpus"
)

func runCorpus(args []string) error {
	if len(args) == 0 {
		return &corpus.Error{Code: "CORPUS_COMMAND_REQUIRED", Exit: 2}
	}
	action := args[0]
	args = args[1:]
	if action == "candidate" || action == "exposure" || action == "policy" || action == "index" {
		if len(args) == 0 {
			return &corpus.Error{Code: "CORPUS_SUBCOMMAND_REQUIRED", Exit: 2}
		}
		action += " " + args[0]
		args = args[1:]
	}
	flags := flag.NewFlagSet("corpus", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("store", "", "trusted local corpus root")
	input := flags.String("input", "", "canonical request JSON file")
	digest := flags.String("digest", "", "SHA-256 object hex")
	head := flags.String("head", "", "expected current ledger SHA-256 hex")
	at := flags.String("at", "", "supplied canonical UTC policy time")
	attemptID := flags.String("attempt-id", "", "safe opaque ID for malformed submission envelopes")
	prepare := flags.Bool("prepare", false, "compute a freeze request without publishing")
	historical := flags.Bool("historical", false, "historical-only verification")
	eroot := flags.String("evidence-store", "", "separate Phase 2A root")
	var parents corpusParentFiles
	flags.Var(&parents, "parent-file", "explicit permitted transient Phase 2A parent file; repeatable")
	eid := flags.String("evidence-store-id", "", "declared Phase 2A store identity")
	if flags.Parse(args) != nil || flags.NArg() != 0 || *root == "" {
		return &corpus.Error{Code: "INVALID_CORPUS_ARGUMENTS", Exit: 2}
	}
	s := corpus.Store{Root: *root, EvidenceRoot: *eroot, EvidenceStoreID: *eid}
	totalParentBytes := 0
	for _, path := range parents {
		b, e := corpus.ReadInput(path)
		if e != nil {
			return e
		}
		totalParentBytes += len(b)
		if totalParentBytes > corpus.MaxClosureBytes {
			return &corpus.Error{Code: "PARENT_LIMIT", Exit: 2}
		}
		s.Parents = append(s.Parents, b)
	}
	output := func(v any) error {
		b, e := json.Marshal(v)
		if e != nil {
			return e
		}
		_, e = fmt.Fprintln(os.Stdout, string(b))
		return e
	}
	var h *corpus.Digest
	if *head != "" {
		h = &corpus.Digest{Algorithm: "sha256", Hex: *head}
	}
	d := corpus.Digest{Algorithm: "sha256", Hex: *digest}
	switch action {
	case "index rebuild":
		v, e := s.RebuildIndex()
		if e != nil {
			return e
		}
		return output(v)
	case "head":
		v, e := s.Head()
		if e != nil {
			return e
		}
		return output(v)
	case "recover":
		v, e := s.Recover()
		if e != nil {
			return e
		}
		return output(v)
	case "inspect":
		v, e := s.Inspect(d)
		if e != nil {
			return e
		}
		return output(v)
	case "verify":
		v, e := s.Verify(d, h, *at, !*historical)
		if e != nil {
			return e
		}
		return output(v)
	}
	if !slices.Contains([]string{"policy register", "candidate add", "candidate revise", "candidate exclude", "cluster", "assign-family", "assign-split", "exposure append", "freeze", "attempt", "retire", "withdraw"}, action) {
		return &corpus.Error{Code: "UNSUPPORTED_CORPUS_COMMAND", Exit: 2}
	}
	b, e := corpus.ReadInput(*input)
	if e != nil {
		if !*prepare && e.Error() == "OBJECT_LIMIT" {
			v, err := s.RejectOversizedInput(*attemptID)
			if v.Sequence > 0 {
				if oe := output(v); oe != nil {
					return oe
				}
			}
			return err
		}
		return e
	}
	q, e := corpus.DecodeRequest(b)
	if e != nil {
		v, err := s.ApplyBytes(b, *attemptID)
		if v.Sequence > 0 {
			if oe := output(v); oe != nil {
				return oe
			}
		}
		return err
	}
	if q.Action != action {
		v, err := s.ApplyCommand(action, b, *attemptID)
		if v.Sequence > 0 {
			if oe := output(v); oe != nil {
				return oe
			}
		}
		return err
	}
	if *prepare {
		if action != "freeze" {
			return &corpus.Error{Code: "PREPARE_REQUIRES_FREEZE", Exit: 2}
		}
		v, e := s.PrepareFreeze(q.Object)
		if e != nil {
			return e
		}
		q.Object = v
		q.ExpectedHead = v.Manifest.LedgerHead
		b, e := json.Marshal(q)
		if e != nil {
			return e
		}
		_, e = fmt.Fprint(os.Stdout, string(b))
		return e
	}
	v, e := s.ApplyCommand(action, b, *attemptID)
	if e != nil {
		if v.Sequence > 0 {
			if oe := output(v); oe != nil {
				return oe
			}
		}
		return e
	}
	return output(v)
}

type corpusParentFiles []string

func (p *corpusParentFiles) String() string { return "explicit local files" }
func (p *corpusParentFiles) Set(v string) error {
	if len(*p) >= corpus.MaxSources {
		return &corpus.Error{Code: "PARENT_LIMIT", Exit: 2}
	}
	*p = append(*p, v)
	return nil
}
