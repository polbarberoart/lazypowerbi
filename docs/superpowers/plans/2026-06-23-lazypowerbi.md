# lazypowerbi Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline, not subagent-driven — see note in Task 0). Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `lazypowerbi`, a TUI for exploring Power BI workspaces, as a new standalone Go module at `C:\Repos\lazypowerbi`, reusing lazyazure's architecture (gocui-based 2-tier list sidebar + detail panel, interface-driven Azure/PowerBI client, demo mode).

**Architecture:** Same layered design as lazyazure: `pkg/domain` (plain data types), `pkg/powerbi` (HTTP client hitting the Power BI REST API directly — no SDK exists), `pkg/demo` (mock data behind the same interfaces), `pkg/gui` (gocui controller + reusable `panels` subpackage), `pkg/utils` (logging/clipboard/browser/metrics, mostly copied verbatim), `vendor_gocui` (vendored gocui fork, copied verbatim), wired together in `main.go`.

**Tech Stack:** Go 1.26+, `github.com/jesseduffield/gocui` (vendored fork via `replace`), `github.com/Azure/azure-sdk-for-go/sdk/azidentity` + `azcore`, `github.com/golang-jwt/jwt/v5`, `github.com/atotto/clipboard`, `github.com/pkg/browser`, standard library `net/http`/`encoding/json` for the Power BI REST calls.

## Global Constraints

- Module path: `github.com/polbarbero/lazypowerbi`
- Repo root: `C:\Repos\lazypowerbi` (new, independent repo — never touch `C:\Repos\lazyazure`, only read from it)
- Power BI auth scope: `https://analysis.windows.net/powerbi/api/.default`
- Power BI base URL: `https://api.powerbi.com/v1.0/myorg`
- User-level API only (no Admin API) in this version
- Env var prefix: `LAZYPOWERBI_` (e.g. `LAZYPOWERBI_DEBUG`, `LAZYPOWERBI_DEMO`)
- Sidebar has 2 list panels (`workspaces`, `items`) + `main` detail panel — not 3
- Every new/modified Go file must pass `go build ./...` and `go vet ./...` before a task is considered done
- This is a teaching project: when executing each task, explain the Go concepts involved (pointers, interfaces, goroutines, error handling, structs, generics, packages) before/while writing the code — don't just paste finished code

---

## Task 0: Bootstrap the repository

**Files:**
- Create: `C:\Repos\lazypowerbi\go.mod`
- Create: `C:\Repos\lazypowerbi\.gitignore`
- Create: `C:\Repos\lazypowerbi\main.go` (minimal, just `func main() { fmt.Println("lazypowerbi") }`)
- Copy: `C:\Repos\lazyazure\vendor_gocui` → `C:\Repos\lazypowerbi\vendor_gocui` (verbatim, no changes)
- Move: `C:\Repos\lazyazure\docs\superpowers\specs\2026-06-23-lazypowerbi-design.md` → `C:\Repos\lazypowerbi\docs\superpowers\specs\2026-06-23-lazypowerbi-design.md`
- Move: `C:\Repos\lazyazure\docs\superpowers\plans\2026-06-23-lazypowerbi.md` → `C:\Repos\lazypowerbi\docs\superpowers\plans\2026-06-23-lazypowerbi.md` (this file)

**Interfaces:**
- Produces: a buildable Go module other tasks add packages to.

- [ ] **Step 1: Create the directory and initialize git**

```bash
mkdir -p /c/Repos/lazypowerbi
cd /c/Repos/lazypowerbi
git init
```

- [ ] **Step 2: Create `go.mod`**

```
module github.com/polbarbero/lazypowerbi

go 1.26.2
```

Explain: a Go module is the unit of versioning/dependency resolution — equivalent to a `pyproject.toml`/`setup.py` root, except the import path *is* the module path (no separate "package name" vs "import path" distinction like in Python).

- [ ] **Step 3: Copy `vendor_gocui` verbatim from lazyazure**

```bash
cp -r /c/Repos/lazyazure/vendor_gocui /c/Repos/lazypowerbi/vendor_gocui
```

This is a vendored fork of gocui, used via a `replace` directive (added in Task 9 when gocui is first imported). Explain: `replace` in `go.mod` lets you swap a dependency's source for a local path — same idea as a local pip `-e` install, but module-path-scoped.

- [ ] **Step 4: Create `.gitignore`**

```
*.exe
*.log
.lazypowerbi/
```

- [ ] **Step 5: Create a minimal `main.go`**

```go
package main

import "fmt"

func main() {
	fmt.Println("lazypowerbi")
}
```

- [ ] **Step 6: Verify it builds and runs**

Run: `go build -o lazypowerbi.exe . && ./lazypowerbi.exe`
Expected: prints `lazypowerbi`

Explain here: `go build` compiles to a single static binary — no virtualenv, no interpreter needed at runtime, unlike Python.

- [ ] **Step 7: Move the design doc and this plan into the new repo**

```bash
mkdir -p /c/Repos/lazypowerbi/docs/superpowers/specs /c/Repos/lazypowerbi/docs/superpowers/plans
mv /c/Repos/lazyazure/docs/superpowers/specs/2026-06-23-lazypowerbi-design.md /c/Repos/lazypowerbi/docs/superpowers/specs/
mv /c/Repos/lazyazure/docs/superpowers/plans/2026-06-23-lazypowerbi.md /c/Repos/lazypowerbi/docs/superpowers/plans/
rmdir /c/Repos/lazyazure/docs/superpowers/specs /c/Repos/lazyazure/docs/superpowers/plans /c/Repos/lazyazure/docs/superpowers 2>/dev/null || true
```

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "chore: bootstrap lazypowerbi module"
```

**Note on execution mode:** because this project is explicitly for learning Go, execute this plan with `superpowers:executing-plans` in the *current conversation* (inline), not `subagent-driven-development`. Each task below should be worked through together, with explanations, not delegated to a fresh subagent that would just hand back finished code.

---

## Task 1: Domain types — `Workspace`, `Item`, `User`

**Files:**
- Create: `pkg/domain/workspace.go`
- Create: `pkg/domain/item.go`
- Create: `pkg/domain/user.go` (copy verbatim from `C:\Repos\lazyazure\pkg\domain\user.go`)
- Test: `pkg/domain/domain_test.go`

**Interfaces:**
- Produces:
  - `domain.Workspace{ID, Name, Type, IsReadOnly, IsOnDedicatedCapacity, CapacityID string/bool}` with methods `DisplayString() string`, `GetID() string`, `GetDisplaySuffix() string`
  - `domain.Item{ID, Name, Kind, WorkspaceID, WebURL string; Properties map[string]interface{}}` with the same three methods (`GetDisplaySuffix()` returns `Kind`)
  - `domain.User` (copied as-is from lazyazure)

- [ ] **Step 1: Write the failing test for `Workspace`**

```go
// pkg/domain/domain_test.go
package domain

import "testing"

func TestWorkspaceDisplayString(t *testing.T) {
	ws := &Workspace{ID: "ws-1", Name: "Sales Analytics"}
	if got := ws.DisplayString(); got != "Sales Analytics" {
		t.Errorf("DisplayString() = %q, want %q", got, "Sales Analytics")
	}
}

func TestWorkspaceGetID(t *testing.T) {
	ws := &Workspace{ID: "ws-1", Name: "Sales Analytics"}
	if got := ws.GetID(); got != "ws-1" {
		t.Errorf("GetID() = %q, want %q", got, "ws-1")
	}
}

func TestWorkspaceGetDisplaySuffix(t *testing.T) {
	ws := &Workspace{ID: "ws-1", Type: "Workspace"}
	if got := ws.GetDisplaySuffix(); got != "Workspace" {
		t.Errorf("GetDisplaySuffix() = %q, want %q", got, "Workspace")
	}
}

func TestItemDisplayString(t *testing.T) {
	it := &Item{ID: "item-1", Name: "Q1 Report", Kind: "Report"}
	if got := it.DisplayString(); got != "Q1 Report" {
		t.Errorf("DisplayString() = %q, want %q", got, "Q1 Report")
	}
}

func TestItemGetDisplaySuffix(t *testing.T) {
	it := &Item{ID: "item-1", Kind: "Report"}
	if got := it.GetDisplaySuffix(); got != "Report" {
		t.Errorf("GetDisplaySuffix() = %q, want %q", got, "Report")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/domain/... -v`
Expected: FAIL — `Workspace`/`Item` undefined (the package doesn't compile yet)

Explain: this is the "red" step of TDD — Go's compiler failure *is* a valid test failure here, since `Workspace`/`Item` don't exist yet.

- [ ] **Step 3: Implement `pkg/domain/workspace.go`**

```go
package domain

// Workspace represents a Power BI workspace.
type Workspace struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"` // "Workspace" | "PersonalGroup"
	IsReadOnly            bool   `json:"isReadOnly"`
	IsOnDedicatedCapacity bool   `json:"isOnDedicatedCapacity"`
	CapacityID            string `json:"capacityId"`
}

// DisplayString returns a string representation for the UI.
func (w *Workspace) DisplayString() string {
	return w.Name
}

// GetID returns the workspace ID.
func (w *Workspace) GetID() string {
	return w.ID
}

// GetDisplaySuffix returns the suffix to display (workspace type).
func (w *Workspace) GetDisplaySuffix() string {
	return w.Type
}
```

Explain while writing this: `(w *Workspace)` is a *method receiver* — `w` is a pointer to the struct the method is called on. Using a pointer receiver (vs a value receiver) means the method works on the original struct, not a copy, and is the convention lazyazure already follows for `Subscription`/`Resource`. There's no `self` keyword in Go; the receiver name is just a regular parameter you choose.

- [ ] **Step 4: Implement `pkg/domain/item.go`**

```go
package domain

// Item represents a generic Power BI artifact: dataset, report, dashboard, or dataflow.
type Item struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Kind        string                 `json:"kind"` // "Dataset" | "Report" | "Dashboard" | "Dataflow"
	WorkspaceID string                 `json:"workspaceId"`
	WebURL      string                 `json:"webUrl"`
	Properties  map[string]interface{} `json:"properties"`
}

// DisplayString returns a string representation for the UI.
func (i *Item) DisplayString() string {
	return i.Name
}

// GetID returns the item ID.
func (i *Item) GetID() string {
	return i.ID
}

// GetDisplaySuffix returns the suffix to display (item kind).
func (i *Item) GetDisplaySuffix() string {
	return i.Kind
}
```

- [ ] **Step 5: Copy `pkg/domain/user.go` verbatim**

Run: `cp /c/Repos/lazyazure/pkg/domain/user.go /c/Repos/lazypowerbi/pkg/domain/user.go`

Open it together and read it: it's a plain struct (`DisplayName`, `UserPrincipalName`, `Type`, `TenantID`) with no methods — explain why it needs no `DisplayString()`/`GetID()`: it's never put in a `FilteredList`, only shown in the auth panel.

- [ ] **Step 6: Run the tests again to verify they pass**

Run: `go test ./pkg/domain/... -v`
Expected: PASS (all 5 tests)

- [ ] **Step 7: Commit**

```bash
git add pkg/domain
git commit -m "feat: add Workspace, Item, User domain types"
```

---

## Task 2: Power BI HTTP client core — `pkg/powerbi/client.go`

**Files:**
- Create: `pkg/powerbi/client.go`
- Test: `pkg/powerbi/client_test.go`

**Interfaces:**
- Consumes: `azcore.TokenCredential`, `azidentity.NewDefaultAzureCredential` (from `github.com/Azure/azure-sdk-for-go/sdk/azidentity` and `.../azcore`), `domain.User` (Task 1)
- Produces:
  - `powerbi.Client{}` struct
  - `func NewClient() (*Client, error)`
  - `func (c *Client) Credential() azcore.TokenCredential`
  - `func (c *Client) VerifyAuthentication(ctx context.Context) error`
  - `func (c *Client) GetUserInfo(ctx context.Context) (*domain.User, error)`
  - `func (c *Client) doRequest(ctx context.Context, method, path string) ([]byte, error)` (unexported, used by Tasks 3 and 4)

- [ ] **Step 1: Add the Azure SDK dependencies**

Run:
```bash
go get github.com/Azure/azure-sdk-for-go/sdk/azcore@v1.22.0
go get github.com/Azure/azure-sdk-for-go/sdk/azidentity@v1.14.0
go get github.com/golang-jwt/jwt/v5@v5.3.1
```

Explain: `go get` adds the dependency to `go.mod`/`go.sum` and downloads it to the local module cache — same role as `pip install` + writing into `requirements.txt`, except `go.sum` also pins exact cryptographic hashes of every dependency (Python's closest equivalent is a lockfile like `poetry.lock`, but Go does this by default).

- [ ] **Step 2: Write the failing test for `GetUserInfo`'s token parsing**

This test exercises the *pure* part of the client — parsing a JWT — without needing network access, by building a fake unsigned token.

```go
// pkg/powerbi/client_test.go
package powerbi

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func fakeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal claims: %v", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return strings.Join([]string{header, payload, ""}, ".")
}

func TestParseAzureTokenUser(t *testing.T) {
	token := fakeJWT(t, map[string]any{
		"tid":  "tenant-1",
		"name": "Ada Lovelace",
		"upn":  "ada@example.com",
	})

	claims, err := parseAzureToken(token)
	if err != nil {
		t.Fatalf("parseAzureToken() error = %v", err)
	}
	if claims.TenantID != "tenant-1" {
		t.Errorf("TenantID = %q, want %q", claims.TenantID, "tenant-1")
	}
	if claims.Name != "Ada Lovelace" {
		t.Errorf("Name = %q, want %q", claims.Name, "Ada Lovelace")
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./pkg/powerbi/... -v`
Expected: FAIL — package doesn't exist / `parseAzureToken` undefined

- [ ] **Step 4: Implement `pkg/powerbi/client.go`**

This is adapted from `C:\Repos\lazyazure\pkg\azure\client.go`: same `azureTokenClaims` struct, same `parseAzureToken`/`mapClaimsToAzureClaims`/`GetUserInfo`/`VerifyAuthentication`/`Credential` logic, but the scope passed to `GetToken` changes, and a new `doRequest` method is added for the REST calls Tasks 3–4 will need.

```go
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
	"github.com/polbarbero/lazypowerbi/pkg/domain"
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
```

Explain while reading this together: `doRequest` is unexported (lowercase `d`) — Go's visibility rule is purely about case, no `private`/`public` keywords. Lowercase = package-private, uppercase = exported. `defer resp.Body.Close()` is Go's mechanism for "run this when the function returns, no matter how it returns" — closest Python analogue is a `try/finally`, but `defer` doesn't need the surrounding `try`.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./pkg/powerbi/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum pkg/powerbi
git commit -m "feat: add powerbi.Client with auth and HTTP request core"
```

---

## Task 3: `WorkspacesClient` — `pkg/powerbi/workspaces.go`

**Files:**
- Create: `pkg/powerbi/workspaces.go`
- Test: `pkg/powerbi/workspaces_test.go`

**Interfaces:**
- Consumes: `(c *Client) doRequest(ctx, method, path string) ([]byte, error)` (Task 2)
- Produces:
  - `powerbi.WorkspacesClient{}`
  - `func NewWorkspacesClient(client *Client) *WorkspacesClient`
  - `func (c *WorkspacesClient) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error)`

- [ ] **Step 1: Write the failing test using a fake HTTP transport**

Rather than hitting the real Power BI API, inject a fake `http.RoundTripper` so the test is fast and deterministic.

```go
// pkg/powerbi/workspaces_test.go
package powerbi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type stubTransport struct {
	statusCode int
	body       string
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: s.statusCode,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

func TestListWorkspaces(t *testing.T) {
	responseBody := `{"value":[
		{"id":"ws-1","name":"Sales Analytics","type":"Workspace","isReadOnly":false,"isOnDedicatedCapacity":true,"capacityId":"cap-1"},
		{"id":"ws-2","name":"My workspace","type":"PersonalGroup","isReadOnly":false,"isOnDedicatedCapacity":false,"capacityId":""}
	]}`

	client := &Client{
		credential: &fakeCredential{},
		httpClient: &http.Client{Transport: &stubTransport{statusCode: 200, body: responseBody}},
	}

	wsClient := NewWorkspacesClient(client)
	workspaces, err := wsClient.ListWorkspaces(context.Background())
	if err != nil {
		t.Fatalf("ListWorkspaces() error = %v", err)
	}
	if len(workspaces) != 2 {
		t.Fatalf("len(workspaces) = %d, want 2", len(workspaces))
	}
	if workspaces[0].Name != "Sales Analytics" {
		t.Errorf("workspaces[0].Name = %q, want %q", workspaces[0].Name, "Sales Analytics")
	}
	if !workspaces[0].IsOnDedicatedCapacity {
		t.Errorf("workspaces[0].IsOnDedicatedCapacity = false, want true")
	}
}
```

This test needs a `fakeCredential` implementing `azcore.TokenCredential` — add it to `pkg/powerbi/client_test.go`:

```go
// add to pkg/powerbi/client_test.go
import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type fakeCredential struct{}

func (f *fakeCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "fake-token"}, nil
}
```

Explain: this is Go's version of mocking — `azcore.TokenCredential` is an interface (a set of method signatures), and any type that implements `GetToken` with that exact signature satisfies it automatically. There's no explicit "implements" keyword like in Java/C#; Go checks this structurally at compile time. This is what makes `doRequest`/`ListWorkspaces` testable without a real network call.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./pkg/powerbi/... -v -run TestListWorkspaces`
Expected: FAIL — `WorkspacesClient`/`NewWorkspacesClient` undefined

- [ ] **Step 3: Implement `pkg/powerbi/workspaces.go`**

```go
package powerbi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/polbarbero/lazypowerbi/pkg/domain"
)

// WorkspacesClient lists Power BI workspaces accessible to the authenticated user.
type WorkspacesClient struct {
	client *Client
}

// NewWorkspacesClient creates a new workspaces client.
func NewWorkspacesClient(client *Client) *WorkspacesClient {
	return &WorkspacesClient{client: client}
}

// workspaceDTO mirrors the JSON shape returned by GET /groups.
// Power BI's field names don't match domain.Workspace 1:1, so we decode into
// this intermediate type first, then map it explicitly.
type workspaceDTO struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	IsReadOnly            bool   `json:"isReadOnly"`
	IsOnDedicatedCapacity bool   `json:"isOnDedicatedCapacity"`
	CapacityID            string `json:"capacityId"`
}

type workspacesResponse struct {
	Value []workspaceDTO `json:"value"`
}

// ListWorkspaces retrieves all workspaces the authenticated user is a member of.
func (c *WorkspacesClient) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error) {
	body, err := c.client.doRequest(ctx, "GET", "/groups")
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	var resp workspacesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse workspaces response: %w", err)
	}

	workspaces := make([]*domain.Workspace, 0, len(resp.Value))
	for _, dto := range resp.Value {
		workspaces = append(workspaces, &domain.Workspace{
			ID:                    dto.ID,
			Name:                  dto.Name,
			Type:                  dto.Type,
			IsReadOnly:            dto.IsReadOnly,
			IsOnDedicatedCapacity: dto.IsOnDedicatedCapacity,
			CapacityID:            dto.CapacityID,
		})
	}

	return workspaces, nil
}
```

Explain: this DTO (data transfer object) → domain mapping step is the same role `deref()` plays in lazyazure's `subscriptions.go` — except there the Azure SDK already gives typed `*string` fields, so `deref` just unwraps pointers; here we're doing the JSON decoding ourselves first via `encoding/json`, the standard library's JSON package (struct tags like `` `json:"id"` `` tell it which JSON key maps to which field — similar to Pydantic field aliases in Python).

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/powerbi/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/powerbi
git commit -m "feat: add WorkspacesClient.ListWorkspaces"
```

---

## Task 4: `ItemsClient` — `pkg/powerbi/items.go`

**Files:**
- Create: `pkg/powerbi/items.go`
- Test: `pkg/powerbi/items_test.go`

**Interfaces:**
- Consumes: `(c *Client) doRequest` (Task 2), `domain.Item` (Task 1)
- Produces:
  - `powerbi.ItemsClient{}`
  - `func NewItemsClient(client *Client) *ItemsClient`
  - `func (c *ItemsClient) ListItemsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Item, error)`

- [ ] **Step 1: Write the failing test with a transport that routes by URL path**

Datasets/reports/dashboards/dataflows are 4 separate endpoints, so the stub transport needs to answer differently per path.

```go
// pkg/powerbi/items_test.go
package powerbi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type routingTransport struct {
	responses map[string]string // path suffix -> JSON body
}

func (r *routingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for suffix, body := range r.responses {
		if strings.HasSuffix(req.URL.Path, suffix) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"value":[]}`)),
		Header:     make(http.Header),
	}, nil
}

func TestListItemsByWorkspace(t *testing.T) {
	transport := &routingTransport{responses: map[string]string{
		"/datasets":   `{"value":[{"id":"ds-1","name":"Sales Dataset","webUrl":"https://x/ds-1"}]}`,
		"/reports":    `{"value":[{"id":"rp-1","name":"Sales Report","webUrl":"https://x/rp-1"}]}`,
		"/dashboards": `{"value":[{"id":"db-1","name":"Sales Dashboard","webUrl":"https://x/db-1"}]}`,
		"/dataflows":  `{"value":[{"objectId":"df-1","name":"Sales Dataflow"}]}`,
	}}

	client := &Client{
		credential: &fakeCredential{},
		httpClient: &http.Client{Transport: transport},
	}

	itemsClient := NewItemsClient(client)
	items, err := itemsClient.ListItemsByWorkspace(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("ListItemsByWorkspace() error = %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("len(items) = %d, want 4", len(items))
	}

	kinds := map[string]bool{}
	for _, item := range items {
		kinds[item.Kind] = true
		if item.WorkspaceID != "ws-1" {
			t.Errorf("item %q WorkspaceID = %q, want %q", item.Name, item.WorkspaceID, "ws-1")
		}
	}
	for _, want := range []string{"Dataset", "Report", "Dashboard", "Dataflow"} {
		if !kinds[want] {
			t.Errorf("missing item with Kind = %q", want)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./pkg/powerbi/... -v -run TestListItemsByWorkspace`
Expected: FAIL — `ItemsClient` undefined

- [ ] **Step 3: Implement `pkg/powerbi/items.go`**

```go
package powerbi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/polbarbero/lazypowerbi/pkg/domain"
)

// ItemsClient lists Power BI items (datasets, reports, dashboards, dataflows)
// within a workspace.
type ItemsClient struct {
	client *Client
}

// NewItemsClient creates a new items client.
func NewItemsClient(client *Client) *ItemsClient {
	return &ItemsClient{client: client}
}

// itemDTO mirrors the common shape of dataset/report/dashboard entries.
// Dataflows use "objectId" instead of "id", so they're decoded separately.
type itemDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	WebURL string `json:"webUrl"`
}

type itemsResponse struct {
	Value []itemDTO `json:"value"`
}

type dataflowDTO struct {
	ObjectID string `json:"objectId"`
	Name     string `json:"name"`
}

type dataflowsResponse struct {
	Value []dataflowDTO `json:"value"`
}

// ListItemsByWorkspace retrieves all datasets, reports, dashboards, and
// dataflows in the given workspace, merged into a single slice.
func (c *ItemsClient) ListItemsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Item, error) {
	var items []*domain.Item

	standardKinds := []struct {
		kind string
		path string
	}{
		{"Dataset", "/datasets"},
		{"Report", "/reports"},
		{"Dashboard", "/dashboards"},
	}

	for _, k := range standardKinds {
		body, err := c.client.doRequest(ctx, "GET", "/groups/"+workspaceID+k.path)
		if err != nil {
			return nil, fmt.Errorf("failed to list %ss for workspace %s: %w", k.kind, workspaceID, err)
		}

		var resp itemsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse %s response: %w", k.kind, err)
		}

		for _, dto := range resp.Value {
			items = append(items, &domain.Item{
				ID:          dto.ID,
				Name:        dto.Name,
				Kind:        k.kind,
				WorkspaceID: workspaceID,
				WebURL:      dto.WebURL,
			})
		}
	}

	dfBody, err := c.client.doRequest(ctx, "GET", "/groups/"+workspaceID+"/dataflows")
	if err != nil {
		return nil, fmt.Errorf("failed to list dataflows for workspace %s: %w", workspaceID, err)
	}

	var dfResp dataflowsResponse
	if err := json.Unmarshal(dfBody, &dfResp); err != nil {
		return nil, fmt.Errorf("failed to parse dataflows response: %w", err)
	}

	for _, dto := range dfResp.Value {
		items = append(items, &domain.Item{
			ID:          dto.ObjectID,
			Name:        dto.Name,
			Kind:        "Dataflow",
			WorkspaceID: workspaceID,
		})
	}

	return items, nil
}
```

Explain: the `standardKinds` slice of anonymous structs is a way to avoid repeating the same fetch-and-map block 3 times — a small loop-driven DRY technique. Note dataflows need their own struct because the API uses a different field name (`objectId` vs `id`) — a good example of why the DTO/intermediate-type pattern matters: each endpoint's quirks stay isolated from `domain.Item`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/powerbi/... -v`
Expected: PASS (all tests in the package)

- [ ] **Step 5: Commit**

```bash
git add pkg/powerbi
git commit -m "feat: add ItemsClient.ListItemsByWorkspace"
```

---

## Task 5: `pkg/gui/interfaces.go` and `factory.go`

**Files:**
- Create: `pkg/gui/interfaces.go`
- Create: `pkg/powerbi/factory.go`

**Interfaces:**
- Consumes: `domain.Workspace`, `domain.Item`, `domain.User` (Task 1); `*powerbi.Client`, `*powerbi.WorkspacesClient`, `*powerbi.ItemsClient` (Tasks 2–4)
- Produces:
  - `gui.WorkspacesClient` interface: `ListWorkspaces(ctx) ([]*domain.Workspace, error)`
  - `gui.ItemsClient` interface: `ListItemsByWorkspace(ctx, workspaceID string) ([]*domain.Item, error)`
  - `gui.PowerBIClient` interface: `GetUserInfo(ctx) (*domain.User, error)`, `VerifyAuthentication(ctx) error`, `Credential() azcore.TokenCredential`
  - `gui.PowerBIClientFactory` interface: `NewWorkspacesClient() (WorkspacesClient, error)`, `NewItemsClient() (ItemsClient, error)`
  - `powerbi.ClientFactory` struct implementing `gui.PowerBIClientFactory`

This task has no separate unit test: it's pure interface declarations plus thin factory wiring, verified by the build succeeding and by Task 9 (GUI wiring) compiling against it.

- [ ] **Step 1: Write `pkg/gui/interfaces.go`**

```go
package gui

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/polbarbero/lazypowerbi/pkg/domain"
)

// WorkspacesClient provides workspace operations.
type WorkspacesClient interface {
	ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error)
}

// ItemsClient provides item operations within a workspace.
type ItemsClient interface {
	ListItemsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Item, error)
}

// PowerBIClient combines authentication and identity operations.
type PowerBIClient interface {
	GetUserInfo(ctx context.Context) (*domain.User, error)
	VerifyAuthentication(ctx context.Context) error
	Credential() azcore.TokenCredential
}

// PowerBIClientFactory creates resource-specific clients.
type PowerBIClientFactory interface {
	NewWorkspacesClient() (WorkspacesClient, error)
	NewItemsClient() (ItemsClient, error)
}
```

Explain: this is the same dependency-inversion pattern as lazyazure's `gui.AzureClientFactory` — the `gui` package only knows about these interfaces, never about `powerbi.Client` directly. That's what lets `pkg/demo` provide a fake implementation later (Task 8) without `gui` caring which one it's talking to. In Go, you don't declare "this struct implements this interface" anywhere — it's implicit, checked at the call site where a concrete type is assigned to an interface-typed variable.

- [ ] **Step 2: Write `pkg/powerbi/factory.go`**

```go
package powerbi

import (
	"github.com/polbarbero/lazypowerbi/pkg/gui"
)

// ClientFactory implements gui.PowerBIClientFactory for real Power BI clients.
type ClientFactory struct {
	client *Client
}

// NewClientFactory creates a new client factory wrapping the given client.
func NewClientFactory(client *Client) *ClientFactory {
	return &ClientFactory{client: client}
}

// NewWorkspacesClient creates a workspaces client.
func (f *ClientFactory) NewWorkspacesClient() (gui.WorkspacesClient, error) {
	return NewWorkspacesClient(f.client), nil
}

// NewItemsClient creates an items client.
func (f *ClientFactory) NewItemsClient() (gui.ItemsClient, error) {
	return NewItemsClient(f.client), nil
}
```

- [ ] **Step 3: Verify everything still builds**

Run: `go build ./... && go vet ./...`
Expected: no errors (this task introduces a `pkg/powerbi` → `pkg/gui` import; `pkg/gui` must NOT import `pkg/powerbi` back, or you'd get an import cycle — explain why Go forbids cycles and how interfaces are exactly the tool used here to avoid one)

- [ ] **Step 4: Commit**

```bash
git add pkg/gui/interfaces.go pkg/powerbi/factory.go
git commit -m "feat: add gui interfaces and powerbi.ClientFactory"
```

---

## Task 6: `pkg/utils` — copy and adapt

**Files:**
- Copy verbatim: `pkg/utils/clipboard.go`, `pkg/utils/browser.go`, `pkg/utils/metrics.go`, `pkg/utils/logger.go`, plus their `_test.go` files
- Create (adapted): `pkg/utils/portal_urls.go`, `pkg/utils/portal_urls_test.go`

**Interfaces:**
- Produces: `utils.Log(format string, args ...any)`, `utils.IsDebugEnabled() bool`, `utils.InitLogger() error`, `utils.CloseLogger()`, `utils.CopyToClipboard(text string) error`, `utils.OpenBrowser(url string) error`, `utils.StartAPITimer(name string) func(error)`, `utils.LogMetrics()`, `utils.BuildWorkspacePortalURL(workspaceID string) string`, `utils.BuildItemPortalURL(workspaceID, kind, itemID string) string`

- [ ] **Step 1: Copy the domain-agnostic utils files verbatim**

```bash
cp /c/Repos/lazyazure/pkg/utils/clipboard.go /c/Repos/lazypowerbi/pkg/utils/clipboard.go
cp /c/Repos/lazyazure/pkg/utils/clipboard_test.go /c/Repos/lazypowerbi/pkg/utils/clipboard_test.go
cp /c/Repos/lazyazure/pkg/utils/browser.go /c/Repos/lazypowerbi/pkg/utils/browser.go
cp /c/Repos/lazyazure/pkg/utils/browser_test.go /c/Repos/lazypowerbi/pkg/utils/browser_test.go
cp /c/Repos/lazyazure/pkg/utils/metrics.go /c/Repos/lazypowerbi/pkg/utils/metrics.go
cp /c/Repos/lazyazure/pkg/utils/metrics_test.go /c/Repos/lazypowerbi/pkg/utils/metrics_test.go
cp /c/Repos/lazyazure/pkg/utils/logger.go /c/Repos/lazypowerbi/pkg/utils/logger.go
cp /c/Repos/lazyazure/pkg/utils/logger_test.go /c/Repos/lazypowerbi/pkg/utils/logger_test.go
```

- [ ] **Step 2: Rename the env var and log path inside `logger.go`**

Open `pkg/utils/logger.go`. It contains references to `LAZYAZURE_DEBUG` and a `.lazyazure` directory name (read it together first to find the exact lines). Replace:
- `LAZYAZURE_DEBUG` → `LAZYPOWERBI_DEBUG`
- `.lazyazure` → `.lazypowerbi`

Do the same check/replace in `logger_test.go` if the env var name appears there.

- [ ] **Step 3: Write the failing test for portal URLs**

```go
// pkg/utils/portal_urls_test.go
package utils

import "testing"

func TestBuildWorkspacePortalURL(t *testing.T) {
	got := BuildWorkspacePortalURL("ws-1")
	want := "https://app.powerbi.com/groups/ws-1/list"
	if got != want {
		t.Errorf("BuildWorkspacePortalURL() = %q, want %q", got, want)
	}
}

func TestBuildItemPortalURL(t *testing.T) {
	cases := []struct {
		kind string
		want string
	}{
		{"Report", "https://app.powerbi.com/groups/ws-1/reports/item-1"},
		{"Dataset", "https://app.powerbi.com/groups/ws-1/datasets/item-1"},
		{"Dashboard", "https://app.powerbi.com/groups/ws-1/dashboards/item-1"},
		{"Dataflow", "https://app.powerbi.com/groups/ws-1/dataflows/item-1"},
	}
	for _, c := range cases {
		got := BuildItemPortalURL("ws-1", c.kind, "item-1")
		if got != c.want {
			t.Errorf("BuildItemPortalURL(%q) = %q, want %q", c.kind, got, c.want)
		}
	}
}
```

- [ ] **Step 4: Run the test to verify it fails**

Run: `go test ./pkg/utils/... -v -run TestBuild`
Expected: FAIL — `BuildWorkspacePortalURL` undefined

- [ ] **Step 5: Implement `pkg/utils/portal_urls.go`**

```go
package utils

import (
	"fmt"
	"strings"
)

// BuildWorkspacePortalURL builds the Power BI portal URL for a workspace.
func BuildWorkspacePortalURL(workspaceID string) string {
	return fmt.Sprintf("https://app.powerbi.com/groups/%s/list", workspaceID)
}

// BuildItemPortalURL builds the Power BI portal URL for an item (dataset,
// report, dashboard, or dataflow) within a workspace.
func BuildItemPortalURL(workspaceID, kind, itemID string) string {
	segment := strings.ToLower(kind) + "s" // "Report" -> "reports", "Dataset" -> "datasets", etc.
	return fmt.Sprintf("https://app.powerbi.com/groups/%s/%s/%s", workspaceID, segment, itemID)
}
```

- [ ] **Step 6: Run all utils tests to verify they pass**

Run: `go test ./pkg/utils/... -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add pkg/utils
git commit -m "feat: add pkg/utils (copied from lazyazure, portal URLs adapted for Power BI)"
```

---

## Task 7: `pkg/tasks` and `pkg/gui/panels` — copy verbatim

**Files:**
- Copy verbatim: `pkg/tasks/tasks.go`, `pkg/tasks/tasks_test.go`
- Copy verbatim: `pkg/gui/panels/filtered_list.go`, `pkg/gui/panels/filtered_list_test.go`, `pkg/gui/panels/search_bar.go`, `pkg/gui/panels/search_bar_test.go`, `pkg/gui/panels/main_panel_search.go`, `pkg/gui/panels/main_panel_search_test.go`

**Interfaces:**
- Produces: `tasks.TaskManager` (used by Task 9's `gui.go`); `panels.FilteredList[T any]`, `panels.SearchBar`, `panels.MainPanelSearch` (used by Task 9's `gui.go`)

These packages are domain-agnostic (they operate on `T any` via generics, or on raw text), so they are copied with zero changes — verified purely by their existing tests passing unmodified.

- [ ] **Step 1: Copy the files**

```bash
mkdir -p /c/Repos/lazypowerbi/pkg/tasks /c/Repos/lazypowerbi/pkg/gui/panels
cp /c/Repos/lazyazure/pkg/tasks/*.go /c/Repos/lazypowerbi/pkg/tasks/
cp /c/Repos/lazyazure/pkg/gui/panels/*.go /c/Repos/lazypowerbi/pkg/gui/panels/
```

- [ ] **Step 2: Fix the import paths**

These files import `github.com/matsest/lazyazure/...` somewhere (check with a search) — replace with `github.com/polbarbero/lazypowerbi/...` everywhere it appears. Confirm there are no such references in `panels/*.go` (they likely have none, being fully generic) — if `tasks.go` has none either, this step is a no-op; verify by searching rather than assuming.

- [ ] **Step 3: Run the tests**

Run: `go test ./pkg/tasks/... ./pkg/gui/panels/... -v`
Expected: PASS, unchanged from lazyazure's own test results

Explain while looking at `filtered_list.go`: `FilteredList[T any]` is a *generic* type — `T` is a type parameter, filled in at the call site (e.g. `FilteredList[*domain.Workspace]`). This is Go's answer to "write once, use with any type safely," roughly analogous to Python's `TypeVar`/`Generic`, except Go checks it at compile time with zero runtime cost.

- [ ] **Step 4: Commit**

```bash
git add pkg/tasks pkg/gui/panels
git commit -m "feat: copy pkg/tasks and pkg/gui/panels (domain-agnostic, unchanged)"
```

---

## Task 8: `pkg/demo` — mock data and demo client

**Files:**
- Create: `pkg/demo/data.go`
- Create: `pkg/demo/client.go`
- Test: `pkg/demo/client_test.go`

**Interfaces:**
- Consumes: `domain.Workspace`, `domain.Item`, `domain.User` (Task 1); `gui.WorkspacesClient`, `gui.ItemsClient`, `gui.PowerBIClient`, `gui.PowerBIClientFactory` (Task 5)
- Produces:
  - `demo.DemoData{User, Workspaces, Items}`
  - `func NewDemoData() *DemoData`
  - `demo.Client{}` implementing `gui.PowerBIClient` + `gui.PowerBIClientFactory`
  - `func NewClientWithMode(mode string) *Client`

- [ ] **Step 1: Write the failing test**

```go
// pkg/demo/client_test.go
package demo

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func TestNewClientWithModeImplementsInterfaces(t *testing.T) {
	client := NewClientWithMode("1")

	user, err := client.GetUserInfo(context.Background())
	if err != nil {
		t.Fatalf("GetUserInfo() error = %v", err)
	}
	if user.DisplayName == "" {
		t.Error("expected a non-empty demo user DisplayName")
	}

	wsClient, err := client.NewWorkspacesClient()
	if err != nil {
		t.Fatalf("NewWorkspacesClient() error = %v", err)
	}
	workspaces, err := wsClient.ListWorkspaces(context.Background())
	if err != nil {
		t.Fatalf("ListWorkspaces() error = %v", err)
	}
	if len(workspaces) == 0 {
		t.Fatal("expected at least one demo workspace")
	}

	itemsClient, err := client.NewItemsClient()
	if err != nil {
		t.Fatalf("NewItemsClient() error = %v", err)
	}
	items, err := itemsClient.ListItemsByWorkspace(context.Background(), workspaces[0].ID)
	if err != nil {
		t.Fatalf("ListItemsByWorkspace() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one demo item for the first workspace")
	}

	var _ azcore.TokenCredential = client.Credential()
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./pkg/demo/... -v`
Expected: FAIL — package doesn't exist yet

- [ ] **Step 3: Implement `pkg/demo/data.go`**

```go
package demo

import "github.com/polbarbero/lazypowerbi/pkg/domain"

// DemoData holds all mock data for the demo mode.
type DemoData struct {
	User       *domain.User
	Workspaces []*domain.Workspace
	Items      map[string][]*domain.Item // key: workspace ID
}

// NewDemoData creates a complete set of demo data.
func NewDemoData() *DemoData {
	data := &DemoData{
		User: &domain.User{
			DisplayName:       "Demo User",
			UserPrincipalName: "demo.user@example.com",
			Type:              "user",
			TenantID:          "00000000-0000-0000-0000-000000000000",
		},
		Workspaces: createDemoWorkspaces(),
		Items:      make(map[string][]*domain.Item),
	}

	for _, ws := range data.Workspaces {
		data.Items[ws.ID] = createDemoItems(ws.ID)
	}

	return data
}

func createDemoWorkspaces() []*domain.Workspace {
	return []*domain.Workspace{
		{
			ID:                    "00000000-0000-0000-0000-000000000001",
			Name:                  "Sales Analytics",
			Type:                  "Workspace",
			IsOnDedicatedCapacity: true,
			CapacityID:            "00000000-0000-0000-0000-000000000099",
		},
		{
			ID:   "00000000-0000-0000-0000-000000000002",
			Name: "My workspace",
			Type: "PersonalGroup",
		},
	}
}

func createDemoItems(workspaceID string) []*domain.Item {
	return []*domain.Item{
		{ID: workspaceID + "-ds-1", Name: "Sales Dataset", Kind: "Dataset", WorkspaceID: workspaceID, WebURL: "https://app.powerbi.com/groups/" + workspaceID + "/datasets/" + workspaceID + "-ds-1"},
		{ID: workspaceID + "-rp-1", Name: "Quarterly Report", Kind: "Report", WorkspaceID: workspaceID, WebURL: "https://app.powerbi.com/groups/" + workspaceID + "/reports/" + workspaceID + "-rp-1"},
		{ID: workspaceID + "-db-1", Name: "Executive Dashboard", Kind: "Dashboard", WorkspaceID: workspaceID, WebURL: "https://app.powerbi.com/groups/" + workspaceID + "/dashboards/" + workspaceID + "-db-1"},
		{ID: workspaceID + "-df-1", Name: "Raw Sales Dataflow", Kind: "Dataflow", WorkspaceID: workspaceID},
	}
}
```

- [ ] **Step 4: Implement `pkg/demo/client.go`**

```go
package demo

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/polbarbero/lazypowerbi/pkg/domain"
	"github.com/polbarbero/lazypowerbi/pkg/gui"
)

// Client is a fake PowerBIClient + PowerBIClientFactory backed by in-memory demo data.
type Client struct {
	data *DemoData
}

// NewClientWithMode creates a demo client. mode is currently unused beyond
// validating it's "1" or "2" (small/large dataset); both produce the same
// fixed dataset today — kept as a parameter for parity with lazyazure's
// LAZYAZURE_DEMO and to leave room for a larger dataset later.
func NewClientWithMode(mode string) *Client {
	return &Client{data: NewDemoData()}
}

// GetUserInfo returns the fixed demo user.
func (c *Client) GetUserInfo(ctx context.Context) (*domain.User, error) {
	return c.data.User, nil
}

// VerifyAuthentication always succeeds in demo mode.
func (c *Client) VerifyAuthentication(ctx context.Context) error {
	return nil
}

// Credential returns a no-op credential; demo mode never makes real HTTP calls.
func (c *Client) Credential() azcore.TokenCredential {
	return &noopCredential{}
}

// NewWorkspacesClient returns a workspaces client backed by demo data.
func (c *Client) NewWorkspacesClient() (gui.WorkspacesClient, error) {
	return &demoWorkspacesClient{data: c.data}, nil
}

// NewItemsClient returns an items client backed by demo data.
func (c *Client) NewItemsClient() (gui.ItemsClient, error) {
	return &demoItemsClient{data: c.data}, nil
}

type demoWorkspacesClient struct {
	data *DemoData
}

func (d *demoWorkspacesClient) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error) {
	return d.data.Workspaces, nil
}

type demoItemsClient struct {
	data *DemoData
}

func (d *demoItemsClient) ListItemsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Item, error) {
	return d.data.Items[workspaceID], nil
}

// noopCredential is a stub azcore.TokenCredential for demo mode.
type noopCredential struct{}

func (n *noopCredential) GetToken(ctx context.Context, opts interface {
}) (azcore.AccessToken, error) {
	return azcore.AccessToken{}, nil
}
```

When writing this together, point out the `noopCredential.GetToken` signature is wrong as drafted (the real interface needs `policy.TokenRequestOptions`, not `interface{}`) — fix it before running:

```go
import "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

func (n *noopCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{}, nil
}
```

Explain: this is a good moment to show how the Go compiler catches an interface mismatch — assigning `&noopCredential{}` to a variable typed `azcore.TokenCredential` (as the test's `var _ azcore.TokenCredential = client.Credential()` line does) fails to compile if the method signature doesn't match exactly, with a clear compiler error naming the missing/mismatched method.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./pkg/demo/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/demo
git commit -m "feat: add pkg/demo with mock workspaces/items client"
```

---

## Task 9: `pkg/gui` — GUI controller (adapted from lazyazure's `gui.go`/`cache.go`)

**Files:**
- Create: `pkg/gui/gui.go`
- Create: `pkg/gui/cache.go`
- Test: `pkg/gui/gui_test.go`, `pkg/gui/cache_test.go`

**Interfaces:**
- Consumes: `gui.PowerBIClient`, `gui.PowerBIClientFactory`, `gui.WorkspacesClient`, `gui.ItemsClient` (Task 5); `domain.Workspace`, `domain.Item`, `domain.User` (Task 1); `panels.FilteredList[T]`, `panels.SearchBar`, `panels.MainPanelSearch` (Task 7); `tasks.TaskManager` (Task 7); `utils.Log`, `utils.BuildWorkspacePortalURL`, `utils.BuildItemPortalURL`, `utils.CopyToClipboard`, `utils.OpenBrowser` (Task 6)
- Produces:
  - `gui.VersionInfo{Version, Commit, Date string}`
  - `func NewGui(client PowerBIClient, factory PowerBIClientFactory, versionInfo VersionInfo) (*Gui, error)`
  - `func (g *Gui) Run() error`

This is the largest task, because lazyazure's `gui.go` is ~4000 lines built for a 3-tier sidebar. The approach: copy the file, then perform exact, listed transformations rather than rewriting from scratch — most of the file's logic (search, scrolling, tabs, mouse handling) is unchanged.

- [ ] **Step 1: Copy the source files as a starting point**

```bash
cp /c/Repos/lazyazure/pkg/gui/gui.go /c/Repos/lazypowerbi/pkg/gui/gui.go
cp /c/Repos/lazyazure/pkg/gui/cache.go /c/Repos/lazypowerbi/pkg/gui/cache.go
cp /c/Repos/lazyazure/pkg/gui/gui_test.go /c/Repos/lazypowerbi/pkg/gui/gui_test.go
cp /c/Repos/lazyazure/pkg/gui/cache_test.go /c/Repos/lazypowerbi/pkg/gui/cache_test.go
```

- [ ] **Step 2: Global identifier renames across all 4 files**

Apply these exact renames (case-sensitive, whole-identifier) throughout `gui.go`, `cache.go`, and both test files:

| From | To |
|---|---|
| `AzureClient` | `PowerBIClient` |
| `AzureClientFactory` | `PowerBIClientFactory` |
| `azureClient` | `powerBIClient` |
| `Subscription` | `Workspace` |
| `subscriptions` | `workspaces` |
| `subscriptionsView` | `workspacesView` |
| `subClient` | `wsClient` |
| `subList` | `workspaceList` |
| `selectedSub` | `selectedWorkspace` |
| `loadingSubs` | `loadingWorkspaces` |
| `Resource` (as in `domain.Resource`, `resources` views, `resList`) | `Item` |
| `resourcesView` | `itemsView` |
| `resClient` | `itemsClient` |
| `resList` | `itemList` |
| `selectedRes` | `selectedItem` |
| `loadingRes` | `loadingItems` |
| `github.com/matsest/lazyazure/...` (any import) | `github.com/polbarbero/lazypowerbi/...` |

Do this with your editor's find-and-replace, one identifier at a time, and rebuild after each batch of related renames to catch typos early (`go build ./pkg/gui/...`). Explain why doing this incrementally (rather than all at once blind) matters: Go's compiler will immediately flag any leftover unrenamed reference as "undefined," which is a fast feedback loop you should lean on.

- [ ] **Step 3: Delete all resource-group-tier code**

Search `gui.go` for every identifier containing `RG` or `ResourceGroup` (e.g. `resourceGroupsView`, `rgClient`, `rgList`, `selectedRG`, `loadingRGs`, `resourceGroups`, `onRGEnter`, `nextRG`, `prevRG`, `updateRGSelection`, `pageDownRG`, `pageUpRG`, `refreshResourceGroupsPanel`, `sortResourceGroups`). Delete:
- The corresponding fields from the `Gui` struct
- The corresponding view setup block in `setupViews` (the "3. Resource Groups panel" block)
- The corresponding keybinding registrations in `setupKeybindings` (the "Resource Groups panel navigation" block)
- The corresponding handler functions in full (`onRGEnter`, `nextRG`, `prevRG`, `updateRGSelection`, `pageDownRG`, `pageUpRG`, `refreshResourceGroupsPanel`, `sortResourceGroups`, and any `RG`-named helper not listed)
- Any reference to these from `switchPanel`/`switchPanelReverse` (the panel cycle becomes `workspaces → items → main → workspaces`, 3 stops instead of 4)
- Any reference from `onPanelClick`'s `viewName` switch
- Any reference from `gui.refreshResourceGroupsPanel()` calls inside `setupViews`/elsewhere

Rename `onWorkspaceEnter` (was `onSubEnter`) so that, instead of loading resource groups, it loads items directly: it should call `gui.loadItems(workspaceID)` (the renamed `loadResourceGroups`/`loadResources` merge into a single `loadItems`, fed by the new `ItemsClient`) and switch focus straight to `"items"`.

- [ ] **Step 4: Adjust panel layout math in `setupViews`**

The original divides remaining vertical space into 20% (subscriptions) / 30% (resource groups) / 50% (resources). With only 2 list panels, replace with a 2-way split — e.g. 35% workspaces / 65% items:

```go
authHeight := AuthViewHeight
remainingHeight := maxY - authHeight - 2 // -2 for status bar
workspaceHeight := (remainingHeight * 35) / 100
```

Remove the `rgHeight` calculation entirely, and adjust the Y-coordinate chain so the `items` panel's `Y0` starts right after `workspaces`' `Y1` (previously `resources`' `Y0` started after `resourcegroups`' `Y1`).

- [ ] **Step 5: Adjust `cache.go` to a single cache tier**

The original `PreloadCache` has two tiers: an RG cache (`GetRGs`/key: subscription ID) and a resource cache (`GetRes`/key: subscription+RG). Collapse to one tier: an items cache keyed by workspace ID only.

- Delete: `rgCacheTTL`, `baseRGCache`, `smallRGCache`, `mediumRGCache`, `largeRGCache`, the `RGCacheSize` field on `CacheConfig`, and any method/field referencing RGs (`GetRGs`, `SetRGs`, the RG half of the cache struct)
- Keep one tier, renamed: `itemCacheTTL` (use the original `resCacheTTL`'s value, 10 minutes), `baseItemCache` (from `baseResCache`, 500), with `GetItems(workspaceID string) ([]*domain.Item, bool)` / `SetItems(workspaceID string, items []*domain.Item)` replacing `GetRes`/`SetRes` (which were keyed by subscription+RG — now keyed by workspace ID alone, since there's one less hierarchy level)

- [ ] **Step 6: Build and fix remaining compile errors**

Run: `go build ./pkg/gui/...`

Work through any remaining errors one at a time — at this file size some will surface only after earlier ones are fixed. Common ones to expect: leftover `domain.Resource`/`domain.Subscription` references missed by the rename pass, and `FilteredList[*domain.ResourceGroup]` instantiations that need deleting rather than renaming.

- [ ] **Step 7: Adjust the test files to match**

Run: `go test ./pkg/gui/... -v`

Fix any test referencing deleted RG concepts (delete those test cases/assertions) or renamed identifiers (apply the same rename table from Step 2). Re-run until green.

- [ ] **Step 8: Manual smoke test in demo mode**

This package has heavy terminal-UI logic that's impractical to unit-test exhaustively (mouse coordinates, gocui rendering) — verify it by actually running it, same as lazyazure's own testing approach relies on `LAZYAZURE_DEMO` for this. Skip running it for real here since Task 11 (main.go) is what wires `LAZYPOWERBI_DEMO` — note this step as deferred to Task 11's Step 4 instead of trying to run an incomplete binary now.

- [ ] **Step 9: Commit**

```bash
git add pkg/gui
git commit -m "feat: adapt gui.go and cache.go to 2-tier Workspaces/Items sidebar"
```

---

## Task 10: `main.go` — wire everything together

**Files:**
- Modify: `main.go` (replace the bootstrap from Task 0)
- Modify: `main_test.go` (copy and adapt from lazyazure's `main_test.go`)

**Interfaces:**
- Consumes: `powerbi.NewClient()`, `powerbi.NewClientFactory()` (Task 2/5); `demo.NewClientWithMode()` (Task 8); `gui.NewGui()`, `gui.VersionInfo` (Task 9); `utils.IsDebugEnabled()`, `utils.InitLogger()`, `utils.CloseLogger()`, `utils.Log()` (Task 6)

- [ ] **Step 1: Copy `main.go` and `main_test.go` from lazyazure as a starting point**

```bash
cp /c/Repos/lazyazure/main.go /c/Repos/lazypowerbi/main.go
cp /c/Repos/lazyazure/main_test.go /c/Repos/lazypowerbi/main_test.go
```

- [ ] **Step 2: Apply renames**

| From | To |
|---|---|
| `lazyazure` (binary name in help text, repo URL `github.com/matsest/lazyazure`) | `lazypowerbi` (and `github.com/polbarbero/lazypowerbi`) |
| `LAZYAZURE_DEBUG` | `LAZYPOWERBI_DEBUG` |
| `LAZYAZURE_DEMO` | `LAZYPOWERBI_DEMO` |
| `LAZYAZURE_CACHE_SIZE` | `LAZYPOWERBI_CACHE_SIZE` |
| `azure.NewClient()` | `powerbi.NewClient()` |
| `azure.NewClientFactory(client)` | `powerbi.NewClientFactory(client)` |
| import `"github.com/matsest/lazyazure/pkg/azure"` | `"github.com/polbarbero/lazypowerbi/pkg/powerbi"` |
| import `"github.com/matsest/lazyazure/pkg/demo"` / `/gui` / `/utils` | `"github.com/polbarbero/lazypowerbi/pkg/demo"` / `/gui` / `/utils` |
| `"Failed to ensure you're logged in with 'az login'"` message and any other Azure-specific copy | adapt wording to mention Power BI / Entra ID sign-in instead of "Azure Portal" |

Update `checkUpdate`'s default API URL from `https://api.github.com/repos/matsest/lazyazure/releases/latest` to `https://api.github.com/repos/polbarbero/lazypowerbi/releases/latest`.

- [ ] **Step 3: Run the tests**

Run: `go test . -v`
Expected: PASS — fix any leftover reference the table above missed (same fast-feedback approach as Task 9)

- [ ] **Step 4: Manual smoke test — demo mode**

```bash
go build -o lazypowerbi.exe .
LAZYPOWERBI_DEMO=1 ./lazypowerbi.exe
```

Verify interactively: the auth panel shows "Demo User"; the `workspaces` panel lists "Sales Analytics" and "My workspace"; pressing Enter on a workspace loads its 4 items (Dataset/Report/Dashboard/Dataflow) directly — no intermediate resource-group panel; Tab cycles `workspaces → items → main → workspaces`; `/` filters the active list panel; `o`/`c` open/copy the Power BI portal URL for the selected item; `q` quits cleanly.

- [ ] **Step 5: Commit**

```bash
git add main.go main_test.go go.mod go.sum
git commit -m "feat: wire main.go for lazypowerbi (demo mode smoke-tested)"
```

---

## Task 11: Project scaffolding — README, Makefile, CI

**Files:**
- Create: `README.md` (lazypowerbi-specific, not copied verbatim)
- Create: `Makefile` (adapted from lazyazure's)
- Create: `.github/workflows/` (adapted: build/test/release workflows, binary name and repo path updated)
- Create: `.goreleaser.yml` (adapted)

**Interfaces:** none — this task doesn't add Go code, only project tooling/docs.

- [ ] **Step 1: Copy and adapt the `Makefile`**

```bash
cp /c/Repos/lazyazure/Makefile /c/Repos/lazypowerbi/Makefile
```
Replace every occurrence of `lazyazure` with `lazypowerbi` (binary name, ldflags version path `github.com/polbarbero/lazypowerbi` instead of `github.com/matsest/lazyazure`).

- [ ] **Step 2: Run `make build` and `make test` (or equivalents) to confirm the Makefile works**

Run: `make build && make test`
Expected: builds `lazypowerbi.exe`/`lazypowerbi`, runs the full test suite successfully

- [ ] **Step 3: Copy and adapt `.goreleaser.yml` and `.github/workflows/`**

```bash
cp /c/Repos/lazyazure/.goreleaser.yml /c/Repos/lazypowerbi/.goreleaser.yml
cp -r /c/Repos/lazyazure/.github /c/Repos/lazypowerbi/.github
```
Replace `lazyazure`/`matsest` with `lazypowerbi`/`polbarbero` throughout both. Don't try to actually run a release — just get the config internally consistent; CI execution is validated once pushed to GitHub, which is outside this plan's scope.

- [ ] **Step 4: Write a minimal `README.md`**

```markdown
# lazypowerbi

A TUI application for exploring Power BI workspaces, inspired by [lazyazure](https://github.com/matsest/lazyazure) and [lazydocker](https://github.com/jesseduffield/lazydocker).

## Usage

```bash
go build -o lazypowerbi .
./lazypowerbi
```

Authenticate first via `az login` (or any method supported by Azure's `DefaultAzureCredential`).

### Environment Variables

- `LAZYPOWERBI_DEBUG=1` — enable debug logging (`~/.lazypowerbi/debug.log`)
- `LAZYPOWERBI_DEMO=1` / `=2` — run with mock data (small / large dataset)
```

- [ ] **Step 5: Commit**

```bash
git add README.md Makefile .goreleaser.yml .github
git commit -m "chore: add README, Makefile, and CI scaffolding"
```

---

## Self-Review Notes

- **Spec coverage:** Sections 2–7 of the design doc map to Tasks 1 (domain), 2–4 (powerbi client), 5 (interfaces/factory), 9 (GUI), 8 (demo), 6 (utils/portal URLs), 0/11 (folder structure). Section 8 (auth decision: user-level API only) is enforced by Task 2/3/4 only ever calling `/v1.0/myorg` endpoints — no admin path exists anywhere in this plan. Section 9 (teaching approach) is reflected in every task's explanatory asides and in the Task 0 note recommending inline execution.
- **Type consistency check:** `gui.WorkspacesClient`/`gui.ItemsClient` (Task 5) match the method names `ListWorkspaces`/`ListItemsByWorkspace` defined in Tasks 3–4 exactly. `gui.PowerBIClientFactory`'s `NewWorkspacesClient()/NewItemsClient()` match both `powerbi.ClientFactory` (Task 5) and `demo.Client` (Task 8). `domain.Item.Kind` (not `Type`) is used consistently across Tasks 1, 4, 8, and the Task 9 rename table.
- **No placeholders:** Task 9's renames/deletions are listed as exact identifier tables/lists rather than "adapt as needed," because the source file (lazyazure's `gui.go`) is too large to fully reproduce in this plan — this is a deliberate exception to inlining full code, scoped narrowly to mechanical copy+rename+delete work on an existing, already-reviewed file.
