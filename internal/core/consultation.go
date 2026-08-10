package core

import (
	"fmt"
	"strings"
	"time"
)

// ConsultationResult is the structured consultant output (no free chat).
type ConsultationResult struct {
	Version         string   `yaml:"version" json:"version"`
	TaskID          string   `yaml:"task_id" json:"task_id"`
	ConsultantID    string   `yaml:"consultant_id" json:"consultant_id"`
	Summary         string   `yaml:"summary" json:"summary"`
	Recommendations []string `yaml:"recommendations" json:"recommendations"`
	Blockers        []string `yaml:"blockers,omitempty" json:"blockers,omitempty"`
	SubmittedAt     string   `yaml:"submitted_at" json:"submitted_at"`
}

// ConsultationWriter persists consultation YAML under the task directory.
type ConsultationWriter interface {
	WriteConsultation(taskID string, result *ConsultationResult) (path string, err error)
}

// SubmitConsultation validates and stores a consultant opinion for the executor.
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
	c, _, err := s.Store.Get(taskID)
	if err != nil {
		return nil, "", err
	}
	if IsTerminal(c.State) {
		return nil, "", errf("EXECUTION.TERMINAL_STATE_IMMUTABLE",
			fmt.Sprintf("task is terminal (%s)", c.State),
			map[string]any{"task_id": taskID})
	}
	if c.Limits.MaxConsultations == 0 || c.Counters.Consultations >= c.Limits.MaxConsultations {
		return nil, "", errf("EXECUTION.CONSULTATION_LIMIT_REACHED",
			"max_consultations reached for this work class",
			map[string]any{"task_id": taskID, "max": c.Limits.MaxConsultations})
	}
	allowed := false
	for _, x := range c.Participants.Consultants {
		if x == consultantID {
			allowed = true
			break
		}
	}
	if !allowed && consultantID != c.PrimaryExecutor {
		// Allow listed consultants only; primary is not a consultant channel.
		return nil, "", errf("EXECUTION.CONSULTANT_NOT_PARTICIPANT",
			"consultant_id is not a participant on this contract",
			map[string]any{"task_id": taskID, "consultant_id": consultantID})
	}
	if consultantID == c.PrimaryExecutor {
		return nil, "", errf("EXECUTION.EXECUTOR_CANNOT_SUBMIT_CONSULTATION",
			"primary_executor cannot submit consultation as consultant",
			map[string]any{"task_id": taskID})
	}
	if s.Consultations == nil {
		return nil, "", errf("STATE.CONSULTATION_STORE_UNAVAILABLE", "consultation writer not configured", nil)
	}
	before := *c
	res := &ConsultationResult{
		Version:         "1.0",
		TaskID:          taskID,
		ConsultantID:    consultantID,
		Summary:         summary,
		Recommendations: append([]string{}, recommendations...),
		Blockers:        append([]string{}, blockers...),
		SubmittedAt:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	path, err := s.Consultations.WriteConsultation(taskID, res)
	if err != nil {
		return nil, "", wrapStore(err)
	}
	c.Counters.Consultations++
	if _, err := s.Store.Save(c); err != nil {
		if rerr := removeFileIfExists(path); rerr != nil {
			return res, path, errf("STATE.ROLLBACK_FAILED", rerr.Error(), map[string]any{"task_id": taskID, "cause": err.Error()})
		}
		return res, path, wrapStore(err)
	}
	if err := s.emitCorrelated(c, "consultation.submitted", map[string]any{
		"consultant_id": consultantID,
		"path":          path,
	}); err != nil {
		if rerr := s.rollbackConsultation(&before, path); rerr != nil {
			return res, path, errf("STATE.ROLLBACK_FAILED", rerr.Error(), map[string]any{"task_id": taskID, "kind": "consultation.submitted", "cause": err.Error()})
		}
		return res, path, errf("STATE.EVENT_APPEND_FAILED", err.Error(), map[string]any{"task_id": taskID, "kind": "consultation.submitted"})
	}
	return res, path, nil
}

func (s *TaskService) rollbackConsultation(before *Contract, path string) error {
	var errs []string
	if err := removeFileIfExists(path); err != nil {
		errs = append(errs, "remove consultation: "+err.Error())
	}
	if before != nil {
		if _, err := s.Store.Save(before); err != nil {
			errs = append(errs, "restore contract: "+err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(errs, "; "))
}
