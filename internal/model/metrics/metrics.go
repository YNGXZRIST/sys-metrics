package metrics

// Metrics is the metric DTO for the JSON API and internal repositories.
// It uses a flat model without nested structs.
// Delta and Value are pointers so zero can be distinguished from "not set" in JSON.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
