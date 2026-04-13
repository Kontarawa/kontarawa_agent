package prompt

import (
	"fmt"
	"strings"

	"kontarawa_agent/internal/retrieval"
)

func BuildSystem(profile string, retrieved []retrieval.Retrieved) string {
	var b strings.Builder
	b.WriteString("You are a local coding assistant. Follow the author's profile and preferences.\n")
	b.WriteString("If memory contains prior decisions, prefer them. If unclear, make a safe assumption and state it.\n")
	b.WriteString("Be concise and concrete.\n\n")

	if strings.TrimSpace(profile) != "" {
		b.WriteString("## Author profile\n")
		b.WriteString(strings.TrimSpace(profile))
		b.WriteString("\n\n")
	}

	if len(retrieved) > 0 {
		b.WriteString("## Retrieved memory\n")
		for i, r := range retrieved {
			b.WriteString(fmt.Sprintf("### Memory %d — %s — score=%.3f\n", i+1, r.DocID, r.Score))
			b.WriteString(r.Text)
			b.WriteString("\n\n")
		}
	}

	return strings.TrimSpace(b.String())
}

