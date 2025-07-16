package transport

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cherrydra/mcpurl/parser"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	ErrNoTransport = errors.New("no transport specified")
)

func Transport(args parser.Arguments) (mcp.Transport, error) {
	if len(args.TransportArgs) == 0 {
		return nil, ErrNoTransport
	}
	transportURL, err := url.Parse(args.TransportArgs[0])
	if err != nil {
		return nil, fmt.Errorf("parse transport url: %w", err)
	}
	switch transportURL.Scheme {
	case "stdio":
		cmd := cmp.Or(transportURL.Host, transportURL.Path)
		command := exec.Command(cmd, args.TransportArgs[1:]...)
		if !args.Silent {
			command.Stderr = os.Stderr
		}
		return mcp.NewCommandTransport(command), nil
	case "http", "https":
		return mcp.NewStreamableClientTransport(transportURL.String(), &mcp.StreamableClientTransportOptions{
			HTTPClient: &http.Client{Transport: &mcpurlRoundTripper{headers: args.Headers}},
		}), nil
	case "":
		switch filepath.Base(transportURL.Path) {
		case "mcp":
			return mcp.NewStreamableClientTransport(fmt.Sprintf("https://%s", transportURL.String()), &mcp.StreamableClientTransportOptions{
				HTTPClient: &http.Client{Transport: &mcpurlRoundTripper{headers: args.Headers}},
			}), nil
		case "sse":
			return mcp.NewSSEClientTransport(fmt.Sprintf("https://%s", transportURL.String()), &mcp.SSEClientTransportOptions{
				HTTPClient: &http.Client{Transport: &mcpurlRoundTripper{headers: args.Headers}},
			}), nil
		default:
			cmd := cmp.Or(transportURL.Host, transportURL.Path)
			command := exec.Command(cmd, args.TransportArgs[1:]...)
			if !args.Silent {
				command.Stderr = os.Stderr
			}
			return mcp.NewCommandTransport(command), nil
		}
	default:
		return nil, fmt.Errorf("unsupportd transport url scheme: %s", transportURL.Scheme)
	}
}
