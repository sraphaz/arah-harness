package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Portable CLI phase 1 (H-07): read/validate commands with exit-code parity.
// PowerShell remains the full surface; this binary covers doctor + sync-check.

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(10)
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	target := fs.String("target", ".", "repository path")
	_ = fs.Parse(os.Args[2:])

	root, err := filepath.Abs(*target)
	if err != nil {
		fail(1, err.Error())
	}

	switch cmd {
	case "doctor":
		os.Exit(runDoctor(root))
	case "sync-check":
		os.Exit(runSyncCheck(root))
	case "version":
		fmt.Println("arah (go) 0.3.1-phase1")
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q (phase1: doctor|sync-check|version)\n", cmd)
		os.Exit(10)
	}
}

func usage() {
	fmt.Print(`ARAH portable CLI (Go) — phase 1

  arah doctor [-target path]
  arah sync-check [-target path]
  arah version

Exit codes: 0 ok · 1 error · 2 drift · 4 unhealthy · 10 usage
PowerShell CLI remains canonical for write/organism flows.
`)
}

func runDoctor(root string) int {
	checks := []struct {
		rel  string
		name string
	}{
		{"AGENTS.md", "AGENTS.md"},
		{"arah.config.yaml", "arah.config.yaml"},
		{filepath.Join(".agents", "choreography.yaml"), ".agents/choreography.yaml"},
		{".skills", ".skills"},
		{filepath.Join("scripts", "agents", "validate-manifests.ps1"), "validate-manifests.ps1"},
	}
	fmt.Printf("ARAH doctor (go) — %s\n", root)
	bad := 0
	for _, c := range checks {
		p := filepath.Join(root, c.rel)
		if _, err := os.Stat(p); err != nil {
			fmt.Printf("  [missing] %s\n", c.name)
			bad++
		} else {
			fmt.Printf("  [ok] %s\n", c.name)
		}
	}
	if bad > 0 {
		fmt.Println("doctor: unhealthy")
		return 4
	}
	fmt.Println("doctor: OK")
	return 0
}

func runSyncCheck(root string) int {
	graph := filepath.Join(root, "docs", "_meta", "agent-graph.generated.json")
	ver := filepath.Join(root, ".arah-version")
	missing := []string{}
	if _, err := os.Stat(graph); err != nil {
		missing = append(missing, "docs/_meta/agent-graph.generated.json")
	}
	if _, err := os.Stat(ver); err != nil {
		missing = append(missing, ".arah-version")
	}
	if len(missing) > 0 {
		fmt.Printf("sync-check: drift — missing %s\n", strings.Join(missing, ", "))
		return 2
	}
	// Optional: parse graph JSON for sanity
	b, err := os.ReadFile(graph)
	if err != nil {
		fail(1, err.Error())
	}
	var probe any
	if err := json.Unmarshal(b, &probe); err != nil {
		fmt.Println("sync-check: drift — graph JSON invalid")
		return 2
	}
	fmt.Println("sync-check: OK")
	return 0
}

func fail(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}
