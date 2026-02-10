package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/blevesearch/bleve/v2"
)

const (
	indexPath = ".lazyman_index"
)

// ManPageDocument represents a man page document for indexing
type ManPageDocument struct {
	Name        string
	Section     string
	Description string
	Content     string
	Path        string
}

// IndexAllManPages builds or rebuilds the search index with parallel processing
func IndexAllManPages() error {
	fmt.Println("Building search index for all man pages...")
	fmt.Println("This may take a few minutes on first run...")

	// Remove existing index if it exists
	if err := os.RemoveAll(indexPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old index: %w", err)
	}

	// Create a new index
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer index.Close()

	// Get all man pages
	pages, err := GetManPages()
	if err != nil {
		return fmt.Errorf("failed to get man pages: %w", err)
	}

	fmt.Printf("Found %d man pages to index\n", len(pages))
	fmt.Println("Fetching man page content in parallel...")

	// Use worker pool for parallel content fetching
	numWorkers := 100
	jobs := make(chan ManPage, len(pages))
	results := make(chan ManPageDocument, 100) // Buffered channel
	var wg sync.WaitGroup
	var processed atomic.Int32

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range jobs {
				// Read raw man page file directly (much faster than calling man command)
				content, err := GetRawManContent(page.Path)
				if err != nil {
					// Skip pages that fail to load
					processed.Add(1)
					continue
				}

				doc := ManPageDocument{
					Name:        page.Name,
					Section:     page.Section,
					Description: page.Description,
					Content:     content,
					Path:        page.Path,
				}

				results <- doc
				processed.Add(1)
			}
		}()
	}

	// Start a goroutine to close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Send jobs
	go func() {
		for _, page := range pages {
			jobs <- page
		}
		close(jobs)
	}()

	// Collect results and batch index in main goroutine
	batch := index.NewBatch()
	batchSize := 100
	count := 0
	lastReport := 0

	for doc := range results {
		// Use name(section) as document ID
		docID := fmt.Sprintf("%s(%s)", doc.Name, doc.Section)
		if err := batch.Index(docID, doc); err != nil {
			return fmt.Errorf("failed to index %s: %w", docID, err)
		}

		count++

		// Report progress every 100 processed items
		currentProcessed := int(processed.Load())
		if currentProcessed-lastReport >= 100 {
			fmt.Printf("Progress: %d/%d man pages processed...\n", currentProcessed, len(pages))
			lastReport = currentProcessed
		}

		// Commit batch every batchSize documents
		if count%batchSize == 0 {
			if err := index.Batch(batch); err != nil {
				return fmt.Errorf("failed to commit batch: %w", err)
			}
			batch = index.NewBatch()
		}
	}

	// Commit remaining documents
	if batch.Size() > 0 {
		if err := index.Batch(batch); err != nil {
			return fmt.Errorf("failed to commit final batch: %w", err)
		}
	}

	fmt.Printf("✓ Successfully indexed %d man pages (processed %d total)\n", count, processed.Load())
	return nil
}

// SearchResult represents a search result with context
type SearchResult struct {
	ManPage   ManPage
	Matches   []string // Lines containing matches
	Score     float64
	TotalHits int
}

// SearchIndexedManPages searches the index for the given query with fuzzy matching
func SearchIndexedManPages(query string) ([]SearchResult, error) {
	// Open existing index
	index, err := bleve.Open(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("index not found. Run 'lazyman -S' first to build the index")
		}
		return nil, fmt.Errorf("failed to open index at '%s': %w", indexPath, err)
	}
	defer index.Close()

	// Create a simple query string query (most flexible)
	searchQuery := bleve.NewQueryStringQuery(query)

	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = 100 // Increase to top 100 results for fuzzy matching
	searchRequest.Highlight = bleve.NewHighlight()
	searchRequest.Fields = []string{"Name", "Section", "Description", "Content"}

	// Execute search
	searchResults, err := index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search execution failed (query: '%s'): %w", query, err)
	}

	// Convert results
	results := make([]SearchResult, 0, len(searchResults.Hits))
	for _, hit := range searchResults.Hits {
		// Extract name and section from document ID "name(section)"
		docID := hit.ID
		name, section := parseDocID(docID)

		// Extract matching lines from content
		var matches []string
		if content, ok := hit.Fields["Content"].(string); ok {
			matches = extractMatchingLines(content, query, 3)
		}

		result := SearchResult{
			ManPage: ManPage{
				Name:        name,
				Section:     section,
				Description: getFieldString(hit.Fields, "Description"),
			},
			Matches:   matches,
			Score:     hit.Score,
			TotalHits: len(matches),
		}

		results = append(results, result)
	}

	return results, nil
}

// parseDocID extracts name and section from "name(section)" format
func parseDocID(docID string) (string, string) {
	if idx := strings.Index(docID, "("); idx != -1 {
		name := docID[:idx]
		section := strings.Trim(docID[idx:], "()")
		return name, section
	}
	return docID, ""
}

// getFieldString safely extracts a string field
func getFieldString(fields map[string]interface{}, key string) string {
	if val, ok := fields[key].(string); ok {
		return val
	}
	return ""
}

// extractMatchingLines finds lines containing the query and returns them with context
func extractMatchingLines(content, query string, contextLines int) []string {
	lines := strings.Split(content, "\n")
	queryLower := strings.ToLower(query)
	matches := []string{}
	matchedLines := make(map[int]bool)

	// Find all matching lines
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), queryLower) {
			matchedLines[i] = true
		}
	}

	// Extract matches with context
	for lineNum := range matchedLines {
		start := lineNum - contextLines
		if start < 0 {
			start = 0
		}
		end := lineNum + contextLines + 1
		if end > len(lines) {
			end = len(lines)
		}

		// Build context block
		contextBlock := strings.Join(lines[start:end], "\n")
		matches = append(matches, strings.TrimSpace(contextBlock))

		// Limit to top 3 matches per document
		if len(matches) >= 3 {
			break
		}
	}

	return matches
}

// IndexExists checks if the search index exists
func IndexExists() bool {
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetIndexPath returns the absolute path to the index
func GetIndexPath() (string, error) {
	absPath, err := filepath.Abs(indexPath)
	if err != nil {
		return "", err
	}
	return absPath, nil
}
