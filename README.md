# Weather MCP Server

A Model Context Protocol (MCP) server implemented in Go that provides a `get_weather` tool for checking current weather conditions in a given location.

## Features

- MCP-compliant service with JSON API
- `get_weather` tool that returns weather data for a location
- Dockerized for easy deployment and running

## Prerequisites

- Docker and Docker Compose (for containerized deployment)
- Go 1.21+ (for local development)

## Running the Server

### Using Docker Compose (Recommended)

```bash
docker-compose up --build
```

This will build and start the MCP server, making it available at http://localhost:8080.

### Building and Running Locally

1. Install dependencies:
   ```bash
   go mod download
   go mod tidy
   ```

2. Build the application:
   ```bash
   go build -o server ./cmd/server
   ```

3. Run the server:
   ```bash
   ./server
   ```

## API Usage

The server exposes an MCP-compliant endpoint at `/mcp` that accepts POST requests with the following format:

```json
{
  "tools": [
    {
      "id": "unique-request-id",
      "name": "get_weather",
      "params": {
        "location": "New York, NY",
        "units": "metric"
      }
    }
  ]
}
```

### get_weather Tool

Parameters:
- `location` (required): City or location name
- `units` (optional): "metric" (default) or "imperial"

Example Response:

```json
{
  "results": [
    {
      "id": "unique-request-id",
      "result": {
        "location": "New York, NY",
        "temperature": 22.5,
        "condition": "Partly Cloudy",
        "humidity": 65,
        "wind_speed": 10.5,
        "units": "metric"
      }
    }
  ]
}
```

## Health Check

The server provides a health check endpoint at `/health` that returns "OK" when the server is running properly.

## Development

Currently, the weather service returns stubbed data. To implement actual weather data, update the `GetWeather` function in `pkg/weather/weather.go` to integrate with a weather API.

## License

MIT