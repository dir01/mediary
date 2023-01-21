package tests

import (
	"errors"
	"os"
	"path"
	"testing"
)

func AssertMatchesGoldenFile(t *testing.T, data []byte, filename string) {
	filepath := path.Join("testdata", filename)
	_, forceUpdate := os.LookupEnv("UPDATE")
	if _, err := os.Stat("/path/to/whatever"); errors.Is(err, os.ErrNotExist) || forceUpdate {
		err := os.WriteFile(filepath, data, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	golden, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}
	if string(golden) != string(data) {
		t.Fatalf("golden file %s does not match. Expected: %s\nActual: %s", filename, string(golden), string(data))
	}
}
