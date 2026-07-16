package extensions

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// extensionsSubdir is the directory under MENTAL_DIR for external extensions.
	extensionsSubdir = "extensions"

	// manifestFile is the name of the extension manifest within each extension dir.
	manifestFile = "extension.yaml"
)

// DiscoverExternal scans mentalDir/extensions/ for subdirectories
// containing an extension.yaml manifest. Each discovered extension is
// registered with the provided manager.
//
// Discovery rules:
//   - The directory name is the extension's identifier.
//   - Duplicate names: first found wins; subsequent entries are skipped
//     with a warning written to stderr.
//   - Directories without a valid extension.yaml are silently skipped.
func DiscoverExternal(m *Manager, mentalDir, version string) error {
	base := filepath.Join(mentalDir, extensionsSubdir)

	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no extensions directory is fine
		}
		return fmt.Errorf("read extensions dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifest, bin, err := loadManifest(base, entry.Name())
		if err != nil {
			// Malformed manifest — skip silently.
			continue
		}

		ext := NewExternalExtension(manifest, bin, mentalDir, version)
		if err := m.Register(ext); err != nil {
			// Duplicate name — warn and continue.
			log.Printf(
				"warning: skipping extension %q: %v",
				entry.Name(), err,
			)
		}
	}
	return nil
}

// loadManifest reads and parses extension.yaml from an extension directory.
// Returns the manifest and the absolute path to the executable.
func loadManifest(base, name string) (Manifest, string, error) {
	manifestPath := filepath.Join(base, name, manifestFile)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return Manifest{}, "", fmt.Errorf(
			"read %s: %w", manifestPath, err,
		)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, "", fmt.Errorf(
			"parse %s: %w", manifestPath, err,
		)
	}

	if m.Executable == "" {
		return Manifest{}, "", fmt.Errorf(
			"%s: executable field is required", manifestPath,
		)
	}

	bin := filepath.Join(base, name, m.Executable)
	if _, err := os.Stat(bin); err != nil {
		return Manifest{}, "", fmt.Errorf(
			"executable %s not found: %w", bin, err,
		)
	}

	return m, bin, nil
}
