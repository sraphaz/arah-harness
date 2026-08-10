package choreography_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
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
