package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SamudralaAjaykumarrr/failuremesh/internal/historical/iceberg16282"
	"github.com/SamudralaAjaykumarrr/failuremesh/internal/pfc2"
	"github.com/SamudralaAjaykumarrr/failuremesh/internal/pfc3"
	"github.com/SamudralaAjaykumarrr/failuremesh/internal/phase1"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: failuremesh init|reset|match|plan|run [vulnerable|remediated], or failuremesh pfc2|pfc3 init|reset|match|plan|run [vulnerable|remediated]")
	}
	if os.Args[1] == "historical" {
		return runHistorical()
	}
	if os.Args[1] == "pfc3" {
		return runPFC3()
	}
	if os.Args[1] == "pfc2" {
		return runPFC2()
	}
	command := os.Args[1]
	mode := "vulnerable"
	if len(os.Args) > 2 {
		mode = os.Args[2]
	}
	if mode != "vulnerable" && mode != "remediated" {
		return fmt.Errorf("unsupported mode")
	}
	p, e := phase1.LoadPFC("contracts/timeout-after-external-commit.v1.json")
	if e != nil {
		return e
	}
	a := phase1.ReferenceAC(mode == "remediated")
	m := phase1.Evaluate(p, a)
	out := func(v any) error {
		b, e := json.MarshalIndent(v, "", "  ")
		if e != nil {
			return e
		}
		fmt.Println(string(b))
		return nil
	}
	if command == "match" {
		return out(m)
	}
	manifest, e := phase1.Compile(p, a, m)
	if e != nil {
		return e
	}
	s := phase1.Plan(p, manifest)
	if command == "plan" {
		return out(struct {
			Manifest phase1.Manifest   `json:"manifest"`
			Safety   phase1.SafetyPlan `json:"safety"`
		}{manifest, s})
	}
	if command != "init" && command != "reset" && command != "run" {
		return fmt.Errorf("unknown command")
	}
	dsn := os.Getenv("FAILUREMESH_DATABASE_URL")
	db, e := phase1.Open(dsn)
	if e != nil {
		return e
	}
	defer db.Close()
	ctx := context.Background()
	if command == "init" {
		return phase1.Init(ctx, db)
	}
	if command == "reset" {
		return phase1.Reset(ctx, db)
	}
	if !s.Approved {
		return fmt.Errorf("safety plan denied: %s", s.Reason)
	}
	ev := phase1.Run(ctx, db, manifest, s, mode == "remediated")
	return out(phase1.Verify(ev))
}

func runHistorical() error {
	if len(os.Args) != 4 || os.Args[2] != "iceberg-16282" {
		return fmt.Errorf("usage: failuremesh historical iceberg-16282 match|plan|init|reset|run|verify")
	}
	command := os.Args[3]
	b, e := os.ReadFile("docs/validation/historical-reproductions/001-iceberg-16282/source-record.md")
	if e != nil {
		return e
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(b))
	a := iceberg16282.ReferenceAC()
	d := iceberg16282.Evaluate(a)
	out := func(v any) error {
		b, e := json.MarshalIndent(v, "", "  ")
		if e != nil {
			return e
		}
		fmt.Println(string(b))
		return nil
	}
	if command == "match" {
		return out(d)
	}
	m, e := iceberg16282.Compile(a, d, digest)
	if e != nil {
		return e
	}
	p := iceberg16282.Plan(m)
	if command == "plan" {
		return out(struct {
			Manifest iceberg16282.Manifest   `json:"manifest"`
			Safety   iceberg16282.PlanResult `json:"safety"`
		}{m, p})
	}
	if command != "init" && command != "reset" && command != "run" && command != "verify" {
		return fmt.Errorf("unknown historical command")
	}
	db, e := phase1.Open(os.Getenv("FAILUREMESH_DATABASE_URL"))
	if e != nil {
		return e
	}
	defer db.Close()
	ctx := context.Background()
	if command == "init" {
		return iceberg16282.Init(ctx, db)
	}
	if command == "reset" {
		return iceberg16282.Reset(ctx, db)
	}
	if !p.Approved {
		return fmt.Errorf("historical safety plan denied")
	}
	if command == "run" {
		return out(iceberg16282.VerifyAgainstDB(ctx, db, iceberg16282.Run(ctx, db, m, p)))
	}
	return out(iceberg16282.VerifyAgainstDB(ctx, db, iceberg16282.CaptureCurrent(ctx, db, m, p)))
}

func runPFC2() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: failuremesh pfc2 init|reset|match|plan|run [vulnerable|remediated]")
	}
	command := os.Args[2]
	mode := "vulnerable"
	if len(os.Args) > 3 {
		mode = os.Args[3]
	}
	if mode != "vulnerable" && mode != "remediated" {
		return fmt.Errorf("unsupported mode")
	}
	p, err := pfc2.Load("contracts/duplicate-queue-delivery.v1.json")
	if err != nil {
		return err
	}
	a := pfc2.ReferenceAC(mode == "remediated")
	match := pfc2.Evaluate(p, a)
	out := func(v any) error {
		b, e := json.MarshalIndent(v, "", "  ")
		if e != nil {
			return e
		}
		fmt.Println(string(b))
		return nil
	}
	if command == "match" {
		return out(match)
	}
	m, err := pfc2.Compile(p, a, match)
	if err != nil {
		return err
	}
	s := pfc2.Plan(p, m)
	if command == "plan" {
		return out(struct {
			Manifest pfc2.Manifest   `json:"manifest"`
			Safety   pfc2.SafetyPlan `json:"safety"`
		}{m, s})
	}
	if command != "init" && command != "reset" && command != "run" {
		return fmt.Errorf("unknown command")
	}
	db, err := phase1.Open(os.Getenv("FAILUREMESH_DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	if command == "init" {
		return pfc2.Init(ctx, db)
	}
	if command == "reset" {
		return pfc2.Reset(ctx, db)
	}
	if !s.Approved {
		return fmt.Errorf("safety plan denied: %s", s.Reason)
	}
	return out(pfc2.VerifyAgainstDB(ctx, db, pfc2.Run(ctx, db, m, s, mode == "remediated")))
}

func runPFC3() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: failuremesh pfc3 init|reset|match|plan|run [vulnerable|remediated]")
	}
	command := os.Args[2]
	mode := "vulnerable"
	if len(os.Args) > 3 {
		mode = os.Args[3]
	}
	if mode != "vulnerable" && mode != "remediated" {
		return fmt.Errorf("unsupported mode")
	}
	p, err := pfc3.Load("contracts/stale-worker-after-lease-expiry.v1.json")
	if err != nil {
		return err
	}
	a := pfc3.ReferenceAC(mode == "remediated")
	match := pfc3.Evaluate(p, a)
	out := func(v any) error {
		b, e := json.MarshalIndent(v, "", "  ")
		if e != nil {
			return e
		}
		fmt.Println(string(b))
		return nil
	}
	if command == "match" {
		return out(match)
	}
	m, err := pfc3.Compile(p, a, match)
	if err != nil {
		return err
	}
	s := pfc3.Plan(p, m)
	if command == "plan" {
		return out(struct {
			Manifest pfc3.Manifest   `json:"manifest"`
			Safety   pfc3.SafetyPlan `json:"safety"`
		}{m, s})
	}
	if command != "init" && command != "reset" && command != "run" {
		return fmt.Errorf("unknown command")
	}
	db, err := phase1.Open(os.Getenv("FAILUREMESH_DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	if command == "init" {
		return pfc3.Init(ctx, db)
	}
	if command == "reset" {
		return pfc3.Reset(ctx, db)
	}
	if !s.Approved {
		return fmt.Errorf("safety plan denied: %s", s.Reason)
	}
	return out(pfc3.VerifyAgainstDB(ctx, db, pfc3.Run(ctx, db, m, s, mode == "remediated")))
}
