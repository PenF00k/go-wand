package gocall

import (
	"fmt"
	"gitlab.vmassive.ru/wand/adapter"
)

type TemplateStructData struct {
	Objects []ObjectType
	//Types   []TemplateProtoTypeData
}

type ObjectType struct {
	Pointer     *Pointer
	Slice       *Slice
	Struct      *Struct
	IsPrimitive bool
}

type Pointer struct {
	InnerType ObjectType
}

type Slice struct {
	InnerType ObjectType
}

type Struct struct {
	Name   string
	Fields []ObjectType
}

func BuildObjects(function adapter.Function) []ObjectType {
	res := make([]ObjectType, 0, len(function.Args))

	for _, v := range function.ReturnValues {
		AppendObject(v.Type, res)
	}

	return res
}

func AppendObject(typ adapter.Type, objects []ObjectType) ObjectType {
	var pointer *Pointer
	var slice *Slice
	var str *Struct
	var isPrimitive = true

	if typ.Pointer != nil {
		pointer = &Pointer{
			InnerType: AppendObject(typ.Pointer.InnerType, objects),
		}

		isPrimitive = typ.Pointer.InnerType.IsPrimitive
	}

	if typ.Slice != nil {
		slice = &Slice{
			InnerType: AppendObject(typ.Slice.InnerType, objects),
		}
	}

	if typ.Struct != nil {
		rawFields := typ.Struct.Fields
		fields := make([]ObjectType, len(rawFields))

		for i, v := range rawFields {
			fields[i] = AppendObject(v.Type, objects)
		}

		str = &Struct{
			Name:   string(typ.Struct.Name),
			Fields: fields,
		}
	}

	obj := ObjectType{
		Pointer:     pointer,
		Slice:       slice,
		Struct:      str,
		IsPrimitive: isPrimitive,
	}

	objects = append(objects, obj)

	return obj
}

//////////////

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
	//var f protoFormatter
	//if isPointer {
	//	f = ProtoFormatter
	//} else {
	//	f = BasicFormatter
	//}

	return name
}
