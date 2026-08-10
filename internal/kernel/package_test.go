package kernel_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"sort"
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

func TestInstallFromZip(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root)
	if _, _, err := kernel.Sync(root); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(root, filepath.FromSlash(kernel.PayloadZipRel))
	raw, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	n, err := kernel.Install(target, raw, kernel.InstallOptions{Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if n < 5 {
		t.Fatalf("installed=%d", n)
	}
	if _, err := os.Stat(filepath.Join(target, ".agents", "choreography.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, ".arah", "kernel.manifest.json")); err != nil {
		t.Fatal(err)
	}
	// without force, existing files are skipped
	n2, err := kernel.Install(target, raw, kernel.InstallOptions{Force: false})
	if err != nil {
		t.Fatal(err)
	}
	if n2 != 0 {
		t.Fatalf("expected skip when present, installed=%d", n2)
	}
}

func TestInstallEmbeddedNilAndForceRestore(t *testing.T) {
	if len(kernel.EmbeddedZip()) == 0 {
		t.Skip("embedded kernel zip empty")
	}
	target := t.TempDir()
	n, err := kernel.Install(target, nil, kernel.InstallOptions{Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if n < 5 {
		t.Fatalf("installed=%d", n)
	}
	p := filepath.Join(target, ".agents", "choreography.yaml")
	orig, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("mutated-by-test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	n2, err := kernel.Install(target, nil, kernel.InstallOptions{Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if n2 < 1 {
		t.Fatalf("force reinstall wrote nothing, n=%d", n2)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(orig) {
		t.Fatalf("force reinstall did not restore file")
	}
}

func TestSyncPayloadDeterministic(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root)
	if _, _, err := kernel.Sync(root); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(root, filepath.FromSlash(kernel.PayloadZipRel))
	a, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := kernel.Sync(root); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("payload zip must be stable across syncs with unchanged sources")
	}
}

func TestSyncRestoresManifestWhenPayloadFails(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root)
	if _, _, err := kernel.Sync(root); err != nil {
		t.Fatal(err)
	}
	manPath := filepath.Join(root, "kernel", "manifest.json")
	prev, err := os.ReadFile(manPath)
	if err != nil {
		t.Fatal(err)
	}
	payloadDir := filepath.Join(root, "internal", "kernel", "payload")
	if err := os.RemoveAll(payloadDir); err != nil {
		t.Fatal(err)
	}
	// Replace directory with a file so writing kernel.zip fails.
	if err := os.WriteFile(payloadDir, []byte("not-a-dir\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := kernel.Sync(root); err == nil {
		t.Fatal("expected payload zip failure")
	}
	got, err := os.ReadFile(manPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(prev, got) {
		t.Fatal("manifest must be restored when payload generation fails")
	}
}

func buildTestZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(files[name])); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestInstallRejectsPathEscape(t *testing.T) {
	target := t.TempDir()
	outside := filepath.Join(filepath.Dir(target), "escaped.txt")
	_ = os.Remove(outside)

	raw := buildTestZip(t, map[string]string{
		"../escaped.txt": "pwned\n",
	})
	n, err := kernel.Install(target, raw, kernel.InstallOptions{Force: true})
	if err == nil {
		t.Fatal("expected path escape error")
	}
	if n != 0 {
		t.Fatalf("installed=%d want 0", n)
	}
	if _, err := os.Stat(outside); err == nil {
		t.Fatal("escaped file must not be written outside target")
	}
	if _, err := os.Stat(filepath.Join(target, "escaped.txt")); err == nil {
		t.Fatal("escaped entry must not land inside target either")
	}
}

func TestInstallRejectsAbsoluteEntry(t *testing.T) {
	target := t.TempDir()
	raw := buildTestZip(t, map[string]string{
		"/tmp/arah-kernel-evil": "nope\n",
	})
	n, err := kernel.Install(target, raw, kernel.InstallOptions{Force: true})
	if err == nil {
		t.Fatal("expected absolute entry rejection")
	}
	if n != 0 {
		t.Fatalf("installed=%d want 0", n)
	}
}

func TestInstallRejectsDriveRelativeEntry(t *testing.T) {
	target := t.TempDir()
	raw := buildTestZip(t, map[string]string{
		"C:evil": "nope\n",
	})
	n, err := kernel.Install(target, raw, kernel.InstallOptions{Force: true})
	if err == nil {
		t.Fatal("expected drive-relative entry rejection")
	}
	if n != 0 {
		t.Fatalf("installed=%d want 0", n)
	}
}

func TestInstallRejectsSymlinkDestination(t *testing.T) {
	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.yaml")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	agents := filepath.Join(target, ".agents")
	if err := os.MkdirAll(agents, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(agents, "choreography.yaml")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	raw := buildTestZip(t, map[string]string{
		".agents/choreography.yaml": "from-zip\n",
	})
	n, err := kernel.Install(target, raw, kernel.InstallOptions{Force: true})
	if err == nil {
		t.Fatal("expected symlink destination rejection")
	}
	if n != 0 {
		t.Fatalf("installed=%d want 0", n)
	}
	got, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "outside\n" {
		t.Fatalf("symlink target mutated: %q", got)
	}
}

func TestInstallRollbackOnPartialFailure(t *testing.T) {
	target := t.TempDir()
	// Second entry needs dir "blocked/c.txt", but "blocked" is a file → MkdirAll fails after first write.
	if err := os.WriteFile(filepath.Join(target, "blocked"), []byte("not a dir\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	raw := buildTestZip(t, map[string]string{
		".agents/ok.yaml": "ok\n",
		"blocked/c.txt":   "fail\n",
	})
	n, err := kernel.Install(target, raw, kernel.InstallOptions{Force: true})
	if err == nil {
		t.Fatal("expected install failure")
	}
	if n != 0 {
		t.Fatalf("installed=%d want 0 after rollback", n)
	}
	if _, err := os.Stat(filepath.Join(target, ".agents", "ok.yaml")); err == nil {
		t.Fatal("partial install must roll back .agents/ok.yaml")
	}
}

func TestInstallRollbackRestoresOverwrite(t *testing.T) {
	target := t.TempDir()
	seed := filepath.Join(target, ".agents", "keep.yaml")
	if err := os.MkdirAll(filepath.Dir(seed), 0o755); err != nil {
		t.Fatal(err)
	}
	const original = "original-content\n"
	if err := os.WriteFile(seed, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "blocked"), []byte("not a dir\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	raw := buildTestZip(t, map[string]string{
		".agents/keep.yaml": "replacement\n",
		"blocked/c.txt":     "fail\n",
	})
	n, err := kernel.Install(target, raw, kernel.InstallOptions{Force: true})
	if err == nil {
		t.Fatal("expected install failure")
	}
	if n != 0 {
		t.Fatalf("installed=%d want 0 after rollback", n)
	}
	got, err := os.ReadFile(seed)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("overwrite must be restored on rollback, got %q", got)
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
