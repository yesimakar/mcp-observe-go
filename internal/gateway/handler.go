package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yesimakar/mcp-observe-go/internal/audit"
	"github.com/yesimakar/mcp-observe-go/internal/budget"
	"github.com/yesimakar/mcp-observe-go/internal/policy"
	"github.com/yesimakar/mcp-observe-go/internal/schema"
	"github.com/yesimakar/mcp-observe-go/internal/tools"
	"github.com/yesimakar/mcp-observe-go/internal/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	registry *tools.Registry
	policy   policy.Engine
	audit    *audit.Logger
	tracer   trace.Tracer
}

func NewServer(registry *tools.Registry, policyEngine policy.Engine, auditLogger *audit.Logger) *Server {
	return &Server{
		registry: registry,
		policy:   policyEngine,
		audit:    auditLogger,
		tracer:   otel.Tracer("github.com/yesimakar/mcp-observe-go"),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /tools", s.listTools)
	mux.HandleFunc("POST /tool-call", s.callTool)
	mux.HandleFunc("GET /audit", s.getAudit)
	mux.HandleFunc("GET /mcp/capabilities", s.capabilities)
	return loggingMiddleware(mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "mcp-observe-go",
		"version": version.Version,
	})
}

func (s *Server) listTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"tools": s.registry.List()})
}

func (s *Server) capabilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":        "mcp-observe-go",
		"description": "MCP-style tool gateway with policy checks, audit logs, and OpenTelemetry-compatible traces.",
		"capabilities": []string{
			"tool_registry",
			"schema_validation",
			"policy_decisions",
			"approval_gates",
			"budget_checks",
			"jsonl_audit_log",
			"opentelemetry_traces",
		},
	})
}

func (s *Server) getAudit(w http.ResponseWriter, r *http.Request) {
	records, err := audit.ReadLast("audit/tool_calls.jsonl", 50)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to read audit log", Details: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"records": records})
}

func (s *Server) callTool(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	var req ToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON body", Details: err.Error()})
		return
	}

	if req.SessionID == "" {
		req.SessionID = "anonymous-session"
	}
	if req.ClientName == "" {
		req.ClientName = "unknown-client"
	}
	if req.Environment == "" {
		req.Environment = "development"
	}
	if req.Arguments == nil {
		req.Arguments = map[string]any{}
	}

	ctx, span := s.tracer.Start(
		r.Context(),
		"mcp.tool_call",
		trace.WithAttributes(
			attribute.String("mcp.session_id", req.SessionID),
			attribute.String("mcp.client_name", req.ClientName),
			attribute.String("mcp.tool.name", req.ToolName),
			attribute.String("mcp.environment", req.Environment),
			attribute.Int("mcp.arguments.count", len(req.Arguments)),
			attribute.String("mcp.arguments.hash", argumentsHash(req.Arguments)),
		),
	)
	defer span.End()

	resp := s.execute(ctx, req, started)
	span.SetAttributes(
		attribute.String("mcp.policy.decision", resp.Decision),
		attribute.String("mcp.tool.risk_level", resp.RiskLevel),
		attribute.Bool("mcp.tool.executed", resp.Executed),
		attribute.String("mcp.audit_id", resp.AuditID),
	)
	if resp.Error != "" {
		span.SetStatus(codes.Error, resp.Error)
	} else {
		span.SetStatus(codes.Ok, resp.Decision)
	}

	status := http.StatusOK
	if resp.Decision == policy.DecisionDeny {
		status = http.StatusForbidden
	} else if resp.Decision == policy.DecisionRequireApproval {
		status = http.StatusAccepted
	} else if resp.Error != "" {
		status = http.StatusInternalServerError
	}

	resp.TraceID = span.SpanContext().TraceID().String()
	writeJSON(w, status, resp)
}

func (s *Server) execute(ctx context.Context, req ToolCallRequest, started time.Time) ToolCallResponse {
	tool, ok := s.registry.Get(req.ToolName)
	if !ok {
		return s.auditAndReturn(ctx, req, started, policy.Decision{
			Decision:  policy.DecisionDeny,
			RiskLevel: "unknown",
			Reason:    fmt.Sprintf("tool %q is not registered", req.ToolName),
		}, false, nil, errors.New("unknown tool"))
	}

	ctx, validationSpan := s.tracer.Start(ctx, "mcp.validate_schema")
	validation := schema.Validate(tool, req.Arguments)
	validationSpan.SetAttributes(
		attribute.Bool("mcp.schema.valid", validation.Valid),
		attribute.Int("mcp.schema.error_count", len(validation.Errors)),
	)
	validationSpan.End()
	if !validation.Valid {
		return s.auditAndReturn(ctx, req, started, policy.Decision{
			Decision:  policy.DecisionDeny,
			RiskLevel: tool.RiskLevel,
			Reason:    strings.Join(validation.Errors, "; "),
		}, false, nil, errors.New("schema validation failed"))
	}

	ctx, policySpan := s.tracer.Start(ctx, "mcp.policy_check")
	decision := s.policy.Evaluate(req, tool)
	policySpan.SetAttributes(
		attribute.String("mcp.policy.decision", decision.Decision),
		attribute.String("mcp.policy.reason", decision.Reason),
		attribute.String("mcp.tool.risk_level", decision.RiskLevel),
	)
	policySpan.End()

	ctx, budgetSpan := s.tracer.Start(ctx, "mcp.budget_check")
	budgetResult := budget.Check(req, tool)
	budgetSpan.SetAttributes(
		attribute.Bool("mcp.budget.allowed", budgetResult.Allowed),
		attribute.Int("mcp.budget.max_units", budgetResult.MaxUnits),
		attribute.Int("mcp.budget.cost_units", budgetResult.Cost),
	)
	budgetSpan.End()
	if !budgetResult.Allowed {
		return s.auditAndReturn(ctx, req, started, policy.Decision{
			Decision:  policy.DecisionDeny,
			RiskLevel: decision.RiskLevel,
			Reason:    budgetResult.Reason,
		}, false, nil, errors.New("budget exceeded"))
	}

	ctx, approvalSpan := s.tracer.Start(ctx, "mcp.approval_check")
	approvalSpan.SetAttributes(
		attribute.Bool("mcp.approval.required", decision.Decision == policy.DecisionRequireApproval),
		attribute.Bool("mcp.approval.token_present", req.ApprovalToken != ""),
	)
	approvalSpan.End()
	if decision.Decision != policy.DecisionAllow {
		return s.auditAndReturn(ctx, req, started, decision, false, nil, nil)
	}

	ctx, executionSpan := s.tracer.Start(ctx, "mcp.execute_tool")
	result, err := tool.Handler(ctx, req.Arguments)
	executionSpan.SetAttributes(
		attribute.String("mcp.execution.status", statusFromErr(err)),
		attribute.String("mcp.tool.name", tool.Name),
	)
	if err != nil {
		executionSpan.RecordError(err)
		executionSpan.SetStatus(codes.Error, err.Error())
	} else {
		executionSpan.SetStatus(codes.Ok, "tool executed")
	}
	executionSpan.End()

	return s.auditAndReturn(ctx, req, started, decision, err == nil, result, err)
}

func (s *Server) auditAndReturn(ctx context.Context, req ToolCallRequest, started time.Time, decision policy.Decision, executed bool, result map[string]any, err error) ToolCallResponse {
	traceID := trace.SpanContextFromContext(ctx).TraceID().String()
	_, auditSpan := s.tracer.Start(ctx, "mcp.audit_write")
	defer auditSpan.End()

	message := ""
	if err != nil {
		message = err.Error()
		auditSpan.RecordError(err)
	}

	auditID, writeErr := s.audit.Write(audit.Record{
		TraceID:     traceID,
		SessionID:   req.SessionID,
		ClientName:  req.ClientName,
		Actor:       req.Actor,
		ToolName:    req.ToolName,
		Environment: req.Environment,
		Decision:    decision.Decision,
		RiskLevel:   decision.RiskLevel,
		Reason:      decision.Reason,
		Executed:    executed,
		DurationMS:  time.Since(started).Milliseconds(),
		Arguments:   audit.RedactArguments(req.Arguments),
		Error:       message,
		ObservedAt:  time.Now().UTC(),
	})
	if writeErr != nil {
		auditSpan.RecordError(writeErr)
		auditSpan.SetStatus(codes.Error, writeErr.Error())
	}
	auditSpan.SetAttributes(attribute.String("mcp.audit_id", auditID))

	return ToolCallResponse{
		Decision:   decision.Decision,
		RiskLevel:  decision.RiskLevel,
		Reason:     decision.Reason,
		TraceID:    traceID,
		AuditID:    auditID,
		ToolName:   req.ToolName,
		Executed:   executed,
		DurationMS: time.Since(started).Milliseconds(),
		Result:     result,
		Error:      message,
		ObservedAt: time.Now().UTC(),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func statusFromErr(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

func argumentsHash(args map[string]any) string {
	encoded, _ := json.Marshal(args)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])[:16]
}
