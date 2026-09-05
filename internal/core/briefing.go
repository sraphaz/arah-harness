package core

import (
	"fmt"
	"strings"
)

// RenderBriefing builds the deterministic BRIEFING.md body for the primary executor.
func RenderBriefing(c *Contract) string {
	if c == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Execution briefing — %s\n\n", c.TaskID)
	fmt.Fprintf(&b, "You are the **primary_executor**: `%s`.\n\n", c.PrimaryExecutor)
	b.WriteString("## Objective\n")
	b.WriteString(c.Objective)
	b.WriteString("\n\n## Rules\n")
	b.WriteString("- You alone conduct execution.\n")
	b.WriteString("- Consultants may only return structured consultation-result YAML under consultations/.\n")
	b.WriteString("- Consultants must not talk to each other or redefine the executor.\n")
	b.WriteString("- Do not return to routed after executing.\n")
	b.WriteString("- Finish only as **done** (with concrete evidence) or **blocked** (with a specific reason).\n")
	b.WriteString("- Analysis alone is not completion when intent_type=execution.\n\n")
	b.WriteString("## Scope\n")
	fmt.Fprintf(&b, "Area: %s\n", c.Scope.Area)
	b.WriteString("Allowed paths:\n")
	for _, p := range c.Scope.AllowedPaths {
		fmt.Fprintf(&b, "- %s\n", p)
	}
	b.WriteString("\n## Limits\n")
	fmt.Fprintf(&b, "- max_handoffs: %d\n", c.Limits.MaxHandoffs)
	fmt.Fprintf(&b, "- max_consultations: %d\n", c.Limits.MaxConsultations)
	fmt.Fprintf(&b, "- max_analysis_cycles: %d\n\n", c.Limits.MaxAnalysisCycles)
	b.WriteString("## Consultants\n")
	if len(c.Participants.Consultants) == 0 {
		b.WriteString("(none)\n")
	} else {
		for _, x := range c.Participants.Consultants {
			fmt.Fprintf(&b, "- %s\n", x)
		}
	}
	b.WriteString("\n## Reviewers\n")
	if len(c.Participants.Reviewers) == 0 {
		b.WriteString("(none)\n")
	} else {
		for _, x := range c.Participants.Reviewers {
			fmt.Fprintf(&b, "- %s\n", x)
		}
	}
	b.WriteString("\n## Complete\n```\n")
	fmt.Fprintf(&b, "arah task complete --task-id %s --evidence \"…\"\n", c.TaskID)
	b.WriteString("```\n")
	return b.String()
}
