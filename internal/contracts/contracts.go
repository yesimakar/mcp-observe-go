package contracts

import "time"

// ToolCallRequest is the public request shape accepted by the gateway.
// It lives in a dependency-neutral package so policy, budget, and gateway
// packages can share it without creating Go import cycles.
type ToolCallRequest struct {
	SessionID     string         `json:"session_id"`
	ClientName    string         `json:"client_name"`
	Actor         string         `json:"actor"`
	ToolName      string         `json:"tool_name"`
	Environment   string         `json:"environment"`
	Arguments     map[string]any `json:"arguments"`
	Budget        *BudgetInput   `json:"budget,omitempty"`
	ApprovalToken string         `json:"approval_token,omitempty"`
}

type BudgetInput struct {
	MaxUnits int `json:"max_units"`
}

type ToolCallResponse struct {
	Decision   string         `json:"decision"`
	RiskLevel  string         `json:"risk_level"`
	Reason     string         `json:"reason"`
	TraceID    string         `json:"trace_id"`
	AuditID    string         `json:"audit_id"`
	ToolName   string         `json:"tool_name"`
	Executed   bool           `json:"executed"`
	DurationMS int64          `json:"duration_ms"`
	Result     map[string]any `json:"result,omitempty"`
	Error      string         `json:"error,omitempty"`
	ObservedAt time.Time      `json:"observed_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
