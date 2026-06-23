package domain

// Item represents a generic Power BI artifact: dataset, report, dashboard, or dataflow.
type Item struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Kind        string                 `json:"kind"` // "Dataset" | "Report" | "Dashboard" | "Dataflow"
	WorkspaceID string                 `json:"workspaceId"`
	WebURL      string                 `json:"webUrl"`
	Properties  map[string]interface{} `json:"properties"`
}

// DisplayString returns a string representation for the UI.
func (i *Item) DisplayString() string {
	return i.Name
}

// GetID returns the item ID.
func (i *Item) GetID() string {
	return i.ID
}

// GetDisplaySuffix returns the suffix to display (item kind).
func (i *Item) GetDisplaySuffix() string {
	return i.Kind
}
