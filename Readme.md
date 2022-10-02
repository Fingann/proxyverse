# Proxyverse

Proxyverse is a reverse proxy. The goal is to provide a simple, easy to use, and secure reverse proxy.
Made as an alternative to nginx, apache, and other reverse proxies.

Benefits:
- Simple route configuration
- One binary


## route.yaml example
    
```yaml
- host: "example.com"
  locations:
    - uri: /gest
      target: http://localhost:4444/test
      headers:
        - name: "X-Real-IP"
          value: "{{ .RemoteIp }}"

- host: example.com
  addr: ":443"
  locations:
    - uri: /api-v[0-9]+
      target: http://localhost:4444

```



[![License](https://img.shields.io/badge/license-MIT-blue.svg)](