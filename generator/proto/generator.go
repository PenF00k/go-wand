package proto

import (
	"gitlab.vmassive.ru/wand/adapter"
	"gitlab.vmassive.ru/wand/util"
	"io"
	"os"
	"path"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gitlab.vmassive.ru/wand/generator"
)

type CodeGenerator struct {
	outDirectory string
	packageName  string
	Adapter      *adapter.Adapter
	CodeList     *generator.CodeList
}

func New(outDirectory string, packageName string, adapter *adapter.Adapter, codeList *generator.CodeList) generator.Generator {
	return &CodeGenerator{
		outDirectory: outDirectory,
		packageName:  packageName,
		Adapter:      adapter,
		CodeList:     codeList,
	}
}

func (gen CodeGenerator) CreateCode() error {
	return gen.writeGeneral()
}

func (gen CodeGenerator) writeGeneral() error {
	outFile := gen.packageName + ".proto"
	log.Printf("creating %s", outFile)

	f, err := os.Create(path.Join(gen.outDirectory, outFile))
	if err != nil {
		log.Errorf("failed to create file %s", outFile)
		return err
	}

	defer util.Close(f, outFile)

	gen.writeHeader(f)
	gen.writeFunctionFieldStructures(f)
	gen.writeStructures(f)

	return nil
}

func (gen CodeGenerator) writeHeader(f io.Writer) {
	tpath := "templates/head.proto.tmpl"
	base := path.Base(tpath)

	headTemplate, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed to write head with error %v", err)
	}

	if err := headTemplate.Execute(f, gen); err != nil {
		log.Errorf("failed to execute head.proto.tmpl with error %v", err)
	}
}

func (gen CodeGenerator) writeFunctionFieldStructures(wr io.Writer) {
	for _, fn := range gen.Adapter.Functions {
		writeFunctionFieldStructure(wr, fn)
	}
}

func writeFunctionFieldStructure(wr io.Writer, functionData adapter.Function) {
	tpath := "templates/struct.proto.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	s := createFunctionArgs(functionData)
	err = t.Execute(wr, s)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func createFunctionArgs(functionData adapter.Function) TemplateStructData {
	types := make([]TemplateProtoTypeData, 0, len(functionData.Args))

	for i, v := range functionData.Args {
		typ := createFunctionArg(v, i+1)
		types = append(types, typ)
	}
	return TemplateStructData{
		MessageName: functionData.FunctionName + "Args",
		Types:       types,
	}
}

func createFunctionArg(functionData adapter.Field, n int) TemplateProtoTypeData {
	return TemplateProtoTypeData{
		Name:        functionData.Name,
		Type:        functionData.Type,
		FieldNumber: n,
	}
}

func (gen CodeGenerator) writeStructures(wr io.Writer) {
	for _, strct := range gen.Adapter.Structures {
		writeStructure(wr, strct)
	}
}

func writeStructure(wr io.Writer, structType adapter.Struct) {
	tpath := "templates/struct.proto.tmpl"
	base := path.Base(tpath)

	t, err := template.New(base).ParseFiles(tpath)
	if err != nil {
		log.Errorf("failed with error %v", err)
		return
	}

	s := createStructure(structType)
	err = t.Execute(wr, s)
	if err != nil {
		log.Errorf("template failed with error %v", err)
	}
}

func createStructure(structData adapter.Struct) TemplateStructData {
	types := make([]TemplateProtoTypeData, 0, len(structData.Fields))

	for i, v := range structData.Fields {
		typ := createStructureField(v, i+1)
		types = append(types, typ)
	}
	return TemplateStructData{
		MessageName: string(structData.Name),
		Types:       types,
	}
}

func createStructureField(fieldData adapter.Field, n int) TemplateProtoTypeData {
	typ := fieldData.Type
	for typ.Pointer != nil {
		it := typ.Pointer.InnerType
		if it.IsPrimitive {
			break
		} else {
			typ = typ.Pointer.InnerType
		}
	}

	return TemplateProtoTypeData{
		Name:        fieldData.Name,
		Type:        typ,
		FieldNumber: n,
	}
}

//func createProtoType(tp ast.Expr) string {
//	switch x := tp.(type) {
//	case *ast.Ident:
//		return toProtoName(x.Name)
//	case *ast.SelectorExpr:
//		return toProtoName(x.Sel.Name)
//
//	case *ast.MapType:
//		return createMapProtoType(x)
//
//	case *ast.StarExpr:
//		return createProtoType(x.X)
//
//	case *ast.ArrayType:
//		arrElType := createProtoType(x.Elt)
//		if arrElType == "byte" {
//			return toProtoName("[]byte")
//		}
//
//		eltType := x.Elt
//
//		//proto restrictions: Map fields cannot be repeated.
//		if _, ok := eltType.(*ast.MapType); ok {
//			return ""
//		}
//
//		return "repeated " + createProtoType(eltType)
//	}
//
//	return ""
//}
//
//func createMapProtoType(m *ast.MapType) string {
//	//proto restrictions on key type
//	keyType := createProtoType(m.Key)
//	if keyType != "int32" && keyType != "int64" && keyType != "bool" && keyType != "string" {
//		return ""
//	}
//
//	//proto restrictions on value type
//	valueType := createProtoType(m.Value)
//	if valueType == "map" {
//		return ""
//	}
//
//	return fmt.Sprintf("map<%s, %s>", keyType, valueType)
//}

//func createStructure(structType generator.StructData) Structure {
//	return Structure{
//		Name:       structType.Name,
//		Field:      createListOfFields(structType.Field),
//		Comments:   structType.Comments,
//		Annotation: structType.Annotation,
//	}
//}
//
//func createListOfFields(list *ast.FieldList) []adapter.Field {
//	fields := make([]Field, 0, 100)
//	for i, field := range list.List {
//		typeName := createProtoType(field.Type)
//		// Skip empty and callback types
//		if typeName == "" || typeName == "JsCallback" || typeName == "EventCallback" {
//			continue
//		}
//
//		for _, name := range field.Names {
//			fieldInfo := Field{
//				Name:        name.Name,
//				Type:        typeName,
//				Comment:     getComments(field.Doc),
//				FieldNumber: i + 1,
//				//FieldNumber: getFieldNumber(field, name.Name), //TODO
//			}
//
//			fields = append(fields, fieldInfo)
//		}
//	}
//
//	return fields
//}
//
//func getFieldNumber(field *ast.Field, name string) (fieldNumber int) {
//	if field.Tag == nil {
//		return
//	}
//
//	tagString := field.Tag.Value
//
//	if len(tagString) < 2 {
//		return
//	}
//
//	var trimmed = tagString[1 : len(tagString)-1]
//
//	tags := strings.Split(trimmed, ",")
//
//	for _, t := range tags {
//		if !strings.HasPrefix(t, "proto") {
//			return
//		}
//
//		s := strings.Split(t, ":")
//
//		if len(s) <= 1 {
//			return
//		}
//
//		number := s[1]
//
//		trimmed = number[1 : len(number)-1]
//
//		fieldNumber, _ = strconv.Atoi(trimmed)
//	}
//
//	if fieldNumber == 0 {
//		log.Error(errors.New(fmt.Sprintf("you must provide an int proto tag (fieldNumber) for field %s", name)))
//		//panic("you must provide a ")
//	}
//
//	return
//}

//func toProtoName(name string) string {
//	switch name {
//	case "float32":
//		return "float"
//	case "float64":
//		return "double"
//	case "int":
//		fallthrough
//	case "int8":
//		fallthrough
//	case "int16":
//		fallthrough
//	case "int32":
//		return "int32"
//	case "int64":
//		return "int64"
//	case "bool":
//		return "bool"
//	case "[]byte":
//		return "bytes"
//	case "Time":
//		return "google.protobuf.Timestamp"
//	}
//
//	return name
//}
//
//func toProtoName(name string) string {
//	switch name {
//	case "float32":
//		return "google.protobuf.FloatValue"
//	case "float64":
//		return "google.protobuf.DoubleValue"
//	case "int":
//		fallthrough
//	case "int8":
//		fallthrough
//	case "int16":
//		fallthrough
//	case "int32":
//		return "google.protobuf.Int32Value"
//	case "int64":
//		return "google.protobuf.Int64Value"
//	case "bool":
//		return "google.protobuf.BoolValue"
//	case "string":
//		return "google.protobuf.StringValue"
//	case "[]byte":
//		return "google.protobuf.BytesValue"
//	case "Time":
//		return "google.protobuf.Timestamp"
//	}
//
//	return name
//}
//
//func getComments(commGroup *ast.CommentGroup) []string {
//	comments := make([]string, 0, 6)
//	if commGroup != nil {
//		for _, comment := range commGroup.List {
//			comments = append(comments, strings.TrimLeft(strings.TrimPrefix(comment.Text, "//"), " "))
//		}
//	}
//
//	return comments
//}
