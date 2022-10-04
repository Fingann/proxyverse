package main

import (
	//"bytes"
	"flag"
	"io"
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
	flag.StringVar(&rewritesFile, "r", "yaml2.yaml", "rewrites file")
	flag.Parse()

	log, _ := zap.NewProduction()
	defer log.Sync()

	requestLogger := createRequestLogger()
	//logS := logger.Sugar()

	conf, err := config.ParseRewritesFile(rewritesFile)
	if err != nil {
		log.Sugar().Fatalf(err.Error(), "path", rewritesFile)
	}

	wg := sync.WaitGroup{}
	for _, listener := range conf.Listeners {
		listener, err := proxy.NewListener(requestLogger, listener)
		if err != nil {
			log.Sugar().Fatalf(err.Error(), "listener", listener)
		}
		wg.Add(1)
		go func(listener *proxy.Listener) {
			log.Sugar().Fatal(listener.Start())
		}(listener)

	}

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
