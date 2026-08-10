package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/envelope"
	"github.com/sraphaz/arah-harness/internal/evidence"
	arahmcp "github.com/sraphaz/arah-harness/internal/mcp"
)

// Version tracks runtime cohesion work (0.5 foundation on 0.4.4 tree).
const Version = "0.5.0-dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(10)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	jsonOut := hasFlag(args, "--json")
	target := flagValue(args, "-target", "--target", ".")
	root, err := filepath.Abs(target)
	if err != nil {
		failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, err.Error(), nil))
	}

	switch cmd {
	case "doctor":
		os.Exit(runDoctor(root, jsonOut))
	case "sync-check":
		os.Exit(runSyncCheck(root, jsonOut))
	case "version":
		if jsonOut {
			os.Exit(envelope.WriteJSON(os.Stdout, envelope.OK(map[string]any{
				"version": Version,
				"runtime": "arah-core",
			})))
		}
		fmt.Printf("arah (go) %s\n", Version)
	case "task":
		os.Exit(runTask(root, args, jsonOut))
	case "evidence":
		os.Exit(runEvidence(root, args, jsonOut))
	case "mcp":
		sub := stripGlobalFlags(args)
		if len(sub) == 0 || sub[0] != "serve" {
			failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah mcp serve [-target path]", nil))
		}
		svc, err := newTaskService(root)
		if err != nil {
			failEnv(jsonOut, envelope.Fail(envelope.CodeStore, err.Error(), nil))
		}
		srv := &arahmcp.Server{Tasks: svc, Version: Version, Evidence: evidenceBuilder(root, svc)}
		if err := srv.Run(); err != nil {
			failEnv(true, envelope.Fail(envelope.CodeInternal, err.Error(), nil))
		}
	case "help", "-h", "--help":
		usage()
	default:
		msg := fmt.Sprintf("unknown command %q", cmd)
		failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, msg, nil, "arah doctor|sync-check|version|task|evidence|mcp"))
	}
}

func newTaskService(root string) (*core.TaskService, error) {
	store, err := sqlitestore.New(root)
	if err != nil {
		return nil, err
	}
	return &core.TaskService{
		Store:  store,
		Events: store,
		Router: choreography.New(root),
	}, nil
}

func evidenceBuilder(root string, svc *core.TaskService) *evidence.Builder {
	return &evidence.Builder{RepoRoot: root, Store: svc.Store, Events: svc.Events}
}

func runTask(root string, args []string, jsonOut bool) int {
	if len(args) == 0 {
		failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah task <create|status|complete|block|timeline>", nil))
	}
	dryRun := boolFlag(args, "--dry-run", "-dry-run")
	subArgs := stripGlobalFlags(args)
	if len(subArgs) == 0 {
		failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah task <create|status|complete|block|timeline>", nil))
	}
	action := subArgs[0]
	rest := subArgs[1:]
	svc, err := newTaskService(root)
	if err != nil {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeStore, err.Error(), nil))
	}

	switch action {
	case "create":
		obj := flagValue(rest, "-objective", "--objective", "")
		area := flagValue(rest, "-area", "--area", "backend")
		wc := flagValue(rest, "-class", "--class", "standard")
		intent := flagValue(rest, "-intent", "--intent", "execution")
		opts := core.MutateOptions{DryRun: dryRun}
		c, path, err := svc.Create(obj, area, core.WorkClass(wc), core.IntentType(intent), opts)
		return emitTask(jsonOut, c, path, err, opts.DryRun)
	case "status":
		id := flagValue(rest, "-task-id", "--task-id", "")
		if id == "" && len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
			id = rest[0]
		}
		c, path, err := svc.Get(id)
		return emitTask(jsonOut, c, path, err, false)
	case "complete":
		id := flagValue(rest, "-task-id", "--task-id", "")
		ev := flagValue(rest, "-evidence", "--evidence", "")
		opts := core.MutateOptions{DryRun: dryRun}
		c, path, err := svc.Complete(id, []string{ev}, opts)
		return emitTask(jsonOut, c, path, err, opts.DryRun)
	case "block":
		id := flagValue(rest, "-task-id", "--task-id", "")
		reason := flagValue(rest, "-reason", "--reason", "")
		opts := core.MutateOptions{DryRun: dryRun}
		c, path, err := svc.Block(id, reason, opts)
		return emitTask(jsonOut, c, path, err, opts.DryRun)
	case "timeline":
		id := flagValue(rest, "-task-id", "--task-id", "")
		evs, err := svc.Timeline(id)
		if err != nil {
			return failEnv(jsonOut, domainEnv(err))
		}
		if jsonOut {
			return envelope.WriteJSON(os.Stdout, envelope.OK(map[string]any{"task_id": id, "events": evs}))
		}
		fmt.Printf("timeline %s (%d events)\n", id, len(evs))
		for _, e := range evs {
			fmt.Printf("  %s  %s\n", e.At, e.Kind)
		}
		return 0
	default:
		return failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "unknown task action: "+action, nil,
			"arah task create|status|complete|block|timeline"))
	}
}

func runEvidence(root string, args []string, jsonOut bool) int {
	sub := stripGlobalFlags(args)
	if len(sub) == 0 || sub[0] != "graph" {
		failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah evidence graph [--json]", nil))
	}
	svc, err := newTaskService(root)
	if err != nil {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeStore, err.Error(), nil))
	}
	g, err := evidenceBuilder(root, svc).Build()
	if err != nil {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, err.Error(), nil))
	}
	if jsonOut || true {
		// evidence graph is always structured
		return envelope.WriteJSON(os.Stdout, envelope.OK(g))
	}
	return 0
}

func emitTask(jsonOut bool, c *core.Contract, path string, err error, dryRun bool) int {
	if err != nil {
		return failEnv(jsonOut, domainEnv(err))
	}
	data := map[string]any{
		"task_id":           c.TaskID,
		"state":             c.State,
		"primary_executor":  c.PrimaryExecutor,
		"objective":         c.Objective,
		"work_class":        c.WorkClass,
		"intent_type":       c.IntentType,
		"path":              path,
		"dry_run":           dryRun || strings.HasPrefix(path, "dry-run"),
		"choreography_rule": c.ChoreographyRule,
		"evidence":          c.Execution.CompletionEvidence,
		"blocking_reason":   c.Result.BlockingReason,
	}
	if jsonOut {
		return envelope.WriteJSON(os.Stdout, envelope.OK(data))
	}
	if dryRun || strings.HasPrefix(path, "dry-run") {
		fmt.Printf("task %s: %s (dry-run)\n", c.TaskID, c.State)
	} else {
		fmt.Printf("task %s: %s\n", c.TaskID, c.State)
	}
	fmt.Printf("  executor: %s\n", c.PrimaryExecutor)
	fmt.Printf("  objective: %s\n", c.Objective)
	fmt.Printf("  path: %s\n", path)
	if c.Result.BlockingReason != nil {
		fmt.Printf("  blocked: %s\n", *c.Result.BlockingReason)
	}
	return 0
}

func domainEnv(err error) envelope.Envelope {
	if de, ok := err.(*core.DomainError); ok {
		return envelope.Fail(de.Code, de.Message, de.Details, de.Remediation...)
	}
	return envelope.Fail(envelope.CodeInternal, err.Error(), nil)
}

func failEnv(jsonOut bool, env envelope.Envelope) int {
	if jsonOut {
		code := envelope.WriteJSON(os.Stdout, env)
		os.Exit(code)
	}
	if !env.OK {
		fmt.Fprintf(os.Stderr, "%s: %s\n", env.Code, env.Message)
		if env.Code == envelope.CodeUsage {
			os.Exit(10)
		}
		os.Exit(1)
	}
	return 0
}

func runDoctor(root string, jsonOut bool) int {
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
	missing := []string{}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(root, c.rel)); err != nil {
			missing = append(missing, c.name)
		}
	}
	if jsonOut {
		ok := len(missing) == 0
		env := envelope.OK(map[string]any{"root": root, "missing": missing, "healthy": ok})
		if !ok {
			env = envelope.Fail("DOCTOR.UNHEALTHY", "doctor: unhealthy", map[string]any{"missing": missing})
			_ = envelope.WriteJSON(os.Stdout, env)
			return 4
		}
		return envelope.WriteJSON(os.Stdout, env)
	}
	fmt.Printf("ARAH doctor (go) — %s\n", root)
	bad := 0
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(root, c.rel)); err != nil {
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

func runSyncCheck(root string, jsonOut bool) int {
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
		if jsonOut {
			_ = envelope.WriteJSON(os.Stdout, envelope.Fail("SYNC.DRIFT", "sync-check: drift", map[string]any{"missing": missing}))
			return 2
		}
		fmt.Printf("sync-check: drift — missing %s\n", strings.Join(missing, ", "))
		return 2
	}
	raw, err := os.ReadFile(graph)
	if err != nil {
		if jsonOut {
			_ = envelope.WriteJSON(os.Stdout, envelope.Fail("SYNC.DRIFT", "sync-check: drift — cannot read graph", map[string]any{"error": err.Error()}))
			return 2
		}
		fmt.Printf("sync-check: drift — cannot read graph: %v\n", err)
		return 2
	}
	var probe any
	if err := json.Unmarshal(raw, &probe); err != nil {
		if jsonOut {
			_ = envelope.WriteJSON(os.Stdout, envelope.Fail("SYNC.DRIFT", "sync-check: drift — graph JSON invalid", nil))
			return 2
		}
		fmt.Println("sync-check: drift — graph JSON invalid")
		return 2
	}
	if jsonOut {
		return envelope.WriteJSON(os.Stdout, envelope.OK(map[string]any{"drift": false}))
	}
	fmt.Println("sync-check: OK")
	return 0
}

func usage() {
	fmt.Print(`ARAH portable CLI (Go) — arah-core 0.5 foundation

  arah doctor [-target path] [--json]
  arah sync-check [-target path] [--json]
  arah version [--json]
  arah task create -objective "…" [-area backend] [-class standard] [--dry-run] [--json]
  arah task status -task-id ID [--json]
  arah task complete -task-id ID -evidence "…" [--dry-run] [--json]
  arah task block -task-id ID -reason "…" [--dry-run] [--json]
  arah task timeline -task-id ID [--json]
  arah evidence graph [--json]
  arah mcp serve [-target path]

Hot state: .arah/local/runtime.db (SQLite WAL) + YAML mirror for PS.
Exit codes: 0 ok · 1 error · 2 drift · 4 unhealthy · 10 usage
`)
}

func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

func flagValue(args []string, names ...string) string {
	def := ""
	if len(names) > 0 {
		// last optional default if passed as final non-dash? we use separate def param pattern
	}
	// signature: flagValue(args, "-a", "--a", default)
	lookup := map[string]bool{}
	n := len(names)
	if n >= 1 {
		def = names[n-1]
		if strings.HasPrefix(def, "-") {
			def = ""
		} else {
			names = names[:n-1]
		}
	}
	for _, name := range names {
		lookup[name] = true
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if lookup[a] && i+1 < len(args) {
			return args[i+1]
		}
		for name := range lookup {
			prefix := name + "="
			if strings.HasPrefix(a, prefix) {
				return strings.TrimPrefix(a, prefix)
			}
		}
	}
	return def
}

func stripGlobalFlags(args []string) []string {
	out := make([]string, 0, len(args))
	skip := false
	for i := 0; i < len(args); i++ {
		if skip {
			skip = false
			continue
		}
		a := args[i]
		switch a {
		case "--json", "--dry-run", "-dry-run":
			continue
		case "-target", "--target":
			skip = true
			continue
		default:
			if strings.HasPrefix(a, "-target=") || strings.HasPrefix(a, "--target=") {
				continue
			}
			out = append(out, a)
		}
	}
	return out
}
