# ratproxy
Ratproxy is a simple reverse proxy server to let you simulate something like AWS cloudfront sitting in front of your services.

![golint](https://github.com/streamingrat/ratproxy/workflows/golint/badge.svg)

Installation:

```sh
go install github.com/streamingrat/ratproxy@latest
```

By default ratproxy will use the environment variable `RATPROXY_FILENAME` to open the config file.  If not set uses ratproxy.yaml.

Configuration:
You configure the ratproxy with a yaml file like the following, which configures two services listening on two different ports.

`ratproxy.yaml`
---------------
```yaml
listen: 0.0.0.0:1414
useTLS: true
targets:
 - name: server1
   type: lambda
   target: http://localhost:10000
   path: /service1/
 - name: server2
   target: http://localhost:1313
   path: /
```
