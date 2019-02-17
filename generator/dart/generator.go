package dart

import (
	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/wand/adapter"
	"gitlab.vmassive.ru/wand/generator"
	"gitlab.vmassive.ru/wand/util"
	"io"
	"os"
	"os/exec"
	"path"
	"text/template"
)

type CodeGenerator struct {
	outDirectory string
	Package      string
	Adapter      *adapter.Adapter
	CodeList     *generator.CodeList
}

func New(outDirectory string, packageName string, adapter *adapter.Adapter, codeList *generator.CodeList) generator.Generator {
	return &CodeGenerator{
		outDirectory: outDirectory,
		Package:      packageName,
		Adapter:      adapter,
		CodeList:     codeList,
	}
}

func (gen CodeGenerator) CreateCode() error {
	err := gen.writeCode()
	if err != nil {
		return err
	}

	err = gen.writeWithFunctionsCode()
	if err != nil {
		return err
	}

	if !gen.CodeList.Dev {
		cmd := exec.Command("dartfmt")
		cmd.Dir = gen.outDirectory

		err := cmd.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (gen CodeGenerator) writeCode() error {
	outFile := gen.Package + ".dart"
	log.Printf("creating %s", outFile)

	f, err := os.Create(path.Join(gen.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer util.Close(f, outFile)
	gen.writeHeader(f)
	gen.writeFunctions(f)
	return nil
}

func (gen CodeGenerator) writeHeader(f io.Writer) {
	tpath := "templates/head.dart.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	data := HeadData{
		CodeList: gen.CodeList,
	}
	err = t.Execute(f, data)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func (gen CodeGenerator) writeFunctions(wr io.Writer) {
	for _, function := range gen.Adapter.Functions {
		gen.writeFunction(wr, gen.Package, function, gen.CodeList.PackageMap.ProtoPackageName)
	}
}

func (gen CodeGenerator) writeFunction(wr io.Writer, pack string, function adapter.Function, protoPackageName string) {
	tpath := "templates/func.dart.tmpl"
	base := path.Base(tpath)

	t, err := template.
		New(base).
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

func (gen CodeGenerator) createFunction(function adapter.Function) TemplateStructData {
	//var flatten []*adapter.Type

	//if function.IsSubscription {
	//	flatten = flattenFieldsResult(function.ReturnValues)
	//} else {
	//	flatten = flattenFieldsResult(function.ReturnValues)
	//}
	//flattenArgs := flattenArgFieldsResult(function.Args)
	//flatten := flattenResultFieldsResult(function.ReturnValues)

	return TemplateStructData{
		//FlatArgFields: flattenArgs,
		//FlatFields:    flatten,
		Adapter:  gen.Adapter,
		CodeList: gen.CodeList,
		Function: function,
		Package:  gen.Package,
	}
}

func (gen CodeGenerator) writeWithFunctionsCode() error {
	outFile := gen.Package + ".with.dart"
	log.Printf("creating %s", outFile)

	f, err := os.Create(path.Join(gen.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer util.Close(f, outFile)
	gen.writeWithHeader(f)
	gen.writeWithFunctions(f)
	return nil
}

func (gen CodeGenerator) writeWithHeader(f io.Writer) {
	tpath := "templates/head.with.dart.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	data := HeadData{
		CodeList: gen.CodeList,
	}
	err = t.Execute(f, data)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func (gen CodeGenerator) writeWithFunctions(wr io.Writer) {
	for _, function := range gen.Adapter.Functions {
		gen.writeWithFunction(wr, gen.Package, function, gen.CodeList.PackageMap.ProtoPackageName)
	}
}

func (gen CodeGenerator) writeWithFunction(wr io.Writer, pack string, function adapter.Function, protoPackageName string) {
	tpath := "templates/with.dart.tmpl"
	base := path.Base(tpath)

	t, err := template.
		New(base).
		Funcs(template.FuncMap{
			"GetDartClassFieldForArg": func(field *adapter.Field) string {
				return GetDartClassFieldForArg(field)
			},
			"GetDartClassConstructorPartForArg": func(field *adapter.Field) string {
				return GetDartClassConstructorPartForArg(field)
			},
			"GetDartClassAssertForArg": func(field *adapter.Field) string {
				return GetDartClassAssertForArg(field)
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
