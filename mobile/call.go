package mobile

import (
	"encoding/json"
	"errors"
	"log"

	"gitlab.vmassive.ru/gocallgen/goapi"
	demo "gitlab.vmassive.ru/rn-demo"
)

// JsCallback the interface for any callbacks
type JsCallback interface {
	OnSuccess(json string)
	OnError(json string)
}

type callbackCaller struct {
	callback JsCallback
}

func (caller callbackCaller) OnSuccess(data interface{}) {
	bytes, _ := json.Marshal(data)
	caller.callback.OnSuccess(string(bytes))
}

func (caller callbackCaller) OnError(data interface{}) {
	bytes, _ := json.Marshal(data)
	caller.callback.OnError(string(bytes))
}

func newCaller(callback JsCallback) goapi.JsCallback {
	return &callbackCaller{}
}

// CallMethod - call from JS
func CallMethod(callData string, callback JsCallback) {
	methodCallData := make(map[string]interface{})
	json.Unmarshal([]byte(callData), &methodCallData)

	methodName := methodCallData["method"]
	log.Printf(">>> methodName %s", methodName)

	caller := newCaller(callback)
	functionCall := callmap[methodName]
	if functionCall != nil {
		functionCall(methodCallData, caller)
	} else {
		caller.OnError("no such method: " + functionCall)
	}
}

type CallFun func(map[string]interface{}, callbackCaller) error

var callmap map[string]CallFun = make(map[string]CallFun)

func init() {

	callmap["SimpleCall"] = callAdapterForSimpleCall

}

func callAdapterForSimpleCall(callData map[string]interface{}, callback callbackCaller) error {
	args, ok := callData["args"].([]interface{})
	if !ok {
		return errors.New("not able to cast args, wrong type")
	}

	body, err := func(arg interface{}) (string, error) {
		return arg.(string), nil

	}(args[0])
	if err != nil {
		return err
	}

	demo.SimpleCall(body, callback)
	return nil
}
