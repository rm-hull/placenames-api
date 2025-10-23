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
	Places   []*Place // Store pointers instead of values to reduce memory duplication
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{root: &TrieNode{Children: make(map[rune]*TrieNode)}}
}

func (t *Trie) Insert(place *Place) {
	node := t.root

	lower := strings.ToLower(place.Name)
	for _, r := range lower {
		if node.Children[r] == nil {
			node.Children[r] = &TrieNode{Children: make(map[rune]*TrieNode)}
		}
		node = node.Children[r]
		// Store a pointer to the place at every node along the path
		// to support efficient ranked prefix search.
		node.Places = append(node.Places, place)
	}
}

// After all insertions, sort the Places slice at every node by (relevancy DESC, name length ASC)
func (t *Trie) SortAllNodes() {
	var dfs func(*TrieNode)
	dfs = func(n *TrieNode) {
		if n == nil {
			return
		}
		sort.SliceStable(n.Places, func(i, j int) bool {
			pi, pj := n.Places[i], n.Places[j]
			if pi.Relevancy == pj.Relevancy {
				return len(pi.Name) < len(pj.Name)
			}
			return pi.Relevancy > pj.Relevancy
		})
		for _, child := range n.Children {
			dfs(child)
		}
	}
	dfs(t.root)
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

	return node.Places
}

func LoadData(filename string) (*Trie, error) {
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
	trie := NewTrie()
	line := 0

	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record on line %d: %w", line+1, err)
		}
		// csv.Reader counts lines starting from 1; we keep our own counter to track header
		line++
		if line == 1 {
			continue // skip header
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
	trie.SortAllNodes()
	log.Printf("Loaded %d place names", line)

	return trie, nil
}
