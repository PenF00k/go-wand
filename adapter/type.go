package adapter

import "github.com/iancoleman/strcase"

type TypeName string
type StructName string
type FunctionName string

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

func (f Field) GetUpperCamelCaseName(prefix string, target string) string {
	n := prefix + strcase.ToCamel(f.Name)

	if f.Type.IsPrimitivePointer() {
		n += ".Value"
	}

	if target == "" {
		return n
	}

	var formatter fieldFormatter
	switch target {
	case "pro":
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
