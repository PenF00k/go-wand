package adapter

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"gitlab.vmassive.ru/wand/generator/formatter"
	"go/ast"
	"strings"
)

type TypeName string
type StructName string
type FunctionName string

func (tn TypeName) ToUpperCamelCase() string {
	return strcase.ToCamel(string(tn))
}

func (tn TypeName) ToLowerCamelCase() string {
	return strcase.ToLowerCamel(string(tn))
}

func (sn StructName) ToUpperCamelCase() string {
	return strcase.ToCamel(string(sn))
}

func (sn StructName) ToLowerCamelCase() string {
	return strcase.ToLowerCamel(string(sn))
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

func (t Type) IsPointer() bool {
	return t.Pointer != nil
}

func (t Type) ToUpperCamelCase() string {
	return strcase.ToCamel(string(t.Name))
}

func (t Type) GetGenFuncName() string {
	if t.IsPrimitivePointer() {
		return t.Pointer.InnerType.GetGenFuncName() + "Wrapper"
	} else if t.Pointer != nil {
		return t.Pointer.InnerType.GetGenFuncName() + "Pointer"
	} else if t.Slice != nil {
		return t.Slice.InnerType.GetGenFuncName() + "Slice"
	} else if t.Selector != nil {
		x := strcase.ToCamel(string(t.Selector.Package))
		sel := strcase.ToCamel(string(t.Selector.TypeName))
		return fmt.Sprintf("%s%s", x, sel)
	}

	return t.ToUpperCamelCase()
}

func (t Type) GetWrapperName(toPointer *bool, reverse bool) string {
	sign := getPointerSign(toPointer)

	if t.IsPrimitivePointer() {
		protoType := format.BasicProtoTypeFormatter.Format(string(t.Pointer.InnerType.Name))
		return fmt.Sprintf("%swrappers.%vValue", sign, strcase.ToCamel(protoType))
	} else if t.IsPrimitive {
		//return string(t.Name)
		return format.BasicGoTypeFormatter.Format(string(t.Name))
	}
	//if t.Selector != nil {
	//	sel := t.Selector
	//	if sel.Package == "time" && sel.TypeName == "Time" {
	//		return fmt.Sprintf("%stimestamp.Timestamp", sign)
	//	}
	//}

	return t.ToUpperCamelCase()
}

func (t Type) WrapToProtoType(s string) string {
	if t.IsPrimitivePointer() {
		return format.GoFormatter.Format(string(t.Pointer.InnerType.Name), s)
	} else if t.IsPrimitive {
		return format.GoFormatter.Format(string(t.Name), s)
	}
	//if t.Selector != nil {
	//	sel := t.Selector
	//	if sel.Package == "time" && sel.TypeName == "Time" {
	//		return fmt.Sprintf("timestamp.Timestamp")
	//	}
	//}

	return ""
}

func (t Type) WrapToDefaultGoType(s string) string {
	if t.IsPrimitivePointer() {
		return format.DefaultFormatter.Format(string(t.Pointer.InnerType.Name), s)
	} else if t.IsPrimitive {
		return format.DefaultFormatter.Format(string(t.Name), s)
	}
	//if t.Selector != nil {
	//	sel := t.Selector
	//	if sel.Package == "time" && sel.TypeName == "Time" {
	//		return fmt.Sprintf("timestamp.Timestamp")
	//	}
	//}

	return ""
}

func (t Type) GetProtoType() string {
	if t.IsPrimitivePointer() {
		return format.BasicProtoTypeFormatter.Format(string(t.Pointer.InnerType.Name))
	} else if t.IsPrimitive {
		return format.BasicProtoTypeFormatter.Format(string(t.Name))
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

func (t Type) GetParamTypeName(pack string, reverse bool) string {
	if t.Pointer != nil {
		return "*" + t.Pointer.InnerType.GetParamTypeName(pack, reverse)
	} else if t.Slice != nil {
		return "[]" + t.Slice.InnerType.GetParamTypeName(pack, reverse)
	} else if t.Selector != nil {
		s := t.Selector
		if !reverse {
			return fmt.Sprintf("%v.%v", s.Package, s.TypeName)
		}
	} else if t.IsPrimitivePointer() {
		toPointer := false
		return t.GetWrapperName(&toPointer, false)
	} else if t.IsPrimitive {
		return string(t.Name)
	}

	if pack != "" {
		return fmt.Sprintf("%v.%v", pack, t.Name)
	}

	return string(t.Name)
}

func (t Type) GetReturnTypeName(pack string, toPointer *bool, reverse bool) string {
	return t.getReturnTypeInner(pack, false, toPointer, reverse)
}

func (t Type) getReturnTypeInner(pack string, skipStructure bool, toPointer *bool, reverse bool) string {
	sign := getPointerSign(toPointer)

	if t.IsPrimitivePointer() {
		if reverse {
			return sign + string(t.Pointer.InnerType.Name)
		} else {
			return t.GetWrapperName(toPointer, reverse)
		}
	} else if t.Pointer != nil {
		it := t.Pointer.InnerType
		return sign + it.getReturnTypeInner(pack, it.Struct != nil, toPointer, reverse)
	} else if t.Struct != nil && !skipStructure {
		return sign + t.getReturnTypeInner(pack, true, toPointer, reverse)
	} else if t.Slice != nil {
		var ps string
		t := t.Slice.InnerType
		if t.Pointer == nil && !t.IsPrimitive {
			ps = sign
		}

		return fmt.Sprintf("[]%s%s", ps, t.getReturnTypeInner(pack, true, toPointer, reverse))
		//return "[]" + t.Slice.InnerType.GetReturnTypeName(pack)
	} else if t.Selector != nil {
		sel := t.Selector
		if sel.IsTime() {
			if reverse {
				return fmt.Sprintf("time.Time")
			} else {
				return fmt.Sprintf("timestamp.Timestamp")
			}
		}
	} else if t.IsPrimitive {
		return format.BasicProtoTypeFormatter.Format(string(t.Name))
		//formatter.Format(string(f.Type.Name), n)
		//return string(t.Name)
	}

	if pack != "" {
		return fmt.Sprintf("%v.%v", pack, t.Name)
	}

	return string(t.Name)
}

type Pointer struct {
	InnerType *Type
}

func (p *Pointer) ToBool() *bool {
	if p == nil {
		return nil
	}

	res := true
	return &res
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

func (f Field) GetParamTypeName(pack string, reverse bool) string {
	return f.Type.GetParamTypeName(pack, reverse)
}

func (f Field) GetReturnTypeName(pack string, toPointer *bool, reverse bool) string {
	return f.Type.GetReturnTypeName(pack, toPointer, reverse)
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
	case "default":
		formatter = format.DefaultFormatter
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

func (f Function) GetLowerCaseName() string {
	return strcase.ToLowerCamel(f.FunctionName)
}

func (f Function) GetSubscribeFunctionBaseName() (string, error) {
	fn := f.FunctionName
	if !f.IsSubscription {
		return "", fmt.Errorf("function %s is not a subscription", fn)
	}

	trimmed := strings.TrimPrefix(fn, "Subscribe")
	return trimmed, nil
}

func (f Function) HasReturnTypes() bool {
	return f.ReturnValues != nil && len(f.ReturnValues) > 0
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
	Type     *Type
}

func (s *Selector) IsTime() bool {
	return s.Package == "time" && s.TypeName == "Time"
}

func getPointerSign(toPointer *bool) string {
	if toPointer == nil {
		return ""
	}
	if *toPointer {
		return "*"
	}
	return "&"
}
