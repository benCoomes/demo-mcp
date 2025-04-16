package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bencoomes/demo-mcp/pkg/weather"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Weather MCP Server",  // Server name
		"1.0.0",               // Server version
		server.WithLogging(),  // Enable logging
		server.WithRecovery(), // Enable panic recovery
	)

	// Create the get_weather tool
	weatherTool := mcp.NewTool(
		"get_weather", // Tool name
		mcp.WithDescription("Get current weather conditions for a specified location"),
		mcp.WithString(
			"location",
			mcp.Required(),
			mcp.Description("City or location name to get weather data for"),
		),
		mcp.WithString(
			"units",
			mcp.Description("Units system to use for measurements"),
			mcp.Enum("metric", "imperial"),
			mcp.DefaultString("metric"), // hallucinated mcp.DefaultValue
		),
	)

	// Add the get_weather tool handler
	s.AddTool(weatherTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Convert the parameters to the format expected by our weather package
		params := make(map[string]interface{})

		// Add location parameter
		if location, ok := request.Params.Arguments["location"].(string); ok {
			params["location"] = location
		}

		// Add units parameter if provided
		if units, ok := request.Params.Arguments["units"].(string); ok {
			params["units"] = units
		}

		// Call the existing weather function
		weatherData, err := weather.GetWeather(params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return the weather data as the result
		return mcp.NewToolResultText(fmt.Sprintf("%v", weatherData)), nil
	})

	fmt.Fprintln(os.Stderr, "Starting Weather MCP Server in stdio mode")
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
	return
}
