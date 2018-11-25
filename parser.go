package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	log "github.com/sirupsen/logrus"

	"gitlab.vmassive.ru/gocallgen/generator"
	"gitlab.vmassive.ru/gocallgen/gocall"
	"gitlab.vmassive.ru/gocallgen/js"
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

	codeList.Functions = make([]generator.FunctionData, 0, len(codeList.Functions)+8)
	codeList.Structures = make([]generator.ExportedStucture, 0, len(codeList.Functions)+8)

	packageName := "unknown"

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

				case *ast.FuncDecl:
					createFuction(codeList, x)
				}

				return true
			})

		}
	}

	jsGen := js.New(codeList.PathMap.Js, packageName)
	jsGen.CreateCode(codeList)

	goGen := gocall.New(codeList.PathMap.Target, packageName)
	goGen.CreateCode(codeList)

	return nil
}

func createFunctionParameters(funcDecl *ast.FuncDecl) generator.FunctionData {
	comments := getComments(funcDecl.Doc)
	comments, subscription := getSubriptionAnnotatedType(comments)
	comments, returnType := getCallbackAnnotatedType(comments)

	return generator.FunctionData{
		Subscription: subscription,
		Comments:     comments,
		ReturnType:   returnType,
		Name:         funcDecl.Name.Name,
		Params:       funcDecl.Type.Params,
	}
}

func getCallbackAnnotatedType(comments []string) ([]string, string) {
	if comments == nil || len(comments) == 0 {
		return comments, "any"
	}

	lastString := comments[len(comments)-1]

	if strings.HasPrefix(lastString, "@callback:") {
		callbackType := strings.TrimPrefix(lastString, "@callback:")
		otherComments := comments[0 : len(comments)-1]
		return otherComments, strings.TrimSpace(callbackType)
	}

	return comments, "any"
}

func getSubriptionAnnotatedType(comments []string) ([]string, *string) {
	if comments == nil || len(comments) == 0 {
		return comments, nil
	}

	lastString := comments[len(comments)-1]

	if strings.HasPrefix(lastString, "@subsription:") {
		callbackType := strings.TrimPrefix(lastString, "@subsription:")
		otherComments := comments[0 : len(comments)-1]
		subName := strings.TrimSpace(callbackType)
		return otherComments, &subName
	}

	return comments, nil
}

func createFuction(codeList *generator.CodeList, funcDecl *ast.FuncDecl) {
	if !funcDecl.Name.IsExported() {
		return
	}

	function := createFunctionParameters(funcDecl)
	codeList.AddFunction(function)
}

func createType(codeList *generator.CodeList, typeSpec *ast.TypeSpec) {
	if !typeSpec.Name.IsExported() {
		log.Warnf("skipping %s", typeSpec.Name.Name)
		return
	}

	switch x := typeSpec.Type.(type) {
	case *ast.StructType:
		strct := createStructure(x, typeSpec.Name.Name, typeSpec.Doc)
		codeList.AddStructure(strct)

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

func createStructure(structType *ast.StructType, name string, commGroup *ast.CommentGroup) generator.ExportedStucture {

	return generator.ExportedStucture{
		Comments: getComments(commGroup),
		Name:     name,
		Field:    structType.Fields,
	}
}
