package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/yesimakar/mcp-observe-go/internal/audit"
)

func main() {
	path := flag.String("path", "audit/tool_calls.jsonl", "path to JSONL audit log")
	limit := flag.Int("limit", 20, "number of recent records to show")
	flag.Parse()

	records, err := audit.ReadLast(*path, *limit)
	if err != nil {
		log.Fatalf("read audit log: %v", err)
	}

	if len(records) == 0 {
		fmt.Println("No audit records found.")
		return
	}

	for _, record := range records {
		fmt.Printf("%s | %s | %s | %s | executed=%v | trace=%s | audit=%s\n",
			record.ObservedAt.Format("2006-01-02 15:04:05"),
			record.Decision,
			record.RiskLevel,
			record.ToolName,
			record.Executed,
			record.TraceID,
			record.AuditID,
		)
	}
}
