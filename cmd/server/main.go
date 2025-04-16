package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bencoomes/demo-mcp/pkg/weather"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// MCPRequest represents an incoming MCP request
type MCPRequest struct {
	Tools []ToolCall `json:"tools"`
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// MCPResponse represents the response to an MCP request
type MCPResponse struct {
	Results []ToolResult `json:"results"`
}

// ToolResult represents the result of a tool invocation
type ToolResult struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stderr) // Use stderr for logs to avoid interfering with stdio communication

	// Check for stdio mode
	if len(os.Args) > 1 && os.Args[1] == "stdio" {
		logger.Info("Starting in stdio mode")
		handleStdioMode()
		return
	}

	// Default to HTTP mode
	logger.Info("Starting in HTTP mode")
	startHTTPServer()
}

func startHTTPServer() {
	// Create a new router
	r := mux.NewRouter()

	// Register the MCP endpoint
	r.HandleFunc("/mcp", handleMCPRequest).Methods("POST")

	// Register a health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}).Methods("GET")

	// Get the port from the environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	logger.Infof("Starting MCP server on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}

func handleStdioMode() {
	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	for {
		// Read request
		request, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("Stdin closed, exiting")
				return
			}
			logger.Errorf("Error reading from stdin: %v", err)
			continue
		}

		// Process the request
		var req MCPRequest
		if err := json.Unmarshal(request, &req); err != nil {
			logger.Errorf("Failed to decode request: %v", err)
			sendError(writer, "Invalid request format: "+err.Error())
			continue
		}

		// Process each tool call
		var resp MCPResponse
		for _, tool := range req.Tools {
			result := processTool(tool)
			resp.Results = append(resp.Results, result)
		}

		// Send the response
		responseJSON, err := json.Marshal(resp)
		if err != nil {
			logger.Errorf("Failed to encode response: %v", err)
			sendError(writer, "Failed to encode response: "+err.Error())
			continue
		}

		// Add a newline to the response
		responseJSON = append(responseJSON, '\n')
		if _, err := writer.Write(responseJSON); err != nil {
			logger.Errorf("Failed to write response: %v", err)
			return
		}
	}
}

func sendError(w io.Writer, errMsg string) {
	resp := MCPResponse{
		Results: []ToolResult{
			{
				Error: errMsg,
			},
		},
	}
	respJSON, _ := json.Marshal(resp)
	respJSON = append(respJSON, '\n')
	w.Write(respJSON)
}

func handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		logger.Errorf("Failed to decode request: %v", err)
		return
	}

	// Process each tool call
	var resp MCPResponse
	for _, tool := range req.Tools {
		result := processTool(tool)
		resp.Results = append(resp.Results, result)
	}

	// Set the content type and write the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}

func processTool(tool ToolCall) ToolResult {
	logger.Infof("Processing tool call: %s", tool.Name)

	result := ToolResult{ID: tool.ID}

	switch strings.ToLower(tool.Name) {
	case "get_weather":
		weatherData, err := weather.GetWeather(tool.Params)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Result = weatherData
		}
	default:
		result.Error = fmt.Sprintf("Unknown tool: %s", tool.Name)
	}

	return result
}
