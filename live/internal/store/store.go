package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Event struct {
	ID            int64           `json:"id,omitempty"`
	V             int             `json:"v"`
	TS            time.Time       `json:"ts"`
	Type          string          `json:"type"`
	Message       string          `json:"message"`
	Route         string          `json:"route,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	AgentID       string          `json:"agent_id,omitempty"`
	Outcome       string          `json:"outcome,omitempty"`
	Source        string          `json:"source"`
	Raw           json.RawMessage `json:"-"`
}

type Gate struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Summary  string `json:"summary,omitempty"`
	TS       string `json:"ts,omitempty"`
}

type Store struct {
	db     *sql.DB
	mu     sync.RWMutex
	errors []string
	path   string
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "live-index.sqlite")
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Path() string { return s.path }

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts TEXT NOT NULL,
  type TEXT NOT NULL,
  message TEXT NOT NULL,
  route TEXT,
  correlation_id TEXT,
  agent_id TEXT,
  outcome TEXT,
  source TEXT NOT NULL,
  raw TEXT NOT NULL,
  file_key TEXT UNIQUE
);
CREATE INDEX IF NOT EXISTS idx_events_ts ON events(ts);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
`)
	return err
}

func (s *Store) Reindex() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM events; DELETE FROM meta;`)
	s.errors = nil
	return err
}

func (s *Store) AddError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.errors) > 50 {
		s.errors = s.errors[len(s.errors)-49:]
	}
	s.errors = append(s.errors, msg)
}

func (s *Store) Errors() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.errors))
	copy(out, s.errors)
	return out
}

func (s *Store) InsertEvent(fileKey string, ev Event, raw []byte) (bool, error) {
	if ev.V == 0 {
		ev.V = 1
	}
	if ev.TS.IsZero() {
		ev.TS = time.Now().UTC()
	}
	res, err := s.db.Exec(`
INSERT OR IGNORE INTO events(ts,type,message,route,correlation_id,agent_id,outcome,source,raw,file_key)
VALUES(?,?,?,?,?,?,?,?,?,?)`,
		ev.TS.UTC().Format(time.RFC3339Nano), ev.Type, ev.Message, ev.Route,
		ev.CorrelationID, ev.AgentID, ev.Outcome, ev.Source, string(raw), fileKey)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *Store) Feed(filter string, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `SELECT id,ts,type,message,route,correlation_id,agent_id,outcome,source FROM events`
	args := []any{}
	if filter != "" && filter != "todos" && filter != "all" {
		q += ` WHERE type LIKE ?`
		args = append(args, filter+"%")
	}
	q += ` ORDER BY ts DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var ev Event
		var ts string
		if err := rows.Scan(&ev.ID, &ts, &ev.Type, &ev.Message, &ev.Route, &ev.CorrelationID, &ev.AgentID, &ev.Outcome, &ev.Source); err != nil {
			return nil, err
		}
		ev.TS, _ = time.Parse(time.RFC3339Nano, ts)
		ev.V = 1
		out = append(out, ev)
	}
	return out, rows.Err()
}

func (s *Store) CountSince(prefix string, since time.Time) (int, error) {
	q := `SELECT COUNT(*) FROM events WHERE ts >= ?`
	args := []any{since.UTC().Format(time.RFC3339Nano)}
	if prefix != "" {
		q += ` AND type LIKE ?`
		args = append(args, prefix+"%")
	}
	var n int
	err := s.db.QueryRow(q, args...).Scan(&n)
	return n, err
}

func (s *Store) GateStats(since time.Time) (passed, failed int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM events WHERE ts >= ? AND type LIKE 'gates.passed%'`, since.UTC().Format(time.RFC3339Nano)).Scan(&passed)
	if err != nil {
		return
	}
	err = s.db.QueryRow(`SELECT COUNT(*) FROM events WHERE ts >= ? AND type LIKE 'gates.failed%'`, since.UTC().Format(time.RFC3339Nano)).Scan(&failed)
	return
}

func (s *Store) RecentGates(limit int) ([]Gate, error) {
	rows, err := s.db.Query(`
SELECT type, message, ts FROM events
WHERE type LIKE 'gates.%'
ORDER BY ts DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Gate
	for rows.Next() {
		var typ, msg, ts string
		if err := rows.Scan(&typ, &msg, &ts); err != nil {
			return nil, err
		}
		status := "unknown"
		if strings.Contains(typ, "passed") {
			status = "ok"
		} else if strings.Contains(typ, "failed") {
			status = "fail"
		}
		slug := strings.TrimPrefix(typ, "gates.")
		slug = strings.TrimSuffix(slug, ".passed")
		slug = strings.TrimSuffix(slug, ".failed")
		out = append(out, Gate{
			Slug: slug, Name: slug, Status: status, Summary: msg, TS: ts, Duration: "",
		})
	}
	return out, rows.Err()
}

func NormalizeType(source, action, signalType, outcome string) string {
	a := strings.ToLower(action)
	st := strings.ToLower(signalType)
	switch {
	case strings.HasPrefix(a, "signal.consult") || st == "consult":
		return "consultation.request"
	case strings.HasPrefix(a, "signal.attract") || st == "attract":
		return "consultation.attract"
	case strings.HasPrefix(a, "signal.propose") || st == "propose":
		return "evolution.propose"
	case strings.HasPrefix(a, "signal.evolve") || st == "evolve":
		return "evolution.cycle"
	case strings.Contains(a, "gate") && (outcome == "blocked" || outcome == "denied" || outcome == "error"):
		return "gates.failed"
	case strings.Contains(a, "gate"):
		return "gates.passed"
	case outcome == "blocked" || outcome == "denied":
		return "gates.failed"
	case source == "live" || strings.HasPrefix(a, "session"):
		return "session.event"
	case st != "":
		return "signal." + st
	case a != "":
		return "change." + sanitize(a)
	default:
		return "change.unknown"
	}
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", ".")
	return s
}

func FormatGateRate(passed, failed int) (string, bool) {
	total := passed + failed
	if total == 0 {
		return "—", true
	}
	pct := float64(passed) * 100 / float64(total)
	return fmt.Sprintf("%.1f%%", pct), failed == 0
}
