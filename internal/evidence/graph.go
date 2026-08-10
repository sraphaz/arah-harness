// Package evidence builds a deterministic Evidence Graph from Arah artifacts (H-18).
// No LLM — only schemas, contracts, specs and runtime events.
package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sraphaz/arah-harness/internal/core"
)

// Node is a vertex in the Evidence Graph.
type Node struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label,omitempty"`
}

// Edge is a typed relation between two nodes.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Rel  string `json:"rel"`
}

// Graph is a deterministic evidence graph export (no LLM).
type Graph struct {
	Version string `json:"version"`
	Nodes   []Node `json:"nodes"`
	Edges   []Edge `json:"edges"`
}

// Builder assembles an Evidence Graph from specs, capabilities, and tasks.
type Builder struct {
	RepoRoot string
	Store    core.StateStore
	Events   core.EventStore
}

// Build derives nodes and edges exclusively from Arah schemas and runtime state.
func (b *Builder) Build() (*Graph, error) {
	g := &Graph{Version: "1"}
	idx := map[string]bool{}
	add := func(n Node) {
		if idx[n.ID] {
			return
		}
		idx[n.ID] = true
		g.Nodes = append(g.Nodes, n)
	}
	link := func(from, to, rel string) {
		g.Edges = append(g.Edges, Edge{From: from, To: to, Rel: rel})
	}

	// Specs → covers paths
	specDir := filepath.Join(b.RepoRoot, "docs", "specs")
	entries, _ := os.ReadDir(specDir)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !(strings.HasSuffix(name, ".spec.yaml") || strings.HasSuffix(name, ".yaml")) {
			continue
		}
		if name == "REGISTRY.md" || strings.HasPrefix(name, "_") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(specDir, name))
		if err != nil {
			continue
		}
		var doc struct {
			ID      string   `yaml:"id"`
			Title   string   `yaml:"title"`
			Covers  []string `yaml:"covers"`
			Status  string   `yaml:"status"`
		}
		if yaml.Unmarshal(raw, &doc) != nil || doc.ID == "" {
			continue
		}
		sid := "spec:" + doc.ID
		add(Node{ID: sid, Type: "spec", Label: doc.Title})
		for _, c := range doc.Covers {
			pid := "path:" + c
			add(Node{ID: pid, Type: "path", Label: c})
			link(sid, pid, "covers")
		}
	}

	// Capabilities
	capPath := filepath.Join(b.RepoRoot, "capabilities.yaml")
	if raw, err := os.ReadFile(capPath); err == nil {
		var caps struct {
			Available    []struct{ ID, Name string } `yaml:"available"`
			Experimental []struct{ ID, Name string } `yaml:"experimental"`
			Planned      []struct{ ID, Name string } `yaml:"planned"`
		}
		if yaml.Unmarshal(raw, &caps) == nil {
			for _, list := range [][]struct{ ID, Name string }{caps.Available, caps.Experimental, caps.Planned} {
				for _, c := range list {
					if c.ID == "" {
						continue
					}
					add(Node{ID: "capability:" + c.ID, Type: "capability", Label: c.Name})
				}
			}
		}
	}

	if b.Store == nil {
		return g, nil
	}
	for _, bucket := range []string{"active", "completed", "blocked"} {
		list, err := b.Store.List(bucket)
		if err != nil {
			continue
		}
		for _, c := range list {
			tid := "task:" + c.TaskID
			add(Node{ID: tid, Type: "task", Label: c.Objective})
			if c.PrimaryExecutor != "" {
				aid := "agent:" + c.PrimaryExecutor
				add(Node{ID: aid, Type: "agent", Label: c.PrimaryExecutor})
				link(tid, aid, "assigned_to")
			}
			for _, cons := range c.Participants.Consultants {
				aid := "agent:" + cons
				add(Node{ID: aid, Type: "agent", Label: cons})
				link(tid, aid, "consulted")
			}
			for _, ev := range c.Execution.CompletionEvidence {
				eid := "evidence:" + stableID(ev)
				add(Node{ID: eid, Type: "evidence", Label: ev})
				link(tid, eid, "evidenced_by")
			}
			for _, f := range c.Result.ChangedFiles {
				pid := "path:" + f
				add(Node{ID: pid, Type: "path", Label: f})
				link(tid, pid, "produced")
			}
			if c.State == core.StateBlocked && c.Result.BlockingReason != nil {
				bid := "block:" + stableID(*c.Result.BlockingReason)
				add(Node{ID: bid, Type: "blocker", Label: *c.Result.BlockingReason})
				link(tid, bid, "blocked_by")
			}
			// Heuristic: task implements specs whose covers overlap changed files / evidence
			for _, n := range g.Nodes {
				if n.Type != "spec" {
					continue
				}
				specID := strings.TrimPrefix(n.ID, "spec:")
				blob := strings.ToLower(c.Objective + " " + strings.Join(c.Execution.CompletionEvidence, " "))
				if strings.Contains(blob, strings.ToLower(specID)) || (strings.Contains(blob, "runtime-cohesion") && specID == "arah-runtime-cohesion") {
					link(tid, n.ID, "implements")
				}
			}
		}
	}

	// Runtime events → validated_by edges (deterministic; EventStore when present)
	if b.Events != nil {
		recent, err := b.Events.ListRecent(200)
		if err == nil {
			for _, ev := range recent {
				if ev.TaskID == "" {
					continue
				}
				tid := "task:" + ev.TaskID
				if !idx[tid] {
					continue
				}
				eid := "event:" + ev.ID
				label := ev.Kind
				if ev.CorrelationID != "" {
					label = ev.Kind + " @" + ev.CorrelationID
				}
				add(Node{ID: eid, Type: "event", Label: label})
				link(tid, eid, "validated_by")
				if ev.AgentID != "" {
					aid := "agent:" + ev.AgentID
					add(Node{ID: aid, Type: "agent", Label: ev.AgentID})
					link(eid, aid, "invokes")
				}
			}
		}
	}
	return g, nil
}

// Explain returns a human-readable slice of the graph for one task (H-18 evidence explain).
func (b *Builder) Explain(taskID string) (map[string]any, error) {
	g, err := b.Build()
	if err != nil {
		return nil, err
	}
	tid := "task:" + taskID
	nodes := map[string]Node{}
	for _, n := range g.Nodes {
		nodes[n.ID] = n
	}
	var edges []Edge
	related := map[string]bool{tid: true}
	for _, e := range g.Edges {
		if e.From == tid || e.To == tid {
			edges = append(edges, e)
			related[e.From] = true
			related[e.To] = true
		}
	}
	var outNodes []Node
	for id := range related {
		if n, ok := nodes[id]; ok {
			outNodes = append(outNodes, n)
		}
	}
	return map[string]any{
		"task_id": taskID,
		"nodes":   outNodes,
		"edges":   edges,
		"version": g.Version,
	}, nil
}

// stableID returns a collision-resistant id fragment for free-form strings.
func stableID(s string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(sum[:16])
}
