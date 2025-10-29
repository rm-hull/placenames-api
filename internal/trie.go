package internal

import (
	"fmt"
	"log"
	"sort"
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

func PopulateFrom(filename string, topK int) (*Trie, error) {
	trie := NewTrie(topK)
	count, err := LoadCSV(filename, func(location string, score float64) error {
		trie.Insert(&Place{Name: location, Relevancy: score})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to populate trie: %w", err)
	}
	log.Printf("Loaded %d place names into trie structure", count)

	return trie, nil
}
