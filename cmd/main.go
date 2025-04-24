package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bneil/mcp-go-colly/internal/crawler"
	localmcp "github.com/bneil/mcp-go-colly/internal/mcp"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a context with cancellation - we'll use this for crawling operations
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Handle graceful shutdown
	go handleInterrupt(cancel)

	// Create MCP server
	s := server.NewMCPServer(
		"WebCrawlerServer",
		"1.0.0",
	)

	// Create web crawler tool using our local implementation
	crawlerTool := localmcp.NewCrawlerTool()

	// Add crawler tool handler
	s.AddTool(crawlerTool, server.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Convert to our local format for processing
		localReq := localmcp.CallToolRequest{
			Params: struct{ Arguments map[string]interface{} }{
				Arguments: req.Params.Arguments,
			},
		}

		// Call our handler
		result, err := crawlerHandler(ctx, localReq)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: err.Error(),
					},
				},
			}, nil
		}

		// Convert back to MCP format
		if result.Error != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: result.Error.Error(),
					},
				},
			}, nil
		}

		// Return result as JSON content
		returnContent := []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v", result.Content),
			},
		}
		return &mcp.CallToolResult{Content: returnContent}, nil
	}))

	// Start the server with the correct function signature
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func crawlerHandler(ctx context.Context, request localmcp.CallToolRequest) (*localmcp.CallToolResult, error) {
	// Extract parameters
	urlsInterface := request.Params.Arguments["urls"]
	var urls []string

	// Handle both single URL and array of URLs
	switch v := urlsInterface.(type) {
	case string:
		urls = []string{v}
	case []interface{}:
		urls = make([]string, len(v))
		for i, u := range v {
			urls[i] = u.(string)
		}
	default:
		return localmcp.NewToolResultError("Invalid URLs parameter"), nil
	}

	// Default max depth to 2
	maxDepth := 2
	if depth, ok := request.Params.Arguments["max_depth"].(float64); ok {
		maxDepth = int(depth)
	}

	// Create crawler with configuration
	mcpCrawler, err := crawler.NewMCPCrawler(ctx,
		func(c *crawler.CrawlerConfig) {
			c.MaxDepth = maxDepth
			c.AllowedDomains = localmcp.ExtractDomainsFromURLs(urls)
		},
	)
	if err != nil {
		return localmcp.NewToolResultError(fmt.Sprintf("Failed to create crawler: %v", err)), nil
	}

	// Perform crawling
	results, err := mcpCrawler.CrawlMultiple(ctx, urls)
	if err != nil {
		return localmcp.NewToolResultError(fmt.Sprintf("Crawling failed: %v", err)), nil
	}

	// Return crawled content
	return localmcp.NewToolResultJSON(results), nil
}

// handleInterrupt manages graceful shutdown
func handleInterrupt(cancel context.CancelFunc) {
	// Create a channel to receive interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,  // Interrupt from keyboard (Ctrl+C)
		syscall.SIGTERM, // Termination signal
	)

	// Wait for an interrupt signal
	<-sigChan
	log.Println("Received an interrupt, stopping all operations...")

	// Cancel ongoing operations
	cancel()
}
