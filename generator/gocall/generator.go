package gocall

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/PenF00k/go-wand/adapter"
	"github.com/PenF00k/go-wand/generator"
	"github.com/PenF00k/go-wand/util"
	"io"
	"os"
	"os/exec"
	"path"
	"text/template"
)

type GoCodeGenerator struct {
	outDirectory string
	Package      string
	Adapter      *adapter.Adapter
	CodeList     *generator.CodeList
}

func New(outDirectory string, packageName string, adapter *adapter.Adapter, codeList *generator.CodeList) generator.Generator {
	return &GoCodeGenerator{
		outDirectory: outDirectory,
		Package:      packageName,
		Adapter:      adapter,
		CodeList:     codeList,
	}
}

func (gen GoCodeGenerator) CreateCode() error {
	err := gen.writeCode()
	if err != nil {
		return err
	}

	if !gen.CodeList.Dev {
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
	log.Printf("creating %s", outFile)

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
	for _, function := range gen.Adapter.Functions {
		gen.writeFunction(wr, gen.Package, function, gen.CodeList.PackageMap.ProtoPackageName)
	}
}

func (gen GoCodeGenerator) writeFunction(wr io.Writer, pack string, function adapter.Function, protoPackageName string) {
	tpath := "templates/func.go.tmpl"
	base := path.Base(tpath)

	t, err := template.
		New(base).
		Funcs(template.FuncMap{
			"toBoolPointer": func(b bool) *bool {
				return &b
			},
		}).
		ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		//return
	}

	f := gen.createFunction(function)
	err = t.Execute(wr, f)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func (gen GoCodeGenerator) createFunction(function adapter.Function) TemplateStructData {
	//var flatten []*adapter.Type

	//if function.IsSubscription {
	//	flatten = flattenFieldsResult(function.ReturnValues)
	//} else {
	//	flatten = flattenFieldsResult(function.ReturnValues)
	//}
	flattenArgs := flattenArgFieldsResult(function.Args)
	flatten := flattenResultFieldsResult(function.ReturnValues)

	return TemplateStructData{
		Fields:        function.Args,
		FlatArgFields: flattenArgs,
		FlatFields:    flatten,
		Adapter:       gen.Adapter,
		CodeList:      gen.CodeList,
		Function:      function,
		Package:       gen.Package,
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
