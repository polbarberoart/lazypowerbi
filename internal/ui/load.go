package ui

import (
	"context"
	"time"

	gocui "github.com/jesseduffield/gocui"
)

const apiTimeout = 30 * time.Second

// loadUserInfo fetches the authenticated user in the background and refreshes the status bar.
func (a *App) loadUserInfo() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		user, err := a.powerbiClient.GetUserInfo(ctx)
		if err != nil {
			return
		}

		a.mu.Lock()
		a.user = user
		a.mu.Unlock()

		a.gui.UpdateAsync(func(_ *gocui.Gui) error {
			a.refreshStatusPanel()
			return nil
		})
	}()
}

// loadWorkspaces fetches all workspaces in the background and refreshes the left panel.
func (a *App) loadWorkspaces() {
	a.gui.UpdateAsync(func(_ *gocui.Gui) error {
		a.refreshWorkspacesPanel()
		return nil
	})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		workspaces, err := a.workspacesClient.ListWorkspaces(ctx)
		if err != nil {
			return
		}

		a.mu.Lock()
		a.workspaces = workspaces
		a.selectedWS = 0
		a.mu.Unlock()

		a.gui.UpdateAsync(func(_ *gocui.Gui) error {
			a.refreshWorkspacesPanel()
			a.refreshDetailsPanel()
			return nil
		})
	}()
}

// loadItems fetches items for the currently selected workspace in the background.
func (a *App) loadItems() {
	a.mu.RLock()
	workspaces := a.workspaces
	selected := a.selectedWS
	a.mu.RUnlock()

	if len(workspaces) == 0 {
		return
	}

	workspaceID := workspaces[selected].ID

	a.gui.UpdateAsync(func(_ *gocui.Gui) error {
		a.refreshItemsPanel()
		return nil
	})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		items, err := a.itemsClient.ListItems(ctx, workspaceID)
		if err != nil {
			return
		}

		a.mu.Lock()
		a.items = items
		a.selectedItem = 0
		a.mu.Unlock()

		a.gui.UpdateAsync(func(_ *gocui.Gui) error {
			a.refreshItemsPanel()
			a.refreshDetailsPanel()
			return nil
		})
	}()
}
