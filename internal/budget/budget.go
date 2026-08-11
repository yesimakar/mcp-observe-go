package budget

import (
	"fmt"

	"github.com/yesimakar/mcp-observe-go/internal/contracts"
	"github.com/yesimakar/mcp-observe-go/internal/tools"
)

type Result struct {
	Allowed  bool   `json:"allowed"`
	Reason   string `json:"reason"`
	MaxUnits int    `json:"max_units"`
	Cost     int    `json:"cost_units"`
}

func Check(req contracts.ToolCallRequest, tool tools.Tool) Result {
	maxUnits := 10
	if req.Budget != nil && req.Budget.MaxUnits > 0 {
		maxUnits = req.Budget.MaxUnits
	}

	if tool.CostUnits > maxUnits {
		return Result{
			Allowed:  false,
			Reason:   fmt.Sprintf("tool cost %d exceeds budget %d", tool.CostUnits, maxUnits),
			MaxUnits: maxUnits,
			Cost:     tool.CostUnits,
		}
	}

	return Result{
		Allowed:  true,
		Reason:   "tool call is within budget",
		MaxUnits: maxUnits,
		Cost:     tool.CostUnits,
	}
}
