// Package sqlitestore is the SQLite WAL StateStore + EventStore (H-16).
// Pure Go driver (modernc.org/sqlite) — no CGO, kern-inspired embedded DB.
package sqlitestore

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"

	"github.com/sraphaz/arah-harness/internal/adapters/fsstore"
	"github.com/sraphaz/arah-harness/internal/core"
)

const schemaVersion = 1

// Store is the SQLite-backed StateStore and EventStore for a single repository.
// Path: <repo>/.arah/local/runtime.db (WAL). YAML under execution/ is a best-effort mirror.
type Store struct {
	RepoRoot string
	DBPath   string
	db       *sql.DB
	fs       *fsstore.Store // migration source / dual-read fallback
}

// New opens (or creates) the repository runtime database and imports YAML tasks if empty.
func New(repoRoot string) (*Store, error) {
	s := &Store{
		RepoRoot: repoRoot,
		DBPath:   filepath.Join(repoRoot, ".arah", "local", "runtime.db"),
		fs:       fsstore.New(repoRoot),
	}
	if err := s.EnsureLayout(); err != nil {
		return nil, err
	}
	return s, nil
}

// EnsureLayout creates directories and applies schema migrations.
func (s *Store) EnsureLayout() error {
	if err := os.MkdirAll(filepath.Dir(s.DBPath), 0o755); err != nil {
		return err
	}
	if err := s.fs.EnsureLayout(); err != nil {
		return err
	}
	if s.db != nil {
		return nil
	}
	db, err := sql.Open("sqlite", s.DBPath)
	if err != nil {
		return err
	}
	// WAL + busy timeout for CLI/MCP concurrency
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return fmt.Errorf("pragma: %w", err)
		}
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return err
	}
	s.db = db
	return s.importFilesystemIfEmpty()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS tasks (
  task_id TEXT PRIMARY KEY,
  state TEXT NOT NULL,
  bucket TEXT NOT NULL,
  primary_executor TEXT,
  objective TEXT,
  work_class TEXT,
  intent_type TEXT,
  choreography_rule TEXT,
  contract_yaml TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tasks_bucket ON tasks(bucket);
CREATE TABLE IF NOT EXISTS task_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_id TEXT NOT NULL UNIQUE,
  task_id TEXT,
  kind TEXT NOT NULL,
  at TEXT NOT NULL,
  trace_id TEXT,
  payload_json TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_events_task ON task_events(task_id);
`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT OR IGNORE INTO schema_meta(key, value) VALUES('version', ?)`, fmt.Sprintf("%d", schemaVersion))
	return err
}

func (s *Store) importFilesystemIfEmpty() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for _, bucket := range []string{"active", "completed", "blocked"} {
		list, err := s.fs.List(bucket)
		if err != nil {
			return fmt.Errorf("migrate filesystem bucket %s: %w", bucket, err)
		}
		for _, c := range list {
			if _, err := s.saveDB(c); err != nil {
				return fmt.Errorf("migrate %s: %w", c.TaskID, err)
			}
		}
	}
	return nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func bucketOf(state core.State) string {
	switch state {
	case core.StateDone:
		return "completed"
	case core.StateBlocked:
		return "blocked"
	default:
		return "active"
	}
}

// Save upserts the contract in SQLite (canonical) and mirrors YAML best-effort.
func (s *Store) Save(c *core.Contract) (string, error) {
	if err := s.EnsureLayout(); err != nil {
		return "", err
	}
	ref, err := s.saveDB(c)
	if err != nil {
		return "", err
	}
	// YAML mirror is best-effort for PowerShell strangler compatibility.
	// SQLite is the canonical hot store; mirror failure must not undo a committed row
	// or report failure that invites duplicate Create retries.
	if _, err := s.fs.Save(c); err != nil {
		_ = err // intentional: canonical save already succeeded
	}
	return ref, nil
}

func (s *Store) saveDB(c *core.Contract) (string, error) {
	raw, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	bucket := bucketOf(c.State)
	_, err = s.db.Exec(`
INSERT INTO tasks(task_id, state, bucket, primary_executor, objective, work_class, intent_type, choreography_rule, contract_yaml, updated_at)
VALUES(?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(task_id) DO UPDATE SET
  state=excluded.state,
  bucket=excluded.bucket,
  primary_executor=excluded.primary_executor,
  objective=excluded.objective,
  work_class=excluded.work_class,
  intent_type=excluded.intent_type,
  choreography_rule=excluded.choreography_rule,
  contract_yaml=excluded.contract_yaml,
  updated_at=excluded.updated_at
`, c.TaskID, string(c.State), bucket, c.PrimaryExecutor, c.Objective, string(c.WorkClass), string(c.IntentType),
		c.ChoreographyRule, string(raw), time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return "", err
	}
	return "sqlite:" + s.DBPath + "#" + c.TaskID, nil
}

// Get returns a contract from SQLite, falling back to the filesystem mirror.
func (s *Store) Get(taskID string) (*core.Contract, string, error) {
	if err := s.EnsureLayout(); err != nil {
		return nil, "", err
	}
	var yamlBody string
	err := s.db.QueryRow(`SELECT contract_yaml FROM tasks WHERE task_id = ?`, taskID).Scan(&yamlBody)
	if err == sql.ErrNoRows {
		// Fallback to filesystem (pre-migration callers / dual-write lag)
		return s.fs.Get(taskID)
	}
	if err != nil {
		return nil, "", err
	}
	var c core.Contract
	if err := yaml.Unmarshal([]byte(yamlBody), &c); err != nil {
		return nil, "", err
	}
	return &c, "sqlite:" + s.DBPath + "#" + taskID, nil
}

// List returns contracts in a bucket, merging SQLite rows with filesystem-only tasks.
func (s *Store) List(bucket string) ([]*core.Contract, error) {
	if err := s.EnsureLayout(); err != nil {
		return nil, err
	}
	if bucket == "" {
		bucket = "active"
	}
	rows, err := s.db.Query(`SELECT contract_yaml FROM tasks WHERE bucket = ? ORDER BY updated_at DESC`, bucket)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := map[string]bool{}
	var out []*core.Contract
	for rows.Next() {
		var body string
		if err := rows.Scan(&body); err != nil {
			return nil, err
		}
		var c core.Contract
		if err := yaml.Unmarshal([]byte(body), &c); err != nil {
			continue
		}
		seen[c.TaskID] = true
		out = append(out, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Merge filesystem-only tasks (PS writers / pre-import) so List matches Get fallback.
	fsList, err := s.fs.List(bucket)
	if err != nil {
		return nil, fmt.Errorf("list filesystem bucket %s: %w", bucket, err)
	}
	for _, c := range fsList {
		if c == nil || seen[c.TaskID] {
			continue
		}
		seen[c.TaskID] = true
		out = append(out, c)
	}
	return out, nil
}

// Append implements EventStore.
func (s *Store) Append(ev core.Event) error {
	if err := s.EnsureLayout(); err != nil {
		return err
	}
	if ev.ID == "" {
		var b [4]byte
		_, _ = rand.Read(b[:])
		ev.ID = fmt.Sprintf("ev-%d-%x", time.Now().UnixNano(), b)
	}
	if ev.At == "" {
		ev.At = time.Now().UTC().Format(time.RFC3339Nano)
	}
	payload, _ := json.Marshal(ev.Payload)
	if payload == nil {
		payload = []byte("{}")
	}
	_, err := s.db.Exec(`
INSERT INTO task_events(event_id, task_id, kind, at, trace_id, payload_json)
VALUES(?,?,?,?,?,?)`, ev.ID, nullStr(ev.TaskID), ev.Kind, ev.At, nullStr(ev.TraceID), string(payload))
	return err
}

func (s *Store) ListByTask(taskID string) ([]core.Event, error) {
	if err := s.EnsureLayout(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`
SELECT event_id, task_id, kind, at, COALESCE(trace_id,''), payload_json
FROM task_events WHERE task_id = ? ORDER BY id ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (s *Store) ListRecent(limit int) ([]core.Event, error) {
	if err := s.EnsureLayout(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
SELECT event_id, task_id, kind, at, COALESCE(trace_id,''), payload_json
FROM task_events ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func scanEvents(rows *sql.Rows) ([]core.Event, error) {
	var out []core.Event
	for rows.Next() {
		var ev core.Event
		var taskID sql.NullString
		var payload string
		if err := rows.Scan(&ev.ID, &taskID, &ev.Kind, &ev.At, &ev.TraceID, &payload); err != nil {
			return nil, err
		}
		if taskID.Valid {
			ev.TaskID = taskID.String
		}
		_ = json.Unmarshal([]byte(payload), &ev.Payload)
		out = append(out, ev)
	}
	return out, rows.Err()
}

func nullStr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
