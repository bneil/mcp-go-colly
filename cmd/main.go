package main

import (
	"fmt"
	"log"

	"github.com/bneil/mcp-go-colly/internal/crawler"
)

func main() {
	// Initialize the MCP crawler
	mcpCrawler, err := crawler.NewMCPCrawler(
		crawler.WithDefaultConfig(),
	)
	if err != nil {
		log.Fatalf("Failed to initialize MCP crawler: %v", err)
	}

	// Example crawl targets
	targets := []string{
		"https://example.com",
		"https://another-example.com",
	}

	// Crawl and extract data
	results, err := mcpCrawler.CrawlMultiple(targets)
	if err != nil {
		log.Fatalf("Crawling failed: %v", err)
	}

	// Process and print results
	fmt.Printf("Crawled %d pages\n", len(results))
	for _, result := range results {
		fmt.Printf("URL: %s, Title: %s, Content Length: %d\n", 
			result.URL, result.Title, len(result.Content))
	}
}
