package gotest

import "encoding/json"

// JsCallback the interface for any callbacks
type JsCallback interface {
	OnSuccess(json string)
	OnError(json string)
}

// LovelyInterface My Comment
type Simple struct {
	SuName string
}

// LovelyStructure for your code
// Test my code
type LovelyStructure struct {
	// Type name test
	TypeName string
	// Optional name test
	TypeOptName *string
	// number test
	Number int
	// list test
	NumberList []int
	// Map example
	Map map[string][]map[string]Simple
}

// CallAndGet use for your classes
// @callback:LovelyStructure
func CallAndGet(id string, params []int, callback JsCallback) {
	ls := LovelyStructure{}
	ExecuteCallback(ls, callback)
}

// Receive change notification for a stage with ID
// @subsription:Stage:
func StageSubscription(id int) {

}

func ExecuteCallback(data interface{}, callback JsCallback) {
	bytes, err := json.Marshal(data)
	if err != nil {
		callback.OnError(err.Error())
	} else {
		callback.OnSuccess(string(bytes))
	}
}
