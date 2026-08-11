package repo

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Info struct {
	Root   string
	Name   string
	Kernel string
}

type Domain struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Paths       []string `yaml:"paths"`
}

func Open(root string) (*Info, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info := &Info{Root: abs, Name: filepath.Base(abs), Kernel: "unknown"}
	if b, err := os.ReadFile(filepath.Join(abs, ".arah-version")); err == nil {
		if m := regexp.MustCompile(`(?m)^version:\s*(\S+)`).FindSubmatch(b); len(m) == 2 {
			info.Kernel = string(m[1])
		}
	}
	if b, err := os.ReadFile(filepath.Join(abs, "arah.config.yaml")); err == nil {
		var cfg struct {
			Project struct {
				Name string `yaml:"name"`
			} `yaml:"project"`
		}
		if yaml.Unmarshal(b, &cfg) == nil && cfg.Project.Name != "" {
			info.Name = cfg.Project.Name
		}
	}
	return info, nil
}

func (i *Info) Domains() []Domain {
	b, err := os.ReadFile(filepath.Join(i.Root, "arah.config.yaml"))
	if err != nil {
		return nil
	}
	var cfg struct {
		Domains []Domain `yaml:"domains"`
	}
	if yaml.Unmarshal(b, &cfg) != nil {
		return nil
	}
	return cfg.Domains
}

func (i *Info) GraphPath() string {
	return filepath.Join(i.Root, "docs", "_meta", "agent-graph.generated.json")
}

func (i *Info) EvolutionPath() string {
	return filepath.Join(i.Root, "docs", "_meta", "evolution.proposed.yaml")
}

func (i *Info) WatchDirs() []string {
	cands := []string{
		filepath.Join(i.Root, ".arah", "local", "bus"),
		filepath.Join(i.Root, ".arah", "local", "audit"),
		filepath.Join(i.Root, ".arah", "bus"),
		filepath.Join(i.Root, ".arah", "audit"),
		filepath.Join(i.Root, ".cursor", "arah-live"),
		filepath.Join(i.Root, "docs", "_meta"),
	}
	var out []string
	for _, d := range cands {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			out = append(out, d)
		}
	}
	return out
}

func (i *Info) Drift() (ok bool, label string) {
	graph := i.GraphPath()
	if _, err := os.Stat(graph); err != nil {
		return false, "sem grafo"
	}
	// Lightweight drift: graph mtime older than organism/discovery by >24h is warn-only;
	// for MVP, presence of graph + .arah-version = ok.
	if i.Kernel == "unknown" {
		return false, "kernel desconhecido"
	}
	return true, "sem drift"
}

func CountProposals(evolutionYAML []byte) int {
	if len(evolutionYAML) == 0 {
		return 0
	}
	return strings.Count(string(evolutionYAML), "\n- id:")
}
