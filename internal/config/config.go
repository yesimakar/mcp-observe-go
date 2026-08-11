package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address         string
	AuditPath       string
	ServiceName     string
	OTelExporter    string
	OTLPEndpoint    string
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		Address:         getEnv("MCP_OBSERVE_ADDR", ":8080"),
		AuditPath:       getEnv("MCP_OBSERVE_AUDIT_PATH", "audit/tool_calls.jsonl"),
		ServiceName:     getEnv("OTEL_SERVICE_NAME", "mcp-observe-go"),
		OTelExporter:    getEnv("MCP_OBSERVE_OTEL_EXPORTER", "stdout"),
		OTLPEndpoint:    getEnv("MCP_OBSERVE_OTLP_ENDPOINT", "127.0.0.1:4317"),
		ShutdownTimeout: time.Duration(getEnvInt("MCP_OBSERVE_SHUTDOWN_SECONDS", 5)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
