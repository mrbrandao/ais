package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDir(t *testing.T) {

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	base := filepath.Join(home, ".local", "share")

	tests := []struct {
		name       string
		mentalDir  string
		xdgHome    string
		wantSuffix string
	}{
		{
			name:       "MENTAL_DIR takes priority",
			mentalDir:  "/custom/mental",
			xdgHome:    "/xdg/data",
			wantSuffix: "/custom/mental",
		},
		{
			name:       "XDG_DATA_HOME used when MENTAL_DIR unset",
			mentalDir:  "",
			xdgHome:    "/xdg/data",
			wantSuffix: "/xdg/data/mental",
		},
		{
			name:       "default XDG path when both unset",
			mentalDir:  "",
			xdgHome:    "",
			wantSuffix: filepath.Join(base, "mental"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvDir, tc.mentalDir)
			t.Setenv("XDG_DATA_HOME", tc.xdgHome)

			got, err := resolveDir()
			if err != nil {
				t.Fatalf("resolveDir: %v", err)
			}
			if !strings.HasSuffix(got, tc.wantSuffix) &&
				got != tc.wantSuffix {
				t.Errorf(
					"got %q, want suffix %q",
					got, tc.wantSuffix,
				)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.MemEngine() != "memx" {
		t.Errorf("MemEngine: got %q, want memx", cfg.MemEngine())
	}
	if cfg.SessionProvider() != "opencode" {
		t.Errorf(
			"SessionProvider: got %q, want opencode",
			cfg.SessionProvider(),
		)
	}
	if cfg.LLMModel() != "" {
		t.Errorf("LLMModel: got %q, want empty", cfg.LLMModel())
	}
}

func TestConfigEnvOverride(t *testing.T) {
	// MENTAL_MEM_ENGINE overrides mem.engine via viper AutomaticEnv.
	t.Setenv("MENTAL_MEM_ENGINE", "myhms")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.MemEngine() != "myhms" {
		t.Errorf("MemEngine: got %q, want myhms", cfg.MemEngine())
	}

	t.Setenv("MENTAL_MEM_ENGINE", "")
}
