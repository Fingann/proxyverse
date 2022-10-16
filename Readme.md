# Proxyverse [WIP]

Proxyverse is a reverse proxy. The goal is to provide a simple, easy to use, and secure reverse proxy.
Made as an alternative to nginx, apache, and other reverse proxies.

**Benefits**:
- Simple route configuration
- One binary


## route.yaml example
    
```yaml
listeners:
  - addr: ":443"
    ssl: true
    domains:
      - name: "fingann.dev"
        rewrites:
          - path: /keep
            target: http://localhost:4444/base
          - path: /drop/
            headers:
              - key: "X-Real-IP"
                value: "{{ .RemoteIp }}"
            target: http://localhost:4444/base

      - name: "company.it"
        rewrites:
          - path: /
            redirect: true
            target: https://vg.no
  
  - addr: ":80"
    domains:
      - name: "fingann.dev"
        rewrites:
          - path: /
            redirect: true
            target: http://vg.no


```

### Variables that can be templated
```
  timestamp string
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
	PrivateIp bool
	RemoteAddr string
	RemoteIp string
	RemotePort string
	TransferEncoding []string
	Referer string
	Scheme string
  ```




[![License](https://img.shields.io/badge/license-MIT-blue.svg)](