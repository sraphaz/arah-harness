package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sraphaz/arah-harness/live/internal/api"
	"github.com/sraphaz/arah-harness/live/internal/github"
	"github.com/sraphaz/arah-harness/live/internal/repo"
	"github.com/sraphaz/arah-harness/live/internal/store"
	"github.com/sraphaz/arah-harness/live/internal/watch"
)

func main() {
	repoPath := flag.String("repo", ".", "path to ARAH-governed repository")
	addr := flag.String("addr", "127.0.0.1:8787", "listen address")
	reindex := flag.Bool("reindex", false, "wipe derived SQLite index before scan")
	flag.Parse()

	info, err := repo.Open(*repoPath)
	if err != nil {
		log.Fatalf("repo: %v", err)
	}

	idxDir := filepath.Join(info.Root, ".arah", "local", "index")
	st, err := store.Open(idxDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()
	if *reindex {
		if err := st.Reindex(); err != nil {
			log.Fatalf("reindex: %v", err)
		}
	}

	ing := &watch.Ingestor{Root: info.Root, Store: st}
	n, _ := ing.ScanAll()
	log.Printf("arah-live: indexed %d events from %s", n, info.Root)

	srv := api.New(info, st, github.FromEnv())
	stop := make(chan struct{})
	w := watch.NewWatcher(ing, info.WatchDirs(), func() {
		srv.Broadcast(map[string]any{"type": "session.tick", "message": "artifacts changed"})
	})
	go func() {
		if err := w.Run(stop); err != nil {
			log.Printf("watch stopped: %v", err)
		}
	}()

	httpSrv := &http.Server{Addr: *addr, Handler: srv.Handler()}
	go func() {
		log.Printf("arah-live: read-only console API on http://%s (repo=%s)", *addr, info.Name)
		log.Printf("  GET /api/summary|/api/feed|/api/gates|/api/domains|/api/queue|/api/proposals|/api/graph")
		log.Printf("  WS  /events")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	close(stop)
	_ = httpSrv.Close()
	fmt.Println("arah-live: bye")
}
