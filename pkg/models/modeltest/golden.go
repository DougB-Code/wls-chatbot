// golden.go provides golden file management for captured request/response pairs.
package modeltest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GoldenFileVersion is the schema version for golden files.
const GoldenFileVersion = "1.0"

// GoldenMetadata contains metadata about a golden file.
type GoldenMetadata struct {
	Provider   string    `json:"provider"`
	Capability string    `json:"capability"`
	Model      string    `json:"model"`
	Timestamp  time.Time `json:"timestamp"`
	Version    string    `json:"version"`
}

// GoldenFile represents a complete golden file with metadata and recording.
type GoldenFile struct {
	Metadata GoldenMetadata   `json:"metadata"`
	Request  RecordedRequest  `json:"request"`
	Response RecordedResponse `json:"response"`
}

// NewGoldenFile creates a golden file from a recording.
func NewGoldenFile(provider, capability, model string, rec Recording) *GoldenFile {
	return &GoldenFile{
		Metadata: GoldenMetadata{
			Provider:   provider,
			Capability: capability,
			Model:      model,
			Timestamp:  rec.Timestamp,
			Version:    GoldenFileVersion,
		},
		Request:  rec.Request,
		Response: rec.Response,
	}
}

// Filename generates the standard filename for this golden file.
func (g *GoldenFile) Filename() string {
	// Sanitize model name for filesystem
	model := strings.ReplaceAll(g.Metadata.Model, "/", "_")
	model = strings.ReplaceAll(model, ":", "_")
	
	date := g.Metadata.Timestamp.Format("20060102")
	return fmt.Sprintf("%s_%s_%s.json", g.Metadata.Capability, model, date)
}

// Save writes the golden file to the specified directory.
func (g *GoldenFile) Save(baseDir string) error {
	dir := filepath.Join(baseDir, g.Metadata.Provider)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	path := filepath.Join(dir, g.Filename())
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal golden file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write golden file: %w", err)
	}

	return nil
}

// LoadGoldenFile reads a golden file from disk.
func LoadGoldenFile(path string) (*GoldenFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read golden file: %w", err)
	}

	var g GoldenFile
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("unmarshal golden file: %w", err)
	}

	return &g, nil
}

// LoadGoldenFiles loads all golden files from a directory.
func LoadGoldenFiles(baseDir string) ([]*GoldenFile, error) {
	var files []*GoldenFile

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		g, err := LoadGoldenFile(path)
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		files = append(files, g)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// GoldenIndex provides fast lookup of golden files by provider/capability/model.
type GoldenIndex struct {
	files map[string]*GoldenFile // key: provider/capability/model
}

// NewGoldenIndex creates an index from loaded golden files.
func NewGoldenIndex(files []*GoldenFile) *GoldenIndex {
	idx := &GoldenIndex{
		files: make(map[string]*GoldenFile),
	}
	for _, f := range files {
		key := f.Metadata.Provider + "/" + f.Metadata.Capability + "/" + f.Metadata.Model
		idx.files[key] = f
	}
	return idx
}

// Lookup finds a golden file by provider, capability, and model.
func (idx *GoldenIndex) Lookup(provider, capability, model string) *GoldenFile {
	key := provider + "/" + capability + "/" + model
	return idx.files[key]
}
