package main

import (
	"context"
	"fmt"
	"os"

	"github.com/polbarberoart/lazypowerbi/internal/ui"
	"github.com/polbarberoart/lazypowerbi/pkg/powerbi"
)

func main() {
	os.Exit(run())
}

func run() int {
	client, err := powerbi.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Power BI client: %v\n", err)
		return 1
	}

	if err := client.VerifyAuthentication(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please log in with 'az login'\n")
		return 1
	}

	app, err := ui.New(
		client,
		powerbi.NewWorkspacesClient(client),
		powerbi.NewItemsClient(client),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating UI: %v\n", err)
		return 1
	}

	if err := app.Run(); err != nil && err.Error() != "quit" {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}
