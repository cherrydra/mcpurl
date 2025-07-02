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
  -T, --tools             list tools
  -P, --prompts           list prompts
  -R, --resources         list resources
  -t, --tool <string>     call tool
  -p, --prompt <string>   get prompt
  -r, --resource <string> read resource
  -d, --data <string>     send json data to server

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
