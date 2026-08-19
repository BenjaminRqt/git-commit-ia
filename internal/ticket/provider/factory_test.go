package provider_test

import (
	"testing"

	"git-ai-commit/internal/ticket/provider"
)

func TestNew_NoopWhenEmpty(t *testing.T) {
	p := provider.New(provider.Config{Provider: ""})
	tk, err := p.Fetch("BOZ-909")
	if err != nil {
		t.Fatalf("noop Fetch() erreur inattendue : %v", err)
	}
	if tk != nil {
		t.Error("noop Fetch() doit renvoyer nil")
	}
}

func TestNew_NoopWhenUnknown(t *testing.T) {
	p := provider.New(provider.Config{Provider: "gitlab"})
	tk, err := p.Fetch("BOZ-909")
	if err != nil {
		t.Fatalf("noop Fetch() erreur inattendue : %v", err)
	}
	if tk != nil {
		t.Error("noop Fetch() doit renvoyer nil pour un fournisseur inconnu")
	}
}

func TestNew_NoopWhenJiraMissingConfig(t *testing.T) {
	// Jira configuré mais sans URL ni variables d'env → noop
	p := provider.New(provider.Config{Provider: "jira", JiraBaseURL: ""})
	tk, err := p.Fetch("BOZ-909")
	if err != nil {
		t.Fatalf("noop Fetch() erreur inattendue : %v", err)
	}
	if tk != nil {
		t.Error("noop Fetch() doit renvoyer nil quand la config Jira est incomplète")
	}
}
