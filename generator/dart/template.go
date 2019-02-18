package dart

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/wand/adapter"
	"gitlab.vmassive.ru/wand/generator"
	"gitlab.vmassive.ru/wand/generator/formatter"
	"os"
	"path"
	"strings"
)

type HeadData struct {
	CodeList *generator.CodeList
}

func (t HeadData) GetDartGeneratedPath() string {
	//import 'package:client_pik_remont/go_client/proto/generated/wrappers.pb.dart';
	// remove lib part from path
	fullFlutterPath := t.CodeList.PathMap.FlutterGeneratedRel
	parts := strings.Split(fullFlutterPath, string(os.PathSeparator))
	f := path.Join(parts[2:]...)
	f = path.Join(t.CodeList.PackageMap.FlutterAppPackage, f)
	return f
}

type TemplateStructData struct {
	FlatArgFields []*adapter.Type
	FlatFields    []*adapter.Type
	Adapter       *adapter.Adapter
	CodeList      *generator.CodeList
	Function      adapter.Function
	Package       string
	//Types   []TemplateProtoTypeData
}

func (t TemplateStructData) CompareOldAndNewArgs() string {
	if !t.IfFunctionHasArgs() {
		return ""
	}

	args := t.Function.Args
	res := strings.Builder{}
	for i, v := range args {
		if i != 0 {
			res.WriteString(" && ")
		}
		s := fmt.Sprintf("widget.%s != oldWidget.%[1]s", v.Name)
		res.WriteString(s)
	}

	return res.String()
}

func (t TemplateStructData) IfArgValuesAreNotNil() string {
	if !t.IfFunctionHasArgs() {
		return ""
	}

	return fmt.Sprintf("if (%s) ", t.ArgValuesAreNotNil())
}

func (t TemplateStructData) ArgValuesAreNotNil() string {
	if !t.IfFunctionHasArgs() {
		return ""
	}

	args := t.Function.Args
	res := strings.Builder{}
	for i, v := range args {
		if i != 0 {
			res.WriteString(" && ")
		}
		s := fmt.Sprintf("widget.%s != null", v.Name)
		res.WriteString(s)
	}

	return res.String()
}

func (t TemplateStructData) IfFunctionHasArgs() bool {
	return t.FunctionArgsLen() > 0
}

func (t TemplateStructData) FunctionArgsLen() int {
	return len(t.Function.Args)
}

func (t TemplateStructData) GetReturnTypeName() string {
	rv := t.Function.ReturnValues
	if len(rv) == 0 || rv[0].Type == nil {
		log.Warnf("No return type for function %s", t.Function.FunctionName)
		return "NoReturnType"
	}

	return getDartType(t.Function.ReturnValues[0].Type, false)
}

func (t TemplateStructData) GetCallFunctionNameForSubscription() string {
	return "get" + t.GetWithFunctionName()
}

func (t TemplateStructData) GetCallFunctionArgsForSubscription(withCallback bool) string {
	//widget.stageID
	args := t.Function.Args
	if len(args) == 0 {
		if withCallback {
			return "callback"
		}

		return ""
	}

	res := ""
	for _, v := range args {
		typeAndValue := fmt.Sprintf("widget.%s, ", v.Name)
		res = res + typeAndValue
	}

	if withCallback {
		return res + "callback"
	}

	return strings.TrimSuffix(res, ", ")
}

func (t TemplateStructData) GetWithFunctionName() string {
	fn := t.Function.FunctionName
	if !t.Function.IsSubscription {
		log.Warnf("method GetCallFunctionName must be called for subscription funcs only. Called on %s", fn)
		return t.Function.FunctionName
	}

	trimmed := strings.TrimPrefix(fn, "Subscribe")
	return trimmed
}

func (t TemplateStructData) BindArgs() string {
	// ..stageID = stageID
	args := t.Function.Args
	res := ""
	for _, v := range args {
		bound := fmt.Sprintf("..%s = %[1]s", v.Name)
		res = res + bound
	}

	return res
}

func (t TemplateStructData) GetArgsAsDartTypes() string {
	args := t.Function.Args
	res := ""
	for i, v := range args {
		typeAndValue := fmt.Sprintf("%s %s", getDartType(v.Type, false), v.Name)
		res = res + typeAndValue
		if i != len(args)-1 {
			res = res + ", "
		}
	}

	return res
}

func (t TemplateStructData) GetSubscriptionArgsAsDartTypes() string {
	args := t.Function.Args
	res := ""
	for _, v := range args {
		typeAndValue := fmt.Sprintf("%s %s, ", getDartType(v.Type, false), v.Name)
		res = res + typeAndValue
	}

	res = res + fmt.Sprintf("SubscriptionCallback<%s> typedCallback", t.GetReturnTypeName())

	return res
}

func getDartType(goType *adapter.Type, toLower bool) string {
	if goType.IsPrimitivePointer() {
		return format.WrapperDartTypeFormatter.Format(string(goType.Pointer.InnerType.Name))
	} else if goType.IsPointer() {
		return getDartType(goType.Pointer.InnerType, toLower)
	} else if goType.Struct != nil {
		if toLower {
			return goType.Struct.Name.ToLowerCamelCase()
		}
		return goType.Struct.Name.ToUpperCamelCase()
	} else if goType.Slice != nil {
		return fmt.Sprintf("List<%s>", getDartType(goType.Slice.InnerType, toLower))
	} else if goType.IsPrimitive {
		//tn := goType.Name.ToUpperCamelCase()
		//if toLower {
		//	tn = goType.Name.ToLowerCamelCase()
		//}
		return format.BasicDartTypeFormatter.Format(string(goType.Name))
	}

	return "unknownDartType"
}

func GetDartClassFieldForArg(field *adapter.Field) string {
	dartType := getDartType(field.Type, false)
	return fmt.Sprintf("final %s %s;", dartType, field.Name)
}

func GetDartClassConstructorPartForArg(field *adapter.Field) string {
	return fmt.Sprintf(", @MaterialDartLib.required this.%s", field.Name)
}

func GetDartClassAssertForArg(field *adapter.Field) string {
	return fmt.Sprintf("assert(%s != null),", field.Name)
}

//func (t TemplateStructData) GetEventTypeName() string {
//	var name string
//	var star string
//	pkg := t.Package + "."
//	rvt := t.Function.ReturnValues[0].Type
//	//t.Function.ReturnValues[0].Type.Pointer.InnerType.Name
//	if rvt.IsPrimitivePointer() {
//		pkg = ""
//	}
//	if rvt.Pointer != nil {
//		name = string(rvt.Pointer.InnerType.Name)
//		star = "*"
//	} else {
//		name = string(rvt.Name)
//		star = ""
//	}
//	return fmt.Sprintf("%v%v%v", star, pkg, name)
//}
//
//func (t TemplateStructData) GetLastFunction() *adapter.Type {
//	return t.FlatFields[len(t.FlatFields)-1]
//}
//
//func (t TemplateStructData) GetLastArgFunctionName(index int) string {
//	if t.FlatArgFields == nil || len(t.FlatArgFields) == 0 {
//		return ""
//	}
//	return fmt.Sprintf("%v%v", t.FlatArgFields[len(t.FlatArgFields)-1].GetGenFuncName(), index)
//}
//
////func (t TemplateStructData) GetEventTypeName() string {
////	return fmt.Sprintf("return %v.%v(", t.Package, t.Function.ReturnValues[0].Type.Name)
////	   return {{ .Package }}.{{ .Function.FunctionName }}({{ range $item := .Function.Args }}{{ $item.GetUpperCamelCaseName "args." "go" $item.Type.IsPrimitive }}, {{ end }}func(data *{{ .Package }}.{{ .GetEventTypeName }}) {
////}
//
//func flattenArgFieldsResult(args []adapter.Field) []*adapter.Type {
//	flatten := make([]*adapter.Type, 0, 10)
//	unique := make(map[*adapter.Type]bool)
//
//	for _, v := range args {
//		f := flattenType(v.Type, unique)
//		flatten = append(flatten, f...)
//	}
//
//	return flatten
//}
//
//func flattenResultFieldsResult(returnedFields []adapter.Field) []*adapter.Type {
//	flatten := make([]*adapter.Type, 0, 10)
//	unique := make(map[*adapter.Type]bool)
//
//	if len(returnedFields) > 0 {
//		f := flattenType(returnedFields[0].Type, unique)
//		flatten = append(flatten, f...)
//	}
//
//	return flatten
//}
//
//func flattenType(typ *adapter.Type, unique map[*adapter.Type]bool) []*adapter.Type {
//	flatten := make([]*adapter.Type, 0, 10)
//
//	if typ.IsPrimitive || typ.IsPrimitivePointer() {
//		//просто добавляем
//	} else if typ.Pointer != nil {
//		f := flattenType(typ.Pointer.InnerType, unique)
//		flatten = append(flatten, f...)
//	} else if typ.Struct != nil {
//		f := flattenStructType(typ.Struct.Fields, unique)
//		flatten = append(flatten, f...)
//	} else if typ.Slice != nil {
//		f := flattenType(typ.Slice.InnerType, unique)
//		flatten = append(flatten, f...)
//	} else if typ.Selector != nil && typ.Selector.IsTime() {
//		//просто добавляем
//	} else {
//		log.Warnf("unwanted type %+v", typ)
//	}
//
//	if _, ok := unique[typ]; !ok {
//		unique[typ] = true
//		flatten = append(flatten, typ)
//	}
//
//	return flatten
//}
//
//func flattenStructType(args []adapter.Field, unique map[*adapter.Type]bool) []*adapter.Type {
//	flatten := make([]*adapter.Type, 0, 10)
//	for _, v := range args {
//		f := flattenType(v.Type, unique)
//		flatten = append(flatten, f...)
//	}
//
//	return flatten
//}
//
//type ObjectType struct {
//	Pointer     *Pointer
//	Slice       *Slice
//	Struct      *Struct
//	IsPrimitive bool
//}
//
//type Pointer struct {
//	InnerType ObjectType
//}
//
//type Slice struct {
//	InnerType ObjectType
//}
//
//type Struct struct {
//	Name   string
//	Fields []ObjectType
//}
//
//func BuildObjects(function adapter.Function) []ObjectType {
//	res := make([]ObjectType, 0, len(function.Args))
//
//	for _, v := range function.ReturnValues {
//		AppendObject(v.Type, res)
//	}
//
//	return res
//}
//
//func AppendObject(typ *adapter.Type, objects []ObjectType) ObjectType {
//	var pointer *Pointer
//	var slice *Slice
//	var str *Struct
//	var isPrimitive = true
//
//	if typ.Pointer != nil {
//		pointer = &Pointer{
//			InnerType: AppendObject(typ.Pointer.InnerType, objects),
//		}
//
//		isPrimitive = typ.Pointer.InnerType.IsPrimitive
//	}
//
//	if typ.Slice != nil {
//		slice = &Slice{
//			InnerType: AppendObject(typ.Slice.InnerType, objects),
//		}
//	}
//
//	if typ.Struct != nil {
//		rawFields := typ.Struct.Fields
//		fields := make([]ObjectType, len(rawFields))
//
//		for i, v := range rawFields {
//			fields[i] = AppendObject(v.Type, objects)
//		}
//
//		str = &Struct{
//			Name:   string(typ.Struct.Name),
//			Fields: fields,
//		}
//	}
//
//	obj := ObjectType{
//		Pointer:     pointer,
//		Slice:       slice,
//		Struct:      str,
//		IsPrimitive: isPrimitive,
//	}
//
//	objects = append(objects, obj)
//
//	return obj
//}
//
////////////////
//
//type TemplateProtoTypeData struct {
//	Type        *adapter.Type
//	Name        string
//	FieldNumber int
//}
//
////{{ $item.TypeName }} {{ $item.Name }} = {{ $item.FieldNumber }};
//func (d TemplateProtoTypeData) GetFieldString() string {
//	return getFieldStringInner(d.Type, d.Name, d.FieldNumber)
//	//return fmt.Sprintf("%v %v = %v;", tn, d.Name, d.FieldNumber)
//}
//
//func getFieldStringInner(typ *adapter.Type, name string, fieldNumber int) string {
//	var typeName string
//	if typ.Pointer != nil {
//		if !typ.Pointer.InnerType.IsPrimitive {
//			return getFieldStringInner(typ.Pointer.InnerType, name, fieldNumber)
//		}
//
//		typeName = toProtoName(string(typ.Pointer.InnerType.Name), true)
//	} else {
//		typeName = toProtoName(string(typ.Name), false)
//	}
//
//	if typ.Slice != nil {
//		return "repeated " + getFieldStringInner(typ.Slice.InnerType, name, fieldNumber)
//	}
//
//	if typ.Selector != nil {
//		typeName = toProtoName(fmt.Sprintf("%v.%v", typ.Selector.Package, typ.Selector.TypeName), false)
//	}
//
//	return fmt.Sprintf("%v %v = %v;", typeName, name, fieldNumber)
//}
//
//func toProtoName(name string, isPointer bool) string {
//	//var f protoFormatter
//	//if isPointer {
//	//	f = ProtoFormatter
//	//} else {
//	//	f = BasicFormatter
//	//}
//
//	return name
//}
