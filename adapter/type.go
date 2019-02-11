package adapter

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
	Package     *string
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
