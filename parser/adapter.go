package parser

import (
	"fmt"
	"github.com/PenF00k/go-wand/adapter"
	"github.com/PenF00k/go-wand/generator"
	"go/ast"
)

func Adopt(codeList *generator.CodeList) *adapter.Adapter {
	structures := adoptStructures(codeList)
	functions := adoptFunctions(codeList)
	return &adapter.Adapter{
		Structures:    structures,
		Functions:     functions,
		Subscriptions: nil,
	}
}

func adoptStructures(codeList *generator.CodeList) []adapter.Struct {
	structures := make([]adapter.Struct, 0, 10)
	for _, v := range codeList.StructData {
		structure := adoptStructure(codeList, v)
		structures = append(structures, structure)
	}

	return structures
}

func adoptStructure(codeList *generator.CodeList, structData generator.StructData) adapter.Struct {
	fields := adoptFields(codeList, structData.FieldData, "todo") //TODO pack?

	return adapter.Struct{
		Name:   adapter.StructName(structData.Name),
		Fields: fields,
		//Annotations: nil,
		//Comments:    nil,
	}
}

func adoptFunctions(codeList *generator.CodeList) []adapter.Function {
	functions := make([]adapter.Function, 0, 10)
	for _, v := range codeList.FunctionData {
		function := adoptSourceFunction(codeList, v)
		functions = append(functions, function)
	}

	return functions
}

func adoptSourceFunction(codeList *generator.CodeList, functionData generator.FunctionData) adapter.Function {
	sourceArgs := functionData.Args
	sourceReturnValues := functionData.ReturnValues
	var returnValues []adapter.Field

	if functionData.IsSubscription {
		// remove last arg, it's event function
		newSize := len(sourceArgs.List) - 1
		last := sourceArgs.List[newSize]
		sFields := make([]*ast.Field, newSize)
		copy(sFields, sourceArgs.List[:newSize])
		sourceArgs.List = sFields

		// replace result values
		rFields := make([]*ast.Field, 1)
		rFields[0] = last
		sourceReturnValues.List = rFields

		returnValuesTmp := adoptFields(codeList, sourceReturnValues, "todo") //TODO pack?
		returnValues = []adapter.Field{returnValuesTmp[0].Type.Function.Args[0]}
	} else {
		returnValues = adoptFields(codeList, sourceReturnValues, "todo")
	}

	args := adoptFields(codeList, sourceArgs, "todo") //TODO pack?

	return adapter.Function{
		FunctionName:   functionData.Name,
		Args:           args,
		ReturnValues:   returnValues,
		IsPure:         functionData.IsPure,
		IsSubscription: functionData.IsSubscription,
		//Annotations: nil,
		//Comments:    nil,
	}
}

func adoptFields(codeList *generator.CodeList, list *ast.FieldList, pack string) []adapter.Field {
	fields := make([]adapter.Field, 0, 10)
	if list == nil {
		return fields
	}

	for _, f := range list.List {
		_type := adoptType(codeList, f.Type, pack)

		if len(f.Names) > 0 {
			for _, n := range f.Names {
				field := adapter.Field{
					Name: n.Name,
					Type: _type,
				}

				fields = append(fields, field)
			}
		} else {
			field := adapter.Field{
				Type: _type,
			}
			fields = append(fields, field)
		}
	}

	return fields
}

var typeContext = make(map[string]*adapter.Type)

func adoptType(codeList *generator.CodeList, tp ast.Expr, pack string) *adapter.Type {
	var TypeName string
	var Pointer *adapter.Pointer
	var Slice *adapter.Slice
	var Map *adapter.Map
	var Struct *adapter.Struct
	var Function *adapter.Function
	var IsPrimitive bool
	var Selector *adapter.Selector
	var innerName string

	switch x := tp.(type) {
	case *ast.StarExpr:
		Pointer = &adapter.Pointer{
			InnerType: adoptType(codeList, x.X, pack),
		}
		innerName = fmt.Sprintf("ptr_%s", Pointer.InnerType.InnerName)

	case *ast.ArrayType:
		Slice = &adapter.Slice{
			InnerType: adoptType(codeList, x.Elt, pack),
		}
		innerName = fmt.Sprintf("slc_%s", Slice.InnerType.InnerName)

	case *ast.MapType:
		Map = &adapter.Map{
			KeyType:   adoptType(codeList, x.Key, pack),
			ValueType: adoptType(codeList, x.Value, pack),
		}
		innerName = fmt.Sprintf("mp_%s_%s", Map.KeyType.InnerName, Map.ValueType.InnerName)

	case *ast.StructType:
		Struct = adoptInnerStructure(codeList, x, pack)
		innerName = fmt.Sprintf("strt_%s", Struct.Name)

	case *ast.FuncType:
		Function = adoptInnerFunction(codeList, x, pack)
		fn := Function.FunctionName

		innerName = fmt.Sprintf("func_%s", fn)

	case *ast.Ident:
		TypeName = x.Name
		for _, v := range codeList.StructData {
			if v.Name == TypeName {
				s := adoptStructure(codeList, v)
				Struct = &s
				innerName = fmt.Sprintf("strt_%s", Struct.Name)
				break
			}
		}

		if Struct == nil {
			IsPrimitive = true
			innerName = fmt.Sprintf("prmt_%s", TypeName)
		}

	case *ast.SelectorExpr:
		xType := adoptType(codeList, x.X, pack)
		var selType *adapter.Type

		// let's assume the selector type is always an ast.TypeSpec
		if x.Sel != nil && x.Sel.Obj != nil && x.Sel.Obj.Decl != nil {
			if ts, ok := x.Sel.Obj.Decl.(*ast.TypeSpec); ok {
				selType = adoptType(codeList, ts.Type, pack)
			}
		}

		Selector = &adapter.Selector{
			Package:  string(xType.Name),
			TypeName: adapter.TypeName(x.Sel.Name),
			Type:     selType,
		}

		innerName = fmt.Sprintf("sltr_%s_%s", Selector.Package, Selector.TypeName)

		//Package = &x.Sel.Name //TODO а так ли это?? или pack возьмем?
	}

	if tp, exists := typeContext[innerName]; exists {
		return tp
	}

	res := adapter.Type{
		Name:        adapter.TypeName(TypeName),
		Pointer:     Pointer,
		Slice:       Slice,
		Map:         Map,
		Struct:      Struct,
		Function:    Function,
		IsPrimitive: IsPrimitive,
		Selector:    Selector,
		InnerName:   innerName,
	}

	typeContext[innerName] = &res

	return &res
}

func adoptInnerStructure(codeList *generator.CodeList, structType *ast.StructType, pack string) *adapter.Struct {
	return &adapter.Struct{
		Name:   "todo", //TODO
		Fields: adoptFields(codeList, structType.Fields, pack),
	}
}

func adoptInnerFunction(codeList *generator.CodeList, funcType *ast.FuncType, pack string) *adapter.Function {
	args := adoptFields(codeList, funcType.Params, pack)
	returnValues := adoptFields(codeList, funcType.Results, pack)

	fn := "event"
	for _, v := range args {
		fn = fn + "_" + v.Type.InnerName
	}
	for _, v := range returnValues {
		fn = fn + "_" + v.Type.InnerName
	}

	return &adapter.Function{
		FunctionName:   fn,
		Args:           args,
		ReturnValues:   returnValues,
		IsPure:         true,
		IsSubscription: false,
	}
}

//
//
//func toTypeName(name string) string {
//	return name
//}
//
//func GetPrimitive(name string) PrimitiveType {
//	var isPrimitive bool
//	var wrapperTypeName string
//
//	switch name {
//	case "float32":
//		isPrimitive = true
//		wrapperTypeName = "FloatValue"
//	case "float64":
//		isPrimitive = true
//		wrapperTypeName = "DoubleValue"
//	case "int":
//		fallthrough
//	case "int8":
//		fallthrough
//	case "int16":
//		fallthrough
//	case "int32":
//		isPrimitive = true
//		wrapperTypeName = "Int32Value"
//	case "int64":
//		isPrimitive = true
//		wrapperTypeName = "Int64Value"
//	case "bool":
//		isPrimitive = true
//		wrapperTypeName = "BoolValue"
//	case "string":
//		isPrimitive = true
//		wrapperTypeName = "StringValue"
//	case "[]byte":
//		isPrimitive = true
//		wrapperTypeName = "BytesValue"
//	}
//
//	return PrimitiveType{
//		IsPrimitive:     isPrimitive,
//		WrapperTypeName: wrapperTypeName,
//	}
//}
