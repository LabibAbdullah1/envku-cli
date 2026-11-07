package sitemanager

import (
	"os"
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

func TestEnsureStarterFile(t *testing.T) {
	tempDir := t.TempDir()

	cfgPHP := SiteConfig{
		Domain:       "phptest.test",
		Type:         SiteTypePHP,
		DocumentRoot: tempDir,
	}

	if err := EnsureStarterFile(cfgPHP, tempDir); err != nil {
		t.Fatalf("EnsureStarterFile PHP failed: %v", err)
	}

	// Verify index.php exists
	indexPath := tempDir + "/index.php"
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatalf("expected index.php to be created, but it was not")
	}

	// Test static
	tempStatic := t.TempDir()
	cfgStatic := SiteConfig{
		Domain:       "statictest.test",
		Type:         SiteTypeStatic,
		DocumentRoot: tempStatic,
	}

	if err := EnsureStarterFile(cfgStatic, tempStatic); err != nil {
		t.Fatalf("EnsureStarterFile Static failed: %v", err)
	}

	htmlPath := tempStatic + "/index.html"
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Fatalf("expected index.html to be created, but it was not")
	}
}

