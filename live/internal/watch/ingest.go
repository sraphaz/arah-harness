package watch

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sraphaz/arah-harness/live/internal/store"
)

type Ingestor struct {
	Root  string
	Store *store.Store
}

func (ing *Ingestor) ScanAll() (int, error) {
	n := 0
	roots := []struct {
		dir    string
		source string
	}{
		{filepath.Join(ing.Root, ".arah", "local", "bus"), "bus"},
		{filepath.Join(ing.Root, ".arah", "local", "audit"), "audit"},
		{filepath.Join(ing.Root, ".arah", "bus"), "bus"},
		{filepath.Join(ing.Root, ".arah", "audit"), "audit"},
		{filepath.Join(ing.Root, ".cursor", "arah-live"), "live"},
	}
	for _, r := range roots {
		c, _ := ing.scanTree(r.dir, r.source)
		n += c
	}
	return n, nil
}

func (ing *Ingestor) scanTree(dir, source string) (int, error) {
	if _, err := os.Stat(dir); err != nil {
		return 0, nil
	}
	n := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".jsonl") {
			c, e := ing.ingestFile(path, source)
			if e != nil {
				ing.Store.AddError(path + ": " + e.Error())
			}
			n += c
		}
		return nil
	})
	return n, nil
}

func (ing *Ingestor) IngestPath(path string) {
	source := "change"
	p := filepath.ToSlash(path)
	switch {
	case strings.Contains(p, "/bus/"):
		source = "bus"
	case strings.Contains(p, "/audit/"):
		source = "audit"
	case strings.Contains(p, "arah-live"):
		source = "live"
	}
	if _, err := ing.ingestFile(path, source); err != nil {
		ing.Store.AddError(path + ": " + err.Error())
	}
}

func (ing *Ingestor) ingestFile(path, source string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	n := 0
	if strings.HasSuffix(path, ".jsonl") {
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
		lineNo := 0
		for sc.Scan() {
			lineNo++
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			key := path + "#" + itoa(lineNo)
			ok, err := ing.insertRaw(key, source, []byte(line))
			if err != nil {
				ing.Store.AddError(key + ": " + err.Error())
				continue
			}
			if ok {
				n++
			}
		}
		return n, sc.Err()
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	b = []byte(strings.TrimSpace(string(b)))
	if len(b) == 0 {
		return 0, nil
	}
	ok, err := ing.insertRaw(path, source, b)
	if err != nil {
		return 0, err
	}
	if ok {
		return 1, nil
	}
	return 0, nil
}

func (ing *Ingestor) insertRaw(key, source string, raw []byte) (bool, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false, err
	}
	ev := store.Event{Source: source, V: 1}
	if v, ok := m["v"].(float64); ok {
		ev.V = int(v)
	}
	if ts, ok := m["ts"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			ev.TS = t
		} else if t, err := time.Parse(time.RFC3339, ts); err == nil {
			ev.TS = t
		}
	}
	action, _ := m["action"].(string)
	sigType, _ := m["type"].(string)
	outcome, _ := m["outcome"].(string)
	ev.Type = store.NormalizeType(source, action, sigType, outcome)
	ev.AgentID, _ = m["agent_id"].(string)
	if ev.AgentID == "" {
		ev.AgentID, _ = m["from"].(string)
	}
	ev.CorrelationID, _ = m["correlation_id"].(string)
	ev.Outcome = outcome
	if to, ok := m["to"].(string); ok {
		ev.Route = to
	}
	if topic, ok := m["topic"].(string); ok && ev.Route == "" {
		ev.Route = topic
	}
	if details, ok := m["details"].(string); ok && details != "" {
		ev.Message = details
	} else if msg, ok := m["message"].(string); ok {
		ev.Message = msg
	} else {
		ev.Message = ev.Type
		if ev.AgentID != "" {
			ev.Message = ev.AgentID + " → " + ev.Type
		}
	}
	return ing.Store.InsertEvent(key, ev, raw)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [32]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
