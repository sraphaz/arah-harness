// Package mcp serves arah-core use cases over MCP stdio (kern ADR-0007 pattern).
// Agents talk to a stable tool contract; CLI and MCP share TaskService.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sraphaz/arah-harness/internal/core"
	"github.com/sraphaz/arah-harness/internal/envelope"
)

type Server struct {
	Tasks   *core.TaskService
	Version string
	Reader  io.Reader
	Writer  io.Writer
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
			},
		}),
		tool("arah_complete_task", "Complete a task with concrete evidence", map[string]any{
			"type":     "object",
			"required": []string{"task_id", "evidence"},
			"properties": map[string]any{
				"task_id":  map[string]any{"type": "string"},
				"evidence": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			},
		}),
		tool("arah_block_task", "Block a task with a concrete reason", map[string]any{
			"type":     "object",
			"required": []string{"task_id", "reason"},
			"properties": map[string]any{
				"task_id": map[string]any{"type": "string"},
				"reason":  map[string]any{"type": "string"},
			},
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
			"runtime": "arah-core",
			"version": s.Version,
			"surfaces": []string{"cli", "mcp"},
			"commands": []string{"doctor", "sync-check", "version", "task", "mcp"},
			"inspired_by": "https://github.com/rafaelnicolett/kern",
		}), false
	case "arah_get_task":
		id, _ := args["task_id"].(string)
		c, path, err := s.Tasks.Get(id)
		return mapResult(c, path, err)
	case "arah_create_task":
		obj, _ := args["objective"].(string)
		area, _ := args["area"].(string)
		wc, _ := args["work_class"].(string)
		it, _ := args["intent_type"].(string)
		c, path, err := s.Tasks.Create(obj, area, core.WorkClass(wc), core.IntentType(it))
		return mapResult(c, path, err)
	case "arah_complete_task":
		id, _ := args["task_id"].(string)
		ev := stringList(args["evidence"])
		c, path, err := s.Tasks.Complete(id, ev)
		return mapResult(c, path, err)
	case "arah_block_task":
		id, _ := args["task_id"].(string)
		reason, _ := args["reason"].(string)
		c, path, err := s.Tasks.Block(id, reason)
		return mapResult(c, path, err)
	default:
		return envelope.Fail(envelope.CodeUsage, "unknown tool: "+name, nil), true
	}
}

func mapResult(c *core.Contract, path string, err error) (envelope.Envelope, bool) {
	if err != nil {
		return domainToEnvelope(err), true
	}
	return envelope.OK(map[string]any{
		"task_id":          c.TaskID,
		"state":            c.State,
		"primary_executor": c.PrimaryExecutor,
		"objective":        c.Objective,
		"path":             path,
		"blocking_reason":  c.Result.BlockingReason,
		"evidence":         c.Execution.CompletionEvidence,
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

func mustJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
