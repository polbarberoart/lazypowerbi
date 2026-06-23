package domain

// Workspace represents a Power BI workspace.
type Workspace struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"` // "Workspace" | "PersonalGroup"
	IsReadOnly            bool   `json:"isReadOnly"`
	IsOnDedicatedCapacity bool   `json:"isOnDedicatedCapacity"`
	CapacityID            string `json:"capacityId"`
}

// DisplayString returns a string representation for the UI.
func (w *Workspace) DisplayString() string {
	return w.Name
}

// GetID returns the workspace ID.
func (w *Workspace) GetID() string {
	return w.ID
}

// GetDisplaySuffix returns the suffix to display (workspace type).
func (w *Workspace) GetDisplaySuffix() string {
	return w.Type
}
