package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
)

// MCPTool represents a tool that can be exposed through the MCP protocol
type MCPTool struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, request CallToolRequest) (*CallToolResult, error)
	Parameters  map[string]ToolParameter
}

// ToolParameter defines the structure of a tool's input parameter
type ToolParameter struct {
	Type        string
	Description string
	Required    bool
	Enum        []string
}

// CallToolRequest represents the input for a tool call
type CallToolRequest struct {
	Params struct {
		Arguments map[string]interface{}
	}
}

// CallToolResult represents the output of a tool call
type CallToolResult struct {
	Type    string
	Content interface{}
	Error   error
}

// MCPServer manages MCP protocol interactions
type MCPServer struct {
	Name    string
	Version string
	tools   map[string]*MCPTool
	mutex   sync.RWMutex
}

// NewMCPServer creates a new MCP server
func NewMCPServer(name, version string) *MCPServer {
	return &MCPServer{
		Name:    name,
		Version: version,
		tools:   make(map[string]*MCPTool),
	}
}

// AddTool registers a new tool with the MCP server
func (s *MCPServer) AddTool(tool *MCPTool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.tools[tool.Name]; exists {
		return fmt.Errorf("tool with name %s already exists", tool.Name)
	}

	s.tools[tool.Name] = tool
	return nil
}

// NewToolResultText creates a text-based tool result
func NewToolResultText(content string) *CallToolResult {
	return &CallToolResult{
		Type:    "text",
		Content: content,
	}
}

// NewToolResultJSON creates a JSON-based tool result
func NewToolResultJSON(content interface{}) *CallToolResult {
	return &CallToolResult{
		Type:    "json",
		Content: content,
	}
}

// NewToolResultError creates an error result
func NewToolResultError(message string) *CallToolResult {
	return &CallToolResult{
		Type:  "error",
		Error: fmt.Errorf(message),
	}
}

// NewCrawlerTool creates a tool for web crawling that's compatible with the external MCP library
func NewCrawlerTool() mcp.Tool {
	return mcp.NewTool("web_crawler",
		mcp.WithDescription("Crawl and extract information from websites"),
		mcp.WithString("urls", mcp.Required(), mcp.Description("List of URLs to crawl")),
		mcp.WithNumber("max_depth", mcp.Description("Maximum crawl depth")),
	)
}

// ExtractDomainsFromURLs extracts unique domains from a list of URLs
func ExtractDomainsFromURLs(urls []string) []string {
	domainMap := make(map[string]bool)
	var domains []string

	for _, urlStr := range urls {
		u, err := url.Parse(urlStr)
		if err != nil {
			continue
		}

		domain := u.Hostname()
		if domain != "" && !domainMap[domain] {
			domainMap[domain] = true
			domains = append(domains, domain)
		}
	}

	return domains
}

// Serialize converts the tool result to a JSON-friendly format
func (r *CallToolResult) Serialize() ([]byte, error) {
	if r.Error != nil {
		return json.Marshal(map[string]interface{}{
			"type":  "error",
			"error": r.Error.Error(),
		})
	}

	return json.Marshal(map[string]interface{}{
		"type":    r.Type,
		"content": r.Content,
	})
}
