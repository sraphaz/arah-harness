package watch

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	ingestor *Ingestor
	dirs     []string
	onEvent  func()
	deb      time.Duration
	mu       sync.Mutex
	pending  map[string]struct{}
	timer    *time.Timer
}

func NewWatcher(ing *Ingestor, dirs []string, onEvent func()) *Watcher {
	return &Watcher{
		ingestor: ing,
		dirs:     dirs,
		onEvent:  onEvent,
		deb:      250 * time.Millisecond,
		pending:  map[string]struct{}{},
	}
}

func (w *Watcher) Run(stop <-chan struct{}) error {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer fw.Close()

	for _, d := range w.dirs {
		_ = addRecursive(fw, d)
	}

	for {
		select {
		case <-stop:
			return nil
		case err := <-fw.Errors:
			if err != nil {
				log.Printf("watch: %v", err)
			}
		case ev := <-fw.Events:
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			if st, err := os.Stat(ev.Name); err == nil && st.IsDir() {
				_ = addRecursive(fw, ev.Name)
			}
			w.queue(ev.Name)
		}
	}
}

func (w *Watcher) queue(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending[path] = struct{}{}
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.deb, w.flush)
}

func (w *Watcher) flush() {
	w.mu.Lock()
	batch := w.pending
	w.pending = map[string]struct{}{}
	w.mu.Unlock()
	for p := range batch {
		w.ingestor.IngestPath(p)
	}
	if w.onEvent != nil {
		w.onEvent()
	}
}

func addRecursive(fw *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			_ = fw.Add(path)
		}
		return nil
	})
}
