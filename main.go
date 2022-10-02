package main

import (
	"backslide/proxy"
	"bytes"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "",
		LevelKey:       "",
		NameKey:        "request",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), os.Stdout, zap.DebugLevel)
	requestLogger := zap.New(core).WithOptions(zap.WithCaller(false))
	//logS := logger.Sugar()

	routes, err := proxy.ReadRouteFile("routes.yaml")
	if err != nil {
		log.Sugar().Fatalf(err.Error(),
			"path", "routes.yaml")
	}
	wg := &sync.WaitGroup{}

	addrProxyMap := proxy.AddrProxyMappings(routes)
	for addr, proxyHandler := range addrProxyMap {
		wg.Add(1)
		go func(addr string, proxyHandler []*proxy.ProxyServerHandler) {
			defer wg.Done()
			log.Info("Listening:", zap.String("addr", addr))
			serverMux := http.NewServeMux()
			serverMux.HandleFunc("/", ProxyRequestHandler(requestLogger, proxyHandler))
			srv := http.Server{
				Addr:    addr,
				Handler: serverMux,
			}
			log.Sugar().Fatal(srv.ListenAndServe())
		}(addr, proxyHandler)
	}

	wg.Wait()
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(log *zap.Logger, proxyHandlers []*proxy.ProxyServerHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		match := &proxy.ProxyMatch{}
		for _, proxyHandler := range proxyHandlers {

			// check if host matches
			if !proxyHandler.ServerRegex.MatchString(r.Host) {
				continue
			}
			match.Host = r.Host
			

			for _, location := range proxyHandler.Locations {
				// check if path matches
				if !location.UriRegex.MatchString(r.URL.RequestURI()) {
					continue
				}
				match.Match=location.Uri

				if strings.HasSuffix(location.Uri, "/") {
					r.URL.Path = strings.TrimPrefix(r.URL.Path, location.Uri)
				}
				match.Uri=r.URL.Path
				match.Addr=proxyHandler.Server.Addr
				logReq:= proxy.NewRequestLog(r, match)
				proxy.LogRequest(log, logReq)

				for _, v := range location.Headers {
					buf := &bytes.Buffer{}
					k,_ := template.New("header").Parse(v.Value)
					k.Execute(buf, logReq.OriginalRequest)
					r.Header.Add(v.Name, buf.String())
				}
					
				


				location.Proxy.ServeHTTP(w, r)
				return
			}

		}
		proxy.LogRequest(log, proxy.NewRequestLog(r, match))
		// if no proxy matches, then return 404
		w.WriteHeader(http.StatusNotFound)
	}
}

/*
routes := []proxy.Server{
		{
			Host: "example.com",
			Location: []proxy.Location{
				{From: "/test", To: "http://localhost:4444"},
				{From: "/fest", To: "http://localhost:4444/"},
				{From: "/best", To: "http://localhost:4444/test"},
				{From: "/gest/", To: "http://localhost:4444/test"},
				{From: "/pest", To: "http://localhost:4444/test/"},
				{From: "/jest/", To: "http://localhost:4444/test/"},
			}},
		{
			Host: "example.com",
			Location: []proxy.Location{
				{From: "/t[de]st", To: "http://localhost:4444"},
			}},
		{
			Host: "example.com",
			Location: []proxy.Location{
				{From: "/t[de]st", To: "http://localhost:4444"},
			},
		},
	}

*/
