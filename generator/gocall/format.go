package gocall

import "fmt"

type fieldFormatter interface {
	format(typ, val string) string
}

var ProtoFormatter = protoFormatter{}
var GoFormatter = goFormatter{}

type protoFormatter struct{}

func (f protoFormatter) format(typ, val string) string {
	switch typ {
	case "int":
		return fmt.Sprintf("int32(%v)", val)
	default:
		return val
	}
}

type goFormatter struct{}

func (f goFormatter) format(typ, val string) string {
	switch typ {
	case "int":
		return fmt.Sprintf("int(%v)", val)
	default:
		return val
	}
}
