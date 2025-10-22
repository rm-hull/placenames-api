package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Depado/ginprom"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rm-hull/place-names/internal"

	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

type PlaceResponse struct {
	Results []*internal.Place `json:"results"`
}

func main() {

	trie, err := loadData("data/placenames_with_relevancy.csv.gz")
	if err != nil {
		log.Fatalf("Error loading data: %v", err)
	}

	r, err := setupServer(trie)
	if err != nil {
		log.Fatalf("Error setting up server: %v", err)
	}
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
	trie := internal.NewTrie()
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
		trie.Insert(&internal.Place{Name: name, Relevancy: rel})
	}
	trie.SortAllNodes()

	return trie, nil
}

func setupServer(trie *internal.Trie) (*gin.Engine, error) {
	r := gin.New()

	prometheus := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz"),
	)

	r.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz", "/metrics"),
		prometheus.Instrument(),
		cors.Default(),
	)

	err := healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize healthcheck: %w", err)
	}

	v1 := r.Group("/v1")
	v1.GET("/place-names/prefix/:query", func(c *gin.Context) {
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
	})

	return r, nil
}
