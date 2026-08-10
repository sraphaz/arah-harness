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

type Store struct {
	RepoRoot string
}

func New(repoRoot string) *Store {
	return &Store{RepoRoot: repoRoot}
}

func (s *Store) root() string {
	return filepath.Join(s.RepoRoot, ".arah", "local", "execution")
}

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

func (s *Store) Save(c *core.Contract) (string, error) {
	if err := s.EnsureLayout(); err != nil {
		return "", err
	}
	bucket := s.bucket(c.State)
	for _, b := range []string{"active", "completed", "blocked"} {
		old := filepath.Join(s.root(), b, c.TaskID+".yaml")
		if b != bucket {
			_ = os.Remove(old)
		}
	}
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
	consult := filepath.Join(s.root(), c.TaskID, "consultations")
	_ = os.MkdirAll(consult, 0o755)
	return dest, nil
}

func (s *Store) Get(taskID string) (*core.Contract, string, error) {
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
