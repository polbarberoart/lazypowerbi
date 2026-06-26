package ui

import (
	"sync"

	"github.com/polbarberoart/lazypowerbi/pkg/domain"
	gocui "github.com/jesseduffield/gocui"
)

// App holds all runtime state for the TUI.
type App struct {
	gui           *gocui.Gui
	powerbiClient PowerBIClient
	workspacesClient WorkspacesClient
	itemsClient   ItemsClient
	user          *domain.User

	// views — referencias directas para evitar búsquedas repetidas
	headerView     *gocui.View
	workspacesView *gocui.View
	itemsView      *gocui.View
	detailsView    *gocui.View
	statusView     *gocui.View

	// estado de navegación
	workspaces   []domain.Workspace
	selectedWS   int
	items        []domain.Item
	selectedItem int
	activePanel  string

	mu sync.RWMutex
}

// New creates an App wired to the given Power BI clients.
func New(
	powerbiClient PowerBIClient,
	workspacesClient WorkspacesClient,
	itemsClient ItemsClient,
) (*App, error) {
	g, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
	})
	if err != nil {
		return nil, err
	}

	app := &App{
		gui:              g,
		powerbiClient:    powerbiClient,
		workspacesClient: workspacesClient,
		itemsClient:      itemsClient,
		activePanel:      "workspaces",
	}

	g.SetManagerFunc(app.layout)

	return app, nil
}

// Run starts the TUI main loop. Blocks until the user quits.
func (a *App) Run() error {
	defer a.gui.Close()

	if err := a.setupKeybindings(); err != nil {
		return err
	}

	a.loadUserInfo()
	a.loadWorkspaces()

	return a.gui.MainLoop()
}
