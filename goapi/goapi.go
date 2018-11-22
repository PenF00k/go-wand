package goapi

// JsCallback the interface for any callbacks
type JsCallback interface {
	OnSuccess(result interface{})
	OnError(err interface{})
}
