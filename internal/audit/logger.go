package audit

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Record struct {
	AuditID     string         `json:"audit_id"`
	TraceID     string         `json:"trace_id"`
	SessionID   string         `json:"session_id"`
	ClientName  string         `json:"client_name"`
	Actor       string         `json:"actor"`
	ToolName    string         `json:"tool_name"`
	Environment string         `json:"environment"`
	Decision    string         `json:"decision"`
	RiskLevel   string         `json:"risk_level"`
	Reason      string         `json:"reason"`
	Executed    bool           `json:"executed"`
	DurationMS  int64          `json:"duration_ms"`
	Arguments   map[string]any `json:"arguments_redacted"`
	Error       string         `json:"error,omitempty"`
	ObservedAt  time.Time      `json:"observed_at"`
}

type Logger struct {
	path string
	mu   sync.Mutex
}

func NewLogger(path string) *Logger {
	return &Logger{path: path}
}

func (l *Logger) Write(record Record) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if record.AuditID == "" {
		record.AuditID = "audit_" + randomHex(16)
	}

	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return "", err
	}

	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoded, err := json.Marshal(record)
	if err != nil {
		return "", err
	}

	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return "", err
	}

	return record.AuditID, nil
}

func ReadLast(path string, limit int) ([]Record, error) {
	if limit <= 0 {
		limit = 20
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Record{}, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	records := make([]Record, 0)
	for scanner.Scan() {
		var record Record
		if err := json.Unmarshal(scanner.Bytes(), &record); err == nil {
			records = append(records, record)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(records) <= limit {
		return records, nil
	}
	return records[len(records)-limit:], nil
}

func RedactArguments(args map[string]any) map[string]any {
	redacted := make(map[string]any, len(args))
	for key, value := range args {
		normalized := key
		if normalized == "token" || normalized == "api_key" || normalized == "password" || normalized == "secret" {
			redacted[key] = "[redacted]"
			continue
		}
		redacted[key] = value
	}
	return redacted
}

func randomHex(bytes int) string {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}
