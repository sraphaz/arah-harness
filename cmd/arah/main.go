package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/adapters/sqlitestore"
	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/envelope"
	"github.com/sraphaz/arah-harness/internal/evidence"
	"github.com/sraphaz/arah-harness/internal/kernel"
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
	case "economy":
		os.Exit(runEconomy(root, args, jsonOut))
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
	case "kernel":
		os.Exit(runKernel(root, args, jsonOut))
	case "help", "-h", "--help":
		usage()
	default:
		msg := fmt.Sprintf("unknown command %q", cmd)
		failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, msg, nil, "arah doctor|sync-check|version|task|evidence|economy|mcp|kernel"))
	}
}

func newTaskService(root string) (*core.TaskService, error) {
	store, err := sqlitestore.New(root)
	if err != nil {
		return nil, err
	}
	return &core.TaskService{
		Store:         store,
		Events:        store,
		Router:        choreography.New(root),
		Briefings:     fsstore.New(root),
		Consultations: fsstore.New(root),
	}, nil
}

func evidenceBuilder(root string, svc *core.TaskService) *evidence.Builder {
	return &evidence.Builder{RepoRoot: root, Store: svc.Store, Events: svc.Events}
}

func runTask(root string, args []string, jsonOut bool) int {
	if len(args) == 0 {
		failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah task <create|status|complete|block|timeline|context>", nil))
	}
	dryRun := boolFlag(args, "--dry-run", "-dry-run")
	subArgs := stripGlobalFlags(args)
	if len(subArgs) == 0 {
		failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah task <create|status|complete|block|timeline|context>", nil))
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
		res, err := svc.Create(obj, area, core.WorkClass(wc), core.IntentType(intent), opts)
		return emitMutation(jsonOut, res, err)
	case "status":
		id := flagValue(rest, "-task-id", "--task-id", "")
		if id == "" && len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
			id = rest[0]
		}
		c, path, err := svc.Get(id)
		return emitTask(jsonOut, c, path, err, false, "", false)
	case "complete":
		id := flagValue(rest, "-task-id", "--task-id", "")
		ev := flagValue(rest, "-evidence", "--evidence", "")
		opts := core.MutateOptions{DryRun: dryRun}
		res, err := svc.Complete(id, []string{ev}, opts)
		return emitMutation(jsonOut, res, err)
	case "block":
		id := flagValue(rest, "-task-id", "--task-id", "")
		reason := flagValue(rest, "-reason", "--reason", "")
		opts := core.MutateOptions{DryRun: dryRun}
		res, err := svc.Block(id, reason, opts)
		return emitMutation(jsonOut, res, err)
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
			fmt.Printf("  %s  %s  run=%s corr=%s agent=%s\n", e.At, e.Kind, e.RunID, e.CorrelationID, e.AgentID)
		}
		return 0
	case "context":
		id := flagValue(rest, "-task-id", "--task-id", "")
		budget := flagValue(rest, "-budget", "--budget", "standard")
		tc, err := svc.Context(id, core.ParseContextBudget(budget))
		if err != nil {
			return failEnv(jsonOut, domainEnv(err))
		}
		if jsonOut {
			return envelope.WriteJSON(os.Stdout, envelope.OK(tc))
		}
		fmt.Printf("context %s budget=%s tokens≈%d\n", tc.TaskID, tc.Budget, tc.EstimatedTokens)
		fmt.Printf("  executor=%s state=%s\n", tc.PrimaryExecutor, tc.State)
		fmt.Printf("  objective=%s\n", tc.Objective)
		for _, n := range tc.DisclosureNotes {
			fmt.Printf("  note: %s\n", n)
		}
		return 0
	default:
		return failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "unknown task action: "+action, nil,
			"arah task create|status|complete|block|timeline|context"))
	}
}

func runKernel(root string, args []string, jsonOut bool) int {
	sub := stripGlobalFlags(args)
	if len(sub) == 0 {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah kernel <sync|verify|install>", nil,
			"arah kernel sync|verify|install [-target path] [--force] [--json]"))
	}
	switch sub[0] {
	case "sync":
		m, n, err := kernel.Sync(root)
		if err != nil {
			return failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, err.Error(), nil))
		}
		data := map[string]any{
			"files":        len(m.Files),
			"copied":       n,
			"manifest":     kernel.ManifestRel,
			"payload_zip":  kernel.PayloadZipRel,
			"generated_at": m.GeneratedAt,
		}
		if jsonOut {
			return envelope.WriteJSON(os.Stdout, envelope.OK(data))
		}
		fmt.Printf("kernel sync: %d files → kernel/ + %s\n", len(m.Files), kernel.PayloadZipRel)
		return 0
	case "verify":
		drifts, err := kernel.Verify(root)
		if err != nil {
			return failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, err.Error(), nil))
		}
		if len(drifts) > 0 {
			details := map[string]any{"drift_count": len(drifts), "drifts": drifts}
			if jsonOut {
				_ = envelope.WriteJSON(os.Stdout, envelope.Fail("KERNEL.DRIFT", "kernel/ out of sync with sources", details,
					"arah kernel sync -target ."))
				return 2
			}
			fmt.Fprintf(os.Stderr, "kernel verify: %d drift(s)\n", len(drifts))
			for _, d := range drifts {
				fmt.Fprintf(os.Stderr, "  %s\n", d)
			}
			return 2
		}
		if jsonOut {
			return envelope.WriteJSON(os.Stdout, envelope.OK(map[string]any{"drift": false}))
		}
		fmt.Println("kernel verify: OK")
		return 0
	case "install":
		force := hasFlag(args, "--force") || hasFlag(args, "-force")
		n, err := kernel.Install(root, nil, kernel.InstallOptions{Force: force})
		if err != nil {
			return failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, err.Error(), nil))
		}
		data := map[string]any{
			"installed": n,
			"target":    root,
			"force":     force,
			"source":    "embed",
		}
		if jsonOut {
			return envelope.WriteJSON(os.Stdout, envelope.OK(data))
		}
		fmt.Printf("kernel install: %d files → %s\n", n, root)
		return 0
	default:
		return failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "unknown kernel action: "+sub[0], nil,
			"arah kernel sync|verify|install"))
	}
}

func runEvidence(root string, args []string, jsonOut bool) int {
	sub := stripGlobalFlags(args)
	if len(sub) == 0 {
		failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah evidence <graph|explain>", nil))
	}
	svc, err := newTaskService(root)
	if err != nil {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeStore, err.Error(), nil))
	}
	b := evidenceBuilder(root, svc)
	switch sub[0] {
	case "graph":
		g, err := b.Build()
		if err != nil {
			return failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, err.Error(), nil))
		}
		return envelope.WriteJSON(os.Stdout, envelope.OK(g))
	case "explain":
		id := flagValue(sub[1:], "-task-id", "--task-id", "")
		data, err := b.Explain(id)
		if err != nil {
			return failEnv(jsonOut, domainEnv(err))
		}
		return envelope.WriteJSON(os.Stdout, envelope.OK(data))
	default:
		return failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah evidence <graph|explain>", nil))
	}
}

func runEconomy(root string, args []string, jsonOut bool) int {
	sub := stripGlobalFlags(args)
	if len(sub) == 0 || sub[0] != "context-compare" {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeUsage, "usage: arah economy context-compare [--task-id id] [--json]", nil))
	}
	rest := sub[1:]
	id := flagValue(rest, "-task-id", "--task-id", "")
	svc, err := newTaskService(root)
	if err != nil {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeStore, err.Error(), nil))
	}
	// If no task id, create a dry-run sample for measurement.
	var c *core.Contract
	if id == "" {
		res, err := svc.Create("measure context budget for runtime cohesion", "backend", core.WorkStandard, core.IntentExecution, core.MutateOptions{DryRun: true})
		if err != nil {
			return failEnv(jsonOut, domainEnv(err))
		}
		c = res.Contract
	} else {
		c, _, err = svc.Get(id)
		if err != nil {
			return failEnv(jsonOut, domainEnv(err))
		}
	}
	agentsMD, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	execMD, _ := os.ReadFile(filepath.Join(root, "docs", "EXECUTION_CONTROL.md"))
	raw, _ := yamlMarshal(c)
	briefing := core.RenderBriefing(c)
	before := core.BaselinePromptTokens(string(agentsMD), string(execMD), string(raw), briefing)

	var events []core.Event
	if id != "" {
		events, _ = svc.Timeline(id)
	}
	minimal := core.BuildTaskContext(c, events, core.BudgetMinimal, "")
	standard := core.BuildTaskContext(c, events, core.BudgetStandard, "")
	full := core.BuildTaskContext(c, events, core.BudgetFull, briefing)

	savingsStandard := 0.0
	if before > 0 {
		savingsStandard = float64(before-standard.EstimatedTokens) / float64(before) * 100
	}
	data := map[string]any{
		"task_id": c.TaskID,
		"before": map[string]any{
			"mode":             "legacy_dump",
			"estimated_tokens": before,
			"includes":         []string{"AGENTS.md", "docs/EXECUTION_CONTROL.md", "full_contract_yaml", "briefing"},
		},
		"after": map[string]any{
			"minimal":  minimal.EstimatedTokens,
			"standard": standard.EstimatedTokens,
			"full":     full.EstimatedTokens,
		},
		"savings_pct_standard_vs_before": savingsStandard,
		"proxy":                          "chars/4",
		"note":                           "Deterministic token proxy for harness context budget (Economy M2-compatible).",
	}
	// Persist cold evidence under docs/_meta/runs when writing to this repo.
	outDir := filepath.Join(root, "docs", "_meta", "runs", "context-budget")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, "cannot create context-budget run dir: "+err.Error(), map[string]any{"path": outDir}))
	}
	summaryPath := filepath.Join(outDir, "summary.json")
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, err.Error(), nil))
	}
	if err := os.WriteFile(summaryPath, b, 0o644); err != nil {
		return failEnv(jsonOut, envelope.Fail(envelope.CodeInternal, "cannot write context-budget summary: "+err.Error(), map[string]any{"path": summaryPath}))
	}
	data["summary_path"] = "docs/_meta/runs/context-budget/summary.json"
	if jsonOut {
		return envelope.WriteJSON(os.Stdout, envelope.OK(data))
	}
	fmt.Printf("economy context-compare task=%s\n", c.TaskID)
	fmt.Printf("  before (legacy dump) ≈ %d tokens\n", before)
	fmt.Printf("  after  minimal≈%d  standard≈%d  full≈%d\n", minimal.EstimatedTokens, standard.EstimatedTokens, full.EstimatedTokens)
	fmt.Printf("  savings standard vs before: %.1f%%\n", savingsStandard)
	fmt.Printf("  wrote %s\n", summaryPath)
	return 0
}

func yamlMarshal(v any) ([]byte, error) {
	return json.Marshal(v) // size proxy; avoid yaml dep in main for estimate
}

func emitMutation(jsonOut bool, res *core.MutationResult, err error) int {
	if err != nil {
		return failEnv(jsonOut, domainEnv(err))
	}
	return emitTask(jsonOut, res.Contract, res.Path, nil, res.DryRun, res.Diff, res.Idempotent)
}

func emitTask(jsonOut bool, c *core.Contract, path string, err error, dryRun bool, diff string, idempotent bool) int {
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
		"idempotent":        idempotent,
		"diff":              diff,
		"choreography_rule": c.ChoreographyRule,
		"evidence":          c.Execution.CompletionEvidence,
		"blocking_reason":   c.Result.BlockingReason,
	}
	if jsonOut {
		return envelope.WriteJSON(os.Stdout, envelope.OK(data))
	}
	label := ""
	if dryRun || strings.HasPrefix(path, "dry-run") {
		label = " (dry-run)"
	}
	if idempotent {
		label += " (idempotent)"
	}
	fmt.Printf("task %s: %s%s\n", c.TaskID, c.State, label)
	fmt.Printf("  executor: %s\n", c.PrimaryExecutor)
	fmt.Printf("  objective: %s\n", c.Objective)
	fmt.Printf("  path: %s\n", path)
	if diff != "" {
		fmt.Printf("  diff:\n")
		for _, line := range strings.Split(diff, "\n") {
			fmt.Printf("    %s\n", line)
		}
	}
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
  arah task context -task-id ID [-budget minimal|standard|full] [--json]
  arah evidence graph|explain [-task-id ID] [--json]
  arah economy context-compare [-task-id ID] [--json]
  arah mcp serve [-target path]
  arah kernel sync|verify|install [-target path] [--force] [--json]

Hot state: .arah/local/runtime.db (SQLite WAL) + YAML mirror for PS.
Context budget: progressive disclosure via task context / MCP arah_get_task_context.
Kernel: edit root sources, then "arah kernel sync"; CI runs "arah kernel verify";
install extracts the go:embed payload zip without a harness checkout.
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

// boolFlag accepts bare flags (--dry-run) and equals forms (--dry-run=true|false|1|0|yes|no).
func boolFlag(args []string, names ...string) bool {
	for _, a := range args {
		for _, name := range names {
			if a == name {
				return true
			}
			prefix := name + "="
			if strings.HasPrefix(a, prefix) {
				switch strings.ToLower(strings.TrimSpace(strings.TrimPrefix(a, prefix))) {
				case "true", "1", "yes", "y", "on", "":
					return true
				default:
					return false
				}
			}
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
		case "--json":
			continue
		case "-target", "--target":
			skip = true
			continue
		default:
			if strings.HasPrefix(a, "-target=") || strings.HasPrefix(a, "--target=") {
				continue
			}
			if a == "--dry-run" || a == "-dry-run" ||
				strings.HasPrefix(a, "--dry-run=") || strings.HasPrefix(a, "-dry-run=") {
				continue
			}
			out = append(out, a)
		}
	}
	return out
}
