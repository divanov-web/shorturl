// Package main Добавляет multichecker в проект.
//
// как сбилдить под windows:
//
//	go build -o staticlint.exe ./cmd/staticlint
//
// Как запустить (PowerShell):
//
// go vet -vettool="$PWD\staticlint.exe" ./...
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/analysis/lint"

	// std passes
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	// staticcheck analyzers
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	// public analyzers
	"github.com/gostaticanalysis/nilerr"
	"github.com/sonatard/noctx"

	// кастомный анализатор - noosexit
	"github.com/divanov-web/shorturl/cmd/staticlint/noosexit"
)

func main() {
	var analyzers []*analysis.Analyzer

	//std passes
	analyzers = append(analyzers,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		unusedresult.Analyzer,
		nilness.Analyzer,
	)

	// всех анализаторов класса SA
	for _, a := range staticcheck.Analyzers {
		if len(a.Analyzer.Name) >= 2 && a.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	// staticcheck.io: S1000, ST1005
	analyzers = append(analyzers,
		findAnalyzer(simple.Analyzers, "S1000"),      // simplify code
		findAnalyzer(stylecheck.Analyzers, "ST1005"), // error strings should not be capitalized
	)

	// публичные анализаторы
	analyzers = append(analyzers,
		nilerr.Analyzer, // returning nil error
		noctx.Analyzer,  // HTTP client.Request without Context
	)

	// кастомный анализатор - noosexit
	analyzers = append(analyzers, noosexit.Analyzer)

	multichecker.Main(analyzers...)
}

// findAnalyzer - хелпер для получения нужного анализатора по его имени
func findAnalyzer(list []*lint.Analyzer, name string) *analysis.Analyzer {
	for _, a := range list {
		if a.Analyzer.Name == name {
			return a.Analyzer
		}
	}
	return nil
}
