package rewriter

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"proxyverse/config"
	"proxyverse/log"
	"regexp"
)

type RedirectRewriter struct {
	target    *url.URL
	pathRegex *regexp.Regexp
	headers   []config.Header
	//Rewrite config.Rewrite
}

// NewProxy takes target host and creates a reverse proxy
func NewRedirectRewriter(rewrite config.Rewrite) (*RedirectRewriter, error) {
	targetUrl, err := url.Parse(rewrite.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target host: %w", err)
	}

	pathRegex, err := regexp.Compile("^" + rewrite.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to create path regex: %w", err)
	}

	return &RedirectRewriter{target: targetUrl, pathRegex: pathRegex, headers: rewrite.Headers}, nil
}

func (p *RedirectRewriter) Handle(w http.ResponseWriter, req *http.Request) {
	req.Header.Set("X-Proxy", "Proxyverse")
	for _, header := range p.headers {
		buf := &bytes.Buffer{}
		k, _ := template.New("header").Parse(header.Value)
		k.Execute(buf, log.FromHttpRequest(req))

		req.Header.Add(header.Key, buf.String())
	}
	redirect := p.target.ResolveReference(req.URL)
	http.Redirect(w, req, redirect.String(), http.StatusMovedPermanently)
}

func (p *RedirectRewriter) Match(req *http.Request) bool {
	return p.pathRegex.MatchString(req.URL.Path)
}
