// Package fsstore is the filesystem StateStore adapter (migration path before SQLite).
package fsstore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sraphaz/arah-harness/internal/core"
)

// Store persists Execution Control contracts as YAML under .arah/local/execution/.
type Store struct {
	RepoRoot string
}

// New returns a filesystem StateStore rooted at repoRoot.
func New(repoRoot string) *Store {
	return &Store{RepoRoot: repoRoot}
}

func (s *Store) root() string {
	return filepath.Join(s.RepoRoot, ".arah", "local", "execution")
}

// EnsureLayout creates active/completed/blocked directories.
func (s *Store) EnsureLayout() error {
	for _, b := range []string{"active", "completed", "blocked"} {
		if err := os.MkdirAll(filepath.Join(s.root(), b), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) bucket(state core.State) string {
	switch state {
	case core.StateDone:
		return "completed"
	case core.StateBlocked:
		return "blocked"
	default:
		return "active"
	}
}

// Save writes the contract into the bucket for its state (write-then-delete across buckets).
func (s *Store) Save(c *core.Contract) (string, error) {
	if err := s.EnsureLayout(); err != nil {
		return "", err
	}
	bucket := s.bucket(c.State)
	dest := filepath.Join(s.root(), bucket, c.TaskID+".yaml")
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	// Remove stale copies only after the new bucket file is durable.
	for _, b := range []string{"active", "completed", "blocked"} {
		if b == bucket {
			continue
		}
		_ = os.Remove(filepath.Join(s.root(), b, c.TaskID+".yaml"))
	}
	consult := filepath.Join(s.root(), c.TaskID, "consultations")
	_ = os.MkdirAll(consult, 0o755)
	return dest, nil
}

// Delete removes a task contract and task-scoped artifacts from the filesystem store.
func (s *Store) Delete(taskID string) error {
	if err := s.EnsureLayout(); err != nil {
		return err
	}
	var firstErr error
	for _, b := range []string{"active", "completed", "blocked"} {
		p := filepath.Join(s.root(), b, taskID+".yaml")
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) && firstErr == nil {
			firstErr = err
		}
	}
	dir := filepath.Join(s.root(), taskID)
	if err := os.RemoveAll(dir); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

// Get loads a contract by scanning active, completed, then blocked buckets.
func (s *Store) Get(taskID string) (*core.Contract, string, error) {
	return s.Peek(taskID)
}

// Peek is identical to Get for the filesystem adapter (no reconcile side effects).
func (s *Store) Peek(taskID string) (*core.Contract, string, error) {
	if err := s.EnsureLayout(); err != nil {
		return nil, "", err
	}
	for _, b := range []string{"active", "completed", "blocked"} {
		p := filepath.Join(s.root(), b, taskID+".yaml")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			raw, err := os.ReadFile(p)
			if err != nil {
				return nil, "", err
			}
			var c core.Contract
			if err := yaml.Unmarshal(raw, &c); err != nil {
				return nil, "", fmt.Errorf("parse %s: %w", p, err)
			}
			return &c, p, nil
		}
	}
	return nil, "", &core.DomainError{
		Code:    "EXECUTION.TASK_NOT_FOUND",
		Message: "task not found: " + taskID,
		Details: map[string]any{"task_id": taskID},
	}
}

// List returns contracts present in the given bucket directory.
func (s *Store) List(bucket string) ([]*core.Contract, error) {
	if err := s.EnsureLayout(); err != nil {
		return nil, err
	}
	if bucket == "" {
		bucket = "active"
	}
	dir := filepath.Join(s.root(), bucket)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []*core.Contract
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var c core.Contract
		if yaml.Unmarshal(raw, &c) == nil {
			out = append(out, &c)
		}
	}
	return out, nil
}
