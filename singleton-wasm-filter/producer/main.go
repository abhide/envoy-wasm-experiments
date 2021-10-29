package main

import (
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

//TODO: Read these variables from config; hard-code for now!
var tickPeriodMs uint32 = 30000

const (
	receiverVMID = "consumer"
	queueName    = "http_headers"
)

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
	queueID         uint32
	requestHeaders  [][2]string
	responseHeaders [][2]string
}

func (vmctx *vmContext) NewPluginContext(ID uint32) types.PluginContext {
	return &pluginContext{
		contextID: ID,
	}
}

func (ctx *pluginContext) NewHttpContext(ID uint32) types.HttpContext {
	queueID, err := proxywasm.ResolveSharedQueue(receiverVMID, queueName)
	if err != nil {
		proxywasm.LogCriticalf("error resolving queue id: %v", err)
		panic(err)
	}

	return &httpContext{
		contextID: ID,
		queueID:   queueID,
	}
}

func (ctx *pluginContext) OnPluginStart(_ int) types.OnPluginStartStatus {
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
	if err := proxywasm.EnqueueSharedQueue(ctx.queueID, payload); err != nil {
		proxywasm.LogCriticalf("Unable to enqueue data to queue: +%v", err)
	}
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

func main() {
	proxywasm.LogInfo("Starting wasm plugin")
	proxywasm.SetVMContext(&vmContext{})
}
