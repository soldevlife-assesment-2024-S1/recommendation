package gorules

import (
	"os"

	"github.com/gorules/zen-go"
)

func Init() (zen.Decision, error) {
	engine := zen.NewEngine(zen.EngineConfig{})

	graph, err := os.ReadFile("./assets/ticket-discounted.json")
	if err != nil {
		return nil, err
	}

	decision, err := engine.CreateDecision(graph)
	if err != nil {
		return nil, err
	}

	return decision, nil
}
