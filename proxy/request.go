package proxy

import (
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type Request struct {
	Host string
	Method string
	Path string
	RequestURI string
	UserAgent string
	ContentLength int64
	Headers map[string][]string
	Proto string
	ProtoMajor int
	ProtoMinor int
	RemoteAddr string
	RemoteIp string
	RemotePort string
	TransferEncoding []string
	Referer string
	Scheme string
}

type ProxyMatch struct {
	Host string
	Addr string
	Match string
	Uri string
}


type RequestLog struct {
	OriginalRequest *Request
	ProxyMatch *ProxyMatch
}

func NewRequestLog(r *http.Request, match *ProxyMatch) *RequestLog {
	return &RequestLog{
		OriginalRequest: FromHttpRequest(r),
		ProxyMatch: match,
	}

}

func LogRequest(logger *zap.Logger, logRequest *RequestLog) {
	logger.Info("request", ToZapField(logRequest)...)
}

func ToZapField(r *RequestLog) []zap.Field{
	fields := []zap.Field{
		zap.String("method", r.OriginalRequest.Method),
		zap.String("host", r.OriginalRequest.Host),
		zap.String("remote", r.OriginalRequest.RemoteAddr),
		zap.String("url", r.OriginalRequest.RequestURI),
		zap.Int64("content_length", r.OriginalRequest.ContentLength),
		zap.String("user_agent", r.OriginalRequest.UserAgent),
		zap.String("referer", r.OriginalRequest.Referer),
		zap.String("proto", r.OriginalRequest.Proto),
	}
	if r.ProxyMatch != nil {
	fields = append(fields, []zap.Field{
		zap.String("proxyHost", r.ProxyMatch.Host),
		zap.String("proxyAddr", r.ProxyMatch.Addr),
		zap.String("proxyMatch", r.ProxyMatch.Match),
		zap.String("proxyUri", r.ProxyMatch.Uri),
		}...)
	}
	fields = append(fields, zap.Namespace("headers"))
	for key, value := range r.OriginalRequest.Headers{
		fields = append(fields, zap.String(key, strings.Join(value, ",")))
	}
	
	return fields
}



func FromHttpRequest(r *http.Request) (*Request) {
	ip,port,_ := net.SplitHostPort(r.RemoteAddr)
	return &Request{
		Path: r.URL.Path,
		Host: r.Host,
		RequestURI: r.RequestURI,
		Headers: r.Header,
		Method: r.Method,
		Proto: r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		RemoteAddr: r.RemoteAddr,
		RemoteIp: ip,
		RemotePort: port,
		ContentLength: r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Referer: r.Referer(),
		UserAgent: r.UserAgent(),
		Scheme: r.URL.Scheme,
		}
	}
