package gotest

import (
	"gitlab.vmassive.ru/gocallgen/goapi"
)

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
func CallAndGet(id string, params []int, callback goapi.JsCallback) {
	ls := LovelyStructure{}
	ExecuteCallback(ls, callback)
}

// Receive change notification for a stage with ID
// @subsription:Stage:
func StageSubscription(id int) {

}
