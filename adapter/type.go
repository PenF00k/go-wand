package adapter

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"gitlab.vmassive.ru/wand/generator/formatter"
	"go/ast"
)

type TypeName string
type StructName string
type FunctionName string

func (tn TypeName) ToUpperCamelCase() string {
	return strcase.ToCamel(string(tn))
}

type Adapter struct {
	Structures    []Struct
	Functions     []Function
	Subscriptions []Subscription
}

type Type struct {
	Name        TypeName
	Pointer     *Pointer
	Slice       *Slice
	Map         *Map
	Struct      *Struct
	Function    *Function
	IsPrimitive bool
	Selector    *Selector
	InnerName   string
}

func (t Type) IsPrimitivePointer() bool {
	return t.Pointer != nil && t.Pointer.InnerType.IsPrimitive
}

func (t Type) ToUpperCamelCase() string {
	return strcase.ToCamel(string(t.Name))
}

func (t Type) GetGenFuncName() string {
	if t.IsPrimitivePointer() {
		return t.Pointer.InnerType.GetGenFuncName() + "Wrapper"
	}
	if t.Pointer != nil {
		return t.Pointer.InnerType.GetGenFuncName() + "Pointer"
	} else if t.Slice != nil {
		return t.Slice.InnerType.GetGenFuncName() + "Slice"
	}

	return t.ToUpperCamelCase()
}

func (t Type) GetWrapperName(toPointer bool) string {
	sign := getPointerSign(toPointer)

	if t.IsPrimitivePointer() {
		protoType := format.BasicGoTypeFormatter.Format(string(t.Pointer.InnerType.Name))
		return fmt.Sprintf("%swrappers.%vValue", sign, strcase.ToCamel(protoType))
	} else if t.IsPrimitive {
		//return string(t.Name)
		return format.BasicGoTypeFormatter.Format(string(t.Name))
	}

	return t.ToUpperCamelCase()
}

func (t Type) WrapToProtoType(s string) string {
	if t.IsPrimitivePointer() {
		return format.GoFormatter.Format(string(t.Pointer.InnerType.Name), s)
	} else if t.IsPrimitive {
		return format.GoFormatter.Format(string(t.Name), s)
	}

	return ""
}

func (t Type) GetActualTypeName(upperCase bool) string {
	if t.IsPrimitivePointer() {
		//return fmt.Sprintf("wrappers.%vValue", t.Pointer.InnerType.GetActualTypeName(upperCase))
		return t.Pointer.InnerType.GetActualTypeName(upperCase)
	}
	if t.Pointer != nil {
		return t.Pointer.InnerType.GetActualTypeName(upperCase)
	} else if t.Slice != nil {
		return t.Slice.InnerType.GetActualTypeName(upperCase)
	}
	if upperCase {
		return t.ToUpperCamelCase()
	} else {
		return string(t.Name)

	}
}

func (t Type) GetParamTypeName(pack string) string {
	if t.Pointer != nil {
		return "*" + t.Pointer.InnerType.GetParamTypeName(pack)
	} else if t.Slice != nil {
		return "[]" + t.Slice.InnerType.GetParamTypeName(pack)
	} else if t.Selector != nil {
		s := t.Selector
		return fmt.Sprintf("%v.%v", s.Package, s.TypeName)
	} else if t.IsPrimitivePointer() {
		return t.GetWrapperName(false)
	} else if t.IsPrimitive {
		return string(t.Name)
	}

	if pack != "" {
		return fmt.Sprintf("%v.%v", pack, t.Name)
	}

	return string(t.Name)
}

func (t Type) GetReturnTypeName(pack string, toPointer bool) string {
	return t.getReturnTypeInner(pack, false, toPointer)
}

func (t Type) getReturnTypeInner(pack string, skipStructure bool, toPointer bool) string {
	sign := getPointerSign(toPointer)

	if t.IsPrimitivePointer() {
		return t.GetWrapperName(toPointer)
	} else if t.Pointer != nil {
		it := t.Pointer.InnerType
		return sign + it.getReturnTypeInner(pack, it.Struct != nil, toPointer)
	} else if t.Struct != nil && !skipStructure {
		return sign + t.getReturnTypeInner(pack, true, toPointer)
	} else if t.Slice != nil {
		var ps string
		t := t.Slice.InnerType
		if t.Pointer == nil && !t.IsPrimitive {
			ps = sign
		}

		return fmt.Sprintf("[]%s%s", ps, t.getReturnTypeInner(pack, true, toPointer))
		//return "[]" + t.Slice.InnerType.GetReturnTypeName(pack)
	} else if t.Selector != nil {
		s := t.Selector
		return fmt.Sprintf("%v.%v", s.Package, s.TypeName)
	} else if t.IsPrimitive {
		return string(t.Name)
	}

	if pack != "" {
		return fmt.Sprintf("%v.%v", pack, t.Name)
	}

	return string(t.Name)
}

type Pointer struct {
	InnerType *Type
}

type Slice struct {
	InnerType *Type
}

type Map struct {
	KeyType   *Type
	ValueType *Type
}

type Struct struct {
	Name   StructName
	Fields []Field
	//Annotations []Annotation
	//Comments    []string
}

type Field struct {
	Name string
	//TypeName TypeName
	Type *Type
}

func (f Field) NotIsLastField(list []Field, i int) bool {
	return i != len(list)-1
}

func (f Field) ToUpperCamelCase() string {
	return strcase.ToCamel(string(f.Name))
}

func (f Field) GetActualTypeName(upperCase bool) string {
	return f.Type.GetActualTypeName(upperCase)
}

func (f Field) GetParamTypeName(pack string) string {
	return f.Type.GetParamTypeName(pack)
}

func (f Field) GetReturnTypeName(pack string, toPointer bool) string {
	return f.Type.GetReturnTypeName(pack, toPointer)
}

func (f Field) IsExported() bool {
	return ast.IsExported(f.Name)
}

func (f Field) GetUpperCamelCaseName(prefix string, target string) string {
	n := prefix + strcase.ToCamel(f.Name)

	//if f.Type.IsPrimitivePointer() {
	//	n += ".Value"
	//}

	if target == "" {
		return n
	}

	var formatter format.FieldFormatter
	switch target {
	case "proto":
		formatter = format.ProtoFormatter
	case "go":
		formatter = format.GoFormatter
	}

	if formatter != nil {
		n = formatter.Format(string(f.Type.Name), n)
	}

	return n
}

func (f Field) GetLowerCamelCaseName() string {
	n := strcase.ToLowerCamel(f.Name)

	return n
}

type Annotation struct {
	Name  string
	Value string
}

type Function struct {
	FunctionName   string
	Args           []Field
	ReturnValues   []Field
	IsPure         bool
	IsSubscription bool
	//Annotations    []Annotation
	//Comments       []string
}

type Primitive struct {
	TypeName TypeName
}

type Subscription struct {
	Field Field
}

type Selector struct {
	Package  string
	TypeName TypeName
}

func getPointerSign(toPointer bool) string {
	if toPointer {
		return "*"
	}
	return "&"
}
