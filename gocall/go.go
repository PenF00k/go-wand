package gocall

import (
	"go/ast"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/gocallgen/generator"
)

type Function struct {
	Name         string
	Comments     []string
	ReturnType   string
	Params       []Field
	Subscription *string
}

type Type struct {
	Name  string
	Inner string
}

type Field struct {
	Name       string
	Type       string
	Comment    []string
	Array      bool
	SimpleType string
}

type GoCodeGenerator struct {
	outDirectory string
	packageName  string
}

func New(outDirectory string, packageName string) generator.Generator {
	return &GoCodeGenerator{
		outDirectory: outDirectory,
		packageName:  packageName,
	}
}

func (generator GoCodeGenerator) CreateCode(source *generator.CodeList) error {
	outFile := "call.go"
	log.Printf("createing %s", outFile)

	f, err := os.Create(path.Join(generator.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer f.Close()
	writeHeader(f)
	writeFunctions(f, source)

	return nil
}

func writeFunctions(wr io.Writer, source *generator.CodeList) {
	for _, function := range source.Functions {
		writeFunction(wr, function)
	}
}

func writeFunction(wr io.Writer, function generator.FunctionData) {
	b, err := ioutil.ReadFile("func.go.tmpl") // just pass the file name
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

func createFunction(function generator.FunctionData) Function {
	return Function{
		Name:         function.Name,
		Comments:     function.Comments,
		ReturnType:   function.ReturnType,
		Params:       createListOfFields(function.Params),
		Subscription: function.Subscription,
	}
}

func writeHeader(f io.Writer) error {
	headBytes, err := ioutil.ReadFile("head.go.tmpl") // just pass the file name
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}

	headTemplate, err := template.New("header").Parse(string(headBytes))
	if err != nil {
		log.Errorf("failed to write head with error %v", err)
		return err
	}

	return headTemplate.Execute(f, nil)
}

func createListOfFields(list *ast.FieldList) []Field {
	fields := make([]Field, 0, 100)
	for _, field := range list.List {
		typeName := createType(field.Type)
		if typeName == "" {
			typeName = "any"
		}

		// Skip callback type
		if typeName == "JsCallback" {
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

func getComments(commGroup *ast.CommentGroup) []string {
	comments := make([]string, 0, 6)
	if commGroup != nil {
		for _, comment := range commGroup.List {
			comments = append(comments, strings.TrimLeft(strings.TrimPrefix(comment.Text, "//"), " "))
		}
	}

	return comments
}

func createType(tp ast.Expr) string {
	switch x := tp.(type) {
	case *ast.Ident:
		return toTypeName(x.Name)

	case *ast.MapType:
		return "map[" + createType(x.Key) + "]" + createType(x.Value)
	case *ast.StarExpr:
		return "*" + createType(x.X)
	case *ast.ArrayType:
		return "[]" + createType(x.Elt)
	}

	return ""
}

func toTypeName(name string) string {
	return name
}
