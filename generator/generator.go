package generator

import (
	"go/ast"

	"gitlab.vmassive.ru/gocallgen/config"
)

type Annotation struct {
	Name  string
	Value string
}

type FunctionData struct {
	Name         string
	Comments     []string
	ReturnType   string
	Params       *ast.FieldList
	Subscription *string
	Annotation   []Annotation
}

type ExportedStucture struct {
	Comments   []string
	Name       string
	Field      *ast.FieldList
	Annotation []Annotation
}

type PathMap struct {
	Source string
	Target string
	Js     string
}

type CodeList struct {
	Package       string
	Dev           bool
	Port          int16
	SourcePackage string
	Structures    []ExportedStucture
	Functions     []FunctionData
	Config        *config.Configuration
	PathMap       PathMap
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
