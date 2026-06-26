package ui

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/polbarberoart/lazypowerbi/pkg/domain"
	"github.com/polbarberoart/lazypowerbi/pkg/powerbi"
)

// Compile-time checks: los tipos concretos deben implementar las interfaces.
var _ PowerBIClient = (*powerbi.Client)(nil)
var _ WorkspacesClient = (*powerbi.WorkspacesClient)(nil)
var _ ItemsClient = (*powerbi.ItemsClient)(nil)

// PowerBIClient provides authentication and user info.
type PowerBIClient interface {
	GetUserInfo(ctx context.Context) (*domain.User, error)
	VerifyAuthentication(ctx context.Context) error
	Credential() azcore.TokenCredential
}

// WorkspacesClient provides workspace operations.
type WorkspacesClient interface {
	ListWorkspaces(ctx context.Context) ([]domain.Workspace, error)
}

// ItemsClient provides item operations within a workspace.
type ItemsClient interface {
	ListItems(ctx context.Context, workspaceID string) ([]domain.Item, error)
}
