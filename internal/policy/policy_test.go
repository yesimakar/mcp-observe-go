package policy

import (
	"testing"

	"github.com/yesimakar/mcp-observe-go/internal/contracts"
	"github.com/yesimakar/mcp-observe-go/internal/tools"
)

func TestDeleteRecordRequiresApprovalInProduction(t *testing.T) {
	registry := tools.NewRegistry()
	tool := registry.MustGet("delete_record")
	decision := NewEngine().Evaluate(contracts.ToolCallRequest{
		Actor:       "engineer@example.com",
		ToolName:    "delete_record",
		Environment: "production",
	}, tool)

	if decision.Decision != DecisionRequireApproval {
		t.Fatalf("expected require_approval, got %s", decision.Decision)
	}
}

func TestApprovedDeleteRecordIsAllowed(t *testing.T) {
	registry := tools.NewRegistry()
	tool := registry.MustGet("delete_record")
	decision := NewEngine().Evaluate(contracts.ToolCallRequest{
		Actor:         "engineer@example.com",
		ToolName:      "delete_record",
		Environment:   "production",
		ApprovalToken: "approved-demo-token",
	}, tool)

	if decision.Decision != DecisionAllow {
		t.Fatalf("expected allow, got %s", decision.Decision)
	}
}

func TestMissingActorIsDenied(t *testing.T) {
	registry := tools.NewRegistry()
	tool := registry.MustGet("search_docs")
	decision := NewEngine().Evaluate(contracts.ToolCallRequest{ToolName: "search_docs"}, tool)

	if decision.Decision != DecisionDeny {
		t.Fatalf("expected deny, got %s", decision.Decision)
	}
}
