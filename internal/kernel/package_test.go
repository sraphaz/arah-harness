package kernel_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/kernel"
)

func writeTree(t *testing.T, root string) {
	t.Helper()
	files := map[string]string{
		".agents/choreography.yaml":            "version: 2\n",
		".agents/backend.agent.yaml":           "id: backend\n",
		".agents/choreography.harness.yaml":    "harness-only: true\n",
		".agents/domain/cli.agent.yaml":        "id: cli\n",
		".agents/domain/test-architect.agent.yaml": "id: test-architect\n",
		".skills/demo.skill.yaml":              "id: demo\n",
		".cursor/hooks.json":                   "{}\n",
		"scripts/agents/hello.ps1":             "# hello\n",
		"scripts/harness/validate-specs.ps1":   "# specs\n",
		"scripts/harness/doctor-harness.ps1":   "# doctor\n",
	}
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSyncAndVerify(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root)

	m, n, err := kernel.Sync(root)
	if err != nil {
		t.Fatal(err)
	}
	if n < 5 {
		t.Fatalf("expected several copies, got %d", n)
	}
	if _, ok := m.Files[".agents/choreography.harness.yaml"]; ok {
		t.Fatal("harness-only file must not enter kernel")
	}
	if _, ok := m.Files[".agents/domain/cli.agent.yaml"]; ok {
		t.Fatal("cli domain agent must not enter kernel")
	}
	if _, ok := m.Files[".agents/domain/test-architect.agent.yaml"]; !ok {
		t.Fatal("distributable domain agent missing")
	}
	if _, err := os.Stat(filepath.Join(root, "kernel", "manifest.json")); err != nil {
		t.Fatal(err)
	}

	drifts, err := kernel.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(drifts) != 0 {
		t.Fatalf("unexpected drift: %v", drifts)
	}
}

func TestVerifyDetectsDrift(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root)
	if _, _, err := kernel.Sync(root); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "scripts", "agents", "hello.ps1")
	if err := os.WriteFile(p, []byte("# changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	drifts, err := kernel.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(drifts) == 0 {
		t.Fatal("expected drift after source change")
	}
}

func TestListSourcesSkipsHarnessOnly(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root)
	entries, err := kernel.ListSources(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Source == ".agents/choreography.harness.yaml" || e.Source == "scripts/harness/doctor-harness.ps1" {
			t.Fatalf("harness-only leaked: %s", e.Source)
		}
	}
}
