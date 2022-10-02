package proxy 

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost *url.URL) (*httputil.ReverseProxy, error) {
  
    proxy := httputil.NewSingleHostReverseProxy(targetHost)
 
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        modifyRequest(req)
    }
 
    proxy.ModifyResponse = modifyResponse()
    proxy.ErrorHandler = errorHandler()
    return proxy, nil
}
 
func modifyRequest(req *http.Request) {
    req.Header.Set("X-Proxy", "Simple-Reverse-Proxy")
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