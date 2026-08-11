package policy

import (
	"strings"

	"github.com/yesimakar/mcp-observe-go/internal/contracts"
	"github.com/yesimakar/mcp-observe-go/internal/tools"
)

const (
	DecisionAllow           = "allow"
	DecisionDeny            = "deny"
	DecisionRequireApproval = "require_approval"
)

type Decision struct {
	Decision  string `json:"decision"`
	RiskLevel string `json:"risk_level"`
	Reason    string `json:"reason"`
}

type Engine struct{}

func NewEngine() Engine { return Engine{} }

func (Engine) Evaluate(req contracts.ToolCallRequest, tool tools.Tool) Decision {
	env := strings.ToLower(strings.TrimSpace(req.Environment))
	if env == "" {
		env = "development"
	}

	if strings.TrimSpace(req.Actor) == "" {
		return Decision{Decision: DecisionDeny, RiskLevel: "high", Reason: "actor is required for auditability"}
	}

	if tool.Name == "delete_record" && env == "production" {
		if req.ApprovalToken == "approved-demo-token" {
			return Decision{Decision: DecisionAllow, RiskLevel: "high", Reason: "destructive production tool approved with demo approval token"}
		}
		return Decision{Decision: DecisionRequireApproval, RiskLevel: "high", Reason: "delete_record is destructive in production and requires approval"}
	}

	if tool.RiskLevel == "high" {
		return Decision{Decision: DecisionRequireApproval, RiskLevel: "high", Reason: "high-risk tool requires approval"}
	}

	return Decision{Decision: DecisionAllow, RiskLevel: tool.RiskLevel, Reason: "tool call passed policy checks"}
}
