// Package mcp serves arah-core use cases over MCP stdio (kern ADR-0007 pattern).
// Agents talk to a stable tool contract; CLI and MCP share TaskService.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/envelope"
	"github.com/sraphaz/arah-harness/internal/evidence"
)

// Server exposes arah-core TaskService (and optional Evidence builder) over MCP stdio.
type Server struct {
	Tasks    *core.TaskService
	Evidence *evidence.Builder
	Version  string
	Reader   io.Reader
	Writer   io.Writer
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Run serves MCP JSON-RPC requests line-delimited on Reader/Writer until EOF.
func (s *Server) Run() error {
	if s.Reader == nil {
		s.Reader = os.Stdin
	}
	if s.Writer == nil {
		s.Writer = os.Stdout
	}
	sc := bufio.NewScanner(s.Reader)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.write(rpcResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &rpcError{Code: -32700, Message: "parse error"},
			})
			continue
		}
		s.handle(req)
	}
	return sc.Err()
}

func (s *Server) write(resp rpcResponse) {
	b, _ := json.Marshal(resp)
	fmt.Fprintf(s.Writer, "%s\n", b)
}

func (s *Server) handle(req rpcRequest) {
	switch req.Method {
	case "initialize":
		s.write(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]any{"name": "arah", "version": s.Version},
			"capabilities":    map[string]any{"tools": map[string]any{}},
		}})
	case "notifications/initialized", "initialized":
		// no response for notifications (id may be null)
		if req.ID != nil {
			s.write(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}})
		}
	case "tools/list":
		s.write(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": toolDefs()}})
	case "tools/call":
		var p toolCallParams
		_ = json.Unmarshal(req.Params, &p)
		result, isErr := s.callTool(p.Name, p.Arguments)
		content := []map[string]any{{"type": "text", "text": mustJSON(result)}}
		s.write(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"content": content,
			"isError": isErr,
		}})
	case "ping":
		s.write(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}})
	default:
		s.write(rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{
			Code:    -32601,
			Message: "method not found: " + req.Method,
		}})
	}
}

func toolDefs() []map[string]any {
	return []map[string]any{
		tool("arah_get_capabilities", "List harness runtime capabilities of this arah-core binary", map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}),
		tool("arah_get_task", "Get an execution contract by task_id", map[string]any{
			"type":     "object",
			"required": []string{"task_id"},
			"properties": map[string]any{
				"task_id": map[string]any{"type": "string"},
			},
		}),
		tool("arah_create_task", "Create and start an execution task", map[string]any{
			"type":     "object",
			"required": []string{"objective"},
			"properties": map[string]any{
				"objective":   map[string]any{"type": "string"},
				"area":        map[string]any{"type": "string"},
				"work_class":  map[string]any{"type": "string"},
				"intent_type": map[string]any{"type": "string"},
				"dry_run":     map[string]any{"type": "boolean"},
			},
		}),
		tool("arah_complete_task", "Complete a task with concrete evidence", map[string]any{
			"type":     "object",
			"required": []string{"task_id", "evidence"},
			"properties": map[string]any{
				"task_id":  map[string]any{"type": "string"},
				"evidence": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"dry_run":  map[string]any{"type": "boolean"},
			},
		}),
		tool("arah_block_task", "Block a task with a concrete reason", map[string]any{
			"type":     "object",
			"required": []string{"task_id", "reason"},
			"properties": map[string]any{
				"task_id": map[string]any{"type": "string"},
				"reason":  map[string]any{"type": "string"},
				"dry_run": map[string]any{"type": "boolean"},
			},
		}),
		tool("arah_get_timeline", "List append-only runtime events for a task", map[string]any{
			"type":     "object",
			"required": []string{"task_id"},
			"properties": map[string]any{
				"task_id": map[string]any{"type": "string"},
			},
		}),
		tool("arah_get_task_context", "Progressive-disclosure task context with token budget (minimal|standard|full)", map[string]any{
			"type":     "object",
			"required": []string{"task_id"},
			"properties": map[string]any{
				"task_id": map[string]any{"type": "string"},
				"budget":  map[string]any{"type": "string", "description": "minimal|standard|full"},
			},
		}),
		tool("arah_explain_route", "Explain choreography routing for an area (model-callable harness API)", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"area":      map[string]any{"type": "string"},
				"preferred": map[string]any{"type": "string"},
			},
		}),
		tool("arah_get_evidence", "Explain Evidence Graph slice for one task", map[string]any{
			"type":     "object",
			"required": []string{"task_id"},
			"properties": map[string]any{
				"task_id": map[string]any{"type": "string"},
			},
		}),
		tool("arah_submit_consultation", "Submit structured consultant opinion (YAML artifact)", map[string]any{
			"type":     "object",
			"required": []string{"task_id", "consultant_id", "summary"},
			"properties": map[string]any{
				"task_id":         map[string]any{"type": "string"},
				"consultant_id":   map[string]any{"type": "string"},
				"summary":         map[string]any{"type": "string"},
				"recommendations": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"blockers":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			},
		}),
		tool("arah_get_evidence_graph", "Deterministic Evidence Graph from specs/tasks (no LLM)", map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}),
	}
}

func tool(name, desc string, schema map[string]any) map[string]any {
	return map[string]any{
		"name":        name,
		"description": desc,
		"inputSchema": schema,
	}
}

func (s *Server) callTool(name string, args map[string]any) (envelope.Envelope, bool) {
	switch name {
	case "arah_get_capabilities":
		return envelope.OK(map[string]any{
			"runtime":     "arah-core",
			"version":     s.Version,
			"surfaces":    []string{"cli", "mcp"},
			"commands":    []string{"doctor", "sync-check", "version", "task", "evidence", "economy", "mcp", "kernel"},
			"state_store": "sqlite-wal",
			"context_budget": []string{"minimal", "standard", "full"},
			"inspired_by": []string{
				"https://github.com/rafaelnicolett/kern",
				"https://github.com/NVIDIA-NeMo/labs-OO-Agents",
			},
		}), false
	case "arah_get_task":
		id, _ := args["task_id"].(string)
		c, path, err := s.Tasks.Get(id)
		return mapResult(c, path, err, "", false, false)
	case "arah_create_task":
		obj, _ := args["objective"].(string)
		area, _ := args["area"].(string)
		wc, _ := args["work_class"].(string)
		it, _ := args["intent_type"].(string)
		opts := core.MutateOptions{DryRun: boolArg(args["dry_run"])}
		res, err := s.Tasks.Create(obj, area, core.WorkClass(wc), core.IntentType(it), opts)
		return mapMutation(res, err)
	case "arah_complete_task":
		id, _ := args["task_id"].(string)
		ev := stringList(args["evidence"])
		opts := core.MutateOptions{DryRun: boolArg(args["dry_run"])}
		res, err := s.Tasks.Complete(id, ev, opts)
		return mapMutation(res, err)
	case "arah_block_task":
		id, _ := args["task_id"].(string)
		reason, _ := args["reason"].(string)
		opts := core.MutateOptions{DryRun: boolArg(args["dry_run"])}
		res, err := s.Tasks.Block(id, reason, opts)
		return mapMutation(res, err)
	case "arah_get_timeline":
		id, _ := args["task_id"].(string)
		evs, err := s.Tasks.Timeline(id)
		if err != nil {
			return domainToEnvelope(err), true
		}
		return envelope.OK(map[string]any{"task_id": id, "events": evs}), false
	case "arah_get_task_context":
		id, _ := args["task_id"].(string)
		budget, _ := args["budget"].(string)
		tc, err := s.Tasks.Context(id, core.ParseContextBudget(budget))
		if err != nil {
			return domainToEnvelope(err), true
		}
		return envelope.OK(tc), false
	case "arah_explain_route":
		area, _ := args["area"].(string)
		preferred, _ := args["preferred"].(string)
		data, err := s.Tasks.ExplainRoute(area, preferred)
		if err != nil {
			return domainToEnvelope(err), true
		}
		return envelope.OK(data), false
	case "arah_get_evidence":
		id, _ := args["task_id"].(string)
		if s.Evidence == nil {
			return envelope.Fail(envelope.CodeInternal, "evidence builder not configured", nil), true
		}
		data, err := s.Evidence.Explain(id)
		if err != nil {
			return domainToEnvelope(err), true
		}
		return envelope.OK(data), false
	case "arah_submit_consultation":
		id, _ := args["task_id"].(string)
		cid, _ := args["consultant_id"].(string)
		summary, _ := args["summary"].(string)
		res, path, err := s.Tasks.SubmitConsultation(id, cid, summary, stringList(args["recommendations"]), stringList(args["blockers"]))
		if err != nil {
			return domainToEnvelope(err), true
		}
		return envelope.OK(map[string]any{"consultation": res, "path": path}), false
	case "arah_get_evidence_graph":
		if s.Evidence == nil {
			return envelope.Fail(envelope.CodeInternal, "evidence builder not configured", nil), true
		}
		g, err := s.Evidence.Build()
		if err != nil {
			return envelope.Fail(envelope.CodeInternal, err.Error(), nil), true
		}
		return envelope.OK(g), false
	default:
		return envelope.Fail(envelope.CodeUsage, "unknown tool: "+name, nil), true
	}
}

func mapMutation(res *core.MutationResult, err error) (envelope.Envelope, bool) {
	if err != nil {
		return domainToEnvelope(err), true
	}
	return mapResult(res.Contract, res.Path, nil, res.Diff, res.Idempotent, res.DryRun)
}

func mapResult(c *core.Contract, path string, err error, diff string, idempotent, dryRun bool) (envelope.Envelope, bool) {
	if err != nil {
		return domainToEnvelope(err), true
	}
	return envelope.OK(map[string]any{
		"task_id":           c.TaskID,
		"state":             c.State,
		"primary_executor":  c.PrimaryExecutor,
		"objective":         c.Objective,
		"path":              path,
		"dry_run":           dryRun || strings.HasPrefix(path, "dry-run"),
		"idempotent":        idempotent,
		"diff":              diff,
		"blocking_reason":   c.Result.BlockingReason,
		"evidence":          c.Execution.CompletionEvidence,
		"choreography_rule": c.ChoreographyRule,
	}), false
}

func domainToEnvelope(err error) envelope.Envelope {
	if de, ok := err.(*core.DomainError); ok {
		return envelope.Fail(de.Code, de.Message, de.Details, de.Remediation...)
	}
	return envelope.Fail(envelope.CodeInternal, err.Error(), nil)
}

func stringList(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	default:
		return nil
	}
}

func boolArg(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1" || t == "yes"
	default:
		return false
	}
}

func mustJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
