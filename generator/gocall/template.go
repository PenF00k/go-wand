package gocall

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/wand/adapter"
	"gitlab.vmassive.ru/wand/generator"
)

type TemplateStructData struct {
	Fields        []adapter.Field
	FlatArgFields []*adapter.Type
	FlatFields    []*adapter.Type
	Adapter       *adapter.Adapter
	CodeList      *generator.CodeList
	Function      adapter.Function
	Package       string
	//Types   []TemplateProtoTypeData
}

func (t TemplateStructData) GetEventTypeName() string {
	var name string
	var star string
	pkg := t.Package + "."
	rvt := t.Function.ReturnValues[0].Type
	//t.Function.ReturnValues[0].Type.Pointer.InnerType.Name
	if rvt.IsPrimitivePointer() {
		pkg = ""
	}
	if rvt.Pointer != nil {
		name = string(rvt.Pointer.InnerType.Name)
		star = "*"
	} else {
		name = string(rvt.Name)
		star = ""
	}
	return fmt.Sprintf("%v%v%v", star, pkg, name)
}

func (t TemplateStructData) GetLastFunction() *adapter.Type {
	return t.FlatFields[len(t.FlatFields)-1]
}

func (t TemplateStructData) GetLastArgFunctionName(index int) string {
	if t.FlatArgFields == nil || len(t.FlatArgFields) == 0 {
		return ""
	}
	return fmt.Sprintf("%v%v", t.FlatArgFields[len(t.FlatArgFields)-1].GetGenFuncName(),  index)
}

//func (t TemplateStructData) GetEventTypeName() string {
//	return fmt.Sprintf("return %v.%v(", t.Package, t.Function.ReturnValues[0].Type.Name)
//	   return {{ .Package }}.{{ .Function.FunctionName }}({{ range $item := .Function.Args }}{{ $item.GetUpperCamelCaseName "args." "go" $item.Type.IsPrimitive }}, {{ end }}func(data *{{ .Package }}.{{ .GetEventTypeName }}) {
//}

func flattenArgFieldsResult(args []adapter.Field) []*adapter.Type {
	flatten := make([]*adapter.Type, 0, 10)
	unique := make(map[*adapter.Type]bool)

	for _, v := range args {
		f := flattenType(v.Type, unique)
		flatten = append(flatten, f...)
	}

	return flatten
}

func flattenResultFieldsResult(returnedFields []adapter.Field) []*adapter.Type {
	flatten := make([]*adapter.Type, 0, 10)
	unique := make(map[*adapter.Type]bool)

	if len(returnedFields) > 0 {
		f := flattenType(returnedFields[0].Type, unique)
		flatten = append(flatten, f...)
	}

	return flatten
}

func flattenType(typ *adapter.Type, unique map[*adapter.Type]bool) []*adapter.Type {
	flatten := make([]*adapter.Type, 0, 10)

	if typ.IsPrimitive || typ.IsPrimitivePointer() {
		//просто добавляем
	} else if typ.Pointer != nil {
		f := flattenType(typ.Pointer.InnerType, unique)
		flatten = append(flatten, f...)
	} else if typ.Struct != nil {
		f := flattenStructType(typ.Struct.Fields, unique)
		flatten = append(flatten, f...)
	} else if typ.Slice != nil {
		f := flattenType(typ.Slice.InnerType, unique)
		flatten = append(flatten, f...)
	} else {
		log.Warnf("unwanted type %+v", typ)
	}

	if _, ok := unique[typ]; !ok {
		unique[typ] = true
		flatten = append(flatten, typ)
	}

	return flatten
}

func flattenStructType(args []adapter.Field, unique map[*adapter.Type]bool) []*adapter.Type {
	flatten := make([]*adapter.Type, 0, 10)
	for _, v := range args {
		f := flattenType(v.Type, unique)
		flatten = append(flatten, f...)
	}

	return flatten
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

func AppendObject(typ *adapter.Type, objects []ObjectType) ObjectType {
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
	Type        *adapter.Type
	Name        string
	FieldNumber int
}

//{{ $item.TypeName }} {{ $item.Name }} = {{ $item.FieldNumber }};
func (d TemplateProtoTypeData) GetFieldString() string {
	return getFieldStringInner(d.Type, d.Name, d.FieldNumber)
	//return fmt.Sprintf("%v %v = %v;", tn, d.Name, d.FieldNumber)
}

func getFieldStringInner(typ *adapter.Type, name string, fieldNumber int) string {
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
