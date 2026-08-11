package watch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/live/internal/store"
)

func TestIngestJSONLAndPending(t *testing.T) {
	root := t.TempDir()
	bus := filepath.Join(root, ".arah", "local", "bus", "pending")
	if err := os.MkdirAll(bus, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"v":1,"ts":"2026-07-19T12:00:00Z","type":"consult","from":"qa","to":"backend","topic":"craft"}`
	if err := os.WriteFile(filepath.Join(bus, "01TEST.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(root, ".arah", "audit")
	_ = os.MkdirAll(legacy, 0o755)
	_ = os.WriteFile(filepath.Join(legacy, "events.jsonl"), []byte(`{"v":1,"ts":"2026-07-19T12:01:00Z","agent_id":"ci","action":"gates.security","outcome":"ok"}`+"\n"), 0o644)

	st, err := store.Open(filepath.Join(root, ".arah", "local", "index"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ing := &Ingestor{Root: root, Store: st}
	n, err := ing.ScanAll()
	if err != nil {
		t.Fatal(err)
	}
	if n < 2 {
		t.Fatalf("expected >=2 events, got %d (errors=%v)", n, st.Errors())
	}
	feed, _ := st.Feed("", 10)
	if len(feed) < 2 {
		t.Fatalf("feed short: %#v", feed)
	}
}
