// Ratproxy is a simple reverse proxy server to let you simulate something like AWS cloudfront sitting in front of your services.
//
// Installation:
//    `go get github.com/streamingrat/ratproxy
//
// By default ratproxy will use the environment variable `RATPROXY_FILENAME` to open the config file.  If not set uses ratproxy.yaml.
//
// Configuration:
// You configure the ratproxy with a yaml file like the following, which configures two services listening on two different ports.
//
// `ratproxy.yaml`
// ---------------
// ```yaml
// listen: 0.0.0.0:1414
// useTLS: true
// targets:
//  - name: server1
//    target: http://localhost:10000
//    path: /service1/
//    type: lambda
//  - name: server2
//    target: http://localhost:1313
//    path: /
// ```
package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// proxyTarget sets up a reverse proxy with name for logging to target, name is for logging.
func proxyTarget(ctx context.Context, target *Target) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		targetServer, err := url.Parse(target.Target)
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(targetServer)
		proxy.Director = func(req *http.Request) {
			req.Header = r.Header
			req.Host = targetServer.Host
			req.URL.Scheme = targetServer.Scheme
			req.URL.Host = targetServer.Host
			req.URL.Path = r.URL.Path
			log.Printf("ratproxy: %v %v: %v\n", r.Method, target.Name, r.URL.Path)
		}
		proxy.ServeHTTP(w, r)
	}

}

func main() {

	log.Printf("ratproxy: Version %v\n", Version)
	c, err := NewConfig()
	if err != nil {
		log.Printf("ratproxy: Could not read config file %v\n", err)
		return
	}
	ctx := context.Background()

	log.Printf("ratproxy: read %v targets\n", len(c.Targets))
	for i, t := range c.Targets {
		log.Printf("ratproxy: %v: %v %v %v %v\n", i, t.Name, t.Path, t.Target, t.Type)
		if t.Type == TargetTypeLambda {
			http.HandleFunc(t.Path, proxyTargetLambda(ctx, t))
		} else {
			http.HandleFunc(t.Path, proxyTarget(ctx, t))
		}

	}

	log.Printf("ratproxy: Listening %v TLS:%v\n", c.Listen, c.UseTLS)
	if c.UseTLS {
		cert, err := createCert()
		if err != nil {
			log.Fatalf("ratproxy: Could not create TLS cert err=%v\n", err)
		}
		server := http.Server{Addr: c.Listen, TLSConfig: cert}
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(http.ListenAndServe(c.Listen, nil))
	}
}
