# Command line tool `mcpurl`
Like cURL but for MCP.

## Install
```sh
go install github.com/nasuci/mcpurl/cmd/mcpurl@latest
```

## Usage
```
Usage:
  mcpurl <options> <mcp_server>

Accepted <options>:
  -t, --tools             list tools
  -p, --prompts           list prompts
  -r, --resources         list resources
  -T, --tool <string>     call tool
  -P, --prompt <string>   get prompt
  -R, --resource <string> read resource

  -h, --help              show this usage

Currently supported transport:
  stdio (standard input/output)

Accepted <mcp_server> formats:
  stdio:///path/to/mcpserver [args]   # Explicit stdio scheme
  /path/to/mcpserver [args]           # Implicit stdio scheme
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
