# Command line tool `mcpurl`
Like cURL but for MCP.

## Install
```sh
go install github.com/nasuci/mcpurl/cmd/mcpurl@latest
```

## Usage
### list tools
```sh
mcpurl --tools docker run -i --rm mcp/filesystem .
```
### call tool
```sh
mcpurl --tool list_directory -d '{"path": ""}' docker run -i --rm mcp/filesystem .
```
## License
GNU General Public License v3.0
