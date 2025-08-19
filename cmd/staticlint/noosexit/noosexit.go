// Package noosexit анализатор запрещает использовать прямой вызов os.Exit в функции main пакета main
package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer структура нового анализатора
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "forbid direct os.Exit calls in main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// Only check package main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if fn.Name.Name != "main" {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Lparen, "do not call os.Exit in main.main; return an error or log.Fatal instead")
				}
				return true
			})
		}
	}
	return nil, nil
}
