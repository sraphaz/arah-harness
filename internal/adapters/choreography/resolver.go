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

// Resolver selects primary_executor from area shortcuts and .agents/choreography.yaml.
type Resolver struct {
	RepoRoot string
}

// New returns a choreography resolver for repoRoot.
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

// Resolve returns the primary executor and participants for an area or file path.
// When area contains path separators it is matched against choreography rule paths.
func (r *Resolver) Resolve(area, preferred string) (core.ResolvedRouting, error) {
	rawArea := strings.TrimSpace(area)
	area = strings.ToLower(rawArea)
	pathKey := filepath.ToSlash(area)

	path := filepath.Join(r.RepoRoot, ".agents", "choreography.yaml")
	raw, err := os.ReadFile(path)
	var doc fileDoc
	if err == nil {
		if uerr := yaml.Unmarshal(raw, &doc); uerr != nil {
			return core.ResolvedRouting{}, uerr
		}
	}

	// Path-shaped inputs (e.g. apps/web/index.ts) match choreography paths first.
	if looksLikeRepoPath(pathKey) && len(doc.Rules) > 0 {
		if matched := matchRuleByPath(doc.Rules, pathKey); matched != nil {
			out := routingFromRule(matched, preferred)
			if out.PrimaryExecutor == "" {
				return out, &core.DomainError{
					Code:    "EXECUTION.EXACTLY_ONE_PRIMARY_EXECUTOR_REQUIRED",
					Message: "no eligible primary_executor",
					Details: map[string]any{"area": area, "choreography_rule": matched.ID},
				}
			}
			return out, nil
		}
	}

	if preferred == "" {
		preferred = areaExec[area]
	}
	out := core.ResolvedRouting{
		PrimaryExecutor: preferred,
		AllowedPaths:    append([]string{}, areaPaths[area]...),
	}
	if len(out.AllowedPaths) == 0 && area != "" && !looksLikeRepoPath(pathKey) {
		out.AllowedPaths = []string{area + "/**"}
	}

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
		applyRuleAgents(&out, matched)
		out.ChoreographyRule = matched.ID
		if matched.Execution.PrimaryExecutor != "" && preferred == "" {
			out.PrimaryExecutor = matched.Execution.PrimaryExecutor
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

// looksLikeRepoPath detects file/dir paths. Only slash separators — a lone "."
// must not treat dotted area labels (e.g. api.v2) as repo paths.
func looksLikeRepoPath(s string) bool {
	return strings.Contains(s, "/") || strings.Contains(s, `\`)
}

func matchRuleByPath(rules []fileRule, filePath string) *fileRule {
	filePath = strings.TrimPrefix(filepath.ToSlash(filePath), "./")
	for i := range rules {
		rule := &rules[i]
		if rule.When == "pull_request" {
			continue
		}
		for _, p := range rule.Paths {
			if pathMatches(filePath, filepath.ToSlash(p)) {
				return rule
			}
		}
	}
	return nil
}

func pathMatches(filePath, pattern string) bool {
	pattern = strings.TrimPrefix(pattern, "./")
	if pattern == "**" || pattern == filePath {
		return true
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return filePath == prefix || strings.HasPrefix(filePath, prefix+"/")
	}
	if strings.HasSuffix(pattern, "/**/**") {
		prefix := strings.TrimSuffix(pattern, "/**/**")
		return strings.HasPrefix(filePath, prefix+"/")
	}
	// simple * segment: services/*
	if strings.Contains(pattern, "*") {
		ok, _ := filepath.Match(pattern, filePath)
		return ok
	}
	return false
}

func routingFromRule(rule *fileRule, preferred string) core.ResolvedRouting {
	out := core.ResolvedRouting{
		PrimaryExecutor:  rule.Execution.PrimaryExecutor,
		ChoreographyRule: rule.ID,
		AllowedPaths:     append([]string{}, rule.Paths...),
	}
	if preferred != "" {
		out.PrimaryExecutor = preferred
	}
	applyRuleAgents(&out, rule)
	return out
}

func applyRuleAgents(out *core.ResolvedRouting, rule *fileRule) {
	for _, a := range rule.Agents {
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
