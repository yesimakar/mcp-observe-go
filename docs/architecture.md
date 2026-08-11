# Architecture

`mcp-observe-go` is a Go-based MCP-style tool gateway for safe AI tool execution.

The gateway receives tool-call requests, validates arguments, applies policy and budget controls, executes mock tools when allowed, writes a JSONL audit record, and emits OpenTelemetry-compatible traces.

## Request Flow

```text
HTTP request
→ decode tool-call payload
→ resolve tool from registry
→ validate arguments
→ evaluate policy
→ check budget
→ check approval requirement
→ execute tool if allowed
→ write audit record
→ return decision and trace ID
```

## Why Go?

Go is a good fit for this project because MCP gateways and observability collectors are infrastructure-style services. They need fast startup, simple deployment, concurrency, clear APIs, and reliable networking behavior.

## Why OpenTelemetry?

Tool execution is operationally important. A production AI system needs to know which tool was called, which policy decision was made, how long each step took, and where failures happened. OpenTelemetry provides a standard trace pipeline for that visibility.
