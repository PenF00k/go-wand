package goapi

// JsCallback the interface for any callbacks
type JsCallback interface {
	OnSuccess(json string)
	OnError(json string)
}
