package proxy

import (
	"fmt"
	"net/http"
	"proxyverse/config"
	"proxyverse/proxy/rewriter"
	"regexp"
)

func NewDomainProxy(domain config.Domain) (*DomainProxy, error) {
	domainRegex, err := regexp.Compile("^" + domain.Name + "$")
	if err != nil {
		return nil, fmt.Errorf("failed to create host regex: %w", err)
	}
	proxies := make([]rewriter.Rewriter, 0, len(domain.Rewrites))
	for _, rewrite := range domain.Rewrites {

		if rewrite.Redirect {
			redirect, err := rewriter.NewRedirectRewriter(rewrite)
			if err != nil {
				return nil, fmt.Errorf("failed to create redirect proxy: %w", err)
			}
			proxies = append(proxies, redirect)
			continue
		}

		proxy, err := rewriter.NewPathRewriter(rewrite)
		if err != nil {
			return nil, fmt.Errorf("failed to create proxy: %w", err)
		}
		proxies = append(proxies, proxy)

	}

	return &DomainProxy{
		DomainName:  domain.Name,
		DomainRegex: domainRegex,
		Rewrites:    proxies,
	}, nil
}

func (s *DomainProxy) Match(req *http.Request) bool {
	return s.DomainRegex.MatchString(req.Host)
}

type DomainProxy struct {
	DomainName  string
	SSL         bool
	DomainRegex *regexp.Regexp
	Rewrites    []rewriter.Rewriter
}
