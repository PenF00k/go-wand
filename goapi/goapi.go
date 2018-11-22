package goapi

// JsCallback the interface for any callbacks
type JsCallback interface {
	OnSuccess(result interface{})
	OnError(err interface{})
}

// JsEvent the interface for any events
type JsEvent interface {
	OnEvent(eventName string, json string)
}
