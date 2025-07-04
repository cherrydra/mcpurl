# Command line tool `mcpurl`
Like cURL but for MCP.

## Install
```sh
go install github.com/nasuci/mcpurl/cmd/mcpurl@latest
```

## Usage
```sh
Usage:
  mcpurl <options> <mcp_server>

Accepted <options>:
  -T, --tools                 List tools
  -P, --prompts               List prompts
  -R, --resources             List resources
  -t, --tool <string>         Call tool
  -p, --prompt <string>       Get prompt
  -r, --resource <string>     Read resource
  -d, --data <string/@file>   Send json data to server
  -H, --header <header/@file> Pass custom header(s) to server
  -s, --silent                Silent mode

  -h, --help                  Show this usage
  -v, --version               Show version

Accepted <mcp_server> formats:
  https://example.com/mcp [options]
  stdio:///path/to/mcpserver [args]
  /path/to/mcpserver [args]

Currently supported transports:
  http(s) (streamable http)
  stdio   (standard input/output)
```
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
