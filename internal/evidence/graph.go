// Package evidence builds a deterministic Evidence Graph from Arah artifacts (H-18).
// No LLM — only schemas, contracts, specs and runtime events.
package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
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

type specDoc struct {
	ID         string   `yaml:"id"`
	Title      string   `yaml:"title"`
	Covers     []string `yaml:"covers"`
	DependsOn  []string `yaml:"depends_on"`
	Supersedes []string `yaml:"supersedes"`
	Status     string   `yaml:"status"`
}

// Build derives nodes and edges exclusively from Arah schemas and runtime state.
func (b *Builder) Build() (*Graph, error) {
	g := &Graph{Version: "1"}
	idx := map[string]int{}
	edgeIdx := map[string]bool{}
	add := func(n Node) {
		if pos, ok := idx[n.ID]; ok {
			existing := &g.Nodes[pos]
			if existing.Type == "spec" && n.Type == "spec" && n.Label != "" {
				stubLabel := strings.TrimPrefix(n.ID, "spec:")
				if existing.Label == "" || existing.Label == stubLabel {
					existing.Label = n.Label
				}
			}
			return
		}
		idx[n.ID] = len(g.Nodes)
		g.Nodes = append(g.Nodes, n)
	}
	link := func(from, to, rel string) {
		if from == "" || to == "" || rel == "" {
			return
		}
		key := from + "\x00" + to + "\x00" + rel
		if edgeIdx[key] {
			return
		}
		edgeIdx[key] = true
		g.Edges = append(g.Edges, Edge{From: from, To: to, Rel: rel})
	}

	coversBySpec := map[string][]string{}

	// Specs → covers / depends_on / supersedes
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
		var doc specDoc
		if yaml.Unmarshal(raw, &doc) != nil || doc.ID == "" {
			continue
		}
		sid := "spec:" + doc.ID
		add(Node{ID: sid, Type: "spec", Label: doc.Title})
		coversBySpec[doc.ID] = append([]string(nil), doc.Covers...)
		for _, c := range doc.Covers {
			pid := "path:" + c
			add(Node{ID: pid, Type: "path", Label: c})
			link(sid, pid, "covers")
		}
		for _, dep := range doc.DependsOn {
			if dep == "" {
				continue
			}
			did := "spec:" + dep
			add(Node{ID: did, Type: "spec", Label: dep})
			link(sid, did, "depends_on")
		}
		for _, old := range doc.Supersedes {
			if old == "" {
				continue
			}
			oid := "spec:" + old
			add(Node{ID: oid, Type: "spec", Label: old})
			link(sid, oid, "supersedes")
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

	if b.Store != nil {
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

				produced := producedPaths(c)
				for _, f := range produced {
					pid := "path:" + f
					add(Node{ID: pid, Type: "path", Label: f})
					link(tid, pid, "produced")
				}

				if c.State == core.StateBlocked && c.Result.BlockingReason != nil {
					bid := "block:" + stableID(*c.Result.BlockingReason)
					add(Node{ID: bid, Type: "blocker", Label: *c.Result.BlockingReason})
					link(tid, bid, "blocked_by")
				}

				// implements: path-based intersection with spec covers (no free-text heuristic)
				for specID, covers := range coversBySpec {
					if pathsOverlap(produced, covers) {
						link(tid, "spec:"+specID, "implements")
					}
				}
			}
		}
	}

	// Runtime events → run nodes linked to tasks
	if b.Events != nil {
		evs, err := b.Events.ListRecent(500)
		if err == nil {
			for _, ev := range evs {
				if ev.TaskID == "" {
					continue
				}
				rid := runNodeID(ev)
				label := ev.Kind
				if ev.TraceID != "" {
					label = ev.TraceID
				}
				add(Node{ID: rid, Type: "run", Label: label})
				link("task:"+ev.TaskID, rid, "evidenced_by")
			}
		}
	}

	sort.Slice(g.Nodes, func(i, j int) bool { return g.Nodes[i].ID < g.Nodes[j].ID })
	sort.Slice(g.Edges, func(i, j int) bool {
		if g.Edges[i].From != g.Edges[j].From {
			return g.Edges[i].From < g.Edges[j].From
		}
		if g.Edges[i].To != g.Edges[j].To {
			return g.Edges[i].To < g.Edges[j].To
		}
		return g.Edges[i].Rel < g.Edges[j].Rel
	})
	return g, nil
}

func runNodeID(ev core.Event) string {
	if strings.TrimSpace(ev.TraceID) != "" {
		return "run:" + ev.TraceID
	}
	if strings.TrimSpace(ev.ID) != "" {
		return "run:" + ev.ID
	}
	return "run:" + stableID(ev.TaskID+"|"+ev.Kind+"|"+ev.At)
}

func producedPaths(c *core.Contract) []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	for _, f := range c.Result.ChangedFiles {
		add(f)
	}
	for _, ev := range c.Execution.CompletionEvidence {
		for _, p := range pathCandidates(ev) {
			add(p)
		}
	}
	sort.Strings(out)
	return out
}

func pathCandidates(s string) []string {
	var out []string
	for _, tok := range strings.Fields(s) {
		tok = strings.Trim(tok, ",:;()[]\"'")
		if tok == "" {
			continue
		}
		if strings.Contains(tok, "/") || looksLikeFile(tok) {
			out = append(out, tok)
		}
	}
	return out
}

func looksLikeFile(s string) bool {
	lower := strings.ToLower(s)
	for _, ext := range []string{".go", ".yaml", ".yml", ".md", ".json", ".ps1", ".ts", ".tsx"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func pathsOverlap(produced, covers []string) bool {
	for _, p := range produced {
		for _, c := range covers {
			if pathMatchesCover(p, c) {
				return true
			}
		}
	}
	return false
}

func pathMatchesCover(path, cover string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	cover = filepath.ToSlash(strings.TrimSpace(cover))
	if path == "" || cover == "" {
		return false
	}
	if path == cover {
		return true
	}
	if strings.HasSuffix(cover, "/") {
		return strings.HasPrefix(path, cover) || path+"/" == cover
	}
	// directory cover without trailing slash
	if strings.HasPrefix(path, cover+"/") {
		return true
	}
	return false
}

// stableID returns a collision-resistant id fragment for free-form strings.
func stableID(s string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(sum[:16])
}
