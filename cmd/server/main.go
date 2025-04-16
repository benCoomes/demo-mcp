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

// MCPVersion represents the MCP protocol version the server implements
const MCPVersion = "0.3.0"

// Initialize request/response types
type InitializeRequest struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    map[string]bool   `json:"capabilities"`
	Params          map[string]string `json:"params,omitempty"`
}

type InitializeResponse struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Vendor  string `json:"vendor,omitempty"`
}

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
	Result interface{} `json:"result,omitempty"`
	Error  *ToolError  `json:"error,omitempty"`
}

// ToolError represents a standardized error response
type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ToolDefinition represents metadata about a tool available through the MCP server
type ToolDefinition struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Parameters  ToolSchema   `json:"parameters"`
	Returns     *ReturnValue `json:"returns,omitempty"`
}

// ToolSchema defines a JSON schema for tool parameters
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]PropertyDef `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// PropertyDef defines a single property in a JSON schema
type PropertyDef struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ReturnValue defines the return value structure for a tool
type ReturnValue struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]PropertyDef `json:"properties,omitempty"`
}

// MCPIntrospectionResponse represents the response from the introspection endpoint
type MCPIntrospectionResponse struct {
	Version string           `json:"version"`
	Tools   []ToolDefinition `json:"tools"`
}

// Define available tools
var availableTools = []ToolDefinition{
	{
		Name:        "get_weather",
		Description: "Get current weather conditions for a specified location",
		Parameters: ToolSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"location": {
					Type:        "string",
					Description: "City or location name to get weather data for",
				},
				"units": {
					Type:        "string",
					Description: "Units system to use for measurements",
					Enum:        []string{"metric", "imperial"},
				},
			},
			Required: []string{"location"},
		},
		Returns: &ReturnValue{
			Type:        "object",
			Description: "Weather information for the requested location",
			Properties: map[string]PropertyDef{
				"location": {
					Type:        "string",
					Description: "The location name",
				},
				"temperature": {
					Type:        "number",
					Description: "Current temperature in requested units",
				},
				"condition": {
					Type:        "string",
					Description: "Weather condition description",
				},
				"humidity": {
					Type:        "number",
					Description: "Humidity percentage",
				},
				"wind_speed": {
					Type:        "number",
					Description: "Wind speed in km/h or mph depending on units",
				},
				"units": {
					Type:        "string",
					Description: "Units used (metric or imperial)",
				},
			},
		},
	},
}

// Server capabilities
var serverCapabilities = map[string]interface{}{
	"supportsIntrospection": true,
	"supportsStreaming":     false,
	"supportsCancellation":  false,
}

// Server info
var serverInfo = ServerInfo{
	Name:    "Weather MCP Server",
	Version: "1.0.0",
	Vendor:  "bencoomes",
}

// Track if the server has been initialized
var initialized = false

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

	// Register the MCP introspection endpoint
	r.HandleFunc("/mcp/introspection", handleIntrospectionRequest).Methods("GET")

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

		// Check if this is an initialization request
		var rawMessage map[string]interface{}
		if err := json.Unmarshal(request, &rawMessage); err != nil {
			logger.Errorf("Failed to decode request as JSON: %v", err)
			sendError(writer, "Invalid JSON format: "+err.Error())
			continue
		}

		// Check if this looks like an initialize request
		if _, ok := rawMessage["protocolVersion"]; ok {
			// Handle initialization
			var initReq InitializeRequest
			if err := json.Unmarshal(request, &initReq); err != nil {
				logger.Errorf("Failed to decode initialize request: %v", err)
				sendError(writer, "Invalid initialize request: "+err.Error())
				continue
			}
			handleInitialize(writer, initReq)
			continue
		}

		// Process the regular MCP request
		var req MCPRequest
		if err := json.Unmarshal(request, &req); err != nil {
			logger.Errorf("Failed to decode MCP request: %v", err)
			sendError(writer, "Invalid MCP request format: "+err.Error())
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
				Error: &ToolError{
					Code:    "internal_error",
					Message: errMsg,
				},
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
			result.Error = &ToolError{
				Code:    "tool_execution_error",
				Message: err.Error(),
			}
		} else {
			result.Result = weatherData
		}
	default:
		result.Error = &ToolError{
			Code:    "unknown_tool",
			Message: fmt.Sprintf("Unknown tool: %s", tool.Name),
		}
	}

	return result
}

func handleIntrospectionRequest(w http.ResponseWriter, r *http.Request) {
	response := MCPIntrospectionResponse{
		Version: MCPVersion,
		Tools:   availableTools,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode introspection response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func handleInitialize(w io.Writer, req InitializeRequest) {
	logger.Infof("Handling initialize request: %+v", req)

	if req.ProtocolVersion != MCPVersion {
		sendError(w, fmt.Sprintf("Unsupported protocol version: %s", req.ProtocolVersion))
		return
	}

	initialized = true

	resp := InitializeResponse{
		ProtocolVersion: MCPVersion,
		Capabilities:    serverCapabilities,
		ServerInfo:      serverInfo,
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("Failed to encode initialize response: %v", err)
		sendError(w, "Failed to encode initialize response: "+err.Error())
		return
	}

	respJSON = append(respJSON, '\n')
	if _, err := w.Write(respJSON); err != nil {
		logger.Errorf("Failed to write initialize response: %v", err)
	}
}
