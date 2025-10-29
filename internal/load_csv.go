package internal

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

func LoadCSV(filename string, action func(location string, score float64) error) (int, error) {
	log.Printf("Loading data from: %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Printf("Error closing gzip reader: %v", err)
		}
	}()

	csvReader := csv.NewReader(gzReader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields
	count := 0

	for {
		count++
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("failed to read CSV record on line %d: %w", count, err)
		}

		// Skip header
		if count == 1 {
			continue
		}

		if len(rec) < 2 {
			return 0, fmt.Errorf("invalid record on line %d: expected at least 2 fields, got %d", count, len(rec))
		}
		name := rec[0]
		score, err := strconv.ParseFloat(rec[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid score value on line %d: %w", count, err)
		}

		if err := action(name, score); err != nil {
			return 0, fmt.Errorf("failed to action (%s, %f): %w", name, score, err)
		}
	}

	return count, nil
}
