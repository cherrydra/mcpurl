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
  -h, --help                  Show this usage
  -I, --interactive           Start interactive mode
  -l, --log-level <level>     Set log level (debug, info, warn, error)
  -s, --silent                Silent mode
  -v, --version               Show version

Accepted <mcp_server> formats:
  https://example.com/mcp [options]
  stdio:///path/to/mcpserver [args]
  /path/to/mcpserver [args]

Currently supported transports:
  http(s) (streamable http)
  stdio   (standard input/output)
```
### List tools
```sh
mcpurl --tools docker run -i --rm mcp/filesystem .
```
### Call tool
```sh
mcpurl --tool list_directory -d '{"path": ""}' docker run -i --rm mcp/filesystem .
```
## Interactive mode
### Basic usage
```sh
mcpurl docker run -i --rm mcp/filesystem . -I
mcpurl> help
Usage:
  tools                   List tools
  prompts                 List prompts
  resources               List resources
  tool <name> [data]      Call tool
  prompt <name> [data]    Get prompt
  resource <name>         Read resource

  clear                   Clear the screen
  exit                    Exit the interactor
  help                    Show this help message
```
### Pipe operator
```sh
mcpurl docker run -i --rm mcp/filesystem . -I
mcpurl> tools | jq .name
read_file
read_multiple_files
write_file
edit_file
create_directory
list_directory
directory_tree
move_file
search_files
get_file_info
list_allowed_directories
```
## License
GNU General Public License v3.0
