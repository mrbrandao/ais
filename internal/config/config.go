// Package config resolves MENTAL_DIR and manages application
// configuration using Viper.
//
// # MENTAL_DIR Resolution
//
// The base directory for mental's data follows this priority order:
//
//  1. MENTAL_DIR environment variable (if set and non-empty)
//  2. $XDG_DATA_HOME/mental (if XDG_DATA_HOME is set)
//  3. ~/.local/share/mental (XDG default on Linux)
//
// # Config File Resolution
//
// mental.toml is searched in this order (first found wins):
//
//  1. MENTAL_CONFIG environment variable path
//  2. $MENTAL_DIR/mental.toml (user override)
//  3. Embedded default (built-in mental.toml)
//
// # Usage
//
//	cfg, err := config.Load()
//	if err != nil {
//	    return fmt.Errorf("config: %w", err)
//	}
//	dir := cfg.Dir()
//	engine := cfg.MemEngine()   // default: "memx"
//	provider := cfg.SessionProvider() // default: "opencode"
package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

//go:embed mental.toml
var defaultTOML []byte

const (
	// EnvDir is the environment variable that overrides the data directory.
	EnvDir = "MENTAL_DIR"

	// EnvConfig is the environment variable that overrides the config file path.
	EnvConfig = "MENTAL_CONFIG"

	// DefaultDirName is the subdirectory name within XDG_DATA_HOME.
	DefaultDirName = "mental"

	// ConfigFile is the name of the configuration file.
	ConfigFile = "mental.toml"
)

// Config holds resolved application configuration.
// Obtain a Config via Load — do not construct directly.
type Config struct {
	dir string
	v   *viper.Viper
}

// Load resolves MENTAL_DIR and reads mental.toml configuration.
// CLI flags always override config file values; config file
// values override the embedded defaults.
func Load() (*Config, error) {
	dir, err := resolveDir()
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}

	v, err := loadViper(dir)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &Config{dir: dir, v: v}, nil
}

// Dir returns the resolved MENTAL_DIR path.
func (c *Config) Dir() string { return c.dir }

// MemEngine returns the configured memory engine name.
// Default: "memx". Override via [mem] engine in mental.toml.
func (c *Config) MemEngine() string {
	return c.v.GetString("mem.engine")
}

// SessionProvider returns the configured session provider name.
// Default: "opencode". Override via [session] provider in mental.toml.
func (c *Config) SessionProvider() string {
	return c.v.GetString("session.provider")
}

// LLMModel returns the configured LLM model string (provider/model format).
// Empty string means no default LLM — use print/pipe mode.
func (c *Config) LLMModel() string {
	return c.v.GetString("llm.model")
}

// loadViper initialises a Viper instance with embedded defaults,
// then overlays the user's mental.toml if found.
func loadViper(dir string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetEnvPrefix("MENTAL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Load embedded defaults first.
	if err := v.ReadConfig(
		strings.NewReader(string(defaultTOML)),
	); err != nil {
		return nil, fmt.Errorf("read embedded config: %w", err)
	}

	// Overlay user config file if present.
	userCfg := userConfigPath(dir)
	if _, err := os.Stat(userCfg); err == nil {
		v.SetConfigFile(userCfg)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf(
				"merge %s: %w", userCfg, err,
			)
		}
	}

	return v, nil
}

// userConfigPath returns the path to the user's mental.toml.
// Priority: MENTAL_CONFIG env → $MENTAL_DIR/mental.toml.
func userConfigPath(dir string) string {
	if v := os.Getenv(EnvConfig); v != "" {
		return v
	}
	return filepath.Join(dir, ConfigFile)
}

// resolveDir returns the mental data directory.
// Priority: MENTAL_DIR env → XDG_DATA_HOME/mental → ~/.local/share/mental.
func resolveDir() (string, error) {
	if v := os.Getenv(EnvDir); v != "" {
		return v, nil
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, DefaultDirName), nil
	}
	return xdgDataHome()
}

// xdgDataHome returns ~/.local/share/mental (XDG default on Linux).
func xdgDataHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("UserHomeDir: %w", err)
	}
	return filepath.Join(home, ".local", "share", DefaultDirName), nil
}
