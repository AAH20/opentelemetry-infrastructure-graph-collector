package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/AAH20/opentelemetry-infrastructure-graph-collector/internal/model"
	"github.com/AAH20/opentelemetry-infrastructure-graph-collector/internal/pipeline"
)

func main() {
	input := flag.String("input", "", "offline observation batch")
	output := flag.String("output", "generated/graph-batch.json", "normalized output")
	allowInferred := flag.Bool("allow-inferred", false, "retain visibly classified inferred edges")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "--input is required")
		os.Exit(2)
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		fail(err)
	}
	var batch model.Batch
	if err := json.Unmarshal(raw, &batch); err != nil {
		fail(err)
	}
	result, metrics, err := pipeline.Process(batch, *allowInferred)
	if err != nil {
		fail(err)
	}
	encoded, _ := json.MarshalIndent(result, "", "  ")
	if err := os.MkdirAll(dir(*output), 0o755); err != nil {
		fail(err)
	}
	if err := os.WriteFile(*output, append(encoded, '\n'), 0o644); err != nil {
		fail(err)
	}
	metricJSON, _ := json.MarshalIndent(metrics, "", "  ")
	fmt.Println(string(metricJSON))
}

func dir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
