package js

import (
	"errors"
	"go/ast"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/wand/assets"
	"gitlab.vmassive.ru/wand/generator"
)

type Field struct {
	Name    string
	Type    string
	Comment []string
}

type Function struct {
	Name         string
	Comments     []string
	ReturnType   string
	Params       []Field
	Subscription *string
}

type Structure struct {
	Comments []string
	Name     string
	Field    []Field
}

type JsCodeGenerator struct {
	outDirectory string
	packageName  string
	Source       *generator.CodeList
}

func New(source *generator.CodeList, outDirectory string, packageName string) generator.Generator {
	return &JsCodeGenerator{
		outDirectory: outDirectory,
		packageName:  packageName,
		Source:       source,
	}
}

func (generator JsCodeGenerator) CreateCode() error {
	source := generator.Source
	err := generator.writeGeneral(source)
	if err != nil {
		return err
	}

	err = generator.writeWithFunctions(source)
	if err != nil {
		return err
	}

	return nil
}

func (gen JsCodeGenerator) writeWithFunctions(source *generator.CodeList) error {
	outFile := gen.packageName + "HOC.js"

	list := gen.getAnnotatedStructures(source)
	if len(list) == 0 {
		return nil
	}

	log.Printf("createing (with) %s", outFile)

	f, err := os.Create(path.Join(gen.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer f.Close()

	gen.writeWithHeader(f, source)

	for _, item := range list {
		var getFunc, updateFunc *generator.FunctionData
		get := gen.getAnnotation("get", item.Annotation)
		if get != nil {
			getFunc = gen.findFunction(*get, source.Functions)
		}

		update := gen.getAnnotation("update", item.Annotation)
		if update != nil {
			updateFunc = gen.findFunction(*update, source.Functions)
		}

		gen.writeWithFunction(f, getFunc, updateFunc, item)
	}

	return nil
}

type WithData struct {
	VarName string
	Name    string
	Props   []Field
	Update  *Function
	Get     *Function
}

type WithHeader struct {
	Structures  []string
	Functions   []string
	PackageName string
}

func (gen JsCodeGenerator) writeWithHeader(f io.Writer, sourceList *generator.CodeList) error {
	// headBytes, err := ioutil.ReadFile("./templates/headWith.js.tmpl") // just pass the file name
	// if err != nil {
	// 	log.Errorf("read file error %v", err)
	// 	return err
	// }

	file, err := assets.Assets.Open("/templates/headWith.js.tmpl")
	defer file.Close()
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}

	headBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}

	header := WithHeader{
		PackageName: gen.packageName,
	}

	for _, f := range sourceList.Functions {
		header.Functions = append(header.Functions, strcase.ToLowerCamel(f.Name))
	}

	for _, s := range sourceList.Structures {
		header.Structures = append(header.Structures, s.Name)
	}

	headTemplate, err := template.New("structType").Parse(string(headBytes))
	if err != nil {
		log.Errorf("failed to write head with error %v", err)
		return err
	}

	return headTemplate.Execute(f, header)
}

func (gen JsCodeGenerator) writeWithFunction(wr io.Writer, get *generator.FunctionData, update *generator.FunctionData, strct generator.ExportedStucture) error {
	// headBytes, err := ioutil.ReadFile("./templates/with.js.tmpl") // just pass the file name
	// if err != nil {
	// 	log.Errorf("read file error %v", err)
	// 	return err
	// }

	file, err := assets.Assets.Open("/templates/with.js.tmpl")
	defer file.Close()
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}

	headBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}

	funcDecl := get
	if get == nil {
		funcDecl = update
	}

	if funcDecl == nil {
		return errors.New("not declared update or get")
	}

	withData := WithData{
		Name:    strct.Name,
		VarName: strcase.ToLowerCamel(strct.Name),
		Props:   createListOfFields(funcDecl.Params),
	}

	if get != nil {
		jsFunc := createFunction(*get)
		withData.Get = &jsFunc
	}

	if update != nil {
		jsFunc := createFunction(*update)
		withData.Update = &jsFunc
	}

	headTemplate, err := template.New("with").Parse(string(headBytes))
	if err != nil {
		log.Errorf("failed to write head with error %v", err)
		return err
	}

	return headTemplate.Execute(wr, withData)
}

func (gen JsCodeGenerator) findFunction(name string, list []generator.FunctionData) *generator.FunctionData {
	for _, item := range list {
		if item.Name == name {
			return &item
		}
	}

	return nil
}

func (gen JsCodeGenerator) getAnnotation(name string, list []generator.Annotation) *string {
	for _, item := range list {
		if item.Name == name {
			return &item.Value
		}
	}

	return nil
}

func (gen JsCodeGenerator) getAnnotatedStructures(source *generator.CodeList) []generator.ExportedStucture {
	list := make([]generator.ExportedStucture, 0, len(source.Structures))

	for _, stucture := range source.Structures {
		if len(stucture.Annotation) > 0 {
			list = append(list, stucture)
		}
	}

	return list
}

func (generator JsCodeGenerator) writeGeneral(source *generator.CodeList) error {
	outFile := generator.packageName + ".js"
	log.Printf("createing %s", outFile)

	f, err := os.Create(path.Join(generator.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer f.Close()
	writeHeader(f, source)
	writeFunctions(f, source)
	writeStructures(f, source)

	return nil
}

func writeHeader(f io.Writer, sourceList *generator.CodeList) error {
	// headBytes, err := ioutil.ReadFile("head.js.tmpl") // just pass the file name
	// if err != nil {
	// 	log.Errorf("read file error %v", err)
	// 	return err
	// }

	file, err := assets.Assets.Open("/templates/head.js.tmpl")
	defer file.Close()
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}

	headBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}

	headTemplate, err := template.New("structType").Parse(string(headBytes))
	if err != nil {
		log.Errorf("failed to write head with error %v", err)
		return err
	}

	return headTemplate.Execute(f, sourceList)
}

func writeFunctions(wr io.Writer, source *generator.CodeList) {
	for _, function := range source.Functions {
		writeFunction(wr, function)
	}
}

func writeStructures(wr io.Writer, source *generator.CodeList) {
	for _, strct := range source.Structures {
		writeStructure(wr, strct)
	}
}

func createJsType(tp ast.Expr) string {
	switch x := tp.(type) {
	case *ast.Ident:
		return toJsName(x.Name)
	case *ast.SelectorExpr:
		return toJsName(x.Sel.Name)

	case *ast.MapType:
		return "{ [key: " + createJsType(x.Key) + "]: " + createJsType(x.Value) + "}"
	case *ast.StarExpr:
		return "?" + createJsType(x.X)
	case *ast.ArrayType:
		return createJsType(x.Elt) + "[]"
	}

	return ""
}

func writeFunction(wr io.Writer, function generator.FunctionData) {
	// b, err := ioutil.ReadFile("func.js.tmpl") // just pass the file name
	// if err != nil {
	// 	log.Errorf("read file error %v", err)
	// 	return
	// }

	file, err := assets.Assets.Open("/templates/func.js.tmpl")
	defer file.Close()
	if err != nil {
		log.Errorf("read file error %v", err)
		return
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("read file error %v", err)
		return
	}

	t, err := template.New("structType").Parse(string(b))
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	err = t.Execute(wr, createFunction(function))
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func writeStructure(wr io.Writer, structType generator.ExportedStucture) {
	// b, err := ioutil.ReadFile("struct.js.tmpl") // just pass the file name
	// if err != nil {
	// 	log.Errorf("read file error %v", err)
	// 	return
	// }

	file, err := assets.Assets.Open("/templates/struct.js.tmpl")
	defer file.Close()
	if err != nil {
		log.Errorf("read file error %v", err)
		return
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("read file error %v", err)
		return
	}

	t, err := template.New("structType").Parse(string(b))
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	err = t.Execute(wr, createStructure(structType))
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func createFunction(function generator.FunctionData) Function {
	return Function{
		Name:         strcase.ToLowerCamel(function.Name),
		Comments:     function.Comments,
		ReturnType:   toJsName(function.ReturnType),
		Params:       createListOfFields(function.Params),
		Subscription: function.Subscription,
	}
}

func createStructure(structType generator.ExportedStucture) Structure {
	return Structure{
		Name:     structType.Name,
		Field:    createListOfFields(structType.Field),
		Comments: structType.Comments,
	}
}

func createListOfFields(list *ast.FieldList) []Field {
	fields := make([]Field, 0, 100)
	for _, field := range list.List {
		typeName := createJsType(field.Type)
		if typeName == "" {
			typeName = "any"
		}

		// Skip callback type
		if typeName == "JsCallback" || typeName == "EventCallback" {
			continue
		}

		for _, name := range field.Names {
			fieldInfo := Field{
				Name:    name.Name,
				Type:    typeName,
				Comment: getComments(field.Doc),
			}

			fields = append(fields, fieldInfo)
		}
	}

	return fields
}

func toJsName(name string) string {
	switch name {
	case "float32":
		fallthrough
	case "float64":
		fallthrough
	case "int":
		fallthrough
	case "int16":
		fallthrough
	case "int64":
		fallthrough
	case "int8":
		fallthrough
	case "int32":
		return "number"
	case "bool":
		return "boolean"
	}

	return name
}

func getComments(commGroup *ast.CommentGroup) []string {
	comments := make([]string, 0, 6)
	if commGroup != nil {
		for _, comment := range commGroup.List {
			comments = append(comments, strings.TrimLeft(strings.TrimPrefix(comment.Text, "//"), " "))
		}
	}

	return comments
}
