// A low-effort race condition finder. Useful to find functions that are too
// large and cumbersome to work with `go <run|test> -race`.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sync"
)

func racyFunc() {
	type Foo struct {
		a int
		b int
	}

	foo := Foo{a: 1, b: 2}
	doStuff := func(f Foo) Foo { return f }

	var wg sync.WaitGroup

	l1 := make([]Foo, 10)
	l2 := make([]int, 10)
	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo.a = i            // concurrent mutation is bad!
			l1[i] = doStuff(foo) // result will be tainted!
			l2[i] = i            // safe; unaffected by mutation of foo
		}(i)
	}

	var foos []Foo
	for i := range 10 {
		f := foo
		f.a = i
		f.b = i
		foos = append(foos, f)
	}
	for i, f := range foos {
		wg.Add(1)
		go func(f Foo) {
			defer wg.Done()
			foos[i] = doStuff(foo) // safe
		}(f)
	}

	wg.Wait()
}

func parseFile(fname string) {
	fset := token.NewFileSet()
	b, _ := os.ReadFile(fname)

	node, err := parser.ParseFile(fset, "", b, parser.AllErrors)
	// TODO: parser.ParseDir()
	if err != nil {
		return
	}

	var assigns []*ast.AssignStmt

	// https://blog.microfast.ch/refactoring-go-code-using-ast-replacement-e3cbacd7a331?gi=5b30ae55812a
	maybeRacy := func(n ast.Node) bool {
		if goStmt, ok := n.(*ast.GoStmt); ok {
			if call, ok := goStmt.Call.Fun.(*ast.FuncLit); ok {
				for _, stmt := range call.Body.List {
					if assign, ok := stmt.(*ast.AssignStmt); ok {
						// TODO: if lhs is slice[int], probably fine
						if assign.Tok.String() == "=" {
							assigns = append(assigns, assign)
						}
					}
				}
			}
		}
		return true
	}

	ast.Inspect(node, maybeRacy)

	for _, a := range assigns {
		line := fset.Position(a.TokPos).Line
		fmt.Printf("%s:%v:%v = %v\n", fname, line, a.Lhs[0], a.Rhs[0])
	}
}

func main() {
	parseFile("main.go")
}
