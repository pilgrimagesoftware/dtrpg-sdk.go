package openapi

import "testing"

func TestDefaultServerURL(t *testing.T) {
	if DefaultServerURL != "https://api.drivethrurpg.com/api" {
		t.Fatalf("unexpected DefaultServerURL: %s", DefaultServerURL)
	}
}

func TestOperationsContainsAuthKey(t *testing.T) {
	for _, op := range Operations {
		if op.Method == "POST" && op.Path == "/{DTRPG_API_VERSION}/auth_key" {
			return
		}
	}
	t.Fatalf("expected Operations to contain POST /{DTRPG_API_VERSION}/auth_key, got %+v", Operations)
}
