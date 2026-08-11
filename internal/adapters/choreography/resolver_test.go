package choreography_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/core"
)

func TestResolveBackendDefault(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte(`
version: 2
rules:
  - id: craft-backend
    paths: ["backend/**"]
    execution:
      primary_executor: backend
    agents:
      - id: backend
        type: operational
        role: executor
      - id: solutions-architect
        type: operational
        role: consultant
`), 0o644)
	r, err := choreography.New(root).Resolve("backend", "")
	if err != nil {
		t.Fatal(err)
	}
	if r.PrimaryExecutor != "backend" {
		t.Fatalf("executor=%s", r.PrimaryExecutor)
	}
	if r.ChoreographyRule != "craft-backend" {
		t.Fatalf("rule=%s", r.ChoreographyRule)
	}
}

func TestResolveByFilePath(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte(`
version: 2
rules:
  - id: craft-frontend
    paths: ["apps/web/**"]
    execution:
      primary_executor: frontend
    agents:
      - id: frontend
        type: operational
        role: executor
  - id: craft-backend
    paths: ["services/**"]
    execution:
      primary_executor: backend
    agents:
      - id: backend
        type: operational
        role: executor
`), 0o644)
	front, err := choreography.New(root).Resolve("apps/web/index.ts.txt", "")
	if err != nil {
		t.Fatal(err)
	}
	if front.PrimaryExecutor != "frontend" {
		t.Fatalf("got %s", front.PrimaryExecutor)
	}
	back, err := choreography.New(root).Resolve("services/api.go.txt", "")
	if err != nil {
		t.Fatal(err)
	}
	if back.PrimaryExecutor != "backend" {
		t.Fatalf("got %s", back.PrimaryExecutor)
	}
}

func TestResolveMixedCasePathPreserved(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte(`
version: 2
rules:
  - id: craft-web-cased
    paths: ["apps/Web/**"]
    execution:
      primary_executor: frontend
    agents:
      - id: frontend
        type: operational
        role: executor
`), 0o644)
	r, err := choreography.New(root).Resolve("apps/Web/index.ts", "")
	if err != nil {
		t.Fatal(err)
	}
	if r.PrimaryExecutor != "frontend" {
		t.Fatalf("got %s", r.PrimaryExecutor)
	}
}

func TestResolveEmbeddedDoublestar(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte(`
version: 2
rules:
  - id: craft-api-nested
    paths: ["src/**/api/**"]
    execution:
      primary_executor: backend
    agents:
      - id: backend
        type: operational
        role: executor
`), 0o644)
	r, err := choreography.New(root).Resolve("src/payments/api/handler.go", "")
	if err != nil {
		t.Fatal(err)
	}
	if r.PrimaryExecutor != "backend" {
		t.Fatalf("got %s", r.PrimaryExecutor)
	}
}

func TestResolvePathRuleRequiresPrimaryExecutor(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte(`
version: 2
rules:
  - id: catch-all-paths
    paths: ["**"]
    agents:
      - id: backend
        type: operational
        role: executor
`), 0o644)
	_, err := choreography.New(root).Resolve("apps/web/index.ts", "")
	if err == nil {
		t.Fatal("expected empty primary_executor error")
	}
	var de *core.DomainError
	if !errors.As(err, &de) || de.Code != "EXECUTION.EXACTLY_ONE_PRIMARY_EXECUTOR_REQUIRED" {
		t.Fatalf("err=%v", err)
	}
}

func TestResolveDottedAreaGetsPathFallback(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)
	r, err := choreography.New(root).Resolve("api.v2", "backend")
	if err != nil {
		t.Fatal(err)
	}
	if r.PrimaryExecutor != "backend" {
		t.Fatalf("executor=%s", r.PrimaryExecutor)
	}
	if len(r.AllowedPaths) != 1 || r.AllowedPaths[0] != "api.v2/**" {
		t.Fatalf("allowed_paths=%v", r.AllowedPaths)
	}
}

func TestResolveRejectsCorruptChoreographyYAML(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: [\n  - id: broken\n    paths: ["), 0o644)
	_, err := choreography.New(root).Resolve("backend", "")
	if err == nil {
		t.Fatal("expected YAML unmarshal error")
	}
}
