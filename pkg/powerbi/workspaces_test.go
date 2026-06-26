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
