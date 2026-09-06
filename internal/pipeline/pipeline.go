package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AAH20/opentelemetry-infrastructure-graph-collector/internal/model"
)

var allowedEvidence = map[string]bool{"observed": true, "declared": true, "calculated": true, "inferred": true}
var secretKeys = []string{"password", "secret", "token", "private_key", "connection_string"}

func Process(input model.Batch, allowInferred bool) (model.Batch, model.Metrics, error) {
	metrics := model.Metrics{InputObservations: len(input.Observations)}
	byID := map[string]model.Observation{}
	for _, item := range input.Observations {
		if item.ID == "" || item.Kind == "" || item.Source == "" || item.ObservedAt == "" {
			return model.Batch{}, metrics, errors.New("observation requires id, kind, source and observedAt")
		}
		if containsSecret(item.Attributes) {
			metrics.SecretRejectCount++
			return model.Batch{}, metrics, fmt.Errorf("sensitive attribute rejected for %s", item.ID)
		}
		if old, ok := byID[item.ID]; ok {
			if old.Kind != item.Kind || old.Name != item.Name {
				metrics.IdentityCollisions++
				return model.Batch{}, metrics, fmt.Errorf("identity collision for %s", item.ID)
			}
			metrics.DuplicateCount++
			continue
		}
		byID[item.ID] = item
	}
	ids := make([]string, 0, len(byID))
	owners := 0
	for id, item := range byID {
		ids = append(ids, id)
		if owner, ok := item.Attributes["owner"]; ok && fmt.Sprint(owner) != "" {
			owners++
		}
	}
	sort.Strings(ids)
	output := model.Batch{BatchID: input.BatchID}
	for _, id := range ids {
		output.Observations = append(output.Observations, byID[id])
	}
	seen := map[string]bool{}
	for _, edge := range input.Relationships {
		if !allowedEvidence[edge.EvidenceType] {
			return model.Batch{}, metrics, fmt.Errorf("invalid evidence type %s", edge.EvidenceType)
		}
		if edge.EvidenceType == "inferred" && !allowInferred {
			continue
		}
		if _, ok := byID[edge.Source]; !ok {
			return model.Batch{}, metrics, fmt.Errorf("unknown edge source %s", edge.Source)
		}
		if _, ok := byID[edge.Target]; !ok {
			return model.Batch{}, metrics, fmt.Errorf("unknown edge target %s", edge.Target)
		}
		key := edge.Source + "|" + edge.Target + "|" + edge.Relation + "|" + edge.Evidence
		if seen[key] {
			metrics.DuplicateCount++
			continue
		}
		seen[key] = true
		output.Relationships = append(output.Relationships, edge)
	}
	sort.Slice(output.Relationships, func(i, j int) bool {
		a, b := output.Relationships[i], output.Relationships[j]
		return a.Source+a.Target+a.Relation+a.Evidence < b.Source+b.Target+b.Relation+b.Evidence
	})
	canonical, _ := json.Marshal(output)
	digest := sha256.Sum256(canonical)
	output.Receipt = "sha256:" + hex.EncodeToString(digest[:])
	metrics.OutputObservations = len(output.Observations)
	metrics.Relationships = len(output.Relationships)
	if len(output.Observations) > 0 {
		metrics.OwnerCoveragePct = float64(owners) * 100 / float64(len(output.Observations))
	}
	return output, metrics, nil
}

func containsSecret(attributes map[string]any) bool {
	for key := range attributes {
		lower := strings.ToLower(key)
		for _, banned := range secretKeys {
			if strings.Contains(lower, banned) {
				return true
			}
		}
	}
	return false
}
