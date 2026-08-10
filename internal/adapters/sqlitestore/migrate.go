package sqlitestore

import (
	"database/sql"
	"fmt"
	"strconv"
)

// schemaVersion is the latest applied schema. Bump when adding a migration.
const schemaVersion = 2

type migration struct {
	Version int
	Name    string
	Up      func(tx *sql.Tx) error
	Down    func(tx *sql.Tx) error
}

// migrations are applied in order for versions 1..schemaVersion.
var migrations = []migration{
	{
		Version: 1,
		Name:    "baseline_tasks_events",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
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
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
DROP INDEX IF EXISTS idx_events_task;
DROP TABLE IF EXISTS task_events;
DROP INDEX IF EXISTS idx_tasks_bucket;
DROP TABLE IF EXISTS tasks;
`)
			return err
		},
	},
	{
		Version: 2,
		Name:    "events_kind_index",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_events_kind ON task_events(kind);`)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec(`DROP INDEX IF EXISTS idx_events_kind;`)
			return err
		},
	},
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);`); err != nil {
		return err
	}
	cur, err := readSchemaVersion(db)
	if err != nil {
		return err
	}
	if cur > schemaVersion {
		return fmt.Errorf("database schema version %d is newer than supported %d", cur, schemaVersion)
	}
	for _, m := range migrations {
		if m.Version <= cur {
			continue
		}
		if m.Version > schemaVersion {
			break
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if err := m.Up(tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migrate up v%d (%s): %w", m.Version, m.Name, err)
		}
		if err := writeSchemaVersion(tx, m.Version); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func readSchemaVersion(db *sql.DB) (int, error) {
	var raw string
	err := db.QueryRow(`SELECT value FROM schema_meta WHERE key = 'version'`).Scan(&raw)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("schema version %q: %w", raw, err)
	}
	return v, nil
}

func writeSchemaVersion(tx *sql.Tx, v int) error {
	_, err := tx.Exec(`
INSERT INTO schema_meta(key, value) VALUES('version', ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value
`, strconv.Itoa(v))
	return err
}

// SchemaVersion returns the schema version recorded in schema_meta.
func (s *Store) SchemaVersion() (int, error) {
	if err := s.EnsureLayout(); err != nil {
		return 0, err
	}
	return readSchemaVersion(s.db)
}

// RollbackTo reverts schema migrations down to targetVersion (inclusive floor).
// Intended for tests and controlled recovery; not a routine CLI path.
func (s *Store) RollbackTo(targetVersion int) error {
	if err := s.EnsureLayout(); err != nil {
		return err
	}
	if targetVersion < 0 || targetVersion > schemaVersion {
		return fmt.Errorf("target version %d out of range 0..%d", targetVersion, schemaVersion)
	}
	cur, err := readSchemaVersion(s.db)
	if err != nil {
		return err
	}
	for v := cur; v > targetVersion; v-- {
		m, ok := migrationByVersion(v)
		if !ok {
			return fmt.Errorf("no down migration for version %d", v)
		}
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if err := m.Down(tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migrate down v%d (%s): %w", m.Version, m.Name, err)
		}
		if err := writeSchemaVersion(tx, v-1); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func migrationByVersion(v int) (migration, bool) {
	for _, m := range migrations {
		if m.Version == v {
			return m, true
		}
	}
	return migration{}, false
}
