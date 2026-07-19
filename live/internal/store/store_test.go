package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestInsertAndFeed(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "idx"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	ev := Event{
		V: 1, TS: time.Now().UTC(), Type: "consultation.request",
		Message: "payments consult", Source: "bus", AgentID: "orchestrator",
	}
	ok, err := s.InsertEvent("k1", ev, []byte(`{"v":1}`))
	if err != nil || !ok {
		t.Fatalf("insert: ok=%v err=%v", ok, err)
	}
	ok, err = s.InsertEvent("k1", ev, []byte(`{"v":1}`))
	if err != nil || ok {
		t.Fatalf("duplicate should ignore: ok=%v err=%v", ok, err)
	}
	feed, err := s.Feed("consultation", 10)
	if err != nil || len(feed) != 1 {
		t.Fatalf("feed: %#v err=%v", feed, err)
	}
}

func TestNormalizeType(t *testing.T) {
	cases := map[string]string{
		NormalizeType("bus", "signal.consult", "consult", "ok"): "consultation.request",
		NormalizeType("audit", "gates.run", "", "ok"):           "gates.passed",
		NormalizeType("audit", "gates.run", "", "blocked"):      "gates.failed",
		NormalizeType("bus", "", "evolve", ""):                  "evolution.cycle",
	}
	for got, want := range cases {
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}
