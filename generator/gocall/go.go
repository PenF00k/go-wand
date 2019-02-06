package gocall

import (
	"go/ast"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/wand/generator"
)

type Function struct {
	Name             string
	Comments         []string
	ReturnType       string
	Params           []Field
	Subscription     *string
	Package          string
	ProtoPackageName string
}

type Type struct {
	Name       string
	Map        bool
	Array      bool
	SimpleType string
	Pointer    bool
	InnerType  *Type
	Object     bool
}

type Field struct {
	Name       string
	Type       string
	Comment    []string
	RichType   Type
	Array      bool
	SimpleType string
	Package    string
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
	err := generator.writeCode(source)
	if err != nil {
		return err
	}

	if !source.Dev {
		cmd := exec.Command("go", "fmt")
		cmd.Dir = generator.outDirectory

		err := cmd.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func writeMap(f io.Writer, source *generator.CodeList) {
	tpath := "templates/callmap.go.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed to write head with error %v", err)
	}

	err = t.Execute(f, source)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func (generator GoCodeGenerator) writeCode(source *generator.CodeList) error {
	outFile := "call.go"
	log.Printf("createing %s", outFile)

	f, err := os.Create(path.Join(generator.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer f.Close()
	writeHeader(f, source)
	writeMap(f, source)
	writeFunctions(f, generator.packageName, source)
	//writePureFunctions(f, generator.packageName, source)
	return nil
}

func writeFunctions(wr io.Writer, pack string, source *generator.CodeList) {
	for _, function := range source.Functions {
		writeFunction(wr, pack, function, source.ProtoPackageName)
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

func writeFunction(wr io.Writer, pack string, function generator.FunctionData, protoPackageName string) {
	tpath := "templates/func.go.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	err = t.Execute(wr, createFunction(pack, function, protoPackageName))
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func createFunction(pack string, function generator.FunctionData, protoPackageName string) Function {
	return Function{
		Name:             function.Name,
		Comments:         function.Comments,
		ReturnType:       function.ReturnType,
		Params:           createListOfFields(function.Params, pack),
		Subscription:     function.Subscription,
		Package:          pack,
		ProtoPackageName: protoPackageName,
	}
}

func writeHeader(f io.Writer, sourceList *generator.CodeList) {
	tpath := "templates/head.go.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	err = t.Execute(f, sourceList)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func createListOfFields(list *ast.FieldList, pack string) []Field {
	fields := make([]Field, 0, 100)
	for _, field := range list.List {

		typeName := createType(field.Type)
		if typeName == "" {
			typeName = "interface{}"
		}

		richType := createRichType(field.Type)

		// Skip callback type
		if typeName == "JsCallback" || typeName == "EventCallback" {
			continue
		}

		for _, name := range field.Names {
			fieldInfo := Field{
				Name:     name.Name,
				Type:     typeName,
				Comment:  getComments(field.Doc),
				RichType: richType,
				Package:  pack,
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

func isObject(name string) bool {
	return unicode.IsUpper([]rune(name)[0])
}

func createRichType(tp ast.Expr) Type {
	Map := false
	Array := false
	SimpleType := "interface{}"
	Pointer := false

	var InnerType *Type

	switch x := tp.(type) {
	case *ast.Ident:
		SimpleType = toTypeName(x.Name)

	case *ast.MapType:
		// return "map[" + createType(x.Key) + "]" + createType(x.Value)
		Map = true
		SimpleType = createType(x.Key)
		inner := createRichType(x.Value)
		InnerType = &inner

	case *ast.StarExpr:
		SimpleType = createType(x.X)
		Pointer = true

	case *ast.ArrayType:
		SimpleType = createType(x.Elt)
		Array = true
	}

	object := isObject(SimpleType)

	return Type{
		SimpleType: SimpleType,
		Array:      Array,
		Map:        Map,
		Pointer:    Pointer,
		InnerType:  InnerType,
		Object:     object,
	}
}

func createType(tp ast.Expr) string {
	switch x := tp.(type) {
	case *ast.Ident:
		return toTypeName(x.Name)

	case *ast.SelectorExpr:
		return toTypeName(x.Sel.Name)

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
