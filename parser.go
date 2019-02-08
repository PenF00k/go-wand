package main

import (
	"bytes"
	"fmt"
	"gitlab.vmassive.ru/wand/caster"
	"gitlab.vmassive.ru/wand/generator/gocall"
	"gitlab.vmassive.ru/wand/proto"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"gitlab.vmassive.ru/wand/generator"
)

func restoreCommentForType(commentMap *ast.CommentMap, fileSet *token.FileSet, typeSpec *ast.TypeSpec) {
	var list *[]*ast.Comment

	for _, comm := range commentMap.Comments() {

		structStart := fileSet.Position(typeSpec.Pos())
		commEnd := fileSet.Position(comm.End())

		if structStart.Line == commEnd.Line+1 {
			list = &comm.List
			break
		}
	}

	if list == nil {
		return
	}

	typeSpec.Doc = &ast.CommentGroup{List: *list}
}

// ParseFile - Parse file
func Parse(codeList *generator.CodeList) error {
	src := codeList.PathMap.Source
	log.Printf("parsing files in %s", src)

	fset := token.NewFileSet() // positions are relative to fset

	pkgs, err := parser.ParseDir(fset, src, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		log.Errorf("parse file error %s : %v", src, err)
		return err
	}

	oldState := *codeList
	codeList.Functions = make([]generator.FunctionData, 0, len(codeList.Functions)+8)
	codeList.Structures = make([]generator.ExportedStucture, 0, len(codeList.Functions)+8)
	//codeList.Pure = make([]generator.FunctionData, 0, len(codeList.Functions)+8)
	packageName := "unknown"

	packageName = parseStructures(codeList, pkgs, fset)
	packageName = parseFunctions(codeList, pkgs)

	codeList.PackageName = packageName

	if hasChanges(codeList, &oldState) {
		//jsGen := js.New(codeList.PathMap.Js, packageName)
		//jsGen.CreateCode(codeList)

		protoGen := proto.New(codeList.PathMap.Proto, packageName)
		protoGen.CreateCode(codeList)

		generateGoFilesFromProto(codeList, packageName)

		goGen := gocall.New(codeList.PathMap.Target, packageName)
		goGen.CreateCode(codeList)
	} else {
		log.Printf("no changes, skipping")
	}

	return nil
}

func parseStructures(codeList *generator.CodeList, pkgs map[string]*ast.Package, fset *token.FileSet) (packageName string) {
	for name, pkg := range pkgs {
		packageName = name

		for name, file := range pkg.Files {
			log.Printf("file %s", name)
			cmap := ast.NewCommentMap(fset, file, file.Comments)
			file.Comments = cmap.Comments()

			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.Package:
					packageName = x.Name

				case *ast.TypeSpec:
					restoreCommentForType(&cmap, fset, x)
					createType(codeList, x)
				}

				return true
			})
		}
	}
	return
}

func parseFunctions(codeList *generator.CodeList, pkgs map[string]*ast.Package) (packageName string) {
	for name, pkg := range pkgs {
		packageName = name

		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				if x, ok := n.(*ast.FuncDecl); ok {
					createFunction(codeList, x)
				}

				return true
			})
		}
	}
	return
}

func hasChanges(newState *generator.CodeList, oldState *generator.CodeList) bool {
	if len(newState.Functions) != len(oldState.Functions) {
		return true
	}

	if len(newState.Structures) != len(oldState.Structures) {
		return true
	}

	//if len(newState.Pure) != len(oldState.Pure) {
	//	return true
	//}

	if reflect.DeepEqual(newState.Functions, oldState.Functions) &&
	//reflect.DeepEqual(newState.Pure, oldState.Pure) &&
		reflect.DeepEqual(newState.Structures, oldState.Structures) {
		return true
	}

	return false
}

func createFunctionParameters(codeList *generator.CodeList, funcDecl *ast.FuncDecl) *generator.FunctionData {
	comments := getComments(funcDecl.Doc)
	returnType, isSubscription := getReturnDataType(codeList, funcDecl.Type)
	//returnType := getCallbackType(funcDecl)

	return &generator.FunctionData{
		Subscription: isSubscription,
		Comments:     comments,
		ReturnType:   returnType,
		Name:         funcDecl.Name.Name,
		Params:       funcDecl.Type.Params,
		CallName:     funcDecl.Name.Name,
	}
}

func getCallbackType(funcDecl *ast.FuncDecl) *generator.ExportedStucture {
	resultTypes := funcDecl.Type.Results

	if resultTypes == nil {
		return nil
	}

	if resultTypes.List == nil || len(resultTypes.List) == 0 {
		return nil
	}

	var returnedTypeField *ast.FieldList
	if funcArgType, ok := resultTypes.List[0].Type.(*ast.Ident); ok {
		if funcArgType.Obj != nil {
			if decl, ok := funcArgType.Obj.Decl.(*ast.TypeSpec); ok {
				if str, ok := decl.Type.(*ast.StructType); ok {
					returnedTypeField = str.Fields
				}
			}
		}
	}

	return &generator.ExportedStucture{
		Name:  caster.GetFullGoTypeAsString(resultTypes.List[0].Type, ""),
		Field: returnedTypeField,
	}
}

var annotationList = []string{
	"subscription",
	"get",
	"update",
	"callback",
	"ignore",
}

func getAnnotation(comments []string) ([]string, *generator.Annotation) {
	if comments == nil || len(comments) == 0 {
		return comments, nil
	}

	lastString := comments[len(comments)-1]

	for _, annotationName := range annotationList {
		keyword := "@" + annotationName + ":"
		if strings.HasPrefix(lastString, keyword) {

			callbackType := strings.TrimPrefix(lastString, keyword)
			otherComments := comments[0 : len(comments)-1]
			value := strings.TrimSpace(callbackType)

			annotation := generator.Annotation{
				Name:  annotationName,
				Value: value,
			}

			return otherComments, &annotation
		}
	}

	return comments, nil
}

func GetAnnotations(comments []string) ([]string, []generator.Annotation) {
	annotations := make([]generator.Annotation, 0, 2)
	outList, annotation := getAnnotation(comments)
	if annotation != nil {
		outList, annotationList := GetAnnotations(outList)
		list := append(annotationList, *annotation)

		return outList, list
	}

	return outList, annotations
}

func getReturnDataType(codeList *generator.CodeList, funcTypes *ast.FuncType) (returnType *generator.ExportedStucture, isSubscription bool) {
	resultTypes := funcTypes.Results
	paramTypes := funcTypes.Params

	if resultTypes == nil || resultTypes.List == nil || len(resultTypes.List) == 0 {
		return
	}

	for _, t := range resultTypes.List {
		if t != nil {
			if selType, ok := t.Type.(*ast.SelectorExpr); ok {
				if x, ok := selType.X.(*ast.Ident); ok {
					if x.Name == "goapi" && selType.Sel.Name == "Subscription" {
						isSubscription = true
						break
					}
				}
			}
		}
	}

	if !isSubscription {
		returnType = &generator.ExportedStucture{
			Name:  caster.GetFullGoTypeAsString(resultTypes.List[0].Type, ""),
			Field: createReturnTypeFieldsForFunction(resultTypes.List),
		}
		return
	}

	log.Infof("isSubscription %v", isSubscription)

	for _, paramFields := range paramTypes.List {
		if functionType, ok := paramFields.Type.(*ast.FuncType); ok {
			for _, n := range paramFields.Names {
				if n.Name == "onEvent" {
					params := functionType.Params.List
					if len(params) > 0 && len(params[0].Names) > 0 {
						var fields *ast.FieldList
						if subEventTypeName := createReturnTypeFieldsForSubscription(params); subEventTypeName != nil {
							for _, s := range codeList.Structures {
								if s.Name == *subEventTypeName {
									fields = s.Field
									break
								}
							}
						}

						returnType = &generator.ExportedStucture{
							Name:  params[0].Names[0].Name,
							Field: fields,
						}
					}
				}
			}
		}
	}

	return
}

func createReturnTypeFieldsForSubscription(params []*ast.Field) *string {
	switch t := params[0].Type.(type) {
	case *ast.StarExpr:
		if t, ok := t.X.(*ast.Ident); ok {
			return &t.Name
		}
	}
	return nil
}

func createReturnTypeFieldsForFunction(params []*ast.Field) (subField *ast.FieldList) {
	switch t := params[0].Type.(type) {
	case *ast.Ident:
		if t.Obj != nil {
			if decl, ok := t.Obj.Decl.(*ast.TypeSpec); ok {
				if t, ok := decl.Type.(*ast.StructType); ok {
					return t.Fields
				}
			}
		}
	case *ast.StarExpr:
		if t, ok := t.X.(*ast.StructType); ok {
			return t.Fields
		}
	}
	return nil
}

func createFunction(codeList *generator.CodeList, funcDecl *ast.FuncDecl) {
	if !funcDecl.Name.IsExported() {
		return
	}

	f := createFunctionParameters(codeList, funcDecl)
	codeList.AddFunction(*f)
}

func createType(codeList *generator.CodeList, typeSpec *ast.TypeSpec) {
	if !typeSpec.Name.IsExported() {
		log.Warnf("skipping %s", typeSpec.Name.Name)
		return
	}

	switch x := typeSpec.Type.(type) {
	case *ast.StructType:
		if strct, ok := createStructure(x, typeSpec.Name.Name, typeSpec.Doc); ok {
			codeList.AddStructure(strct)
		}

	case *ast.InterfaceType:
	}
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

func createStructure(structType *ast.StructType, name string, commGroup *ast.CommentGroup) (generator.ExportedStucture, bool) {
	comments := getComments(commGroup)
	comments, annotations := GetAnnotations(comments)

	if containsAnnotation("ignore", annotations) {
		log.Infof("ignoring %s ", name)
		return generator.ExportedStucture{}, false
	}

	return generator.ExportedStucture{
		Comments:   comments,
		Name:       name,
		Field:      structType.Fields,
		Annotation: annotations,
	}, true
}

func containsAnnotation(name string, list []generator.Annotation) bool {
	for _, item := range list {
		if item.Name == name {
			return true
		}
	}

	return false
}

func generateGoFilesFromProto(codeList *generator.CodeList, packageName string) {
	cmd := exec.Command("protoc", "--go_out=.", packageName+".proto")
	cmd.Dir = codeList.PathMap.Proto

	log.Infof("executing protoc in dir %v, packageName = %v", cmd.Dir, packageName)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
}
