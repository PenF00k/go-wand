package generator

import (
	"gitlab.vmassive.ru/wand/config"
	"go/ast"
)

type CodeList struct {
	Package       string
	Dev           bool
	ServerIp      string
	Port          int16
	SourcePackage string
	StructData    []StructData
	FunctionData  []FunctionData
	Config        *config.Configuration
	PathMap       PathMap
	PackageMap    PackageMap
}

type StructData struct {
	Name        string
	FieldData   *ast.FieldList
	Annotations []Annotation
	Comments    []string
}

type Annotation struct {
	Name  string
	Value string
}

type FunctionData struct {
	Name           string
	Args           *ast.FieldList
	ReturnValues   *ast.FieldList
	IsSubscription bool
	IsPure         bool
	Annotations    []Annotation
	Comments       []string
}

type PathMap struct {
	Source              string
	Target              string
	Js                  string
	FlutterGenerated    string
	FlutterGeneratedRel string
	Proto               string
	ProtoRel            string
}

type PackageMap struct {
	PackageName       string
	ProtoPackageName  string
	FlutterAppPackage string
}

func (list *CodeList) AddStructData(structure StructData) {
	list.StructData = append(list.StructData, structure)
}

func (list *CodeList) AddFunctionData(function FunctionData) {
	list.FunctionData = append(list.FunctionData, function)
}
