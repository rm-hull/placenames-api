package internal

import (
	"testing"
)

func BenchmarkTrieMemoryUsage(b *testing.B) {
	// Sample data with repeated prefixes to demonstrate memory savings
	testData := []Place{
		{Name: "London", Relevancy: 1.0},
		{Name: "Los Angeles", Relevancy: 0.9},
		{Name: "Liverpool", Relevancy: 0.8},
		{Name: "Leeds", Relevancy: 0.7},
		{Name: "Leicester", Relevancy: 0.6},
	}

	b.Run("Insert and Search", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			trie := NewTrie() // Create fresh trie for each iteration
			b.StartTimer()

			// Insert all test data
			for _, p := range testData {
				trie.Insert(p)
			}

			// Search to ensure correctness
			results := trie.FindByPrefix("L")
			if len(results) != 5 {
				b.Fatalf("expected 5 results, got %d", len(results))
			}
		}
	})

	b.Run("Memory Allocation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			trie := NewTrie()

			// Insert all test data once to measure allocations
			for _, p := range testData {
				trie.Insert(p)
			}

			// Verify trie is working correctly
			results := trie.FindByPrefix("L")
			if len(results) != 5 {
				b.Fatalf("expected 5 results, got %d", len(results))
			}
		}
	})
}

func TestTrieBasics(t *testing.T) {
	trie := NewTrie()

	// Test data
	places := []Place{
		{Name: "London", Relevancy: 1.0},
		{Name: "Los Angeles", Relevancy: 0.9},
		{Name: "Liverpool", Relevancy: 0.8},
	}

	// Insert test data
	for _, p := range places {
		trie.Insert(p)
	}

	// Test exact prefix match
	results := trie.FindByPrefix("Lo")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'Lo', got %d", len(results))
	}

	// Test case insensitivity
	results = trie.FindByPrefix("lo")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'lo', got %d", len(results))
	}

	// Test sorting by relevancy
	if len(results) > 0 && results[0].Name != "London" {
		t.Errorf("expected first result to be London, got %s", results[0].Name)
	}

	// Test no results
	results = trie.FindByPrefix("X")
	if len(results) != 0 {
		t.Errorf("expected 0 results for 'X', got %d", len(results))
	}
}
