package schema

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/yesimakar/mcp-observe-go/internal/tools"
)

type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

func Validate(tool tools.Tool, args map[string]any) ValidationResult {
	errors := make([]string, 0)

	for _, required := range tool.Schema.Required {
		if _, ok := args[required]; !ok {
			errors = append(errors, fmt.Sprintf("missing required argument %q", required))
		}
	}

	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		expected, ok := tool.Schema.Properties[key]
		if !ok {
			errors = append(errors, fmt.Sprintf("unexpected argument %q", key))
			continue
		}
		if !matchesType(args[key], expected) {
			errors = append(errors, fmt.Sprintf("argument %q expected %s but got %s", key, expected, typeName(args[key])))
		}
	}

	return ValidationResult{Valid: len(errors) == 0, Errors: errors}
}

func matchesType(value any, expected string) bool {
	switch expected {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case float64, float32, int, int64, int32:
			return true
		default:
			return false
		}
	case "boolean":
		_, ok := value.(bool)
		return ok
	default:
		return true
	}
}

func typeName(value any) string {
	if value == nil {
		return "null"
	}
	return reflect.TypeOf(value).String()
}
