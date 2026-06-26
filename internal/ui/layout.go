package ui

import (
	"errors"

	gocui "github.com/jesseduffield/gocui"
)

// layout define la posición y tamaño de cada view en cada redibujado.
func (a *App) layout(g *gocui.Gui) error {
	w, h := g.Size()

	leftW := w / 4
	headerH := 2
	statusH := 2
	midH := headerH + (h-headerH-statusH)/2

	if v, err := g.SetView("header", 0, 0, w-1, headerH, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
		a.headerView = v
	}

	if v, err := g.SetView("workspaces", 0, headerH, leftW, midH, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Workspaces "
		v.Highlight = true
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorWhite
		a.workspacesView = v
		if _, err := g.SetCurrentView("workspaces"); err != nil {
			return err
		}
		a.refreshAll()
	}

	if v, err := g.SetView("items", 0, midH, leftW, h-statusH-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Items "
		v.Highlight = true
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorWhite
		a.itemsView = v
	}

	if v, err := g.SetView("details", leftW, headerH, w-1, h-statusH-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Details "
		v.Wrap = true
		a.detailsView = v
	}

	if v, err := g.SetView("status", 0, h-statusH-1, w-1, h-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
		a.statusView = v
	}

	return nil
}
