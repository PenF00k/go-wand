package proto

import (
	"fmt"
	"gitlab.vmassive.ru/wand/adapter"
)

type TemplateStructData struct {
	Types []TemplateProtoTypeData
}

type TemplateProtoTypeData struct {
	TypeName    adapter.TypeName
	Name        string
	FieldNumber int
}

//{{ $item.TypeName }} {{ $item.Name }} = {{ $item.FieldNumber }};
func (gen CodeGenerator) GetFieldString(data TemplateProtoTypeData) string {
	return fmt.Sprintf("%v %v = %v;", data.TypeName, data.Name, data.FieldNumber)
}
