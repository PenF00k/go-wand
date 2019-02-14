package format

import "fmt"

type FieldFormatter interface {
	Format(typ, val string) string
}

var ProtoFormatter = protoFormatter{}
var GoFormatter = goFormatter{}
var DefaultFormatter = defaultFormatter{}

type protoFormatter struct{}

func (f protoFormatter) Format(typ, val string) string {
	protoType := BasicProtoTypeFormatter.Format(typ)
	return fmt.Sprintf("%v(%v)", protoType, val)
}

type goFormatter struct{}

func (f goFormatter) Format(typ, val string) string {
	goType := BasicGoTypeFormatter.Format(typ)

	return fmt.Sprintf("%v(%v)", goType, val)
}

type defaultFormatter struct{}

func (f defaultFormatter) Format(typ, val string) string {
	return fmt.Sprintf("%v(%v)", typ, val)
}
