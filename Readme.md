# Proxyverse

Proxyverse is a reverse proxy. The goal is to provide a simple, easy to use, and secure reverse proxy.
Made as an alternative to nginx, apache, and other reverse proxies.

Benefits:
- Simple route configuration
- One binary


## route.yaml example
    
```yaml
- host: "example.com"
  addr: ":80"
  rewrite:
    - path: /
      redirect: "https://example.com"
- host: example.com
  addr: ":443"
  rewrite:
    - path: /gest
      headers:
        - key: "X-Real-IP"
          value: "{{ .RemoteIp }}"
      target: http://localhost:4444/test
    - path: /pest/
      target: http://localhost:4444/test/

```



[![License](https://img.shields.io/badge/license-MIT-blue.svg)](