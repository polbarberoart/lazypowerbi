package domain

import "testing"

func TestWorkspaceDisplayString(t *testing.T) {
	ws := &Workspace{ID: "ws-1", Name: "Sales Analytics"}
	if got := ws.DisplayString(); got != "Sales Analytics" {
		t.Errorf("DisplayString() = %q, want %q", got, "Sales Analytics")
	}
}

func TestWorkspaceGetID(t *testing.T) {
	ws := &Workspace{ID: "ws-1", Name: "Sales Analytics"}
	if got := ws.GetID(); got != "ws-1" {
		t.Errorf("GetID() = %q, want %q", got, "ws-1")
	}
}

func TestWorkspaceGetDisplaySuffix(t *testing.T) {
	ws := &Workspace{ID: "ws-1", Type: "Workspace"}
	if got := ws.GetDisplaySuffix(); got != "Workspace" {
		t.Errorf("GetDisplaySuffix() = %q, want %q", got, "Workspace")
	}
}

func TestItemDisplayString(t *testing.T) {
	it := &Item{ID: "item-1", Name: "Q1 Report", Kind: "Report"}
	if got := it.DisplayString(); got != "Q1 Report" {
		t.Errorf("DisplayString() = %q, want %q", got, "Q1 Report")
	}
}

func TestItemGetDisplaySuffix(t *testing.T) {
	it := &Item{ID: "item-1", Kind: "Report"}
	if got := it.GetDisplaySuffix(); got != "Report" {
		t.Errorf("GetDisplaySuffix() = %q, want %q", got, "Report")
	}
}
