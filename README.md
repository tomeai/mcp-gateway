## MCP-Gateway

## Remote

1. streamable http

```
{
  "mcpServers": {
    "mcpjungle": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://registry.wemcp.cn/mcp/{mcp_server_name}",
        "--allow-http"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>"
      }
    }
  }
}
```

2. sse

```
{
  "mcpServers": {
    "mcpjungle": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://registry.wemcp.cn/mcp/{mcp_server_name}/sse",
        "--allow-http"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>"
      }
    }
  }
}
```