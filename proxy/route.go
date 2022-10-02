package proxy

import (
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Header struct	{
	Name string `yaml:"name"`
	Value string `yaml:"value"`
}

type Location struct {
	Uri string `yaml:"uri"`
	Target   string `yaml:"target"`
	Headers []Header `yaml:"headers"`
}

type Server struct {
	Host     string     `yaml:"host"`
	Addr     string     `yaml:"addr"`
	Location []Location `yaml:"locations"`
}

func ToProxyHandler(route Server) (*ProxyServerHandler, error) {
	sr := "^" + route.Host + "$"
	serverRegex, err := regexp.Compile(sr)
	if err != nil {
		return nil, fmt.Errorf("failed to create host regex '%s': %w", sr, err)
	}
	locations, err := CreateLocationProxies(route)
	if err != nil {
		return nil, fmt.Errorf("failed to create location proxies: %w", err)
	}

	return &ProxyServerHandler{
		Server:      route,
		ServerRegex: serverRegex,
		Locations:   locations,
	}, nil
}

func CreateLocationProxies(route Server) ([]*LocationProxy, error) {
	locations := make([]*LocationProxy, 0, len(route.Location))
	for _, location := range route.Location {
		targetUrl, err := url.Parse(location.Target)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target host: %w", err)
		}
		proxy, err := NewProxy(targetUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to create new proxy: %w", err)
		}

		pr := "^" + location.Uri
		pathRegex, err := regexp.Compile(pr)
		if err != nil {
			return nil, fmt.Errorf("failed to create path regex '%s': %w", pr, err)
		}
	
		locations = append(locations, &LocationProxy{
			Headers: location.Headers,
			Proxy:    proxy,
			ProxyUrl: targetUrl,
			Uri:      location.Uri,
			UriRegex: pathRegex,
		})
	}
	return locations, nil
}

type LocationProxy struct {
	Headers   []Header
	Proxy    *httputil.ReverseProxy
	ProxyUrl *url.URL
	Uri      string
	UriRegex *regexp.Regexp
}

type ProxyServerHandler struct {
	Server      Server
	ServerRegex *regexp.Regexp
	Locations   []*LocationProxy
}

func AddrProxyMappings(routes []Server) map[string][]*ProxyServerHandler {
	addrProxyMap := make(map[string][]*ProxyServerHandler, len(routes))
	for _, route := range routes {
		proxyHandler, err := ToProxyHandler(route)
		if err != nil {
			log.Fatalf("failed to create proxy handler: %v", err)
		}
		addrProxyMap[route.Addr] = append(addrProxyMap[route.Addr], proxyHandler)
	}
	return addrProxyMap
}

func ReadRouteFile(filename string) ([]Server, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	routes := make([]Server, 0)
	err = yaml.NewDecoder(f).Decode(&routes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	return routes, nil
}
