package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/rm-hull/placenames-api/internal"
)

const (
	apiURL     = "http://hydra.local:8080/v1/chat/completions"
	outputFile = "popularity_scores.csv"
)

//go:embed system_prompt.md
var systemPrompt string

var (
	reNumber   = regexp.MustCompile(`\b0(?:\.\d+)?|1(?:\.0+)?\b`)
	httpClient = &http.Client{Timeout: 30 * time.Second}
)

// Result holds the location, score, original index, and any error message
type Result struct {
	Index    int
	Location string
	Score    string
	Err      error
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model,omitempty"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Messages    []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func getPopularityScore(location string) (float64, error) {
	reqBody := chatRequest{
		// Model:       "local-llama",
		Temperature: 0.1,
		MaxTokens:   6,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: location},
		},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var data chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	content := strings.TrimSpace(data.Choices[0].Message.Content)
	match := reNumber.FindString(content)
	if match == "" {
		return 0, fmt.Errorf("no numeric output")
	}

	score, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0, err
	}

	return math.Min(1, math.Max(0, score)), nil
}

func worker(id int, jobs <-chan [2]interface{}, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		idx := job[0].(int)
		location := job[1].(string)

		score, err := getPopularityScore(location)
		if err != nil {
			log.Printf("Worker %d: %s → ERROR: %v\n", id, location, err)
			results <- Result{Index: idx, Location: location, Err: err}
			continue
		}
		log.Printf("Worker %d: %s → %.2f\n", id, location, score)
		results <- Result{Index: idx, Location: location, Score: fmt.Sprintf("%.2f", score)}
		// time.Sleep(100 * time.Millisecond)
	}
}

func RegenCSV(filename string, numWorkers int) error {
	var locations []string
	count, err := internal.LoadCSV(filename, func(location string, score float64) error {
		locations = append(locations, location)
		return nil
	})
	if err != nil {
		return nil
	}
	log.Printf("Loaded %d locations...", count)

	// Channels for jobs and results
	jobs := make(chan [2]interface{}, numWorkers*2)
	results := make(chan Result, len(locations))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i+1, jobs, results, &wg)
	}

	// Send jobs
	go func() {
		for idx, loc := range locations {
			jobs <- [2]interface{}{idx, loc}
		}
		close(jobs)
	}()

	wg.Wait()
	close(results)

	// Collect and sort results by index
	finalResults := make([]Result, 0, count)
	for res := range results {
		finalResults = append(finalResults, res)
	}
	sort.SliceStable(finalResults, func(i, j int) bool {
		return finalResults[i].Index < finalResults[j].Index
	})

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = outFile.Close()
	}()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()
	if err := writer.Write([]string{"location", "score"}); err != nil {
		return err
	}

	for _, r := range finalResults {
		if r.Err != nil {
			return fmt.Errorf("failed to determine score for %s (%d): %w", r.Location, r.Index, err)
		}
		if err := writer.Write([]string{r.Location, r.Score}); err != nil {
			return err
		}
	}
	writer.Flush()

	log.Printf("Output written to: %s", outputFile)
	return nil
}
