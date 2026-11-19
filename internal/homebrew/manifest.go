package homebrew

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var sha256Re = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

type Asset struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	URL      string `json:"url"`
	SHA256   string `json:"sha256"`
	DestPath string `json:"dest"`
}

type Manifest struct {
	Version string  `json:"version"`
	Assets  []Asset `json:"assets"`
}

func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", path, err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	for i, a := range m.Assets {
		if strings.TrimSpace(a.Name) == "" {
			return nil, fmt.Errorf("asset[%d]: empty name", i)
		}
		if strings.TrimSpace(a.URL) == "" {
			return nil, fmt.Errorf("asset %q: empty URL", a.Name)
		}
		if !sha256Re.MatchString(a.SHA256) {
			return nil, fmt.Errorf("asset %q: invalid sha256 %q (must be 64 hex chars)", a.Name, a.SHA256)
		}
		if strings.TrimSpace(a.DestPath) == "" {
			return nil, fmt.Errorf("asset %q: empty dest path", a.Name)
		}
	}

	return &m, nil
}
