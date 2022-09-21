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
// targets:
//  - name: server1
//    target: http://localhost:10000
//    path: /service1/
//  - name: server2
//    target: http://localhost:1313
//    path: /
// ```
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the ratproxy.yaml.
type Config struct {
	Listen  string   `yaml:"listen"`
	Targets []Target `yaml:"targets"`
}

// Target is a proxy target based on a path.
type Target struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
	Path   string `yaml:"path"`
}

// proxyTarget sets up a reverse proxy with name for logging to target, name is for logging.
func proxyTarget(ctx context.Context, target string, name string) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		showServer, err := url.Parse(target)
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(showServer)
		proxy.Director = func(req *http.Request) {
			req.Header = r.Header
			req.Host = showServer.Host
			req.URL.Scheme = showServer.Scheme
			req.URL.Host = showServer.Host
			req.URL.Path = r.URL.Path
			log.Printf("ratproxy: %v %v: %v\n", r.Method, name, r.URL.Path)
		}
		proxy.ServeHTTP(w, r)
	}

}

func main() {

	configFilename := os.Getenv("RATPROXY_FILENAME")
	if configFilename == "" {
		configFilename = "ratproxy.yaml"
	}

	fmt.Printf("ratproxy: reading config at %v\n", configFilename)
	data, err := ioutil.ReadFile(configFilename)
	if err != nil {
		panic(err)
	}
	c := Config{}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	fmt.Printf("ratproxy: read %v targets\n", len(c.Targets))
	for i, t := range c.Targets {
		fmt.Printf("ratproxy: %v: %v %v %v\n", i, t.Name, t.Path, t.Target)
		http.HandleFunc(t.Path, proxyTarget(ctx, t.Target, t.Name))

	}

	fmt.Printf("ratproxy: Listening %v\n", c.Listen)
	log.Fatal(http.ListenAndServe(c.Listen, nil))
}
