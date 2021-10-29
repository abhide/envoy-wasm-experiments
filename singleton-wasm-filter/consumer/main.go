package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

//TODO: Read these variables from config; hard-code for now!
const httpbinAuthority = "httpbin.org"
const httpbinClusterName = "httpbin"

var requestTimeoutMs uint32 = 5000

const (
	queueName = "http_headers"
)

type vmContext struct {
	types.DefaultVMContext
}

type pluginContext struct {
	types.DefaultPluginContext
	contextID uint32
	queueName string
}

func (vmctx *vmContext) NewPluginContext(ID uint32) types.PluginContext {
	return &pluginContext{
		contextID: ID,
	}
}

func (ctx *pluginContext) OnPluginStart(_ int) types.OnPluginStartStatus {
	queueID, err := proxywasm.RegisterSharedQueue(queueName)
	if err != nil {
		proxywasm.LogCritical("Unable to register queue")
		panic(err)
	}
	proxywasm.LogInfof("Registered queue with ID: [%d]", queueID)
	return types.OnPluginStartStatusOK
}

func (ctx *pluginContext) OnQueueReady(queueID uint32) {
	data, err := proxywasm.DequeueSharedQueue(queueID)
	switch err {
	case types.ErrorStatusEmpty:
		return
	case nil:
		proxywasm.LogInfo("Sending request/response header data to httpbin")
		proxywasm.LogInfof("Read data from queue: %s", string(data))
		httpbinReqHeaders := [][2]string{
			{":method", "POST"}, {":path", "/anything"}, {":authority", httpbinAuthority}, {":scheme", "https"},
			// {":method", "GET"}, {":path", "/uuid"}, {":authority", "httpbin.org"}, {":scheme", "http"},
		}
		if _, err := proxywasm.DispatchHttpCall(
			httpbinClusterName, httpbinReqHeaders, data, [][2]string{}, requestTimeoutMs, httpCallResponseCallback); err != nil {
			proxywasm.LogCriticalf("Failed to dispatch to httpbin: %+v", err)
		}
	default:
		proxywasm.LogCriticalf("error retrieving data from queue %d: %v", queueID, err)
	}
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
func main() {
	proxywasm.LogInfo("Starting wasm plugin")
	proxywasm.SetVMContext(&vmContext{})
}
