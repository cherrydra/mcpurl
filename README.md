# Command line tool `mcpurl`
Like cURL but for MCP.

## Install
```sh
go install github.com/cherrydra/mcpurl/cmd/mcpurl@latest
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
  -K, --llm-api-key <key>     API key for authenticating with the LLM
  -L, --llm-base-url <url>    Base URL of the LLM service
  -M, --llm-name <name>       Name of the LLM model to use
  -l, --log-level <level>     Set log level (debug, info, warn, error)
  -m, --msg <message>         Talk to LLM
  -s, --silent                Silent mode
  -v, --version               Show version

Accepted <mcp_server> formats:
  https://example.com/mcp [options]
  stdio:///path/to/mcpserver [args] (or simply /path/to/mcpserver [args])

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
Available Commands:
  tools                           List tools
  prompts                         List prompts
  resources                       List resources
  tool <name> [options]           Call tool
  prompt <name> [options]         Get prompt
  resource <name>                 Read resource
  ctx <subcmd>                    LLM context operations
  msg <message>                   Talk to LLM
  connect <mcp_server> [options]  Connect to server
  disconnect                      Disconnect from server
  status                          Show connection info

System Commands:
  cat <file>                      Read file
  cd [dir]                        Change working directory
  clear                           Clear the screen
  export [name=value ...]         Set/get environment variables
  exit                            Exit the interactor
  help                            Show this help message
  ls [dir]                        List files in directory
  pwd                             Print working directory
  version                         Show version information

Supports command pipelining and stdout redirection:
  tools | jq .name > tools.txt
```
### Pipe / stdout redirect operator
```sh
mcpurl docker run -i --rm mcp/filesystem . -I
mcpurl> tools | jq -r .name > tools.txt
mcpurl> cat tools.txt
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
## Play with LLM
```
mcpurl -I -L <base_url> -K <api_key> -M <model>
```
Read [mcpurl.io/llm](https://mcpurl.io/llm) for details.
## Play with mcpoly
```sh
mcpurl mcpoly -I
```
Read [mcpurl.io/mcpoly](https://mcpurl.io/mcpoly) for details.

## License
[GNU General Public License v3.0](https://github.com/cherrydra/mcpurl/blob/main/LICENSE)
