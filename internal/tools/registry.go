package tools

import (
	"context"
	"fmt"
	"sort"
)

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	RiskLevel   string      `json:"risk_level"`
	CostUnits   int         `json:"cost_units"`
	Schema      ToolSchema  `json:"schema"`
	Handler     ToolHandler `json:"-"`
}

type ToolSchema struct {
	Required   []string          `json:"required"`
	Properties map[string]string `json:"properties"`
}

type ToolHandler func(ctx context.Context, args map[string]any) (map[string]any, error)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	registry := &Registry{tools: make(map[string]Tool)}
	registry.Register(defaultTools()...)
	return registry
}

func (r *Registry) Register(items ...Tool) {
	for _, item := range items {
		r.tools[item.Name] = item
	}
}

func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) List() []Tool {
	items := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		items = append(items, tool)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func (r *Registry) MustGet(name string) Tool {
	tool, ok := r.Get(name)
	if !ok {
		panic(fmt.Sprintf("tool %s not registered", name))
	}
	return tool
}
