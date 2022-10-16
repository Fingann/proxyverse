package rewriter

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxyverse/config"
	"proxyverse/log"
	"regexp"
	"strings"
)

type Rewriter interface {
	Match(req *http.Request) bool
	Handle(w http.ResponseWriter, req *http.Request)
}

type PathRewriter struct {
	Proxy     *httputil.ReverseProxy
	path      string
	pathRegex *regexp.Regexp
}

func (p *PathRewriter) Handle(w http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(p.path, "/") {
		if match := p.pathRegex.FindString(req.URL.Path); match != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, match)
		}
	}
	p.Proxy.ServeHTTP(w, req)
}

func (p *PathRewriter) Match(req *http.Request) bool {
	return p.pathRegex.MatchString(req.URL.Path)
}

// NewPathRewriter takes target host and creates a reverse proxy
func NewPathRewriter(rewrite config.Rewrite) (*PathRewriter, error) {
	targetUrl, err := url.Parse(rewrite.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target host: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		modifyHeaders(req, rewrite.Headers)
	}
	proxy.ModifyResponse = modifyResponse()
	proxy.ErrorHandler = errorHandler()

	pathRegex, err := regexp.Compile("^" + rewrite.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to create path regex: %w", err)
	}

	return &PathRewriter{Proxy: proxy, pathRegex: pathRegex, path: rewrite.Path}, nil
}

func modifyHeaders(req *http.Request, headers []config.Header) {
	req.Header.Set("X-Proxy", "Proxyverse")
	for _, header := range headers {
		buf := &bytes.Buffer{}
		k, _ := template.New("header").Parse(header.Value)
		k.Execute(buf, log.FromHttpRequest(req))

		req.Header.Add(header.Key, buf.String())
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
