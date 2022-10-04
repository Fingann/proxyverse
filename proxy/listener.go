package proxy

import (
	"context"
	"fmt"
	"net/http"
	"proxyverse/config"
	"proxyverse/log"
	"regexp"

	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)


type Listener struct {
	Addr string
	SSL  bool
	Domains []string
	server *http.Server
}

func NewListener(logger *zap.Logger, listener config.Listener) (*Listener,error) {
	domainProxies:= make([]*DomainProxy, 0, len(listener.Domains))
	domainRegexes := make([]*regexp.Regexp, 0, len(listener.Domains))
	for _, domain := range listener.Domains {
		domainProxy, err := NewDomainProxy(domain)
		if err != nil {
			return nil, fmt.Errorf("failed to create domain proxy: %w", err)
		}
		domainProxies = append(domainProxies, domainProxy)
		domainRegexes = append(domainRegexes, domainProxy.DomainRegex)
	}
	if len(domainProxies) == 0 {
		return nil, fmt.Errorf("no domains configured for listener on addr '%s'", listener.Addr)
	}

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", ProxyRequestHandler(logger, domainProxies))
	hostPolicy := func(ctx context.Context, host string) error {
		
		for _, domainRegex := range domainRegexes {
			if domainRegex.MatchString(host) {
				return nil
			}
		}
		return fmt.Errorf("acme/autocert: only %v host(s) allowed", domainRegexes)
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache("certs"),
		Client:    &acme.Client{DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory"},
	}
	

	return &Listener{
		Addr: listener.Addr,
		SSL:  listener.SSL,
		//Domains: domainRegexes,
		server: &http.Server{
			Addr:    listener.Addr,
			Handler: serverMux,
			TLSConfig: m.TLSConfig(),
		},
	},nil
}

func (l *Listener) Start() error {
	fmt.Println("Starting listener on addr: ", l.Addr)
	return l.server.ListenAndServe()
}


// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(logger *zap.Logger, serverProxies []*DomainProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, serverProxy := range serverProxies {
			// if the request matches the a
			if !serverProxy.Match(r) {
				continue
			}
			// check all rewrite rules
			for _, p := range serverProxy.Rewrites {
				// check if path matches
				if !p.Match(r) {
					continue
				}
				log.LogRequest(logger, log.NewRequestLog(r, nil))
				p.Handle(w, r)
				return
			}

		}
		log.LogRequest(logger, log.NewRequestLog(r, nil))
		// if no proxy matches, then return 404
		w.WriteHeader(http.StatusNotFound)
	}
}