package main

import (
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

//TODO: Read these variables from config; hard-code for now!
var tickPeriodMs uint32 = 30000

const httpbinAuthority = "httpbin.org"
const httpbinClusterName = "httpbin"

var requestTimeoutMs uint32 = 5000

var buffer [][]byte

type vmContext struct {
	types.DefaultVMContext
}

type pluginContext struct {
	types.DefaultPluginContext
	contextID uint32
}

type httpContext struct {
	types.DefaultHttpContext
	contextID       uint32
	requestHeaders  [][2]string
	responseHeaders [][2]string
}

func (vmctx *vmContext) NewPluginContext(ID uint32) types.PluginContext {
	return &pluginContext{
		contextID: ID,
	}
}

func (ctx *pluginContext) NewHttpContext(ID uint32) types.HttpContext {
	return &httpContext{
		contextID: ID,
	}
}

func (ctx *pluginContext) OnPluginStart(_ int) types.OnPluginStartStatus {
	if err := proxywasm.SetTickPeriodMilliSeconds(tickPeriodMs); err != nil {
		proxywasm.LogCriticalf("Unable to set tick interval: %+v", err)
		return types.OnPluginStartStatusFailed
	}
	return types.OnPluginStartStatusOK
}

func (ctx *httpContext) OnHttpRequestHeaders(_ int, _ bool) types.Action {
	headers, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogErrorf("Unable to read HTTP request headers: %+v", err)
	} else {
		ctx.requestHeaders = headers
	}
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpResponseHeaders(_ int, _ bool) types.Action {
	headers, err := proxywasm.GetHttpResponseHeaders()
	if err != nil {
		proxywasm.LogErrorf("Unable to read HTTP response headers: %+v", err)
		return types.ActionContinue
	} else {
		ctx.responseHeaders = headers
	}
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpStreamDone() {
	payload := generatePayload(ctx.requestHeaders, ctx.responseHeaders)
	buffer = append(buffer, payload)
}

func generatePayload(requestHeaders [][2]string, responseHeaders [][2]string) []byte {
	data := []byte("{}")
	for _, header := range requestHeaders {
		newData, err := jsonparser.Set(data, []byte(fmt.Sprintf("\"%s\"", header[1])), header[0])
		if err != nil {
			proxywasm.LogErrorf("Unable to set header: %s", header[0])
		} else {
			data = newData
		}
	}
	for _, header := range responseHeaders {
		newData, err := jsonparser.Set(data, []byte(fmt.Sprintf("\"%s\"", header[1])), header[0])
		if err != nil {
			proxywasm.LogErrorf("Unable to set header: %s", header[0])
		} else {
			data = newData
		}
	}
	return data
}

func httpCallResponseCallback(_ int, _ int, _ int) {
	headers, err := proxywasm.GetHttpCallResponseHeaders()
	if err != nil {
		proxywasm.LogCriticalf("Failed to get httpcall response headers: [%+v]", err)
		return
	}
	for _, header := range headers {
		if header[0] == ":status" {
			proxywasm.LogInfof("Got [%s] response from httpbin", header[1])
			break
		}
	}
}

func (ctx *pluginContext) OnTick() {
	proxywasm.LogInfo("Starting OnTick")
	length := len(buffer)
	if length == 0 {
		return
	}
	proxywasm.LogInfo("Sending request/response header data to httpbin")
	httpbinReqHeaders := [][2]string{
		{":method", "POST"}, {":path", "/anything"}, {":authority", httpbinAuthority}, {":scheme", "http"},
	}
	for i := 0; i < length; i++ {
		if _, err := proxywasm.DispatchHttpCall(

			httpbinClusterName, httpbinReqHeaders, buffer[i], [][2]string{}, requestTimeoutMs, httpCallResponseCallback); err != nil {
			proxywasm.LogCriticalf("Failed to dispatch to httpbin: %+v", err)
		}
	}
	// Drain the buffer
	buffer = buffer[length:]
	proxywasm.LogInfo("Done OnTick")
}

func main() {
	proxywasm.LogInfo("Starting wasm plugin")
	proxywasm.SetVMContext(&vmContext{})
}
