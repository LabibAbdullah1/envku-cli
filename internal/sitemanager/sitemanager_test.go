package sitemanager

import (
	"testing"
)

func TestIsLocalDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"myapp.test", true},
		{"api.myapp.test", true},
		{"project.local", true},
		{"devserver.internal", true},
		{"singleword", true},
		{"example.com", false},
		{"app.myserver.org", false},
		{"myvps.id", false},
	}

	for _, tt := range tests {
		result := IsLocalDomain(tt.domain)
		if result != tt.expected {
			t.Errorf("IsLocalDomain(%q) = %v; want %v", tt.domain, result, tt.expected)
		}
	}
}

func TestGenerateNginxConfig(t *testing.T) {
	cfg := SiteConfig{
		Domain:       "testapp.test",
		Type:         SiteTypePHP,
		PHPVersion:   "8.3",
		DocumentRoot: "/var/www/testapp/public",
	}

	conf := GenerateNginxConfig(cfg)
	if conf == "" {
		t.Fatal("expected generated nginx config, got empty string")
	}
}

