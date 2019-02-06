package caster

import (
	"go/ast"
)

func GetGoType(tp ast.Expr) string {
	switch tp.(type) {
	case *ast.Ident:
		return "ast.Ident"

	case *ast.SelectorExpr:
		return "ast.SelectorExpr"

	case *ast.MapType:
		return "ast.MapType"

	case *ast.StarExpr:
		return "ast.StarExpr"

	case *ast.ArrayType:
		return "ast.ArrayType"
	}

	return ""
}

func GetFullGoTypeAsString(tp ast.Expr, packageName string) string {
	switch x := tp.(type) {
	case *ast.Ident:
		return toPackageName(x.Name, packageName)
	case *ast.SelectorExpr:
		return toPackageName(x.Sel.Name, packageName)

	case *ast.MapType:
		return "map[" + GetFullGoTypeAsString(x.Key, packageName) + "]" + GetFullGoTypeAsString(x.Value, packageName)

	case *ast.StarExpr:
		return "*" + GetFullGoTypeAsString(x.X, packageName)

	case *ast.ArrayType:
		return "[]" + GetFullGoTypeAsString(x.Elt, packageName)
	}

	return ""
}

func toPackageName(name, packageName string) string {
	if ast.IsExported(name) && packageName != "" {
		return packageName + "." + name
	}
	return name
}
