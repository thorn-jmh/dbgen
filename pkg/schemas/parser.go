package schemas

import (
	"encoding/json"
	"github.com/thorn-jmh/errorst"
	"io"
	"os"
)

// FromJSONFile reads from a  JSON file and returns a Schema.
func FromJSONFile(filePath string) (*Schema, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errorst.NewError("failed to open file %s: %w", filePath, err)
	}

	defer func() {
		_ = f.Close()
	}()

	return FromJSON(f)
}

// FromJSON reads from a JSON reader and returns a Schema.
func FromJSON(r io.Reader) (*Schema, error) {
	var schema Schema
	if err := json.NewDecoder(r).Decode(&schema); err != nil {
		return nil, errorst.NewError("failed to unmarshal JSON: %w", err)
	}

	return &schema, nil
}
