// A low-effort race condition finder. Useful to find functions that are too
// large and cumbersome to work with `go <run|test> -race`.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"slices"
	"sync"
)

func racyFunc() {
	type Foo struct {
		a int
		b int
	}

	foo := Foo{a: 99, b: 99}
	doStuff := func(f Foo) Foo { f.a = f.a * 10; return f }

	var wg sync.WaitGroup

	l1 := make([]Foo, 10)
	l2 := make([]int, 10)
	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// concurrent write/read of outer var is bad!
			// commenting this line alone will avoid the data race
			// completely (i.e. go run -race main.go will pass)
			//
			// (a false positive will still be reported, but the
			// code would already look bogus at that point)
			foo.a = i

			l1[i] = doStuff(foo) // result will be tainted by the concurrent write!
			l2[i] = i            // safe; unaffected by mutation of foo
		}(i)
	}
	wg.Wait()
	fmt.Println(l1)
	fmt.Println(l2)

	l3 := make([]Foo, 10)
	for i := range 10 {
		f := foo
		f.a = i
		f.b = i
		l3[i] = f
	}
	for i, f := range l3 {
		wg.Add(1)
		go func(f Foo) {
			defer wg.Done()
			l3[i] = doStuff(f) // safe, because f is different in every iteration
		}(f)
	}
	wg.Wait()
	fmt.Println(l3)

	// same as above, but written in a single loop
	l4 := make([]Foo, 10)
	for i := range 10 {
		f := foo
		f.a = i
		f.b = i
		wg.Add(1)
		go func(f Foo) {
			defer wg.Done()
			l4[i] = doStuff(f)
		}(f)
	}
	wg.Wait()
	fmt.Println(l4)

	//
}

func parseFile(fname string) {
	fset := token.NewFileSet()
	b, _ := os.ReadFile(fname)

	node, err := parser.ParseFile(fset, "", b, parser.AllErrors)
	// TODO: parser.ParseDir()
	if err != nil {
		return
	}

	type GoStmt struct {
		Params      []*ast.Field
		Assigns     map[token.Pos]*ast.AssignStmt
		RacyAssigns []token.Pos
	}

	var goStmts []GoStmt

	// https://blog.microfast.ch/refactoring-go-code-using-ast-replacement-e3cbacd7a331?gi=5b30ae55812a
	maybeRacy := func(n ast.Node) bool {
		if goStmt, ok := n.(*ast.GoStmt); ok {
			var g GoStmt
			g.Assigns = make(map[token.Pos]*ast.AssignStmt)
			call := goStmt.Call.Fun.(*ast.FuncLit)
			g.Params = call.Type.Params.List
			for _, stmt := range call.Body.List {
				if assign, ok := stmt.(*ast.AssignStmt); ok {
					// := is probably fine, since it can only be used to initialise new vars
					if assign.Tok.String() == "=" {
						g.Assigns[assign.TokPos] = assign
					}
				}
			}
			goStmts = append(goStmts, g)

		}
		return true
	}

	ast.Inspect(node, maybeRacy)

	for _, g := range goStmts {

		var paramNames []string
		for _, p := range g.Params {
			paramNames = append(paramNames, p.Names[0].Name)
		}

		for _, a := range g.Assigns {
			// TODO: if lhs is slice[int], probably fine

			for _, rhs := range a.Rhs {
				// rhs can be a simple Ident: foo = x
				// or a CallExpr: foo = fn(x)
				//
				// in either case, if x was not passed directly
				// via go func(...){...}(x), we assume that x
				// came from the outer scope, and mark the
				// whole goStmt as unsafe

				switch rtype := rhs.(type) {
				case *ast.Ident:
					// https://stackoverflow.com/a/65433734
					if !slices.Contains(paramNames, types.ExprString(rhs)) {
						g.RacyAssigns = append(g.RacyAssigns, a.TokPos)
					}
				case *ast.CallExpr:
					// need rtype to access Args
					for _, arg := range rtype.Args {
						if !slices.Contains(paramNames, types.ExprString(arg)) {
							g.RacyAssigns = append(g.RacyAssigns, a.TokPos)
						}
					}
				default:
					panic("Unhandled rhs type in assignment")
				}
			}
		}

		if len(g.RacyAssigns) > 0 {
			fmt.Println("Racy assignment(s) found in goroutine:")
			// ast.Print(fset, g) // full ast
			for _, a := range g.RacyAssigns {
				// ast.FilterDecl(a, ast.Filter())
				fmt.Printf(
					"%s:%d:%v = %v\n",
					// fset.Position(a).Filename, // will probably be ""
					fname,
					fset.Position(a).Line,
					types.ExprString(g.Assigns[a].Lhs[0]),
					types.ExprString(g.Assigns[a].Rhs[0]),
				)
			}
		}

	}
}

func main() {
	parseFile("main.go")
	// racyFunc()
}
