package main

import (
	"bytes"
	"context"
	"encoding/base64"
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
	Body              string              `json:"body"`
	Cookies           []string            `json:"cookies"`
	Headers           map[string]string   `json:"headers"`
	IsBase64Encoded   bool                `json:"isBase64Encoded"`
	MultiValueHeaders []map[string]string `json:"multiValueHeaders"`
	StatusCode        int                 `json:"statusCode"`
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
		log.Printf("ratproxy: transformLambdaBody json.Unmarshal err=%v\n", err)
		return err
	}

	resp.StatusCode = lambdaResp.StatusCode
	var bbody []byte
	if lambdaResp.IsBase64Encoded {
		bbody = make([]byte, base64.StdEncoding.DecodedLen(len(lambdaResp.Body)))
		n, err := base64.StdEncoding.Decode(bbody, []byte(lambdaResp.Body))
		if err != nil {
			log.Printf("ratproxy: transformLambdaBody error %v\n", err)
			return err
		}
		bbody = bbody[:n]
	} else {
		bbody = []byte(lambdaResp.Body)
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bbody))
	for k, v := range lambdaResp.Headers {
		resp.Header.Set(k, v)
	}
	for _, c := range lambdaResp.Cookies {
		resp.Header.Set("Set-Cookie", c)
	}

	resp.ContentLength = int64(len(bbody))
	resp.Header.Set("Content-Length", fmt.Sprintf("%v", len(bbody)))

	return nil
}

// proxyTargetLambda treats the target as a lambda function.
func proxyTargetLambda(ctx context.Context, target *Target) func(http.ResponseWriter, *http.Request) {

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
			req.Method = http.MethodPost
			data := make(map[string]interface{})
			data["rawPath"] = req.URL.Path
			data["cookies"] = req.Header["Cookie"]

			if r.Method == http.MethodPost {
				data["requestContext"] = map[string]interface{}{"http": map[string]interface{}{"method": "POST"}}
				posted, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Printf("ratproxy: could not read posted body %v\n", err)
					posted = []byte("{}")
				}

				data["body"] = string(posted)
			}
			b, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			req.URL.Path = "/2015-03-31/functions/function/invocations"
			req.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			req.ContentLength = int64(len(b))

			log.Printf("ratproxy: %v %v: %v %v\n", r.Method, target.Name, req.URL.Path, string(b))
		}
		proxy.ModifyResponse = transformLambdaBody
		proxy.ServeHTTP(w, r)
	}
}
