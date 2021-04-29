package main

import (
	"math/rand"

	"github.com/buger/jsonparser"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

type rootContext struct {
	// You'd better embed the default root context
	// so that you don't need to reimplement all the methods by yourself.
	proxywasm.DefaultRootContext
}

func newContext(uint32) proxywasm.RootContext { return &rootContext{} }

// Override DefaultRootContext.
func newHttpContext(rootContextID uint32, contextID uint32) proxywasm.HttpContext {
	return &httpHeaders{contextID: contextID}
}

type httpHeaders struct {
	// You'd better embed the default root context
	// so that you don't need to reimplement all the methods by yourself.
	proxywasm.DefaultHttpContext
	contextID uint32
}

func main() {
	proxywasm.LogInfo("Starting wasm plugin")
	proxywasm.SetNewRootContext(newContext)
	proxywasm.SetNewHttpContext(newHttpContext)
}

func (ctx *httpHeaders) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	headers, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("Unable to get request headers: [%v]", err)
	}
	for _, header := range headers {
		proxywasm.LogInfof("Request Header: [%s] :: Value: [%s]", header[0], header[1])
	}

	err = proxywasm.AddHttpRequestHeader("new_wasm_request_header", "woohoo")
	if err != nil {
		proxywasm.LogErrorf("Unable to add headers to request headers")
	}

	if rand.Int()%2 == 0 {
		proxywasm.LogInfof("HIJACK!!! Changing to a bogus path")
		proxywasm.RemoveHttpRequestHeader(":path")
		if err := proxywasm.AddHttpRequestHeader(":path", "/bogus_path"); err != nil {
			proxywasm.LogErrorf("PLOT FAILED! Unable to change to bogus path")
			return types.ActionContinue
		}
	}
	hs := [][2]string{
		{":method", "GET"}, {":path", "/uuid"}, {":authority", "httpbin.org"}, {":scheme", "http"},
	}
	if _, err := proxywasm.DispatchHttpCall(
		"httpbin", hs, "", [][2]string{}, 5000, httpCallResponseCallback); err != nil {
		proxywasm.LogCriticalf("Failed to dispatch http call", err)
		return types.ActionContinue
	}
	proxywasm.LogInfo("Successfully dispatched httpcall")
	return types.ActionPause
}

// Override DefaultHttpContext.
func (ctx *httpHeaders) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	headers, err := proxywasm.GetHttpResponseHeaders()
	if err != nil {
		proxywasm.LogCriticalf("Unable to get response headers: [%v]", err)
	}
	for _, header := range headers {
		proxywasm.LogInfof("Response Header: [%s] :: Value: [%s]", header[0], header[1])
	}

	err = proxywasm.AddHttpResponseHeader("new_wasm_response_header", "woohoo")
	// Ugly!! I am violating bunch of HTTP standards here; but experimentation; so why not!
	// Removing original content-length as we will mutate response body
	// else clients can't read entire mutated response
	// curl returns * Excess found in a read: excess = 30, size = 86, maxdownload = 86, bytecount = 0
	// with truncated response
	proxywasm.RemoveHttpResponseHeader("content-length")
	if err != nil {
		proxywasm.LogErrorf("Unable to add headers to response headers")
	}
	return types.ActionContinue
}

func (ctx *httpHeaders) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	if bodySize <= 0 {
		return types.ActionContinue
	}
	body, err := proxywasm.GetHttpRequestBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("Unable to get request body: [%v]", err)
	}
	proxywasm.LogInfof("Request Body: [%s]", string(body))
	return types.ActionContinue
}

func httpCallResponseCallback(numHeaders int, bodySize int, numTrailers int) {
	headers, err := proxywasm.GetHttpCallResponseHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to get httpcall response headers", err)
		return
	}
	for _, header := range headers {
		proxywasm.LogInfof("httpCallResponse Header: [%s] :: Value: [%s]", header[0], header[1])
	}
	body, err := proxywasm.GetHttpCallResponseBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("failed to get httpcall response body", err)
		return
	}
	proxywasm.LogInfof("httpCall Response Body: [%s]", string(body))
	proxywasm.ResumeHttpRequest()
}

func (ctx *httpHeaders) OnHttpResponseBody(bodySize int, endOfStream bool) types.Action {
	if bodySize <= 0 {
		return types.ActionContinue
	}
	body, err := proxywasm.GetHttpResponseBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("Unable to get response body: [%v]", err)
	}
	proxywasm.LogInfof("Response Body: [%s]", string(body))

	// Convert JSON to map
	// tinygo doesn't support encoding/json
	newBody, err := jsonparser.Set(body, []byte("new_wasm_value"), "new_wasm_key")
	if err != nil {
		proxywasm.LogCriticalf("Unable to marshal response body: [%v]", err)
	}
	proxywasm.SetHttpResponseBody(newBody)
	return types.ActionContinue
}
