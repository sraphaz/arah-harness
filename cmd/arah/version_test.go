package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Garante o invariante de fonte única: a versão embutida no binário é
// exatamente o conteúdo do arquivo VERSION na raiz do repositório.
func TestVersionMatchesVersionFile(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	want := strings.TrimSpace(string(raw))
	if Version != want {
		t.Fatalf("Version = %q, want %q (VERSION file)", Version, want)
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+([.-].+)?$`).MatchString(Version) {
		t.Fatalf("Version %q is not semver X.Y.Z", Version)
	}
}
