package api_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/sraphaz/arah-harness/live/internal/api"
	"github.com/sraphaz/arah-harness/live/internal/repo"
	"github.com/sraphaz/arah-harness/live/internal/store"
)

func TestReadOnlyRejectsWrites(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	info, err := repo.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	h := api.New(info, st, nil).Handler()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		req := httptest.NewRequest(method, "/api/summary", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s /api/summary => %d, want 405", method, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET health => %d", rec.Code)
	}
}
