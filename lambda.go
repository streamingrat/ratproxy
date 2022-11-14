package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// LambdaResponse is the wrapped response from the lambda function that needs to be transformed into an http.Response.
type LambdaResponse struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders []map[string]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
	Cookies           []map[string]string `json:"cookies"`
}

// transformLambdaBody puts the body into the response
func transformLambdaBody(resp *http.Response) error {
	rbody := strings.Builder{}
	if resp.Body != nil {
		_, err := io.Copy(&rbody, resp.Body)
		if err != nil {
			log.Printf("transformLambdaBody err=%v\n", err)
			return err
		}
		resp.Body.Close()
	}
	body := rbody.String()

	lambdaResp := &LambdaResponse{}
	err := json.Unmarshal([]byte(body), &lambdaResp)
	if err != nil {
		log.Printf("transformLambdaBody json.Unmarshal err=%v\n", err)
		return err
	}

	resp.StatusCode = lambdaResp.StatusCode
	bbody := []byte(lambdaResp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(lambdaResp.Body)))
	for k, v := range lambdaResp.Headers {
		resp.Header.Set(k, v)
	}
	resp.ContentLength = int64(len(bbody))
	resp.Header.Set("Content-Length", fmt.Sprintf("%v", len(bbody)))

	return nil
}

// proxyTargetLambda treats the target as a lambda function.
func proxyTargetLambda(ctx context.Context, target *Target) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		showServer, err := url.Parse(target.Target)
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(showServer)
		proxy.Director = func(req *http.Request) {
			req.Header = r.Header
			req.Host = showServer.Host
			req.URL.Scheme = showServer.Scheme
			req.URL.Host = showServer.Host
			req.Method = http.MethodPost
			rawpath := req.URL.Path
			data := make(map[string]string)
			data["rawpath"] = req.URL.Path
			b, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			req.URL.Path = "/2015-03-31/functions/function/invocations"
			req.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			req.ContentLength = int64(len(b))
			log.Printf("ratproxy: %v %v: %v [%v]\n", r.Method, target.Name, req.URL.Path, rawpath)
		}
		proxy.ModifyResponse = transformLambdaBody
		proxy.ServeHTTP(w, r)
	}
}
