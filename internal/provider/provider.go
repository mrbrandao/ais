// Package provider defines the interface every AI
// assistant backend must implement.
package provider

import (
	"context"

	"github.com/mrbrandao/ais/internal/model"
)

// Provider is the contract for an AI assistant
// session backend. Add Export, Import, Save here
// as the tool grows.
type Provider interface {
	// Name returns the assistant identifier,
	// e.g. "opencode", "claude".
	Name() string
	// Search returns sessions matching q.
	Search(
		ctx context.Context,
		q model.Query,
	) ([]model.Session, error)
}
