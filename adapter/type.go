package adapter

import (
	"fmt"
	"github.com/iancoleman/strcase"
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
}

func (t Type) IsPrimitivePointer() bool {
	return t.Pointer != nil && t.Pointer.InnerType.IsPrimitive
}

func (t Type) ToUpperCamelCase() string {
	return strcase.ToCamel(string(t.Name))
}

func (t Type) GetActualTypeName(upperCase bool) string {
	if t.IsPrimitivePointer() {
		return fmt.Sprintf("wrappers.%vValue", t.Pointer.InnerType.GetActualTypeName(upperCase))
	}
	if t.Pointer != nil {
		return t.Pointer.InnerType.GetActualTypeName(upperCase)
	} else if t.Slice != nil {
		return t.Slice.InnerType.GetActualTypeName(upperCase) + "Slice"
	}
	if upperCase {
		return t.ToUpperCamelCase()
	} else {
		return string(t.Name)

	}
}

func (t Type) GetPrintableTypeName() string {
	if t.Pointer != nil {
		return "*" + t.Pointer.InnerType.GetPrintableTypeName()
	} else if t.Slice != nil {
		return "[]" + t.Slice.InnerType.GetPrintableTypeName()
	} else if t.Selector != nil {
		s := t.Selector
		return fmt.Sprintf("%v.%v", s.Package, s.TypeName)
	}

	return string(t.Name)
}

type Pointer struct {
	InnerType Type
}

type Slice struct {
	InnerType Type
}

type Map struct {
	KeyType   Type
	ValueType Type
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
	Type Type
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

func (f Field) GetPrintableTypeName() string {
	return f.Type.GetPrintableTypeName()
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

	var formatter fieldFormatter
	switch target {
	case "proto":
		formatter = ProtoFormatter
	case "go":
		formatter = GoFormatter
	}

	if formatter != nil {
		n = formatter.format(string(f.Type.Name), n)
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
