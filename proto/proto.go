package proto

import (
	"errors"
	"fmt"
	"gitlab.vmassive.ru/wand/assets"
	"go/ast"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/wand/generator"
)

type Field struct {
	Name        string
	Type        string
	Comment     []string
	FieldNumber int
}

type Function struct {
	Name         string
	Comments     []string
	ReturnType   string
	Params       []Field
	Subscription *string
}

type Structure struct {
	Comments   []string
	Name       string
	Field      []Field
	Annotation []generator.Annotation
}

type ProtoCodeGenerator struct {
	outDirectory string
	packageName  string
}

func New(outDirectory string, packageName string) generator.Generator {
	return &ProtoCodeGenerator{
		outDirectory: outDirectory,
		packageName:  packageName,
	}
}

func (generator ProtoCodeGenerator) CreateCode(source *generator.CodeList) error {
	err := generator.writeGeneral(source)
	if err != nil {
		return err
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

func (generator ProtoCodeGenerator) writeGeneral(source *generator.CodeList) error {
	outFile := generator.packageName + ".proto"
	log.Printf("createing %s", outFile)

	f, err := os.Create(path.Join(generator.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer f.Close()
	writeHeader(f, source)
	writeStructures(f, source)

	return nil
}

func writeHeader(f io.Writer, sourceList *generator.CodeList) error {
	file, err := assets.Assets.Open("/templates/head.proto.tmpl")
	if err != nil {
		log.Errorf("read file error %v", err)
		return err
	}
	defer file.Close()

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

func writeStructures(wr io.Writer, source *generator.CodeList) {
	for _, strct := range source.Structures {
		writeStructure(wr, strct)
	}
}

func createProtoType(tp ast.Expr) string {
	switch x := tp.(type) {
	case *ast.Ident:
		return toProtoName(x.Name)
	case *ast.SelectorExpr:
		return toProtoName(x.Sel.Name)

	case *ast.MapType:
		return createMapProtoType(x)

	case *ast.StarExpr:
		return createProtoType(x.X)

	case *ast.ArrayType:
		arrElType := createProtoType(x.Elt)
		if arrElType == "byte" {
			return toProtoName("[]byte")
		}

		eltType := x.Elt

		//proto restrictions: Map fields cannot be repeated.
		if _, ok := eltType.(*ast.MapType); ok {
			return ""
		}

		return "repeated " + createProtoType(eltType)
	}

	return ""
}

func createMapProtoType(m *ast.MapType) string {
	//proto restrictions on key type
	keyType := createProtoType(m.Key)
	if keyType != "int32" && keyType != "int64" && keyType != "bool" && keyType != "string" {
		return ""
	}

	//proto restrictions on value type
	valueType := createProtoType(m.Value)
	if valueType == "map" {
		return ""
	}

	return fmt.Sprintf("map<%s, %s>", keyType, valueType)
}

func writeStructure(wr io.Writer, structType generator.ExportedStucture) {
	file, err := assets.Assets.Open("/templates/struct.proto.tmpl")
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

func createStructure(structType generator.ExportedStucture) Structure {
	return Structure{
		Name:       structType.Name,
		Field:      createListOfFields(structType),
		Comments:   structType.Comments,
		Annotation: structType.Annotation,
	}
}

func createListOfFields(gen generator.ExportedStucture) []Field {
	list := gen.Field
	fields := make([]Field, 0, 100)
	for i, field := range list.List {
		typeName := createProtoType(field.Type)
		// Skip empty and callback types
		if typeName == "" || typeName == "JsCallback" || typeName == "EventCallback" {
			continue
		}

		for _, name := range field.Names {
			fieldInfo := Field{
				Name:        name.Name,
				Type:        typeName,
				Comment:     getComments(field.Doc),
				FieldNumber: i + 1,
				//FieldNumber: getFieldNumber(field, name.Name), //TODO
			}

			fields = append(fields, fieldInfo)
		}
	}

	return fields
}

func getFieldNumber(field *ast.Field, name string) (fieldNumber int) {
	if field.Tag == nil {
		return
	}

	tagString := field.Tag.Value

	if len(tagString) < 2 {
		return
	}

	var trimmed = tagString[1 : len(tagString)-1]

	tags := strings.Split(trimmed, ",")

	for _, t := range tags {
		if !strings.HasPrefix(t, "proto") {
			return
		}

		s := strings.Split(t, ":")

		if len(s) <= 1 {
			return
		}

		number := s[1]

		trimmed = number[1 : len(number)-1]

		fieldNumber, _ = strconv.Atoi(trimmed)
	}

	if fieldNumber == 0 {
		log.Error(errors.New(fmt.Sprintf("you must provide an int proto tag (fieldNumber) for field %s", name)))
		//panic("you must provide a ")
	}

	return
}

func toProtoName(name string) string {
	switch name {
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "int":
		fallthrough
	case "int8":
		fallthrough
	case "int16":
		fallthrough
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "bool":
		return "bool"
	case "[]byte":
		return "bytes"
	case "Time":
		return "google.protobuf.Timestamp"
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
