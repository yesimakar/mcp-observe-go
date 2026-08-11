package budget

import (
	"testing"

	"github.com/yesimakar/mcp-observe-go/internal/contracts"
	"github.com/yesimakar/mcp-observe-go/internal/tools"
)

func TestBudgetRejectsExpensiveTool(t *testing.T) {
	registry := tools.NewRegistry()
	tool := registry.MustGet("delete_record")

	result := Check(contracts.ToolCallRequest{Budget: &contracts.BudgetInput{MaxUnits: 2}}, tool)

	if result.Allowed {
		t.Fatalf("expected budget rejection")
	}
}

func TestBudgetAllowsToolWithinLimit(t *testing.T) {
	registry := tools.NewRegistry()
	tool := registry.MustGet("search_docs")

	result := Check(contracts.ToolCallRequest{Budget: &contracts.BudgetInput{MaxUnits: 2}}, tool)

	if !result.Allowed {
		t.Fatalf("expected budget allowed")
	}
}
