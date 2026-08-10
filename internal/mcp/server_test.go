package mcp_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sraphaz/arah-harness/internal/adapters/choreography"
	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/core"
	arahmcp "github.com/sraphaz/arah-harness/internal/mcp"
)

func TestMCPToolsListAndCreate(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".agents"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".agents", "choreography.yaml"), []byte("version: 2\nrules: []\n"), 0o644)

	in := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"arah_create_task","arguments":{"objective":"from mcp","area":"backend"}}}`,
	}, "\n") + "\n"

	var out bytes.Buffer
	srv := &arahmcp.Server{
		Tasks:   &core.TaskService{Store: fsstore.New(root), Router: choreography.New(root)},
		Version: "test",
		Reader:  strings.NewReader(in),
		Writer:  &out,
	}
	if err := srv.Run(); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 3 {
		t.Fatalf("responses=%d body=%s", len(lines), out.String())
	}
	var createResp map[string]any
	if err := json.Unmarshal([]byte(lines[2]), &createResp); err != nil {
		t.Fatal(err)
	}
	result, _ := createResp["result"].(map[string]any)
	if result["isError"] == true {
		t.Fatalf("create error: %#v", result)
	}
}
