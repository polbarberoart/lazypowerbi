package powerbi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/polbarberoart/lazypowerbi/pkg/domain"
)

// WorkspacesClient provides operations on Power BI workspaces.
type WorkspacesClient struct {
	client *Client
}

// NewWorkspacesClient creates a WorkspacesClient backed by the given Client.
func NewWorkspacesClient(client *Client) *WorkspacesClient {
	return &WorkspacesClient{client: client}
}

// listWorkspacesResponse is the JSON envelope returned by the Power BI groups endpoint.
type listWorkspacesResponse struct {
	Value []domain.Workspace `json:"value"`
}

// ListWorkspaces returns all workspaces the authenticated user has access to.
func (w *WorkspacesClient) ListWorkspaces(ctx context.Context) ([]domain.Workspace, error) {
	body, err := w.client.doRequest(ctx, "GET", "/groups")
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}

	var resp listWorkspacesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("list workspaces: failed to decode response: %w", err)
	}

	return resp.Value, nil
}
