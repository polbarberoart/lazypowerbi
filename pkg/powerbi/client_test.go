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
