package provider

import (
	"context"
	"testing"
)

type registryTestProvider struct{ name string }

func (p *registryTestProvider) Name() string { return p.name }
func (p *registryTestProvider) Translate(context.Context, Request) (*Result, error) {
	return &Result{}, nil
}
func (p *registryTestProvider) TranslateBatch(context.Context, []Request) ([]*Result, []error) {
	return nil, nil
}
func (p *registryTestProvider) MaxTextLen() int               { return 100 }
func (p *registryTestProvider) IsTextTooLongError(error) bool { return false }

func TestCandidatesRoundRobin(t *testing.T) {
	Clear()
	defer Clear()
	Register(&registryTestProvider{name: "first"})
	Register(&registryTestProvider{name: "second"})

	candidates, err := Candidates("auto")
	if err != nil {
		t.Fatal(err)
	}
	if candidates[0].Name() != "first" || candidates[1].Name() != "second" {
		t.Fatalf("first candidates = %v, want first,second", names(candidates))
	}

	candidates, err = Candidates("auto")
	if err != nil {
		t.Fatal(err)
	}
	if candidates[0].Name() != "second" || candidates[1].Name() != "first" {
		t.Fatalf("second candidates = %v, want second,first", names(candidates))
	}
}

func names(providers []Provider) []string {
	result := make([]string, len(providers))
	for i, p := range providers {
		result[i] = p.Name()
	}
	return result
}
