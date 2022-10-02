package main

import (
	//"bytes"
	"flag"
	"io"
	"net/http"
	"os"
	"proxyverse/config"
	"proxyverse/proxy"
	"sync"

	//"text/template"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// user can specify cli args to override config file
	var rewritesFile string
	flag.StringVar(&rewritesFile, "r", "rewrites.yaml", "rewrites file")
	flag.Parse()

	log, _ := zap.NewProduction()
	defer log.Sync()

	requestLogger := createRequestLogger()
	//logS := logger.Sugar()

	servers, err := config.ParseRewritesFile(rewritesFile)
	if err != nil {
		log.Sugar().Fatalf(err.Error(), "path", rewritesFile)
	}

	// group all proxies by address
	addrMap:= make(map[string][]*proxy.ServerProxy)
	for _, server := range servers {
		serverProxy, err := proxy.NewServerProxy(server)
		if err != nil {
			log.Sugar().Fatalf(err.Error(), "server", server)
		}
		addrMap[serverProxy.Addr] = append(addrMap[serverProxy.Addr], serverProxy)
	}

	wg := &sync.WaitGroup{}	
	// start a server for each address
	for addr, serverProxy := range addrMap {
		wg.Add(1)

		go func(addr string,serverProxy []*proxy.ServerProxy) {
			defer wg.Done()
		log.Sugar().Infow("Listening on", "addr",addr)

		serverMux := http.NewServeMux()
			serverMux.HandleFunc("/", ProxyRequestHandler(requestLogger, serverProxy))
			srv := http.Server{
				Addr:    addr,
				Handler: serverMux,
			}
			log.Sugar().Fatal(srv.ListenAndServe())
		}(addr, serverProxy)
	}

	/*
	addrProxyMap := proxy.AddrProxyMappings(servers)
	for addr, proxyHandler := range addrProxyMap {
		wg.Add(1)
		go func(addr string, proxyHandler []*proxy.ServerProxy) {
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
	*/

	wg.Wait()
}

func createRequestLogger() *zap.Logger {

	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "",
		LevelKey:       "",
		NameKey:        "request",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	// lumberjack.Logger is already safe for concurrent use, so we don't need to
	// lock it.
	w := zapcore.AddSync(
		io.MultiWriter(
			&lumberjack.Logger{
				Filename:   "./logs/access.log",
				MaxSize:    500, // megabytes
				MaxBackups: 3,
				MaxAge:     28, // days
			},
			os.Stdout,
		),
	)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		w,
		zap.DebugLevel,
	)

	requestLogger := zap.New(core).WithOptions(zap.WithCaller(false))
	return requestLogger
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(log *zap.Logger, serverProxies []*proxy.ServerProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, serverProxy := range serverProxies {
			// if the request matches the a
			if !serverProxy.Match(r) {
				continue
			}
			// check all rewrite rules
			for _, rewrite := range serverProxy.Rewrites {
				// check if path matches
				if !rewrite.Match(r) {
					continue
				}
				proxy.LogRequest(log, proxy.NewRequestLog(r, nil))
				rewrite.Proxy.ServeHTTP(w, r)
				return
			}

		}
		proxy.LogRequest(log, proxy.NewRequestLog(r, nil))
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
