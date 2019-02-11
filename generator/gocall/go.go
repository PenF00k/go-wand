package gocall

import (
	"fmt"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/wand/adapter"
	"gitlab.vmassive.ru/wand/generator"
	"gitlab.vmassive.ru/wand/util"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path"
)

type Function struct {
	Name             string
	Comments         []string
	ReturnType       *ReturnType
	Params           []Field
	Subscription     bool
	Package          string
	ProtoPackageName string
}

type ReturnType struct {
	Name      string
	EventName string
	Params    []Field
	IsPointer bool
}

type Type struct {
	Name       string
	Map        bool
	Array      bool
	SimpleType string
	Pointer    bool
	InnerType  *Type
	Object     bool
	Primitive  PrimitiveType
}

type PrimitiveType struct {
	IsPrimitive     bool
	WrapperTypeName string
}

type Field struct {
	Name           string
	Type           string
	Comment        []string
	RichType       Type
	Array          bool
	SimpleType     string
	Package        string
	FunctionParams []Field
}

func (f Field) NotIsLastField(list []Field, i int) bool {
	return i != len(list)-1
}

func (f Field) GetUpperCamelCaseName(prefix string, target string, isPrimitive bool) string {
	n := prefix + strcase.ToCamel(f.Name)

	if isPrimitive {
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
		n = formatter.format(f.Type, n)
	}

	return n
}

func (f Field) GetLowerCamelCaseName() string {
	n := strcase.ToLowerCamel(f.Name)

	return n
}

type GoCodeGenerator struct {
	outDirectory string
	packageName  string
	adapter      *adapter.Adapter
	codeList     *generator.CodeList
}

func New(outDirectory string, packageName string, adapter *adapter.Adapter, codeList *generator.CodeList) generator.Generator {
	return &GoCodeGenerator{
		outDirectory: outDirectory,
		packageName:  packageName,
		adapter:      adapter,
		codeList:     codeList,
	}
}

func (gen GoCodeGenerator) CreateCode() error {
	err := gen.writeCode()
	if err != nil {
		return err
	}

	if !gen.codeList.Dev {
		cmd := exec.Command("go", "fmt")
		cmd.Dir = gen.outDirectory

		err := cmd.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (gen GoCodeGenerator) writeCode() error {
	outFile := "call.go"
	log.Printf("createing %s", outFile)

	f, err := os.Create(path.Join(gen.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer util.Close(f, outFile)
	gen.writeHeader(f)
	gen.writeMap(f)
	gen.writeFunctions(f)
	//writePureFunctions(f, generator.packageName, source)
	return nil
}

func (gen GoCodeGenerator) writeHeader(f io.Writer) {
	tpath := "templates/head.go.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	err = t.Execute(f, gen)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func (gen GoCodeGenerator) writeMap(f io.Writer) {
	tpath := "templates/callmap.go.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).Funcs(template.FuncMap{
		"format": func(format string, a ...interface{}) string {
			return fmt.Sprintf(format, a...)
		},
	}).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed to write head with error %v", err)
	}

	err = t.Execute(f, gen)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}
func (gen GoCodeGenerator) writeFunctions(wr io.Writer) {
	for _, function := range gen.adapter.Functions {
		writeFunction(wr, gen.packageName, function, gen.codeList.PackageMap.ProtoPackageName)
	}
}

//func writePureFunctions(wr io.Writer, pack string, source *generator.CodeList) {
//	for _, function := range source.Pure {
//		writePureFunction(wr, pack, function)
//	}
//}

//func writePureFunction(wr io.Writer, pack string, function generator.FunctionData) {
//	tpath := "templates/pure.go.tmpl"
//	base := path.Base(tpath)
//
//	t, err := template.New(base).ParseFiles(tpath)
//	if err != nil {
//		log.Errorf("failed with error %v", err)
//		return
//	}
//
//	err = t.Execute(wr, createFunction(pack, function))
//	if err != nil {
//		log.Errorf("template failed with error %v", err)
//	}
//}

func writeFunction(wr io.Writer, pack string, function adapter.Function, protoPackageName string) {
	tpath := "templates/func.go.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	//f := createFunction(pack, function, protoPackageName)
	err = t.Execute(wr, function)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}
