package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/place-names/internal"
)

func main() {

	trie, err := loadData("data/placenames_with_relevancy.csv.gz")
	if err != nil {
		log.Fatalf("Error loading data: %v", err)
	}

	r := setupServer(trie)
	log.Fatal(r.Run(":8080"))
}

func loadData(filename string) (*internal.Trie, error) {
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
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %w", err)
	}

	trie := internal.NewTrie()
	for i, rec := range records {
		if i == 0 {
			continue // skip header
		}
		if len(rec) < 2 {
			continue
		}
		name := rec[0]
		var rel float64
		rel, err := strconv.ParseFloat(rec[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid relevancy value on line %d: %w", i+1, err)
		}
		trie.Insert(internal.Place{Name: name, Relevancy: rel})
	}
	trie.SortAllNodes()

	return trie, nil
}

func setupServer(trie *internal.Trie) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.GET("/place-names/prefix/:query", func(c *gin.Context) {
			query := c.Param("query")
			maxResults := 10
			if maxStr := c.Query("max_results"); maxStr != "" {
                if max, err := strconv.Atoi(maxStr); err == nil && max > 0 {
                    maxResults = max
                } else {
                    c.JSON(http.StatusBadRequest, gin.H{"error": "max_results must be a positive integer"})
                    return
                }
			}

			results := trie.FindByPrefix(query)
			maxResults = min(maxResults, len(results))

			c.JSON(http.StatusOK, map[string][]internal.Place{"results": results[:maxResults]})
		})
	}
	return r
}
