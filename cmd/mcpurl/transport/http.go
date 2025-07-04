package transport

import (
	"net/http"
	"strings"
	"sync"
)

type mcpurlRoundTripper struct {
	headers []string

	parsedHeaders    http.Header
	parseHeadersOnce sync.Once
}

func (r *mcpurlRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.parseHeadersOnce.Do(func() {
		for _, header := range r.headers {
			kv := strings.Split(header, ":")
			if len(kv) != 2 {
				continue
			}
			r.parsedHeaders.Add(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
		}
	})
	for k, v := range r.parsedHeaders {
		if _, ok := req.Header[k]; ok {
			continue
		}
		for _, hv := range v {
			req.Header.Add(k, hv)
		}
	}
	return http.DefaultTransport.RoundTrip(req)
}
