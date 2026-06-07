package lockfile

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const LockFileName = ".score-hub.lock"

type LockFile struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Entries    []LockEntry `yaml:"entries"`
}

type LockEntry struct {
	Name          string `yaml:"name"`
	Variant       string `yaml:"variant"`
	Platform      string `yaml:"platform"`
	Version       string `yaml:"version"`
	Commit        string `yaml:"commit,omitempty"`
	InstalledFile string `yaml:"installedFile"`
	Filename      string `yaml:"filename"`
	Checksum      string `yaml:"checksum"`
	InstalledAt   string `yaml:"installedAt"`
	Registry      string `yaml:"registry,omitempty"`
	RegistryURL   string `yaml:"registryURL,omitempty"`
	DownloadURL   string `yaml:"downloadURL,omitempty"` // absolute file URL (empty for community/path-based)
}

func (e LockEntry) EffectiveRegistry() string {
	if e.Registry == "" {
		return "public"
	}
	return e.Registry
}

func New() *LockFile {
	return &LockFile{
		APIVersion: "score-hub/v1alpha1",
		Kind:       "LockFile",
		Entries:    []LockEntry{},
	}
}

func Load(projectDir string) (*LockFile, error) {
	lockPath := filepath.Join(projectDir, LockFileName)
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}
	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}
	return &lock, nil
}

func (l *LockFile) Save(projectDir string) error {
	lockPath := filepath.Join(projectDir, LockFileName)
	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}
	header := "# score-hub lock file\n# Commit this to source control.\n\n"
	return os.WriteFile(lockPath, []byte(header+string(data)), 0644)
}

func (l *LockFile) AddEntry(entry LockEntry) {
	for i, e := range l.Entries {
		if e.Name == entry.Name && e.Platform == entry.Platform {
			l.Entries[i] = entry
			return
		}
	}
	l.Entries = append(l.Entries, entry)
}

func (l *LockFile) FindEntry(name string, platform string) *LockEntry {
	for i := range l.Entries {
		if l.Entries[i].Name == name {
			if platform == "" || l.Entries[i].Platform == platform {
				return &l.Entries[i]
			}
		}
	}
	return nil
}
