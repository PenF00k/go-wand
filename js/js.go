package js

import (
	"go/ast"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/gocallgen/generator"
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
}

func New(outDirectory string, packageName string) generator.Generator {
	return &JsCodeGenerator{
		outDirectory: outDirectory,
		packageName:  packageName,
	}
}

func (generator JsCodeGenerator) CreateCode(source *generator.CodeList) error {
	outFile := generator.packageName + ".js"
	log.Printf("createing %s", outFile)

	f, err := os.Create(path.Join(generator.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer f.Close()
	writeHeader(f)
	writeFunctions(f, source)
	writeStructures(f, source)

	return nil
}

func writeHeader(f io.Writer) error {
	headBytes, err := ioutil.ReadFile("head.js.tmpl") // just pass the file name
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}

	headTemplate, err := template.New("structType").Parse(string(headBytes))
	if err != nil {
		log.Errorf("failed to write head with error %v", err)
		return err
	}

	return headTemplate.Execute(f, nil)
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
	b, err := ioutil.ReadFile("func.js.tmpl") // just pass the file name
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
	b, err := ioutil.ReadFile("struct.js.tmpl") // just pass the file name
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
		Name:         function.Name,
		Comments:     function.Comments,
		ReturnType:   function.ReturnType,
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
