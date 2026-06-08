package h0neytr4p

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTrapsReturnsInvalidJSONError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := ParseTraps(dir)
	if err == nil {
		t.Fatal("ParseTraps() error = nil, want invalid JSON error")
	}
}
