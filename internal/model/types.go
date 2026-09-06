package model

type Observation struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	Name       string         `json:"name"`
	Source     string         `json:"source"`
	ObservedAt string         `json:"observedAt"`
	Attributes map[string]any `json:"attributes"`
}

type Relationship struct {
	Source       string `json:"source"`
	Target       string `json:"target"`
	Relation     string `json:"relation"`
	Evidence     string `json:"evidence"`
	EvidenceType string `json:"evidenceType"`
	ObservedAt   string `json:"observedAt"`
}

type Batch struct {
	BatchID       string         `json:"batchId"`
	Observations  []Observation  `json:"observations"`
	Relationships []Relationship `json:"relationships"`
	Receipt       string         `json:"receipt,omitempty"`
}

type Metrics struct {
	InputObservations  int     `json:"inputObservations"`
	OutputObservations int     `json:"outputObservations"`
	Relationships      int     `json:"relationships"`
	DuplicateCount     int     `json:"duplicateCount"`
	SecretRejectCount  int     `json:"secretRejectCount"`
	IdentityCollisions int     `json:"identityCollisions"`
	OwnerCoveragePct   float64 `json:"ownerCoveragePercent"`
}
