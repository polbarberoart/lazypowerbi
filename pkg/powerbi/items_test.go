package powerbi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// routedTransport returns different responses depending on the request URL path.
type routedTransport struct {
	routes map[string]string // path suffix → JSON body
}

func (r *routedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for suffix, body := range r.routes {
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

func TestListItems_Reports(t *testing.T) {
	transport := &routedTransport{
		routes: map[string]string{
			"/reports": `{"value":[
				{"id":"r-1","name":"Sales Report","webUrl":"https://app.powerbi.com/r-1"},
				{"id":"r-2","name":"Finance Report","webUrl":"https://app.powerbi.com/r-2"}
			]}`,
		},
	}

	client := &Client{
		credential: &fakeCredential{},
		httpClient: &http.Client{Transport: transport},
	}

	itemsClient := NewItemsClient(client)
	items, err := itemsClient.ListItems(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("ListItems() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Name != "Sales Report" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "Sales Report")
	}
	if items[0].Kind != "Report" {
		t.Errorf("items[0].Kind = %q, want %q", items[0].Kind, "Report")
	}
	if items[0].WorkspaceID != "ws-1" {
		t.Errorf("items[0].WorkspaceID = %q, want %q", items[0].WorkspaceID, "ws-1")
	}
}
