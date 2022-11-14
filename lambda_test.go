package main

import (
	"bufio"
	"fmt"
	"go/token"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type respTest struct {
	Raw  string
	Resp http.Response
	Body string
}

func dummyReq(method string) *http.Request {
	return &http.Request{Method: method}
}

func dummyReq11(method string) *http.Request {
	return &http.Request{Method: method, Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1}
}

var respTests = []respTest{
	{
		"HTTP/1.1 200 OK\r\n" +
			"\r\n" +
			"{\"statusCode\":200, \"body\": \"\"}\n",

		http.Response{
			Status:        "200 OK",
			StatusCode:    200,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header{"Content-Length": []string{"0"}},
			Request:       dummyReq("POST"),
			Close:         true,
			ContentLength: 0,
		},
		"",
	},
	{
		"HTTP/1.1 200 OK\r\n" +
			"\r\n" +
			`{"statusCode":200,"headers":{},"multiValueHeaders":null,"body":"","cookies":[]}
`,

		http.Response{
			Status:        "200 OK",
			StatusCode:    200,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header{"Content-Length": []string{"0"}},
			Request:       dummyReq("POST"),
			Close:         true,
			ContentLength: 0,
		},
		"",
	},
	{
		"HTTP/1.1 200 OK\r\n" +
			"\r\n" +
			`{"statusCode":200,"headers":{},"multiValueHeaders":null,"body":"<html><head></head><body><h1>something</h1></body></html>","cookies":[]}
`,

		http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				"Content-Length": []string{"57"},
			},
			Request:       dummyReq("POST"),
			Close:         true,
			ContentLength: 57,
		},

		"<html><head></head><body><h1>something</h1></body></html>",
	},
}

func TestModifyResponse(t *testing.T) {
	// body io.ReadCloser
	for i, tt := range respTests {
		resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(tt.Raw)), tt.Resp.Request)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		if transformLambdaBody(resp) != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}

		rbody := resp.Body
		resp.Body = nil
		diff(t, fmt.Sprintf("#%d Response", i), resp, &tt.Resp)
		var bout strings.Builder
		if rbody != nil {
			_, err = io.Copy(&bout, rbody)
			if err != nil {
				t.Errorf("#%d: %v", i, err)
				continue
			}
			rbody.Close()
		}
		body := bout.String()
		if body != tt.Body {
			t.Errorf("#%d: Body = %q want %q", i, body, tt.Body)
		}
	}
}

func TestWriteResponse(t *testing.T) {
	for i, tt := range respTests {
		resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(tt.Raw)), tt.Resp.Request)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		err = resp.Write(io.Discard)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
	}
}

func diff(t *testing.T, prefix string, have, want any) {
	t.Helper()
	hv := reflect.ValueOf(have).Elem()
	wv := reflect.ValueOf(want).Elem()
	if hv.Type() != wv.Type() {
		t.Errorf("%s: type mismatch %v want %v", prefix, hv.Type(), wv.Type())
	}
	for i := 0; i < hv.NumField(); i++ {
		name := hv.Type().Field(i).Name
		if !token.IsExported(name) {
			continue
		}
		hf := hv.Field(i).Interface()
		wf := wv.Field(i).Interface()
		if !reflect.DeepEqual(hf, wf) {
			t.Errorf("%s: %s = %v want %v", prefix, name, hf, wf)
		}
	}
}
