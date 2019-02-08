package gocall

import (
	"fmt"
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

	t, err := template.New(base).Funcs(template.FuncMap{
		"format": func(format string, a ...interface{}) string {
			return fmt.Sprintf(format, a...)
		},
	}).ParseFiles(tpath)
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

	var returnType *ReturnType
	if function.ReturnType != nil {
		returnType = createReturnType(pack, size, function)

		// returnType.Name == "MyProtoCall"
		// trim last param (onEvent)
		if function.Subscription && size > 0 {
			params = params[:size-1]
		}
	}
	//function.ReturnType

	return Function{
		Name:             function.Name,
		Comments:         function.Comments,
		ReturnType:       returnType,
		Params:           params,
		Subscription:     function.Subscription,
		Package:          pack,
		ProtoPackageName: protoPackageName,
	}
}

func createReturnType(pack string, paramsNumber int, function generator.FunctionData) *ReturnType {
	var typ *ReturnType
	paramTypeName := ""
	isPointer := false
	//var returnParam *ast.Field
	//var params []Field

	if function.Subscription {
		returnParam := function.Params.List[paramsNumber-1]

		if tName, ok := returnParam.Type.(*ast.FuncType); ok {
			eventParams := tName.Params.List

			if len(eventParams) > 0 && len(eventParams[0].Names) > 0 {
				switch t := eventParams[0].Type.(type) {
				case *ast.Ident:
					paramTypeName = t.Name
				case *ast.StarExpr:
					isPointer = true
					if tt, ok := t.X.(*ast.Ident); ok {
						paramTypeName = tt.Name
					}
				}
			}

		}
	} else if function.ReturnType != nil && function.ReturnType.Field != nil {
		paramTypeName = function.ReturnType.Name
	}

	params := createListOfFields(function.ReturnType.Field, pack)

	typ = &ReturnType{
		Name:      function.ReturnType.Name,
		EventName: paramTypeName,
		Params:    params,
		IsPointer: isPointer,
	}

	return typ
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
	//isPod := true

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
		Primitive:  GetPrimitive(SimpleType),
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

func GetPrimitive(name string) PrimitiveType {
	var isPrimitive bool
	var wrapperTypeName string

	switch name {
	case "float32":
		isPrimitive = true
		wrapperTypeName = "FloatValue"
	case "float64":
		isPrimitive = true
		wrapperTypeName = "DoubleValue"
	case "int":
		fallthrough
	case "int8":
		fallthrough
	case "int16":
		fallthrough
	case "int32":
		isPrimitive = true
		wrapperTypeName = "Int32Value"
	case "int64":
		isPrimitive = true
		wrapperTypeName = "Int64Value"
	case "bool":
		isPrimitive = true
		wrapperTypeName = "BoolValue"
	case "string":
		isPrimitive = true
		wrapperTypeName = "StringValue"
	case "[]byte":
		isPrimitive = true
		wrapperTypeName = "BytesValue"
	}

	return PrimitiveType{
		IsPrimitive:     isPrimitive,
		WrapperTypeName: wrapperTypeName,
	}
}
