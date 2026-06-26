package ui

import (
	gocui "github.com/jesseduffield/gocui"
)

// setupKeybindings registers all keyboard shortcuts for the app.
func (a *App) setupKeybindings() error {
	// --- global ---
	for _, key := range []any{'q', gocui.KeyCtrlC} {
		if err := a.gui.SetKeybinding("", key, gocui.ModNone, a.quit); err != nil {
			return err
		}
	}

	if err := a.gui.SetKeybinding("", gocui.KeyTab, gocui.ModNone, a.nextPanel); err != nil {
		return err
	}

	// --- workspaces panel ---
	for _, key := range []any{gocui.KeyArrowDown, 'j'} {
		if err := a.gui.SetKeybinding("workspaces", key, gocui.ModNone, a.nextWorkspace); err != nil {
			return err
		}
	}
	for _, key := range []any{gocui.KeyArrowUp, 'k'} {
		if err := a.gui.SetKeybinding("workspaces", key, gocui.ModNone, a.prevWorkspace); err != nil {
			return err
		}
	}

	// --- items panel ---
	for _, key := range []any{gocui.KeyArrowDown, 'j'} {
		if err := a.gui.SetKeybinding("items", key, gocui.ModNone, a.nextItem); err != nil {
			return err
		}
	}
	for _, key := range []any{gocui.KeyArrowUp, 'k'} {
		if err := a.gui.SetKeybinding("items", key, gocui.ModNone, a.prevItem); err != nil {
			return err
		}
	}

	return nil
}

// quit signals gocui to stop the main loop.
func (a *App) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

// nextPanel cycles focus: workspaces → items → workspaces.
func (a *App) nextPanel(_ *gocui.Gui, _ *gocui.View) error {
	switch a.activePanel {
	case "workspaces":
		a.activePanel = "items"
	default:
		a.activePanel = "workspaces"
	}
	_, err := a.gui.SetCurrentView(a.activePanel)
	return err
}

// nextWorkspace moves the selection down and reloads items for the new workspace.
func (a *App) nextWorkspace(_ *gocui.Gui, _ *gocui.View) error {
	a.mu.Lock()
	if len(a.workspaces) > 0 && a.selectedWS < len(a.workspaces)-1 {
		a.selectedWS++
		a.items = nil
		a.selectedItem = 0
	}
	a.mu.Unlock()
	a.refreshAll()
	a.loadItems()
	return nil
}

// prevWorkspace moves the selection up and reloads items for the new workspace.
func (a *App) prevWorkspace(_ *gocui.Gui, _ *gocui.View) error {
	a.mu.Lock()
	if a.selectedWS > 0 {
		a.selectedWS--
		a.items = nil
		a.selectedItem = 0
	}
	a.mu.Unlock()
	a.refreshAll()
	a.loadItems()
	return nil
}

// nextItem moves the selection down in the items list.
func (a *App) nextItem(_ *gocui.Gui, _ *gocui.View) error {
	a.mu.Lock()
	if len(a.items) > 0 && a.selectedItem < len(a.items)-1 {
		a.selectedItem++
	}
	a.mu.Unlock()
	a.refreshAll()
	return nil
}

// prevItem moves the selection up in the items list.
func (a *App) prevItem(_ *gocui.Gui, _ *gocui.View) error {
	a.mu.Lock()
	if a.selectedItem > 0 {
		a.selectedItem--
	}
	a.mu.Unlock()
	a.refreshAll()
	return nil
}
