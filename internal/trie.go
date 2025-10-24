package internal

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Place struct {
	Name      string
	Relevancy float64
}

type TrieNode struct {
	Children map[rune]*TrieNode
	Places   *MinHeap[*Place] // Store pointers instead of values to reduce memory duplication
}

type Trie struct {
	root *TrieNode
	less func(a, b *Place) bool
	topK int
}

func NewTrie(maxPerNode int) *Trie {
	less := func(a, b *Place) bool {
		if a.Relevancy == b.Relevancy {
			return len(a.Name) > len(b.Name)
		}
		return a.Relevancy < b.Relevancy
	}
	return &Trie{
		root: &TrieNode{
			Children: make(map[rune]*TrieNode),
			Places:   NewMinHeap(less),
		},
		less: less,
		topK: maxPerNode,
	}
}

func (t *Trie) TopK() int {
	return t.topK
}

func (t *Trie) Insert(place *Place) {
	node := t.root
	lower := strings.ToLower(place.Name)

	for _, r := range lower {
		if node.Children[r] == nil {
			node.Children[r] = &TrieNode{
				Children: make(map[rune]*TrieNode),
				Places:   NewMinHeap(t.less),
			}
		}
		node = node.Children[r]
		node.Places.PushBounded(place, t.topK)
	}
}

func (t *Trie) FindByPrefix(prefix string) []*Place {
	node := t.root
	lower := strings.ToLower(prefix)
	for _, r := range lower {
		next := node.Children[r]
		if next == nil {
			return []*Place{}
		}
		node = next
	}

	items := node.Places.Items()
	result := make([]*Place, len(items))
	copy(result, items)

sort.SliceStable(result, func(i, j int) bool {
		return t.less(result[j], result[i]) // note: reverse order
	})

	return result
}

func LoadData(filename string, topK int) (*Trie, error) {
	log.Printf("Loading data from: %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Printf("Error closing gzip reader: %v", err)
		}
	}()

	csvReader := csv.NewReader(gzReader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields
	trie := NewTrie(topK)
	line := 0

	for {
		line++
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record on line %d: %w", line, err)
		}

		// Skip header
		if line == 1 {
			continue
		}

		if len(rec) < 2 {
			return nil, fmt.Errorf("invalid record on line %d: expected at least 2 fields, got %d", line, len(rec))
		}
		name := rec[0]
		rel, err := strconv.ParseFloat(rec[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid relevancy value on line %d: %w", line, err)
		}
		trie.Insert(&Place{Name: name, Relevancy: rel})
	}
	log.Printf("Loaded %d place names", line)

	return trie, nil
}
