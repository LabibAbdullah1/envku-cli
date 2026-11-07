package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("expected default config, got nil")
	}

	if cfg.PMAPort != 8080 {
		t.Errorf("expected default PMA port 8080, got %d", cfg.PMAPort)
	}

	if cfg.Environment != ModeProduction && cfg.Environment != ModeLocal {
		t.Errorf("unexpected environment mode: %s", cfg.Environment)
	}
}

