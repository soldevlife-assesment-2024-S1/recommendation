package gorules

import (
	"os"

	"github.com/gorules/zen-go"
)

func Init(filePath string) (zen.Decision, error) {
	engine := zen.NewEngine(zen.EngineConfig{})

	graph, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	decision, err := engine.CreateDecision(graph)
	if err != nil {
		return nil, err
	}

	return decision, nil
}
