package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sraphaz/arah-harness/live/internal/github"
	"github.com/sraphaz/arah-harness/live/internal/repo"
	"github.com/sraphaz/arah-harness/live/internal/store"
	"gopkg.in/yaml.v3"
)

type Server struct {
	Info   *repo.Info
	Store  *store.Store
	GH     *github.Adapter
	mu     sync.Mutex
	hub    map[*websocket.Conn]struct{}
	up     websocket.Upgrader
}

func New(info *repo.Info, st *store.Store, gh *github.Adapter) *Server {
	return &Server{
		Info:  info,
		Store: st,
		GH:    gh,
		hub:   map[*websocket.Conn]struct{}{},
		up: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/summary", s.handleSummary)
	mux.HandleFunc("GET /api/feed", s.handleFeed)
	mux.HandleFunc("GET /api/gates", s.handleGates)
	mux.HandleFunc("GET /api/domains", s.handleDomains)
	mux.HandleFunc("GET /api/queue", s.handleQueue)
	mux.HandleFunc("GET /api/proposals", s.handleProposals)
	mux.HandleFunc("GET /api/graph", s.handleGraph)
	mux.HandleFunc("GET /events", s.handleWS)
	// Reject writes explicitly
	mux.HandleFunc("POST /", s.rejectWrite)
	mux.HandleFunc("PUT /", s.rejectWrite)
	mux.HandleFunc("PATCH /", s.rejectWrite)
	mux.HandleFunc("DELETE /", s.rejectWrite)
	return withCORS(mux)
}

func (s *Server) rejectWrite(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "read-only: write endpoints are disabled", http.StatusMethodNotAllowed)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet && r.URL.Path != "/events" {
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete {
				http.Error(w, "read-only: write endpoints are disabled", http.StatusMethodNotAllowed)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Broadcast(v any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, _ := json.Marshal(v)
	for c := range s.hub {
		_ = c.WriteMessage(websocket.TextMessage, b)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"ok": true, "repo": s.Info.Root})
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	driftOK, drift := s.Info.Drift()
	since := time.Now().Add(-24 * time.Hour)
	signals, _ := s.Store.CountSince("signal", since)
	consult, _ := s.Store.CountSince("consultation", since)
	passed, failed, _ := s.Store.GateStats(since)
	rate, gateOK := store.FormatGateRate(passed, failed)
	evo, _ := os.ReadFile(s.Info.EvolutionPath())
	proposals := repo.CountProposals(evo)
	queueN := 0
	if s.GH != nil {
		if prs, err := s.GH.OpenPRs(); err == nil {
			queueN = len(prs)
		}
	}
	domains := s.Info.Domains()
	writeJSON(w, map[string]any{
		"repo":      s.Info.Name,
		"kernel":    s.Info.Kernel,
		"drift":     drift,
		"drift_ok":  driftOK,
		"live":      true,
		"synced_at": time.Now().UTC().Format(time.RFC3339),
		"kpis": map[string]any{
			"cells":       len(domains),
			"signals_24h": signals + consult,
			"gate_rate":   rate,
			"gate_ok":     gateOK,
			"awaiting":    queueN,
			"proposals":   proposals,
		},
		"errors": s.Store.Errors(),
	})
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	switch filter {
	case "consulta":
		filter = "consultation"
	case "gates":
		filter = "gates"
	case "mudancas", "mudanças", "change":
		filter = "change"
	case "evolucao", "evolução", "evolution":
		filter = "evolution"
	case "todos", "all", "":
		filter = ""
	}
	events, err := s.Store.Feed(filter, 100)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"events": events})
}

func (s *Server) handleGates(w http.ResponseWriter, r *http.Request) {
	gates, err := s.Store.RecentGates(20)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if len(gates) == 0 {
		// Derive placeholder from graph presence / doctor-like heuristics
		gates = []store.Gate{
			{Slug: "manifests", Name: "manifests", Status: "unknown", Summary: "sem eventos de gate ainda"},
		}
	}
	since := time.Now().Add(-24 * time.Hour)
	passed, failed, _ := s.Store.GateStats(since)
	rate, _ := store.FormatGateRate(passed, failed)
	summary := rate + " · últimas 24h"
	if len(gates) > 0 && gates[0].Summary != "" {
		summary = gates[0].Summary + " · " + rate
	}
	writeJSON(w, map[string]any{"gates": gates, "summary": summary})
}

func (s *Server) handleDomains(w http.ResponseWriter, r *http.Request) {
	domains := s.Info.Domains()
	since := time.Now().Add(-24 * time.Hour)
	out := make([]map[string]any, 0, len(domains))
	for _, d := range domains {
		path := ""
		if len(d.Paths) > 0 {
			path = d.Paths[0]
		}
		sig, _ := s.Store.CountSince("", since)
		out = append(out, map[string]any{
			"id":          d.ID,
			"name":        d.Name,
			"path":        path,
			"health":      "ok",
			"agents":      1,
			"signals_24h": sig,
			"autonomy":    "propose-only",
		})
	}
	writeJSON(w, map[string]any{"domains": out})
}

func (s *Server) handleQueue(w http.ResponseWriter, r *http.Request) {
	if s.GH == nil {
		writeJSON(w, map[string]any{
			"items":   []any{},
			"note":    "Set ARAH_GITHUB_REPO=owner/name (and optional GITHUB_TOKEN) for selection queue",
			"enabled": false,
		})
		return
	}
	prs, err := s.GH.OpenPRs()
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	writeJSON(w, map[string]any{"items": prs, "enabled": true})
}

func (s *Server) handleProposals(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile(s.Info.EvolutionPath())
	if err != nil {
		writeJSON(w, map[string]any{"proposals": []any{}})
		return
	}
	var doc struct {
		Proposals []map[string]any `yaml:"proposals"`
	}
	_ = yaml.Unmarshal(b, &doc)
	writeJSON(w, map[string]any{"proposals": doc.Proposals})
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	p := s.Info.GraphPath()
	b, err := os.ReadFile(p)
	if err != nil {
		http.Error(w, "graph not found — run arah export-graph", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := s.up.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.hub[c] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.hub, c)
		s.mu.Unlock()
		_ = c.Close()
	}()
	_ = c.WriteJSON(map[string]any{"type": "session.hello", "message": "arah-live connected", "repo": s.Info.Name})
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
