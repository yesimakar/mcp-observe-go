package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/yesimakar/mcp-observe-go/internal/audit"
	"github.com/yesimakar/mcp-observe-go/internal/policy"
	"github.com/yesimakar/mcp-observe-go/internal/tools"
)

func TestToolCallAllowsSearchDocs(t *testing.T) {
	server := NewServer(tools.NewRegistry(), policy.NewEngine(), audit.NewLogger(filepath.Join(t.TempDir(), "audit.jsonl")))
	body := map[string]any{
		"session_id":  "test-session",
		"client_name": "test-client",
		"actor":       "engineer@example.com",
		"environment": "development",
		"tool_name":   "search_docs",
		"arguments":   map[string]any{"query": "policy", "limit": 3},
		"budget":      map[string]any{"max_units": 10},
	}

	response := postJSON(t, server.Routes(), "/tool-call", body)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}

	var payload ToolCallResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Decision != policy.DecisionAllow || !payload.Executed {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestToolCallRequiresApproval(t *testing.T) {
	server := NewServer(tools.NewRegistry(), policy.NewEngine(), audit.NewLogger(filepath.Join(t.TempDir(), "audit.jsonl")))
	body := map[string]any{
		"session_id":  "test-session",
		"client_name": "test-client",
		"actor":       "engineer@example.com",
		"environment": "production",
		"tool_name":   "delete_record",
		"arguments":   map[string]any{"record_id": "cust_123"},
		"budget":      map[string]any{"max_units": 10},
	}

	response := postJSON(t, server.Routes(), "/tool-call", body)
	if response.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", response.Code, response.Body.String())
	}
}

func postJSON(t *testing.T, handler http.Handler, path string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
