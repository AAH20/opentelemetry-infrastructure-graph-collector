package pipeline

import (
	"github.com/AAH20/opentelemetry-infrastructure-graph-collector/internal/model"
	"testing"
)

func observation(id, kind string, attrs map[string]any) model.Observation {
	return model.Observation{ID: id, Kind: kind, Name: id, Source: "test", ObservedAt: "2026-09-06T08:00:00Z", Attributes: attrs}
}

func TestProcessDeterministic(t *testing.T) {
	input := model.Batch{BatchID: "one", Observations: []model.Observation{observation("b", "resource", map[string]any{"owner": "x"}), observation("a", "service", map[string]any{"owner": "y"})}, Relationships: []model.Relationship{{Source: "a", Target: "b", Relation: "depends_on", Evidence: "test:1", EvidenceType: "observed", ObservedAt: "2026-09-06T08:00:00Z"}}}
	a, metrics, err := Process(input, false)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := Process(input, false)
	if err != nil {
		t.Fatal(err)
	}
	if a.Receipt != b.Receipt {
		t.Fatal("receipt is not deterministic")
	}
	if metrics.OwnerCoveragePct != 100 {
		t.Fatalf("coverage=%v", metrics.OwnerCoveragePct)
	}
}

func TestRejectsSecretAttributes(t *testing.T) {
	input := model.Batch{Observations: []model.Observation{observation("a", "service", map[string]any{"api_token": "leak"})}}
	_, metrics, err := Process(input, false)
	if err == nil || metrics.SecretRejectCount != 1 {
		t.Fatal("expected secret rejection")
	}
}

func TestRejectsIdentityCollision(t *testing.T) {
	input := model.Batch{Observations: []model.Observation{observation("a", "service", nil), observation("a", "database", nil)}}
	_, metrics, err := Process(input, false)
	if err == nil || metrics.IdentityCollisions != 1 {
		t.Fatal("expected collision")
	}
}

func TestRejectsDanglingEdge(t *testing.T) {
	input := model.Batch{Observations: []model.Observation{observation("a", "service", nil)}, Relationships: []model.Relationship{{Source: "a", Target: "missing", Relation: "calls", Evidence: "x", EvidenceType: "observed"}}}
	if _, _, err := Process(input, false); err == nil {
		t.Fatal("expected missing target error")
	}
}

func TestInferredEdgesDisabledByDefault(t *testing.T) {
	input := model.Batch{Observations: []model.Observation{observation("a", "service", nil), observation("b", "service", nil)}, Relationships: []model.Relationship{{Source: "a", Target: "b", Relation: "might_call", Evidence: "model:1", EvidenceType: "inferred"}}}
	result, _, err := Process(input, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Relationships) != 0 {
		t.Fatal("inferred edge retained")
	}
}
