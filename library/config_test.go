package library

import "testing"

func TestNewConfigUsesDefaults(t *testing.T) {
	c := NewConfig("my-app-key")
	if c.ApplicationKey() != "my-app-key" {
		t.Errorf("ApplicationKey() = %q, want %q", c.ApplicationKey(), "my-app-key")
	}
	if c.BaseURL() != DefaultBaseURL {
		t.Errorf("BaseURL() = %q, want %q", c.BaseURL(), DefaultBaseURL)
	}
	if c.APIVersion() != DefaultAPIVersion {
		t.Errorf("APIVersion() = %q, want %q", c.APIVersion(), DefaultAPIVersion)
	}
}

func TestNewConfigWithBaseURLOverridesBaseURL(t *testing.T) {
	c := NewConfigWithBaseURL("my-app-key", "http://localhost:8080/api")
	if c.BaseURL() != "http://localhost:8080/api" {
		t.Errorf("BaseURL() = %q, want %q", c.BaseURL(), "http://localhost:8080/api")
	}
	if c.APIVersion() != DefaultAPIVersion {
		t.Errorf("APIVersion() = %q, want %q", c.APIVersion(), DefaultAPIVersion)
	}
}

func TestConfigEquality(t *testing.T) {
	a := NewConfig("key")
	b := NewConfig("key")
	if a != b {
		t.Errorf("expected equal configs, got %+v != %+v", a, b)
	}

	c := NewConfig("other-key")
	if a == c {
		t.Errorf("expected unequal configs for different application keys")
	}
}
