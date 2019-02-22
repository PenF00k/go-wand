package parser

import (
	"bytes"
	"fmt"
	"gitlab.vmassive.ru/wand/generator"
	"gitlab.vmassive.ru/wand/generator/dart"
	"gitlab.vmassive.ru/wand/generator/gocall"
	"gitlab.vmassive.ru/wand/generator/proto"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ParseFile - Parse file
func Parse(codeList *generator.CodeList) {
	src := codeList.PathMap.Source
	log.Printf("parsing files in %s", src)

	fset := token.NewFileSet() // positions are relative to fset

	pkgs, err := parser.ParseDir(fset, src, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		log.Errorf("parse file error %s : %v", src, err)
	}

	oldState := *codeList
	codeList.FunctionData = make([]generator.FunctionData, 0, len(codeList.FunctionData)+8)
	codeList.StructData = make([]generator.StructData, 0, len(codeList.StructData)+8)
	packageName := "unknown"

	packageName = parseStructures(codeList, pkgs, fset)
	packageName = parseFunctions(codeList, pkgs, fset)

	codeList.PackageMap.PackageName = packageName

	if hasChanges(codeList, &oldState) {
		//jsGen := js.New(codeList.PathMap.Js, packageName)
		//jsGen.CreateCode(codeList)

		ad := Adopt(codeList)

		protoGen := proto.New(codeList.PathMap.Proto, packageName, ad, codeList)
		if err := protoGen.CreateCode(); err != nil {
			log.Errorf("error while generating proto code")
		}

		generateGoFilesFromProto(codeList, packageName)
		generateDartFilesFromProto(codeList, packageName)
		//generateJavaFilesFromProto(codeList, packageName)

		goGen := gocall.New(codeList.PathMap.Target, packageName, ad, codeList)
		if err := goGen.CreateCode(); err != nil {
			log.Errorf("error while generating go code")
		}

		dartGen := dart.New(codeList.PathMap.FlutterGenerated, packageName, ad, codeList)
		if err := dartGen.CreateCode(); err != nil {
			log.Errorf("error while generating dart code")
		}
	} else {
		log.Printf("no changes, skipping")
	}
}

func parseStructures(codeList *generator.CodeList, pkgs map[string]*ast.Package, fset *token.FileSet) (packageName string) {
	for name, pkg := range pkgs {
		packageName = name

		for _, file := range pkg.Files {
			//log.Printf("file %s", name)
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

func createType(codeList *generator.CodeList, typeSpec *ast.TypeSpec) {
	if !typeSpec.Name.IsExported() {
		log.Warnf("skipping unexported struct %s", typeSpec.Name.Name)
		return
	}

	switch x := typeSpec.Type.(type) {
	case *ast.StructType:
		if structData, notSkipped := createStructData(x, typeSpec.Name.Name, typeSpec.Doc); notSkipped {
			codeList.AddStructData(structData)
		}
	}
}

func createStructData(structType *ast.StructType, name string, commGroup *ast.CommentGroup) (generator.StructData, bool) {
	comments := getComments(commGroup)
	comments, annotations := GetAnnotations(comments)

	if containsAnnotation("ignore", annotations) {
		log.Warnf("skipping ignored struct %s ", name)
		return generator.StructData{}, false
	}

	return generator.StructData{
		Name:        name,
		FieldData:   structType.Fields,
		Annotations: annotations,
		Comments:    comments,
	}, true
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

func containsAnnotation(name string, list []generator.Annotation) bool {
	for _, item := range list {
		if item.Name == name {
			return true
		}
	}

	return false
}

var annotationList = []string{
	"subscription",
	"get",
	"update",
	"callback",
	"ignore",
}

func parseFunctions(codeList *generator.CodeList, pkgs map[string]*ast.Package, fset *token.FileSet) (packageName string) {
	for name, pkg := range pkgs {
		packageName = name

		for _, file := range pkg.Files {
			cmap := ast.NewCommentMap(fset, file, file.Comments)
			file.Comments = cmap.Comments()

			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					//restoreCommentForType(&cmap, fset, x) // TODO add annotations for functions
					createFunction(codeList, x)
				}

				return true
			})
		}
	}
	return
}

func createFunction(codeList *generator.CodeList, funcDecl *ast.FuncDecl) {
	if !funcDecl.Name.IsExported() {
		//log.Warnf("skipping unexported function %s", funcDecl.Name.Name)
		return
	}

	functionData := createFunctionData(funcDecl)
	codeList.AddFunctionData(functionData)
}

func createFunctionData(funcDecl *ast.FuncDecl) generator.FunctionData {
	comments := getComments(funcDecl.Doc)
	isSubscription := checkIsSubscription(funcDecl.Type)
	isPure := funcDecl.Type.Results == nil || funcDecl.Type.Results.List == nil || len(funcDecl.Type.Results.List) == 0
	name := funcDecl.Name.Name
	args := funcDecl.Type.Params
	returnValues := funcDecl.Type.Results

	return generator.FunctionData{
		Name:           name,
		Args:           args,
		ReturnValues:   returnValues,
		IsSubscription: isSubscription,
		IsPure:         isPure,
		Comments:       comments,
		//Annotations:
	}
}

func checkIsSubscription(funcTypes *ast.FuncType) bool {
	resultTypes := funcTypes.Results

	if resultTypes == nil || resultTypes.List == nil || len(resultTypes.List) == 0 {
		return false
	}

	for _, t := range resultTypes.List {
		if t != nil {
			if selType, ok := t.Type.(*ast.SelectorExpr); ok {
				if x, ok := selType.X.(*ast.Ident); ok {
					if x.Name == "goapi" && selType.Sel.Name == "Subscription" {
						return true
					}
				}
			}
		}
	}

	return false
}

func hasChanges(newState *generator.CodeList, oldState *generator.CodeList) bool {
	if len(newState.FunctionData) != len(oldState.FunctionData) {
		return true
	}

	if len(newState.StructData) != len(oldState.StructData) {
		return true
	}

	if reflect.DeepEqual(newState.FunctionData, oldState.FunctionData) &&
		reflect.DeepEqual(newState.StructData, oldState.StructData) {
		return true
	}

	return false
}

func generateGoFilesFromProto(codeList *generator.CodeList, packageName string) {
	generateFilesFromProto(codeList, packageName, "go", codeList.PathMap.Proto)
}

func generateDartFilesFromProto(codeList *generator.CodeList, packageName string) {
	generateFilesFromProto(codeList, packageName, "dart", codeList.PathMap.FlutterGenerated)
}

func generateJavaFilesFromProto(codeList *generator.CodeList, packageName string) {
	generateFilesFromProto(codeList, packageName, "java", codeList.PathMap.JavaGenerated)
}

func generateFilesFromProto(codeList *generator.CodeList, packageName, target, outDir string) {
	execDir := codeList.PathMap.Proto

	protoFilesPath := path.Join(execDir, "*.proto")
	files, err := filepath.Glob(protoFilesPath)
	if err != nil {
		log.Errorf("couldn't get proto files from dir with error: %v", err)
	} else {
		for _, v := range files {
			_, fileName := path.Split(v)
			if shouldSkipProtoFile(target, v) {
				continue
			}

			//if target == "java" {
			//	generateJavaFileFromProto(codeList, packageName, v)
			//} else {
			//	generateFileFromProto(codeList, packageName, target, outDir, v, strings.HasPrefix(fileName, "debugwebsocket"))
			//}
			generateFileFromProto(codeList, packageName, target, outDir, v, strings.HasPrefix(fileName, "debugwebsocket"))
		}
	}
}

// skip timestamp.proto and wrappers.proto because they are in gopath
func shouldSkipProtoFile(target, filePath string) bool {
	_, fileName := path.Split(filePath)
	return (target == "go" && (strings.HasPrefix(fileName, "timestamp") || strings.HasPrefix(fileName, "wrappers"))) ||
		strings.HasPrefix(fileName, "debugwebsocket")
}

func generateFileFromProto(codeList *generator.CodeList, packageName, target, outDir, protoSource string, grpc bool) {
	execDir := codeList.PathMap.Proto
	plugin := ""
	if grpc {
		switch target {
		case "go":
			plugin = "plugins=grpc:"
		case "dart":
			plugin = "grpc:"
		}
	}
	targetFlag := fmt.Sprintf("--%s_out=%s%s", target, plugin, outDir)
	protoPathFlag := fmt.Sprintf("--proto_path=%s", execDir)

	cmd := exec.Command("protoc", protoPathFlag, protoSource, targetFlag)
	cmd.Dir = execDir

	log.Infof("executing protoc in dir '%v', packageName = '%v', target = '%v', outDir = '%v', protoPathFlag = '%v' \n protoSource = '%v'", cmd.Dir, packageName, target, outDir, protoPathFlag, protoSource)
	runCommand(cmd)
}

//protoc -I=$SRC_DIR --java_out=$DST_DIR $SRC_DIR/addressbook.proto
//protoc --plugin=protoc-gen-grpc-java --grpc-java_out="$OUTPUT_FILE" --proto_path="$DIR_OF_PROTO_FILE" "$PROTO_FILE"
func generateJavaFileFromProto(codeList *generator.CodeList, packageName, protoSource string) {
	execDir := codeList.PathMap.Proto

	outDir := codeList.PathMap.JavaGenerated
	targetFlag := fmt.Sprintf("--grpc-java_out=%s", outDir)
	protoPathFlag := fmt.Sprintf("--proto_path=%s", execDir)

	cmd := exec.Command("protoc", "--plugin=protoc-gen-grpc-java", targetFlag, protoSource)
	cmd.Dir = execDir

	log.Infof("executing protoc in dir '%v', packageName = '%v', target = '%v', outDir = '%v', protoPathFlag = '%v' \n protoSource = '%v'", cmd.Dir, packageName, "java", outDir, protoPathFlag, protoSource)
	runCommand(cmd)
}

func runCommand(cmd *exec.Cmd) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
	}
}
