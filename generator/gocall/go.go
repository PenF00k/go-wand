package gocall

import (
	"github.com/iancoleman/strcase"
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
	ReturnType       *ReturnType
	Params           []Field
	Subscription     *Subscription
	Package          string
	ProtoPackageName string
}

type ReturnType struct {
	Name string
	//EventName string
	Params []Field
}

type Subscription struct {
	Name      string
	EventName string
	Params    []Field
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

func (f Field) GetUpperCamelCaseName(prefix string, target string) string {
	n := prefix + strcase.ToCamel(f.Name)

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

	f := createFunction(pack, function, protoPackageName)
	err = t.Execute(wr, f)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func createFunction(pack string, function generator.FunctionData, protoPackageName string) Function {
	// function.Name == "SubscribeToMyProtoData"
	params := createListOfFields(function.Params, pack)
	size := len(params)

	var sub *Subscription
	if function.Subscription != nil && size > 0 {
		sub = createSubscription(pack, params, function)

		// trim last param (onEvent)
		params = params[:size-1]
	}

	var returnType *ReturnType
	if function.ReturnType != nil && size > 0 {
		returnType = createReturnType(pack, params, function)

		// trim last param (onEvent)
		params = params[:size-1]
	}
	//function.ReturnType

	return Function{
		Name:             function.Name,
		Comments:         function.Comments,
		ReturnType:       returnType,
		Params:           params,
		Subscription:     sub,
		Package:          pack,
		ProtoPackageName: protoPackageName,
	}
}

func createReturnType(pack string, params []Field, function generator.FunctionData) *ReturnType {
	return nil //TODO
}

func createSubscription(pack string, params []Field, function generator.FunctionData) *Subscription {
	var sub *Subscription
	eventParam := function.Params.List[len(params)-1]

	if tName, ok := eventParam.Type.(*ast.FuncType); ok {
		eventParams := tName.Params.List
		paramTypeName := ""

		if len(eventParams) > 0 && len(eventParams[0].Names) > 0 {
			if paramType, ok := eventParams[0].Type.(*ast.Ident); ok {
				paramTypeName = paramType.Name
			}
		}

		params := createListOfFields(function.Subscription.Field, pack)

		sub = &Subscription{
			Name:      function.Subscription.Name,
			EventName: paramTypeName,
			Params:    params,
		}
	}

	return sub
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
	if list == nil {
		return fields
	}
	for _, field := range list.List {

		typeName := createType(field.Type)
		if typeName == "" {
			typeName = "interface{}"
		}

		// Skip callback type
		if typeName == "JsCallback" || typeName == "EventCallback" {
			continue
		}

		richType := createRichType(field.Type)

		var funcParams []Field
		if fTyre, ok := field.Type.(*ast.FuncType); ok {
			funcParams = createListOfFields(fTyre.Params, pack)
		}

		for _, name := range field.Names {
			fieldInfo := Field{
				Name:           name.Name,
				Type:           typeName,
				Comment:        getComments(field.Doc),
				RichType:       richType,
				Package:        pack,
				FunctionParams: funcParams,
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
