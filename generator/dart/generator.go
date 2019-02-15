package dart

//import (
//	"fmt"
//	log "github.com/sirupsen/logrus"
//	"gitlab.vmassive.ru/wand/adapter"
//	"gitlab.vmassive.ru/wand/generator"
//	"gitlab.vmassive.ru/wand/util"
//	"io"
//	"os"
//	"os/exec"
//	"path"
//	"text/template"
//)
//
//type CodeGenerator struct {
//	outDirectory string
//	Package      string
//	Adapter      *adapter.Adapter
//	CodeList     *generator.CodeList
//}
//
//func New(outDirectory string, packageName string, adapter *adapter.Adapter, codeList *generator.CodeList) generator.Generator {
//	return &CodeGenerator{
//		outDirectory: outDirectory,
//		Package:      packageName,
//		Adapter:      adapter,
//		CodeList:     codeList,
//	}
//}
//
//func (gen CodeGenerator) CreateCode() error {
//	err := gen.writeCode()
//	if err != nil {
//		return err
//	}
//
//	if !gen.CodeList.Dev {
//		cmd := exec.Command("go", "fmt")
//		cmd.Dir = gen.outDirectory
//
//		err := cmd.Start()
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (gen CodeGenerator) writeCode() error {
//	outFile := "call.go"
//	log.Printf("createing %s", outFile)
//
//	f, err := os.Create(path.Join(gen.outDirectory, outFile))
//	if err != nil {
//		log.Errorf("failed to create file %s", outFile)
//		return err
//	}
//
//	defer util.Close(f, outFile)
//	gen.writeHeader(f)
//	gen.writeMap(f)
//	gen.writeFunctions(f)
//	//writePureFunctions(f, generator.packageName, source)
//	return nil
//}
//
//func (gen CodeGenerator) writeHeader(f io.Writer) {
//	tpath := "templates/head.go.tmpl"
//	base := path.Base(tpath)
//
//	t, err := template.New(base).ParseFiles(tpath)
//	if err != nil {
//		log.Errorf("failed with error %v", err)
//		return
//	}
//
//	err = t.Execute(f, gen)
//	if err != nil {
//		log.Errorf("template failed with error %v", err)
//	}
//}
//
//func (gen CodeGenerator) writeMap(f io.Writer) {
//	tpath := "templates/callmap.go.tmpl"
//	base := path.Base(tpath)
//
//	t, err := template.New(base).Funcs(template.FuncMap{
//		"format": func(format string, a ...interface{}) string {
//			return fmt.Sprintf(format, a...)
//		},
//	}).ParseFiles(tpath)
//	if err != nil {
//		log.Errorf("failed to write head with error %v", err)
//	}
//
//	err = t.Execute(f, gen)
//	if err != nil {
//		log.Errorf("template failed with error %v", err)
//	}
//}
//func (gen CodeGenerator) writeFunctions(wr io.Writer) {
//	for _, function := range gen.Adapter.Functions {
//		gen.writeFunction(wr, gen.Package, function, gen.CodeList.PackageMap.ProtoPackageName)
//	}
//}
//
//func (gen CodeGenerator) writeFunction(wr io.Writer, pack string, function adapter.Function, protoPackageName string) {
//	tpath := "templates/func.go.tmpl"
//	base := path.Base(tpath)
//
//	t, err := template.
//		New(base).
//		Funcs(template.FuncMap{
//			"toBoolPointer": func(b bool) *bool {
//				return &b
//			},
//		}).
//		ParseFiles(tpath)
//	if err != nil {
//		log.Errorf("failed with error %v", err)
//		//return
//	}
//
//	f := gen.createFunction(function)
//	err = t.Execute(wr, f)
//	if err != nil {
//		log.Errorf("template failed with error %v", err)
//	}
//}
//
//func (gen CodeGenerator) createFunction(function adapter.Function) TemplateStructData {
//	//var flatten []*adapter.Type
//
//	//if function.IsSubscription {
//	//	flatten = flattenFieldsResult(function.ReturnValues)
//	//} else {
//	//	flatten = flattenFieldsResult(function.ReturnValues)
//	//}
//	flattenArgs := flattenArgFieldsResult(function.Args)
//	flatten := flattenResultFieldsResult(function.ReturnValues)
//
//	return TemplateStructData{
//		Fields:        function.Args,
//		FlatArgFields: flattenArgs,
//		FlatFields:    flatten,
//		Adapter:       gen.Adapter,
//		CodeList:      gen.CodeList,
//		Function:      function,
//		Package:       gen.Package,
//	}
//}