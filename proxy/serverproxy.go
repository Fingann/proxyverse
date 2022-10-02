package proxy

import (
	"fmt"
	"net/http"
	"proxyverse/config"
	"regexp"
)

func NewServerProxy(server config.Server) (*ServerProxy, error) {
	serverRegex, err := regexp.Compile("^" + server.Host + "$")
	if err != nil {
		return nil, fmt.Errorf("failed to create host regex: %w", err)
	}
	proxies := make([]*Proxy, 0, len(server.Rewrites))
	for _, rewrite := range server.Rewrites {
		proxy, err := NewProxy(rewrite)
		if err != nil {
			return nil, fmt.Errorf("failed to create proxy: %w", err)
		}
		proxies = append(proxies, proxy)

	}

	return &ServerProxy{
		Host:    server.Host,
		Addr:    server.Addr,
		hostRegex: serverRegex,
		Rewrites:    proxies,
	}, nil
}

func (s *ServerProxy) Match(req *http.Request) bool {
	return s.hostRegex.MatchString(req.Host)
}

type ServerProxy struct {
	Host        string
	Addr        string
	hostRegex *regexp.Regexp
	Rewrites    []*Proxy
}
