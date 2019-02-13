package format

type TypeFormatter interface {
	Format(TypeName string) string
}

var PointerTypeFormatter = pointerTypeFormatter{}
var BasicProtoTypeFormatter = basicProtoTypeFormatter{}
var BasicGoTypeFormatter = basicGoTypeFormatter{}

type pointerTypeFormatter struct{}

func (f pointerTypeFormatter) Format(TypeName string) string {
	switch TypeName {
	case "float32":
		return "google.protobuf.FloatValue"
	case "float64":
		return "google.protobuf.DoubleValue"
	case "int":
		fallthrough
	case "int8":
		fallthrough
	case "int16":
		fallthrough
	case "int32":
		return "google.protobuf.Int32Value"
	case "int64":
		return "google.protobuf.Int64Value"
	case "bool":
		return "google.protobuf.BoolValue"
	case "string":
		return "google.protobuf.StringValue"
	case "[]byte":
		return "google.protobuf.BytesValue"
	case "time.Time":
		return "google.protobuf.Timestamp"
	}

	return TypeName
}

type basicProtoTypeFormatter struct{}

func (f basicProtoTypeFormatter) Format(TypeName string) string {
	switch TypeName {
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "int":
		fallthrough
	case "int8":
		fallthrough
	case "int16":
		fallthrough
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "bool":
		return "bool"
	case "string":
		return "string"
	case "[]byte":
		return "bytes"
	case "time.Time":
		return "google.protobuf.Timestamp"
	}

	return TypeName
}


type basicGoTypeFormatter struct{}

func (f basicGoTypeFormatter) Format(TypeName string) string {
	switch TypeName {
	case "int":
		fallthrough
	case "int8":
		fallthrough
	case "int16":
		return "int32"
	}

	return TypeName
}
