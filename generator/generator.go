package generator

import (
	"go/ast"

	"gitlab.vmassive.ru/wand/config"
)

type Annotation struct {
	Name  string
	Value string
}

type FunctionData struct {
	Name         string
	Comments     []string
	ReturnType   *ExportedStucture
	Params       *ast.FieldList
	Subscription bool
	Annotation   []Annotation
	CallName     string
}

type ExportedStucture struct {
	Comments   []string
	Name       string
	Field      *ast.FieldList
	Annotation []Annotation
}

type PathMap struct {
	Source   string
	Target   string
	Js       string
	Proto    string
	ProtoRel string
}

type CodeList struct {
	Package          string
	PackageName      string
	ProtoPackageName string
	Dev              bool
	Port             int16
	SourcePackage    string
	Structures       []ExportedStucture
	Functions        []FunctionData
	//Pure          []FunctionData
	Config  *config.Configuration
	PathMap PathMap
}

func (list *CodeList) AddStructure(structure ExportedStucture) {
	list.Structures = append(list.Structures, structure)
}

func (list *CodeList) AddFunction(function FunctionData) {
	list.Functions = append(list.Functions, function)
}

//func (list *CodeList) AddPureFunction(function FunctionData) {
//	list.Pure = append(list.Pure, function)
//}

type Generator interface {
	CreateCode(source *CodeList) error
}
