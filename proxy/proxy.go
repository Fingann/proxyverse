package proxy

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxyverse/config"
	"regexp"
	"strings"
)

type Proxy struct {
	Proxy   *httputil.ReverseProxy
    pathRegex *regexp.Regexp
	//Rewrite config.Rewrite
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.Proxy.ServeHTTP(w, req)
}

func (p *Proxy) Match(req *http.Request) bool{
    return p.pathRegex.MatchString(req.URL.Path)
}

// NewProxy takes target host and creates a reverse proxy
func NewProxy(rewrite config.Rewrite) (*Proxy, error) {
    targetUrl, err := url.Parse(rewrite.Target)
    if err != nil {
        return nil, fmt.Errorf("failed to parse target host: %w", err)
    }
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		modifyHeaders(req, rewrite.Headers)
		modifyPath(req, rewrite.Target)
	}
	proxy.ModifyResponse = modifyResponse()
	proxy.ErrorHandler = errorHandler()

    pathRegex, err := regexp.Compile("^" + rewrite.Path)
    if err != nil {
        return nil, fmt.Errorf("failed to create path regex: %w", err)
    }

	return &Proxy{Proxy: proxy,pathRegex: pathRegex}, nil
}

func modifyHeaders(req *http.Request, headers []config.Header) {
	req.Header.Set("X-Proxy", "Proxyverse")
	for _, header := range headers {
		buf := &bytes.Buffer{}
		k, _ := template.New("header").Parse(header.Value)
		k.Execute(buf, FromHttpRequest(req))

		req.Header.Add(header.Key, buf.String())
	}
}
func modifyPath(req *http.Request, path string) {
	if strings.HasSuffix(path, "/") {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, path)
	}
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
	}
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		return nil
	}
}
