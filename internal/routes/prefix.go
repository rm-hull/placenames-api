package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/placenames-api/internal"
)

type PlaceResponse struct {
	Results []*internal.Place `json:"results"`
}

func Prefix(trie *internal.Trie) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Param("query")
		maxResults := 10
		if maxStr := c.Query("max_results"); maxStr != "" {
			if max, err := strconv.Atoi(maxStr); err == nil && max > 0 && max <= 100 {
				maxResults = max
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "max_results must be a positive integer less than or equal to 100"})
				return
			}
		}

		results := trie.FindByPrefix(query)
		maxResults = min(maxResults, len(results))
		c.JSON(http.StatusOK, PlaceResponse{Results: results[:maxResults]})
	}
}
