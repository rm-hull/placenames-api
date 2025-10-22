package internal

import (
	"sort"
	"strings"
)

type Place struct {
	Name      string  `json:"name"`
	Relevancy float64 `json:"relevancy"`
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
