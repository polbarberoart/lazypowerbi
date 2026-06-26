package powerbi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/golang-jwt/jwt/v5"
	"github.com/polbarberoart/lazypowerbi/pkg/domain"
)

// powerBIScope is the OAuth scope required to call the Power BI REST API.
const powerBIScope = "https://analysis.windows.net/powerbi/api/.default"

// baseURL is the root of the Power BI REST API for the user's own organization.
const baseURL = "https://api.powerbi.com/v1.0/myorg"

// Client wraps an Azure credential and an HTTP client to call the Power BI REST API.
type Client struct {
	credential azcore.TokenCredential
	httpClient *http.Client
}

// NewClient creates a new Power BI client using DefaultAzureCredential.
func NewClient() (*Client, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	return &Client{
		credential: credential,
		httpClient: &http.Client{},
	}, nil
}

// Credential returns the underlying token credential.
func (c *Client) Credential() azcore.TokenCredential {
	return c.credential
}

// VerifyAuthentication checks if the client can authenticate.
func (c *Client) VerifyAuthentication(ctx context.Context) error {
	_, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{powerBIScope},
	})
	return err
}

// doRequest performs an authenticated GET against the Power BI REST API and
// returns the raw response body. path must start with "/" and is appended to baseURL.
func (c *Client) doRequest(ctx context.Context, method, path string) ([]byte, error) {
	token, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{powerBIScope},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request for %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s returned status %d: %s", path, resp.StatusCode, string(body))
	}

	return body, nil
}

// azureTokenClaims represents the claims we expect in an Azure access token.
type azureTokenClaims struct {
	TenantID          string `json:"tid"`
	ObjectID          string `json:"oid"`
	UserPrincipalName string `json:"upn"`
	PreferredUsername string `json:"preferred_username"`
	AppID             string `json:"appid"`
	Azp               string `json:"azp"`
	Name              string `json:"name"`
}

// GetUserInfo retrieves information about the currently authenticated user
// by parsing the access token JWT.
func (c *Client) GetUserInfo(ctx context.Context) (*domain.User, error) {
	tokenResponse, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{powerBIScope},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	claims, err := parseAzureToken(tokenResponse.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	user := &domain.User{TenantID: claims.TenantID}

	hasUserIdentity := claims.UserPrincipalName != "" || claims.PreferredUsername != "" || claims.Name != ""
	if !hasUserIdentity && (claims.AppID != "" || claims.Azp != "") {
		user.Type = "serviceprincipal"
		appID := claims.AppID
		if appID == "" {
			appID = claims.Azp
		}
		user.UserPrincipalName = appID
		user.DisplayName = appID
	} else {
		user.Type = "user"
		switch {
		case claims.UserPrincipalName != "":
			user.UserPrincipalName = claims.UserPrincipalName
		case claims.PreferredUsername != "":
			user.UserPrincipalName = claims.PreferredUsername
		case claims.ObjectID != "":
			user.UserPrincipalName = claims.ObjectID
		}
		if claims.Name != "" {
			user.DisplayName = claims.Name
		} else {
			user.DisplayName = user.UserPrincipalName
		}
	}

	return user, nil
}

// parseAzureToken parses an Azure access token JWT and extracts the claims.
// The token is not verified as it comes from the trusted Azure SDK.
func parseAzureToken(tokenString string) (*azureTokenClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			return mapClaimsToAzureClaims(claims), nil
		}
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims azureTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return &claims, nil
}

// mapClaimsToAzureClaims converts jwt.MapClaims to azureTokenClaims.
func mapClaimsToAzureClaims(mapClaims jwt.MapClaims) *azureTokenClaims {
	claims := &azureTokenClaims{}
	if v, ok := mapClaims["tid"].(string); ok {
		claims.TenantID = v
	}
	if v, ok := mapClaims["oid"].(string); ok {
		claims.ObjectID = v
	}
	if v, ok := mapClaims["upn"].(string); ok {
		claims.UserPrincipalName = v
	}
	if v, ok := mapClaims["preferred_username"].(string); ok {
		claims.PreferredUsername = v
	}
	if v, ok := mapClaims["appid"].(string); ok {
		claims.AppID = v
	}
	if v, ok := mapClaims["azp"].(string); ok {
		claims.Azp = v
	}
	if v, ok := mapClaims["name"].(string); ok {
		claims.Name = v
	}
	return claims
}
