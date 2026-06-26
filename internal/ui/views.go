package ui

import (
	"fmt"

	"github.com/polbarberoart/lazypowerbi/pkg/domain"
	gocui "github.com/jesseduffield/gocui"
)

// refreshAll redraws every panel. Call after any state change.
func (a *App) refreshAll() {
	a.refreshHeader()
	a.refreshWorkspacesPanel()
	a.refreshItemsPanel()
	a.refreshDetailsPanel()
	a.refreshStatusPanel()
}

func (a *App) refreshHeader() {
	if a.headerView == nil {
		return
	}
	a.headerView.Clear()
	fmt.Fprint(a.headerView, " lazypowerbi — Power BI explorer")
}

func (a *App) refreshWorkspacesPanel() {
	if a.workspacesView == nil {
		return
	}
	a.workspacesView.Clear()

	a.mu.RLock()
	workspaces := a.workspaces
	selected := a.selectedWS
	a.mu.RUnlock()

	if len(workspaces) == 0 {
		fmt.Fprint(a.workspacesView, " Loading...")
		return
	}

	for i, ws := range workspaces {
		if i == selected {
			fmt.Fprintf(a.workspacesView, "> %s\n", ws.Name)
		} else {
			fmt.Fprintf(a.workspacesView, "  %s\n", ws.Name)
		}
	}
}

func (a *App) refreshItemsPanel() {
	if a.itemsView == nil {
		return
	}
	a.itemsView.Clear()

	a.mu.RLock()
	items := a.items
	selected := a.selectedItem
	a.mu.RUnlock()

	if len(items) == 0 {
		fmt.Fprint(a.itemsView, " Select a workspace")
		return
	}

	for i, item := range items {
		if i == selected {
			fmt.Fprintf(a.itemsView, "> %-30s [%s]\n", item.Name, item.Kind)
		} else {
			fmt.Fprintf(a.itemsView, "  %-30s [%s]\n", item.Name, item.Kind)
		}
	}
}

func (a *App) refreshDetailsPanel() {
	if a.detailsView == nil {
		return
	}
	a.detailsView.Clear()

	a.mu.RLock()
	activePanel := a.activePanel
	workspaces := a.workspaces
	selectedWS := a.selectedWS
	items := a.items
	selectedItem := a.selectedItem
	a.mu.RUnlock()

	switch activePanel {
	case "workspaces":
		if len(workspaces) == 0 {
			return
		}
		a.renderWorkspaceDetails(workspaces[selectedWS])
	case "items":
		if len(items) == 0 {
			return
		}
		a.renderItemDetails(items[selectedItem])
	}
}

func (a *App) renderWorkspaceDetails(ws domain.Workspace) {
	printField(a.detailsView, "ID", ws.ID)
	printField(a.detailsView, "Name", ws.Name)
	printField(a.detailsView, "Type", ws.Type)
	printField(a.detailsView, "Read Only", fmt.Sprintf("%v", ws.IsReadOnly))
	printField(a.detailsView, "Dedicated Capacity", fmt.Sprintf("%v", ws.IsOnDedicatedCapacity))
	if ws.CapacityID != "" {
		printField(a.detailsView, "Capacity ID", ws.CapacityID)
	}
}

func (a *App) renderItemDetails(item domain.Item) {
	printField(a.detailsView, "ID", item.ID)
	printField(a.detailsView, "Name", item.Name)
	printField(a.detailsView, "Kind", item.Kind)
	printField(a.detailsView, "Workspace ID", item.WorkspaceID)
	if item.WebURL != "" {
		printField(a.detailsView, "Web URL", item.WebURL)
	}
}

func (a *App) refreshStatusPanel() {
	if a.statusView == nil {
		return
	}
	a.statusView.Clear()

	a.mu.RLock()
	user := a.user
	activePanel := a.activePanel
	a.mu.RUnlock()

	var userStr string
	if user != nil {
		userStr = fmt.Sprintf("%s (%s)", user.DisplayName, user.TenantID)
	} else {
		userStr = "Authenticating..."
	}

	var keys string
	switch activePanel {
	case "workspaces":
		keys = "↑↓/jk: navigate  Tab: items  q: quit"
	case "items":
		keys = "↑↓/jk: navigate  Tab: workspaces  q: quit"
	}

	fmt.Fprintf(a.statusView, " %s  |  %s", userStr, keys)
}

// printField writes a key: value line to a view.
func printField(v *gocui.View, key, value string) {
	fmt.Fprintf(v, " %-20s %s\n", key+":", value)
}
