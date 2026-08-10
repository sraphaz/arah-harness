package kernel_test

import (
	"os"
	"path/filepath"
	"strings"
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

func TestListSourcesSkipsArahLiveTelemetry(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root)
	live := filepath.Join(root, ".cursor", "arah-live", "sessions", "chat.json")
	if err := os.MkdirAll(filepath.Dir(live), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(live, []byte(`{"session":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := kernel.ListSources(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Source, ".cursor/arah-live") {
			t.Fatalf("live telemetry leaked: %s", e.Source)
		}
	}
	m, _, err := kernel.Sync(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Files[".cursor/arah-live/sessions/chat.json"]; ok {
		t.Fatal("arah-live must not enter kernel manifest")
	}
	if _, err := os.Stat(filepath.Join(root, "kernel", ".cursor", "arah-live")); !os.IsNotExist(err) {
		t.Fatalf("expected arah-live absent under kernel, err=%v", err)
	}
}

func TestVerifyMalformedManifestDigest(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root)
	if _, _, err := kernel.Sync(root); err != nil {
		t.Fatal(err)
	}
	manPath := filepath.Join(root, "kernel", "manifest.json")
	raw, err := os.ReadFile(manPath)
	if err != nil {
		t.Fatal(err)
	}
	// Corrupt one digest to a short value that would panic on naive [:12].
	corrupted := string(raw)
	// Replace first 64-hex digest occurrence with "dead"
	idx := -1
	for i := 0; i+64 <= len(corrupted); i++ {
		chunk := corrupted[i : i+64]
		ok := true
		for _, c := range chunk {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				ok = false
				break
			}
		}
		if ok {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("no digest found in manifest")
	}
	corrupted = corrupted[:idx] + "dead" + corrupted[idx+64:]
	if err := os.WriteFile(manPath, []byte(corrupted), 0o644); err != nil {
		t.Fatal(err)
	}
	drifts, err := kernel.Verify(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(drifts) == 0 {
		t.Fatal("expected malformed digest drift without panic")
	}
	found := false
	for _, d := range drifts {
		if d.Reason == "manifest digest malformed" || d.Got == "dead" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected malformed digest reason, got %#v", drifts)
	}
}
