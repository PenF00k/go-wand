package generator

import (
	"go/ast"
)

type FunctionData struct {
	Name         string
	Comments     []string
	ReturnType   string
	Params       *ast.FieldList
	Subscription *string
}

type ExportedStucture struct {
	Comments []string
	Name     string
	Field    *ast.FieldList
}

type CodeList struct {
	Structures []ExportedStucture
	Functions  []FunctionData
}

func (list *CodeList) AddStructure(structure ExportedStucture) {
	list.Structures = append(list.Structures, structure)
}

func (list *CodeList) AddFunction(function FunctionData) {
	list.Functions = append(list.Functions, function)
}

type Generator interface {
	CreateCode(source *CodeList) error
}
