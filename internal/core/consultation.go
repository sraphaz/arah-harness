package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// ConsultationResult is the structured consultant output (no free chat).
type ConsultationResult struct {
	Version         string   `yaml:"version" json:"version"`
	ID              string   `yaml:"id" json:"id"`
	TaskID          string   `yaml:"task_id" json:"task_id"`
	ConsultantID    string   `yaml:"consultant_id" json:"consultant_id"`
	Summary         string   `yaml:"summary" json:"summary"`
	Recommendations []string `yaml:"recommendations" json:"recommendations"`
	Blockers        []string `yaml:"blockers,omitempty" json:"blockers,omitempty"`
	SubmittedAt     string   `yaml:"submitted_at" json:"submitted_at"`
}

// SubmitConsultation validates and stores a consultant opinion for the executor.
// Slot reservation prefers ConsultationReserver (SQLite BEGIN IMMEDIATE) so CLI
// and MCP cannot both pass MaxConsultations across processes; falls back to
// mutex + Save for stores without that capability.
func (s *TaskService) SubmitConsultation(taskID, consultantID, summary string, recommendations, blockers []string) (*ConsultationResult, string, error) {
	taskID = strings.TrimSpace(taskID)
	consultantID = strings.TrimSpace(consultantID)
	summary = strings.TrimSpace(summary)
	if taskID == "" {
		return nil, "", errf("EXECUTION.TASK_ID_REQUIRED", "task_id is required", nil)
	}
	if consultantID == "" {
		return nil, "", errf("EXECUTION.CONSULTANT_REQUIRED", "consultant_id is required", nil)
	}
	if summary == "" {
		return nil, "", errf("EXECUTION.CONSULTATION_SUMMARY_REQUIRED", "consultation summary is required", nil)
	}
	if s.Consultations == nil {
		return nil, "", errf("STATE.CONSULTATION_STORE_UNAVAILABLE", "consultation writer not configured", nil)
	}

	s.consultMu.Lock()
	defer s.consultMu.Unlock()

	c, _, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	if IsTerminal(c.State) {
		return nil, "", errf("EXECUTION.TERMINAL_STATE_IMMUTABLE",
			fmt.Sprintf("task is terminal (%s)", c.State),
			map[string]any{"task_id": taskID})
	}
	allowed := false
	for _, x := range c.Participants.Consultants {
		if x == consultantID {
			allowed = true
			break
		}
	}
	if !allowed && consultantID != c.PrimaryExecutor {
		return nil, "", errf("EXECUTION.CONSULTANT_NOT_PARTICIPANT",
			"consultant_id is not a participant on this contract",
			map[string]any{"task_id": taskID, "consultant_id": consultantID})
	}
	if consultantID == c.PrimaryExecutor {
		return nil, "", errf("EXECUTION.EXECUTOR_CANNOT_SUBMIT_CONSULTATION",
			"primary_executor cannot submit consultation as consultant",
			map[string]any{"task_id": taskID})
	}

	max := c.Limits.MaxConsultations
	if r, ok := s.Store.(ConsultationReserver); ok {
		c, err = r.ReserveConsultationSlot(taskID, max)
		if err != nil {
			return nil, "", err
		}
	} else {
		if max == 0 || c.Counters.Consultations >= max {
			return nil, "", errf("EXECUTION.CONSULTATION_LIMIT_REACHED",
				"max_consultations reached for this work class",
				map[string]any{"task_id": taskID, "max": max})
		}
		c.Counters.Consultations++
		if _, err := s.Store.Save(c); err != nil {
			return nil, "", wrapStore(err)
		}
	}

	res := &ConsultationResult{
		Version:         "1.0",
		ID:              newConsultationID(),
		TaskID:          taskID,
		ConsultantID:    consultantID,
		Summary:         summary,
		Recommendations: append([]string{}, recommendations...),
		Blockers:        append([]string{}, blockers...),
		SubmittedAt:     time.Now().UTC().Format(time.RFC3339Nano),
	}

	path, err := s.Consultations.WriteConsultation(taskID, res)
	if err != nil {
		_ = s.rollbackConsultationSlot(c)
		return nil, "", wrapStore(err)
	}
	if err := s.emitCorrelated(c, "consultation.submitted", map[string]any{
		"consultant_id": consultantID,
		"path":          path,
		"id":            res.ID,
	}); err != nil {
		_ = s.Consultations.RemoveConsultation(path)
		_ = s.rollbackConsultationSlot(c)
		return nil, "", errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": taskID, "kind": "consultation.submitted"})
	}
	return res, path, nil
}

func (s *TaskService) rollbackConsultationSlot(c *Contract) error {
	if r, ok := s.Store.(ConsultationReserver); ok {
		_, err := r.ReleaseConsultationSlot(c.TaskID)
		return err
	}
	if c.Counters.Consultations > 0 {
		c.Counters.Consultations--
	}
	_, err := s.Store.Save(c)
	return err
}

func newConsultationID() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("c-%d", time.Now().UnixNano())
	}
	return "c-" + hex.EncodeToString(b[:])
}
