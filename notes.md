Lets try to vibe-code an MCP server.

Here's my initial prompt:
```
I would like to build an MCP server that provides a tool named 'get_weather' that checks the current weather conditions in a given location. The server should be written in golang and runnable as a docker container.

To start, can you scaffold the project, stubbing out the method that returns weather details for now?
```

Agent mode with Claude 3.7 Sonnet did a good job getting _something_ that is containerized, runs, and returns hardcoded weather data.

Next, I wanted copilot to add the server to vs code: 
```
Can you add the weather mcp server to this vs code project? Use a docker image named 'weather-mcp'
```

This was not helpful - copilot only updated the name of the image to `weather-mcp` in `docker-compose.yml`.

However, providing a publicly available URL with documentation worked wonders!
```
There are instructions for setting up an mcp server in vs code here: https://github.com/github/github-mcp-server

Can you do that for this repository and the weather-mcp server?
```

This created the expected `.vscode/mcp.json` file with correct commands to call the `weather-mcp` server. It even added a 'stdio' mode to the server.
One oddity: copilot added an unsupported 'description' field in `.vscode/mcp.json`. I removed it an carried on.

I started the server in vs code (`> MCP: List Servers > weather > Start`):
```
2025-04-16 14:31:04.738 [info] Starting server weather
2025-04-16 14:31:04.738 [info] Connection state: Starting
2025-04-16 14:31:04.738 [info] Starting server from LocalProcess extension host
2025-04-16 14:31:04.742 [info] Connection state: Starting
2025-04-16 14:31:04.742 [info] Connection state: Running
2025-04-16 14:31:04.910 [warning] [server stderr] docker: Error response from daemon: failed to create task for container: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: exec: "stdio": executable file not found in $PATH: unknown
2025-04-16 14:31:04.911 [warning] [server stderr] 
2025-04-16 14:31:04.911 [warning] [server stderr] Run 'docker run --help' for more information
2025-04-16 14:31:04.912 [info] Connection state: Error Process exited with code 127
2025-04-16 14:32:34.515 [info] Starting server weather
2025-04-16 14:32:34.516 [info] Connection state: Starting
2025-04-16 14:32:34.529 [info] Starting server from LocalProcess extension host
2025-04-16 14:32:34.529 [info] Connection state: Starting
2025-04-16 14:32:34.529 [info] Connection state: Running
2025-04-16 14:32:34.680 [warning] [server stderr] {"level":"info","msg":"Starting in stdio mode","time":"2025-04-16T18:32:34Z"}
2025-04-16 14:32:39.533 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:32:44.530 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:32:49.531 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:32:54.531 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:32:59.530 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:04.530 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:09.529 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:14.533 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:19.531 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:24.530 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:29.530 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:34.529 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:39.530 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:44.531 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:33:48.790 [info] Stopping server weather
2025-04-16 14:33:48.794 [info] Connection state: Stopped
```

I don't know anything about the MCP protocol, but it doesn't look like the current code would implement it. For example, I don't see any way to list available tools and their descriptions. That seems to be what is happening in the logs - our server isn't responding to the expected 'initialize' request.

Some more prompts added more of the metadata and introspection I'd expect:
```
It seems like weather-mcp doesn't implement the mcp protocol. What methods must an MCP server provide? Can you update the server to handle all required methods?
```

Method descriptions, `mcp/introspection` endpoint...

```
Do MCP servers need to respond to initialize methods as well?
```

More changes, handling `protocolVersion` requests with a `handleInitialize` function.

But still no luck starting:
```
2025-04-16 14:50:44.332 [info] Starting server weather
2025-04-16 14:50:44.333 [info] Connection state: Starting
2025-04-16 14:50:44.339 [info] Starting server from LocalProcess extension host
2025-04-16 14:50:44.339 [info] Connection state: Starting
2025-04-16 14:50:44.340 [info] Connection state: Running
2025-04-16 14:50:44.520 [warning] [server stderr] {"level":"info","msg":"Starting in stdio mode","time":"2025-04-16T18:50:44Z"}
2025-04-16 14:50:49.342 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:50:54.342 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:50:59.340 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:51:04.341 [info] Waiting for server to respond to `initialize` request...
2025-04-16 14:51:08.090 [info] Stopping server weather
2025-04-16 14:51:08.099 [info] Connection state: Stopped
```