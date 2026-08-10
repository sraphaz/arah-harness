// Package envelope defines the stable CLI/MCP JSON response shape (H-14).
// Deliberately stricter than kern's informal PREFIX.CODE strings (ADR-0007 note).
package envelope

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Envelope is the machine-parseable result of every arah-core command.
type Envelope struct {
	OK          bool           `json:"ok"`
	Code        string         `json:"code,omitempty"`
	Message     string         `json:"message,omitempty"`
	TraceID     string         `json:"trace_id"`
	Details     map[string]any `json:"details,omitempty"`
	Remediation []string       `json:"remediation,omitempty"`
	Data        any            `json:"data,omitempty"`
}

// Error codes — BOUNDED_CONTEXT.CODE (kern-inspired, structured).
const (
	CodeOK                            = "OK"
	CodeUsage                         = "CLI.USAGE"
	CodeNotFound                      = "EXECUTION.TASK_NOT_FOUND"
	CodeInvalidTransition             = "EXECUTION.INVALID_STATE_TRANSITION"
	CodeEvidenceRequired              = "EXECUTION.COMPLETION_EVIDENCE_REQUIRED"
	CodeBlockingReasonRequired        = "EXECUTION.BLOCKING_REASON_REQUIRED"
	CodeTerminalImmutable             = "EXECUTION.TERMINAL_STATE_IMMUTABLE"
	CodePrimaryExecutorRequired       = "EXECUTION.PRIMARY_EXECUTOR_REQUIRED"
	CodeExactlyOnePrimary             = "EXECUTION.EXACTLY_ONE_PRIMARY_EXECUTOR_REQUIRED"
	CodeRerouteForbidden              = "EXECUTION.REROUTE_AFTER_EXECUTING_FORBIDDEN"
	CodeInternal                      = "RUNTIME.INTERNAL"
	CodeStore                         = "STATE.STORE_ERROR"
)

// NewTraceID returns a short unique id (hex timestamp + random).
func NewTraceID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), hex.EncodeToString(b[:]))
}

func OK(data any) Envelope {
	return Envelope{OK: true, Code: CodeOK, TraceID: NewTraceID(), Data: data}
}

func Fail(code, message string, details map[string]any, remediation ...string) Envelope {
	return Envelope{
		OK:          false,
		Code:        code,
		Message:     message,
		TraceID:     NewTraceID(),
		Details:     details,
		Remediation: remediation,
	}
}

// WriteJSON writes the envelope to w and returns a suggested process exit code.
func WriteJSON(w io.Writer, env Envelope) (exitCode int) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(env)
	if env.OK {
		return 0
	}
	switch env.Code {
	case CodeUsage:
		return 10
	default:
		return 1
	}
}
