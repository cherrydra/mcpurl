# mcpoly

An MCP aggregator combines multiple MCP servers.

## Install
```sh
go install github.com/cherrydra/mcpurl/cmd/mcpoly@latest
```

## Get Started
```sh
mcpurl mcpoly -I
```

## Config
> default config path is `$HOME/.config/mcpoly/mcp.json` 

```json
{
    "servers": {
        "mcp1": {
            "type": "stdio",
            "command": "command",
            "args": [],
            "env": {},
        },
        "mcp2": {
            "type": "http",
            "url": "https://example.com/mcp",
            "headers": {}
        },
        "mcp3": {
            "type": "sse",
            "url": "https://example.com/sse",
            "headers": {}
        }
    }
}
```
