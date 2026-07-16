package memx

import (
	"fmt"
	"strings"
)

// GeneratePrompt builds a structured prompt from extracted session data.
// The prompt instructs any LLM to produce a mental checkpoint in the exact
// save-input format that mental mem save reads from stdin.
//
// Usage:
//
//	input, _ := opencode.Extract(sessionID, project, "", "")
//	prompt := GeneratePrompt(input)
//	fmt.Print(prompt)   // pipe to: | claude -p | mental mem save
func GeneratePrompt(input SaveInput) string {
	var b strings.Builder

	b.WriteString("You are creating a cross-session memory checkpoint " +
		"for the mental memory system.\n\n")
	b.WriteString(
		"Here is the session data extracted from the provider:\n\n",
	)

	fmt.Fprintf(&b, "Session ID: %s\n", input.Session.ID)
	fmt.Fprintf(&b, "Client: %s\n", input.Session.Client)
	fmt.Fprintf(&b, "Directory: %s\n", input.Session.Dir)
	fmt.Fprintf(&b, "Title/Summary: %s\n", input.Summary)

	if len(input.Files) > 0 {
		fmt.Fprintf(&b, "Files changed: %s\n",
			strings.Join(input.Files, ", "),
		)
	}
	if len(input.Topics) > 0 {
		fmt.Fprintf(&b, "Extracted keywords: %s\n",
			strings.Join(input.Topics, ", "),
		)
	}

	b.WriteString("\nGenerate a mental checkpoint in EXACTLY this format " +
		"(no extra text before or after):\n\n")

	fmt.Fprintf(&b, "project: %s\n", input.Project)
	fmt.Fprintf(&b, "session.id: %s\n", input.Session.ID)
	fmt.Fprintf(&b, "session.client: %s\n", input.Session.Client)
	fmt.Fprintf(&b, "session.model: <model used in this session>\n")
	fmt.Fprintf(&b, "session.dir: %s\n", input.Session.Dir)
	b.WriteString(
		"topics: <3-5 relevant topic keywords, comma-separated>\n",
	)

	if len(input.Files) > 0 {
		fmt.Fprintf(&b, "files: %s\n",
			strings.Join(input.Files, ", "),
		)
	} else {
		b.WriteString("files: <files changed, comma-separated>\n")
	}

	b.WriteString(
		"summary: <one sentence describing what this session accomplished>\n",
	)
	b.WriteString("memory:\n")
	fmt.Fprintf(&b, "# %s\n", input.Project)
	b.WriteString("status: active\n")
	b.WriteString("updated: <ISO datetime>\n\n")
	b.WriteString("## Context\n")
	b.WriteString("<3-5 sentences: current project state after this session>\n\n")
	b.WriteString("## Decisions\n")
	b.WriteString("- <decision>: <rationale>\n\n")
	b.WriteString("## Next Steps\n")
	b.WriteString("- [ ] <next action>\n\n")
	b.WriteString("## Related\n")
	b.WriteString("<related projects or empty>\n\n")
	b.WriteString("## Search\n")
	fmt.Fprintf(&b, "Topics and past sessions: topics.yaml\n")
	fmt.Fprintf(&b,
		"Command: mental mem search \"<topic>\" --project %s\n",
		input.Project,
	)
	b.WriteString("---\n")
	b.WriteString("## What We Did\n")
	b.WriteString("<paragraph: what happened in this session>\n\n")
	b.WriteString("## Decisions Made\n")
	b.WriteString("- <decision and rationale>\n\n")
	b.WriteString("## Open Questions\n")
	b.WriteString("- <unresolved question if any>\n\n")
	b.WriteString("## Handoff\n")
	b.WriteString(
		"<1-2 sentences: exactly where to pick up in the next session>\n",
	)

	return b.String()
}
