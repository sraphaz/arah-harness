// Package choreography resolves primary_executor from area shortcuts + choreography.yaml.
package choreography

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sraphaz/arah-harness/internal/core"
)

var areaExec = map[string]string{
	"backend":      "backend",
	"frontend":     "frontend",
	"docs":         "docs-steward",
	"spec":         "spec-steward",
	"sdd":          "spec-steward",
	"architecture": "solutions-architect",
	"ops":          "release",
	"security":     "security",
	"planning":     "planner",
	"testing":      "qa",
	"harness":      "docs-steward",
}

var areaPaths = map[string][]string{
	"backend":      {"backend/**", "src/**/api/**", "services/**", "cmd/**", "internal/**"},
	"frontend":     {"frontend/**", "apps/web/**", "apps/mobile/**"},
	"docs":         {"docs/**"},
	"architecture": {"docs/architecture/**", "docs/design/**", "docs/adr/**"},
	"spec":         {"docs/specs/**", "scripts/harness/**"},
	"sdd":          {"docs/specs/**"},
	"testing":      {"tests/**", "e2e/**", "docs/testing/**"},
	"ops":          {".github/workflows/**", "infrastructure/**"},
	"harness":      {"kernel/**", ".agents/**", "scripts/**", "schemas/**"},
}

type Resolver struct {
	RepoRoot string
}

func New(repoRoot string) *Resolver {
	return &Resolver{RepoRoot: repoRoot}
}

type fileRule struct {
	ID        string `yaml:"id"`
	When      string `yaml:"when"`
	Paths     []string `yaml:"paths"`
	Execution struct {
		PrimaryExecutor string `yaml:"primary_executor"`
	} `yaml:"execution"`
	Agents []struct {
		ID       string   `yaml:"id"`
		Type     string   `yaml:"type"`
		Role     string   `yaml:"role"`
		Autonomy []string `yaml:"autonomy"`
	} `yaml:"agents"`
}

type fileDoc struct {
	Rules []fileRule `yaml:"rules"`
}

func (r *Resolver) Resolve(area, preferred string) (core.ResolvedRouting, error) {
	area = strings.ToLower(strings.TrimSpace(area))
	if preferred == "" {
		preferred = areaExec[area]
	}
	out := core.ResolvedRouting{
		PrimaryExecutor: preferred,
		AllowedPaths:    append([]string{}, areaPaths[area]...),
	}
	if len(out.AllowedPaths) == 0 && area != "" {
		out.AllowedPaths = []string{area + "/**"}
	}

	path := filepath.Join(r.RepoRoot, ".agents", "choreography.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		if out.PrimaryExecutor == "" {
			return out, &core.DomainError{
				Code:    "EXECUTION.EXACTLY_ONE_PRIMARY_EXECUTOR_REQUIRED",
				Message: "no eligible primary_executor",
				Details: map[string]any{"area": area},
			}
		}
		return out, nil
	}
	var doc fileDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return out, err
	}

	var matched *fileRule
	for i := range doc.Rules {
		rule := &doc.Rules[i]
		if rule.When == "pull_request" {
			continue
		}
		if rule.Execution.PrimaryExecutor == preferred && preferred != "" {
			matched = rule
			break
		}
		if strings.Contains(rule.ID, area) {
			matched = rule
			break
		}
	}
	if matched != nil {
		out.ChoreographyRule = matched.ID
		if matched.Execution.PrimaryExecutor != "" && preferred == "" {
			out.PrimaryExecutor = matched.Execution.PrimaryExecutor
		}
		for _, a := range matched.Agents {
			if a.ID == out.PrimaryExecutor || a.ID == "orchestrator" {
				continue
			}
			role := a.Role
			isConsultant := a.Type == "domain" || a.Type == "specialist" || role == "consultant" ||
				contains(a.Autonomy, "consult") || contains(a.Autonomy, "consult_post")
			isReviewer := role == "reviewer" || a.ID == "qa" || a.ID == "security" || a.ID == "pr-steward"
			if isConsultant {
				out.Consultants = appendUnique(out.Consultants, a.ID)
			} else if isReviewer {
				out.Reviewers = appendUnique(out.Reviewers, a.ID)
			}
		}
	}
	if out.PrimaryExecutor == "" {
		return out, &core.DomainError{
			Code:    "EXECUTION.EXACTLY_ONE_PRIMARY_EXECUTOR_REQUIRED",
			Message: "no eligible primary_executor",
			Details: map[string]any{"area": area},
		}
	}
	return out, nil
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func appendUnique(xs []string, v string) []string {
	for _, x := range xs {
		if x == v {
			return xs
		}
	}
	return append(xs, v)
}
