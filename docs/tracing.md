# Tracing Design

Each tool call creates a root span named `mcp.tool_call`.

Child spans represent the major control points:

- `mcp.validate_schema`
- `mcp.policy_check`
- `mcp.budget_check`
- `mcp.approval_check`
- `mcp.execute_tool`
- `mcp.audit_write`

## Attribute Strategy

The project avoids placing raw tool arguments into trace attributes. Instead, it records:

- argument count
- argument hash
- tool name
- policy decision
- risk level
- budget status
- approval requirement
- audit ID

This gives observability without leaking sensitive payloads into telemetry backends.

## Exporters

The default exporter is stdout for local development. OTLP gRPC export can be enabled with:

```bash
MCP_OBSERVE_OTEL_EXPORTER=otlp \
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 \
go run ./cmd/server
```
