package schema

import (
	"testing"

	"github.com/yesimakar/mcp-observe-go/internal/tools"
)

func TestValidateRequiredArgument(t *testing.T) {
	registry := tools.NewRegistry()
	tool := registry.MustGet("search_docs")

	result := Validate(tool, map[string]any{})

	if result.Valid {
		t.Fatalf("expected validation failure")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
}

func TestValidateUnexpectedArgument(t *testing.T) {
	registry := tools.NewRegistry()
	tool := registry.MustGet("search_docs")

	result := Validate(tool, map[string]any{"query": "test", "extra": "bad"})

	if result.Valid {
		t.Fatalf("expected validation failure")
	}
}

func TestValidateValidArguments(t *testing.T) {
	registry := tools.NewRegistry()
	tool := registry.MustGet("search_docs")

	result := Validate(tool, map[string]any{"query": "test", "limit": float64(3)})

	if !result.Valid {
		t.Fatalf("expected valid arguments, got %v", result.Errors)
	}
}
