package proto

import (
	"fmt"
	"gitlab.vmassive.ru/wand/adapter"
)

type TemplateStructData struct {
	MessageName string
	Types       []TemplateProtoTypeData
}

type TemplateProtoTypeData struct {
	Type        adapter.Type
	Name        string
	FieldNumber int
}

//{{ $item.TypeName }} {{ $item.Name }} = {{ $item.FieldNumber }};
func (d TemplateProtoTypeData) GetFieldString() string {
	return getFieldStringInner(d.Type, d.Name, d.FieldNumber)
	//return fmt.Sprintf("%v %v = %v;", tn, d.Name, d.FieldNumber)
}

func getFieldStringInner(typ adapter.Type, name string, fieldNumber int) string {
	var typeName string
	if typ.Pointer != nil {
		if !typ.Pointer.InnerType.IsPrimitive {
			return getFieldStringInner(typ.Pointer.InnerType, name, fieldNumber)
		}

		typeName = toProtoName(string(typ.Pointer.InnerType.Name), true)
	} else {
		typeName = toProtoName(string(typ.Name), false)
	}

	if typ.Slice != nil {
		return "repeated " + getFieldStringInner(typ.Slice.InnerType, name, fieldNumber)
	}

	if typ.Selector != nil {
		typeName = toProtoName(fmt.Sprintf("%v.%v", typ.Selector.Package, typ.Selector.TypeName), false)
	}

	return fmt.Sprintf("%v %v = %v;", typeName, name, fieldNumber)
}

func toProtoName(name string, isPointer bool) string {
	var f protoFormatter
	if isPointer {
		f = PointerFormatter
	} else {
		f = BasicFormatter
	}

	return f.format(name)
}
