package tools

import (
	"context"
	"fmt"
	"time"
)

func defaultTools() []Tool {
	return []Tool{
		{
			Name:        "search_docs",
			Description: "Search internal documentation and return matching snippets.",
			RiskLevel:   "low",
			CostUnits:   1,
			Schema: ToolSchema{
				Required: []string{"query"},
				Properties: map[string]string{
					"query": "string",
					"limit": "number",
				},
			},
			Handler: searchDocs,
		},
		{
			Name:        "create_ticket",
			Description: "Create a support ticket for engineering follow-up.",
			RiskLevel:   "medium",
			CostUnits:   2,
			Schema: ToolSchema{
				Required: []string{"title", "severity"},
				Properties: map[string]string{
					"title":       "string",
					"description": "string",
					"severity":    "string",
				},
			},
			Handler: createTicket,
		},
		{
			Name:        "send_notification",
			Description: "Send a notification to a user or channel.",
			RiskLevel:   "medium",
			CostUnits:   2,
			Schema: ToolSchema{
				Required: []string{"recipient", "message"},
				Properties: map[string]string{
					"recipient": "string",
					"message":   "string",
					"channel":   "string",
				},
			},
			Handler: sendNotification,
		},
		{
			Name:        "delete_record",
			Description: "Delete a record. Destructive actions require approval in production.",
			RiskLevel:   "high",
			CostUnits:   5,
			Schema: ToolSchema{
				Required: []string{"record_id"},
				Properties: map[string]string{
					"record_id": "string",
					"reason":    "string",
				},
			},
			Handler: deleteRecord,
		},
	}
}

func searchDocs(ctx context.Context, args map[string]any) (map[string]any, error) {
	_ = ctx
	query := stringValue(args, "query")
	limit := numberValue(args, "limit", 3)
	return map[string]any{
		"query": query,
		"limit": limit,
		"matches": []map[string]any{
			{"source": "runbook.md", "score": 0.91, "snippet": "Validate tool arguments before execution."},
			{"source": "policy.md", "score": 0.84, "snippet": "High-risk tools require approval in production."},
		},
	}, nil
}

func createTicket(ctx context.Context, args map[string]any) (map[string]any, error) {
	_ = ctx
	return map[string]any{
		"ticket_id":   fmt.Sprintf("TCK-%d", time.Now().Unix()%100000),
		"title":       stringValue(args, "title"),
		"severity":    stringValue(args, "severity"),
		"description": stringValue(args, "description"),
		"status":      "created",
	}, nil
}

func sendNotification(ctx context.Context, args map[string]any) (map[string]any, error) {
	_ = ctx
	return map[string]any{
		"recipient": stringValue(args, "recipient"),
		"channel":   stringValueWithDefault(args, "channel", "email"),
		"status":    "queued",
	}, nil
}

func deleteRecord(ctx context.Context, args map[string]any) (map[string]any, error) {
	_ = ctx
	return map[string]any{
		"record_id": stringValue(args, "record_id"),
		"status":    "delete_scheduled",
		"note":      "mock tool; no real data was deleted",
	}, nil
}

func stringValue(args map[string]any, key string) string {
	value, _ := args[key].(string)
	return value
}

func stringValueWithDefault(args map[string]any, key string, fallback string) string {
	value := stringValue(args, key)
	if value == "" {
		return fallback
	}
	return value
}

func numberValue(args map[string]any, key string, fallback int) int {
	value, ok := args[key]
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	default:
		return fallback
	}
}
