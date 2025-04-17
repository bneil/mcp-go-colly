package crawler

import (
	"fmt"
	"log"
	"sync"

	"github.com/gocolly/colly/v2"
)

// CrawlResult represents the data extracted from a single webpage
type CrawlResult struct {
	URL     string
	Title   string
	Content string
	Error   error
}

// MCPCrawler manages web crawling with MCP integration
type MCPCrawler struct {
	collector *colly.Collector
	config    *CrawlerConfig
	results   []CrawlResult
	mutex     sync.Mutex
}

// CrawlerConfig allows customization of crawler behavior
type CrawlerConfig struct {
	MaxDepth       int
	AllowedDomains []string
	UserAgent      string
	Timeout        int
}

// Option type for configuration
type Option func(*CrawlerConfig)

// WithDefaultConfig provides a standard configuration
func WithDefaultConfig() Option {
	return func(c *CrawlerConfig) {
		c.MaxDepth = 2
		c.UserAgent = "MCPCrawler/1.0"
		c.Timeout = 10
	}
}

// WithMaxDepth sets the maximum crawl depth
func WithMaxDepth(depth int) Option {
	return func(c *CrawlerConfig) {
		c.MaxDepth = depth
	}
}

// WithAllowedDomains restricts crawling to specific domains
func WithAllowedDomains(domains []string) Option {
	return func(c *CrawlerConfig) {
		c.AllowedDomains = domains
	}
}

// NewMCPCrawler initializes a new crawler with given options
func NewMCPCrawler(opts ...Option) (*MCPCrawler, error) {
	// Default configuration
	config := &CrawlerConfig{
		MaxDepth:       2,
		UserAgent:      "MCPCrawler/1.0",
		Timeout:        10,
		AllowedDomains: []string{},
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Create collector with configuration
	c := colly.NewCollector(
		colly.MaxDepth(config.MaxDepth),
		colly.UserAgent(config.UserAgent),
	)

	// Add allowed domains if specified
	if len(config.AllowedDomains) > 0 {
		c.AllowedDomains = config.AllowedDomains
	}

	// Initialize crawler
	crawler := &MCPCrawler{
		collector: c,
		config:    config,
		results:   []CrawlResult{},
	}

	// Setup event handlers
	crawler.setupEventHandlers()

	return crawler, nil
}

// setupEventHandlers configures Colly collector event handlers
func (mc *MCPCrawler) setupEventHandlers() {
	// Handle successful page visits
	mc.collector.OnHTML("html", func(e *colly.HTMLElement) {
		result := CrawlResult{
			URL:     e.Request.URL.String(),
			Title:   e.DOM.Find("title").Text(),
			Content: e.Text,
		}

		mc.mutex.Lock()
		mc.results = append(mc.results, result)
		mc.mutex.Unlock()
	})

	// Handle errors
	mc.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("Error on %s: %v", r.Request.URL, err)
		
		result := CrawlResult{
			URL:   r.Request.URL.String(),
			Error: err,
		}

		mc.mutex.Lock()
		mc.results = append(mc.results, result)
		mc.mutex.Unlock()
	})
}

// CrawlMultiple crawls multiple target URLs concurrently
func (mc *MCPCrawler) CrawlMultiple(urls []string) ([]CrawlResult, error) {
	var wg sync.WaitGroup
	
	for _, url := range urls {
		wg.Add(1)
		go func(targetURL string) {
			defer wg.Done()
			err := mc.collector.Visit(targetURL)
			if err != nil {
				log.Printf("Failed to crawl %s: %v", targetURL, err)
			}
		}(url)
	}

	wg.Wait()

	// Clone results to prevent race conditions
	mc.mutex.Lock()
	resultsCopy := make([]CrawlResult, len(mc.results))
	copy(resultsCopy, mc.results)
	mc.mutex.Unlock()

	return resultsCopy, nil
}
