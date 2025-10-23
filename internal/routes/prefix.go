package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/placenames-api/internal"
)

type Result struct {
	Name      string  `json:"name"`
	Relevancy float64 `json:"relevancy"`
}

type PlaceResponse struct {
	Results []Result `json:"results"`
}

func applyPrefixCasing(candidate string, prefix string) string {
	if len(prefix) == 0 {
		return candidate
	}

	result := []rune(candidate)
	prefixRunes := []rune(prefix)
	for i := range min(len(prefixRunes), len(result)) {
		if unicode.IsUpper(prefixRunes[i]) {
			result[i] = unicode.ToUpper(result[i])
		} else {
			result[i] = unicode.ToLower(result[i])
		}
	}

	return string(result)
}

func Prefix(trie *internal.Trie) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Param("query")
		maxResults := 10
		if maxStr := c.Query("max_results"); maxStr != "" {
			if max, err := strconv.Atoi(maxStr); err == nil && max > 0 && max <= trie.TopK() {
				maxResults = max
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("max_results must be a positive integer less than or equal to %d", trie.TopK()),
				})
				return
			}
		}

		matches := trie.FindByPrefix(query)
		maxResults = min(maxResults, len(matches))

		results := make([]Result, maxResults)
		for i, match := range matches[:maxResults] {
			results[i] = Result{
				Name:      applyPrefixCasing(match.Name, query),
				Relevancy: match.Relevancy,
			}
		}

		c.JSON(http.StatusOK, PlaceResponse{Results: results})
	}
}
