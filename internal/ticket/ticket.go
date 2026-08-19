// Package ticket defines the generic interface for ticket providers.
package ticket

// Ticket holds provider-agnostic ticket metadata.
type Ticket struct {
	// Type is the suggested commit type (e.g. "feat", "fix") inferred by the adapter.
	Type        string
	Summary     string
	Description string
}

// Provider is the interface that every ticket adapter must implement.
type Provider interface {
	Fetch(key string) (*Ticket, error)
}

// Noop is a no-op adapter used when no provider is configured.
type Noop struct{}

// Fetch always returns nil, nil (no ticket, no error).
func (Noop) Fetch(_ string) (*Ticket, error) {
	return nil, nil
}
