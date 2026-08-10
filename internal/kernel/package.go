// Package kernel generates and verifies the distributable kernel/ tree (H-15).
//
// Canonical sources live at the repository root (.agents, .skills, .cursor,
// scripts/agents, scripts/harness). Sync produces:
//   - kernel/ + kernel/manifest.json (checkout / PowerShell init)
//   - internal/kernel/payload/kernel.zip (go:embed for `arah kernel install`)
//
// Do not hand-edit files under kernel/ — run `arah kernel sync` after changing sources.
package kernel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ManifestRel = "kernel/manifest.json"
	manifestVer = 1
)

// Manifest records SHA-256 digests of every file under kernel/ (except itself).
type Manifest struct {
	Version     int               `json:"version"`
	GeneratedAt string            `json:"generated_at,omitempty"`
	Files       map[string]string `json:"files"` // path relative to kernel/ → sha256 hex
}

// sourceRoots are repo-relative directories whose allowlisted files feed kernel/.
var sourceRoots = []string{
	".agents",
	".skills",
	".cursor",
	"scripts/agents",
	"scripts/harness",
}

// harnessOnly excludes paths that must not ship in the distributable kernel.
func harnessOnly(rel string) bool {
	rel = filepath.ToSlash(rel)
	switch rel {
	case ".agents/choreography.harness.yaml",
		".agents/choreography.domains.yaml",
		".agents/pr-steward.yaml",
		".agents/domain/cli.agent.yaml",
		".agents/domain/example-domain.agent.yaml",
		".agents/domain/extension-live.agent.yaml",
		".agents/domain/kernel.agent.yaml",
		"scripts/harness/doctor-harness.ps1",
		"scripts/harness/install-harness.ps1":
		return true
	}
	if strings.HasPrefix(rel, ".agents/skills/") {
		return true
	}
	// Machine-local Live telemetry (sessions, events, diagnostics) — never distribute.
	if rel == ".cursor/arah-live" || strings.HasPrefix(rel, ".cursor/arah-live/") {
		return true
	}
	return false
}

func shortDigest(sum string) string {
	sum = strings.TrimSpace(strings.ToLower(sum))
	if len(sum) >= 12 {
		return sum[:12]
	}
	if sum == "" {
		return "(empty)"
	}
	return sum
}

func validSHA256Hex(sum string) bool {
	if len(sum) != 64 {
		return false
	}
	for i := 0; i < len(sum); i++ {
		c := sum[i]
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}

// SourceEntry maps a root-relative source path to its kernel-relative destination.
type SourceEntry struct {
	Source string // relative to repo root
	Dest   string // relative to kernel/
}

// ListSources walks allowlisted roots and returns sorted source→dest pairs.
func ListSources(repoRoot string) ([]SourceEntry, error) {
	var out []SourceEntry
	for _, root := range sourceRoots {
		base := filepath.Join(repoRoot, filepath.FromSlash(root))
		if _, err := os.Stat(base); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if d.IsDir() {
				if harnessOnly(rel) {
					return filepath.SkipDir
				}
				return nil
			}
			if harnessOnly(rel) {
				return nil
			}
			out = append(out, SourceEntry{Source: rel, Dest: rel})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Source < out[j].Source })
	return out, nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func buildManifest(repoRoot string, entries []SourceEntry) (*Manifest, error) {
	m := &Manifest{
		Version:     manifestVer,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Files:       make(map[string]string, len(entries)),
	}
	for _, e := range entries {
		sum, err := fileSHA256(filepath.Join(repoRoot, filepath.FromSlash(e.Source)))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Source, err)
		}
		m.Files[e.Dest] = sum
	}
	return m, nil
}

func writeManifest(repoRoot string, m *Manifest) error {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	path := filepath.Join(repoRoot, filepath.FromSlash(ManifestRel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func loadManifest(repoRoot string) (*Manifest, error) {
	raw, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(ManifestRel)))
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Sync copies allowlisted sources into kernel/, removes stale kernel files, and
// rewrites kernel/manifest.json. Returns the written manifest.
func Sync(repoRoot string) (*Manifest, int, error) {
	entries, err := ListSources(repoRoot)
	if err != nil {
		return nil, 0, err
	}
	m, err := buildManifest(repoRoot, entries)
	if err != nil {
		return nil, 0, err
	}
	if _, err := writePayloadZip(repoRoot, entries, m); err != nil {
		return nil, 0, fmt.Errorf("payload zip: %w", err)
	}
	wanted := make(map[string]bool, len(entries))
	copied := 0
	for _, e := range entries {
		wanted[e.Dest] = true
		src := filepath.Join(repoRoot, filepath.FromSlash(e.Source))
		dst := filepath.Join(repoRoot, "kernel", filepath.FromSlash(e.Dest))
		if err := copyFile(src, dst); err != nil {
			return nil, 0, fmt.Errorf("copy %s: %w", e.Source, err)
		}
		copied++
	}
	kernelRoot := filepath.Join(repoRoot, "kernel")
	if err := filepath.WalkDir(kernelRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(kernelRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "manifest.json" {
			return nil
		}
		if !wanted[rel] {
			return os.Remove(path)
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}
	if err := writeManifest(repoRoot, m); err != nil {
		return nil, 0, err
	}
	return m, copied, nil
}

// Drift describes a verify failure.
type Drift struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	Want   string `json:"want,omitempty"`
	Got    string `json:"got,omitempty"`
}

func (d Drift) String() string {
	if d.Want != "" || d.Got != "" {
		return fmt.Sprintf("%s: %s (want %s got %s)", d.Path, d.Reason, d.Want, d.Got)
	}
	return fmt.Sprintf("%s: %s", d.Path, d.Reason)
}

// Verify ensures kernel/ matches allowlisted sources and the recorded manifest.
func Verify(repoRoot string) ([]Drift, error) {
	entries, err := ListSources(repoRoot)
	if err != nil {
		return nil, err
	}
	m, err := loadManifest(repoRoot)
	if err != nil {
		return []Drift{{Path: ManifestRel, Reason: "missing or unreadable manifest"}}, nil
	}
	var drifts []Drift
	seen := map[string]bool{}
	for _, e := range entries {
		seen[e.Dest] = true
		srcPath := filepath.Join(repoRoot, filepath.FromSlash(e.Source))
		dstPath := filepath.Join(repoRoot, "kernel", filepath.FromSlash(e.Dest))
		srcSum, err := fileSHA256(srcPath)
		if err != nil {
			drifts = append(drifts, Drift{Path: e.Source, Reason: "source unreadable"})
			continue
		}
		dstSum, err := fileSHA256(dstPath)
		if err != nil {
			drifts = append(drifts, Drift{Path: e.Dest, Reason: "missing in kernel/"})
			continue
		}
		if srcSum != dstSum {
			drifts = append(drifts, Drift{Path: e.Dest, Reason: "kernel differs from source", Want: shortDigest(srcSum), Got: shortDigest(dstSum)})
		}
		manSum, ok := m.Files[e.Dest]
		if !ok {
			drifts = append(drifts, Drift{Path: e.Dest, Reason: "missing from manifest"})
		} else if !validSHA256Hex(manSum) {
			drifts = append(drifts, Drift{Path: e.Dest, Reason: "manifest digest malformed", Got: shortDigest(manSum)})
		} else if manSum != srcSum {
			drifts = append(drifts, Drift{Path: e.Dest, Reason: "manifest stale vs source", Want: shortDigest(srcSum), Got: shortDigest(manSum)})
		}
	}
	for path, sum := range m.Files {
		if !seen[path] {
			got := shortDigest(sum)
			reason := "manifest entry not in allowlist"
			if !validSHA256Hex(sum) {
				reason = "manifest entry not in allowlist (digest malformed)"
			}
			drifts = append(drifts, Drift{Path: path, Reason: reason, Got: got})
		}
	}
	kernelRoot := filepath.Join(repoRoot, "kernel")
	_ = filepath.WalkDir(kernelRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(kernelRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "manifest.json" {
			return nil
		}
		if !seen[rel] {
			drifts = append(drifts, Drift{Path: rel, Reason: "extra file under kernel/"})
		}
		return nil
	})
	drifts = append(drifts, verifyPayloadZip(repoRoot, entries, m)...)
	sort.Slice(drifts, func(i, j int) bool { return drifts[i].Path < drifts[j].Path })
	return drifts, nil
}
