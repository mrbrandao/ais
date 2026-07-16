package memx

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// SearchResult is a single match from topics.yaml.
type SearchResult struct {
	Checkpoint string
	Name       string
	Summary    string
	Topics     []string
}

// Search parses topics.yaml and returns entries whose topic list
// contains a case-insensitive substring match for query.
func Search(
	cfg *Config,
	mentalDir,
	project,
	query string,
) ([]SearchResult, error) {
	layout := NewLayout(cfg, mentalDir)

	data, err := os.ReadFile(layout.TopicsFile(project))
	if err != nil {
		return nil, fmt.Errorf(
			"read topics for %q: %w", project, err,
		)
	}

	var tf TopicsFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("parse topics.yaml: %w", err)
	}

	q := strings.ToLower(query)
	var results []SearchResult
	for _, entry := range tf.Memory {
		if matchesTopics(entry.Topics, q) ||
			strings.Contains(
				strings.ToLower(entry.Summary), q,
			) {
			results = append(results, SearchResult{
				Checkpoint: entry.Checkpoint,
				Name:       entry.Name,
				Summary:    entry.Summary,
				Topics:     entry.Topics,
			})
		}
	}
	return results, nil
}

// PrintSearchResults writes search results to stdout in a format
// the LLM can read to decide which checkpoint to load.
func PrintSearchResults(results []SearchResult, query string) {
	if len(results) == 0 {
		fmt.Printf("No checkpoints found for %q\n", query)
		return
	}

	fmt.Printf("Found %d checkpoint(s) for %q:\n\n", len(results), query)
	for i, r := range results {
		fmt.Printf("%d. %s\n", i+1, r.Checkpoint)
		if r.Summary != "" {
			fmt.Printf("   Summary: %s\n", r.Summary)
		}
		if len(r.Topics) > 0 {
			fmt.Printf(
				"   Topics:  %s\n",
				strings.Join(r.Topics, ", "),
			)
		}
		fmt.Println()
	}
	fmt.Println(
		"Load a checkpoint with: mental mem load <project>",
	)
}

// matchesTopics returns true if any topic contains the query substring.
func matchesTopics(topics []string, query string) bool {
	for _, t := range topics {
		if strings.Contains(strings.ToLower(t), query) {
			return true
		}
	}
	return false
}
