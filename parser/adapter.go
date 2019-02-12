package parser

import (
	"gitlab.vmassive.ru/wand/adapter"
	"gitlab.vmassive.ru/wand/generator"
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

func adoptType(codeList *generator.CodeList, tp ast.Expr, pack string) adapter.Type {

	var TypeName string
	var Pointer *adapter.Pointer
	var Slice *adapter.Slice
	var Map *adapter.Map
	var Struct *adapter.Struct
	var Function *adapter.Function
	var IsPrimitive bool
	var Selector *adapter.Selector

Loop:
	switch x := tp.(type) {
	case *ast.StarExpr:
		Pointer = &adapter.Pointer{
			InnerType: adoptType(codeList, x.X, pack),
		}

	case *ast.ArrayType:
		Slice = &adapter.Slice{
			InnerType: adoptType(codeList, x.Elt, pack),
		}

	case *ast.MapType:
		Map = &adapter.Map{
			KeyType:   adoptType(codeList, x.Key, pack),
			ValueType: adoptType(codeList, x.Value, pack),
		}

	case *ast.StructType:
		Struct = adoptInnerStructure(codeList, x, pack)

	case *ast.FuncType:
		Function = adoptInnerFunction(codeList, x, pack)

	case *ast.Ident:
		TypeName = x.Name
		for _, v := range codeList.StructData {
			if v.Name == TypeName {
				s := adoptStructure(codeList, v)
				Struct = &s
				break Loop
			}
		}

		IsPrimitive = true

	case *ast.SelectorExpr:
		n := adoptType(codeList, x.X, pack).Name
		Selector = &adapter.Selector{
			Package:  string(n),
			TypeName: adapter.TypeName(x.Sel.Name),
		}
		//Package = &x.Sel.Name //TODO а так ли это?? или pack возьмем?
	}

	return adapter.Type{
		Name:        adapter.TypeName(TypeName),
		Pointer:     Pointer,
		Slice:       Slice,
		Map:         Map,
		Struct:      Struct,
		Function:    Function,
		IsPrimitive: IsPrimitive,
		Selector:    Selector,
	}
}

func toTypeName(name string) string {
	return name
}

func adoptInnerStructure(codeList *generator.CodeList, structType *ast.StructType, pack string) *adapter.Struct {
	return &adapter.Struct{
		Name:   "todo", //TODO
		Fields: adoptFields(codeList, structType.Fields, pack),
	}
}

func adoptInnerFunction(codeList *generator.CodeList, funcType *ast.FuncType, pack string) *adapter.Function {
	return &adapter.Function{
		//FunctionName:   "todo", //TODO
		Args:           adoptFields(codeList, funcType.Params, pack),
		ReturnValues:   adoptFields(codeList, funcType.Results, pack),
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
