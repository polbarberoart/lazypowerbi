package powerbi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/polbarberoart/lazypowerbi/pkg/domain"
)

// ItemsClient provides operations on Power BI items within a workspace.
type ItemsClient struct {
	client *Client
}

// NewItemsClient creates an ItemsClient backed by the given Client.
func NewItemsClient(client *Client) *ItemsClient {
	return &ItemsClient{client: client}
}

// apiItem is the raw shape returned by the Power BI report/dataset/dashboard endpoints.
type apiItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	WebURL string `json:"webUrl"`
}

// listItemsResponse is the JSON envelope for item list endpoints.
type listItemsResponse struct {
	Value []apiItem `json:"value"`
}

// ListItems returns all items (reports, datasets, dashboards) in the given workspace.
func (c *ItemsClient) ListItems(ctx context.Context, workspaceID string) ([]domain.Item, error) {
	kinds := []string{"Report", "Dataset", "Dashboard"}
	endpoints := map[string]string{
		"Report":    "/groups/" + workspaceID + "/reports",
		"Dataset":   "/groups/" + workspaceID + "/datasets",
		"Dashboard": "/groups/" + workspaceID + "/dashboards",
	}

	var items []domain.Item

	for _, kind := range kinds {
		fetched, err := c.fetchItemsOfKind(ctx, workspaceID, kind, endpoints[kind])
		if err != nil {
			return nil, err
		}
		items = append(items, fetched...)
	}

	return items, nil
}

// fetchItemsOfKind calls a single endpoint and maps the results to domain.Item.
func (c *ItemsClient) fetchItemsOfKind(ctx context.Context, workspaceID, kind, path string) ([]domain.Item, error) {
	body, err := c.client.doRequest(ctx, "GET", path)
	if err != nil {
		return nil, fmt.Errorf("list %s items: %w", kind, err)
	}

	var resp listItemsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("list %s items: failed to decode response: %w", kind, err)
	}

	items := make([]domain.Item, len(resp.Value))
	for i, raw := range resp.Value {
		items[i] = domain.Item{
			ID:          raw.ID,
			Name:        raw.Name,
			Kind:        kind,
			WorkspaceID: workspaceID,
			WebURL:      raw.WebURL,
		}
	}

	return items, nil
}
