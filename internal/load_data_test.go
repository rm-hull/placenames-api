package internal

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestGzipFile creates a temporary gzipped CSV file for testing.
func createTestGzipFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv.gz")

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzWriter := gzip.NewWriter(file)
	defer func() {
		_ = gzWriter.Close()
	}()

	_, err = gzWriter.Write([]byte(content))
	if err != nil {
		t.Fatalf("failed to write to gzip writer: %v", err)
	}

	return path
}

func TestLoadData(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		content := `name,relevancy
London,1.0
Luton,0.8
`
		path := createTestGzipFile(t, content)
		trie, err := PopulateFrom(path, 100)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		results := trie.FindByPrefix("L")
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}

		results = trie.FindByPrefix("Lu")
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
		if results[0].Name != "Luton" {
			t.Errorf("expected Luton, got %s", results[0].Name)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := PopulateFrom("non-existent-file.csv.gz", 100)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if !strings.Contains(err.Error(), "no such file or directory") {
			t.Errorf("expected error to contain 'no such file or directory', got %v", err)
		}
	})

	t.Run("invalid gzip file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(path, []byte("not a gzip file"), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		_, err := PopulateFrom(path, 100)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid header") {
			t.Errorf("expected error to contain 'invalid header', got %v", err)
		}
	})

	t.Run("invalid csv content", func(t *testing.T) {
		// Test case with a row that has a non-numeric relevancy
		content := `name,relevancy
London,1.0
Paris,invalid
`
		path := createTestGzipFile(t, content)
		_, err := PopulateFrom(path, 100)
		if err == nil {
			t.Fatal("expected an error for invalid score, got nil")
		}
		if !strings.Contains(err.Error(), "invalid score value") {
			t.Errorf("expected error to contain 'invalid score value', got %v", err)
		}
	})

	t.Run("csv content with wrong number of columns", func(t *testing.T) {
		content := `name,relevancy
London,1.0
Paris
`
		path := createTestGzipFile(t, content)
		_, err := PopulateFrom(path, 100)
		if err == nil {
			t.Fatal("expected an error for wrong number of columns, got nil")
		}
		if !strings.Contains(err.Error(), "expected at least 2 fields") {
			t.Errorf("expected error to contain 'expected at least 2 fields', got %v", err)
		}
	})

	t.Run("malformed csv record", func(t *testing.T) {
		// Test case with a malformed CSV record (unclosed quote)
		content := `name,relevancy
"London,1.0
`
		path := createTestGzipFile(t, content)
		_, err := PopulateFrom(path, 100)
		if err == nil {
			t.Fatal("expected an error for malformed CSV, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read CSV record") {
			t.Errorf("expected error to contain 'failed to read CSV record', got %v", err)
		}
	})

	t.Run("successful load of large file", func(t *testing.T) {
		// This is more of an integration test and relies on the actual data file.
		// It's useful for ensuring the LoadData function can handle the real data.
		const dataFile = "../data/placenames_with_relevancy.csv.gz"

		if _, err := os.Stat(dataFile); os.IsNotExist(err) {
			t.Skipf("data file not found: %s, skipping test", dataFile)
		}

		trie, err := PopulateFrom(dataFile, 100)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		results := trie.FindByPrefix("Lond")
		if len(results) == 0 {
			t.Fatal("expected at least one result for 'Lond', got 0")
		}

		if results[0].Name != "London" {
			t.Errorf("expected the first result to be 'London', got %s", results[0].Name)
		}
	})
}
