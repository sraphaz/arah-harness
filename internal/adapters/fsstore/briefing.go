package fsstore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/sraphaz/arah-harness/internal/core"
)

// WriteBriefing persists BRIEFING.md under .arah/local/execution/<task-id>/.
func (s *Store) WriteBriefing(c *core.Contract) (string, error) {
	if c == nil || strings.TrimSpace(c.TaskID) == "" {
		return "", fmt.Errorf("contract task_id required")
	}
	dir := filepath.Join(s.root(), c.TaskID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "BRIEFING.md")
	content := core.RenderBriefing(c)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return path, nil
}

// WriteConsultation persists a structured consultant opinion YAML.
func (s *Store) WriteConsultation(taskID string, result *core.ConsultationResult) (string, error) {
	if result == nil {
		return "", fmt.Errorf("consultation result required")
	}
	dir := filepath.Join(s.root(), taskID, "consultations")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	stamp := time.Now().UTC().Format("20060102-150405.000")
	stamp = strings.ReplaceAll(stamp, ".", "")
	name := fmt.Sprintf("%s-%s.yaml", stamp, sanitizeID(result.ConsultantID))
	path := filepath.Join(dir, name)
	raw, err := yaml.Marshal(result)
	if err != nil {
		return "", err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return path, nil
}

// RemoveConsultation deletes a consultation YAML artifact (create/consultation rollback).
func (s *Store) RemoveConsultation(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	// Only allow deletes under this store's execution root.
	root, err := filepath.Abs(s.root())
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if abs != root && !strings.HasPrefix(abs, root+string(filepath.Separator)) {
		return fmt.Errorf("consultation path outside execution root")
	}
	return os.Remove(abs)
}

func sanitizeID(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	s = strings.ReplaceAll(s, " ", "-")
	if s == "" {
		return "consultant"
	}
	return s
}
