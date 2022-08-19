package txl

import (
	"go/ast"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/ssa"
)

var (
	Analyzer = &analysis.Analyzer{
		Name: "txl",
		Doc:  "TODO",
		Run:  run,
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
			buildssa.Analyzer,
		},
	}

	insertExp = regexp.MustCompile(`^.*INSERT\s+INTO\s+(\w+)\s+.*`)
)

func run(pass *analysis.Pass) (interface{}, error) {
	funcs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	if !pkgImports(pass.Pkg, "database/sql") {
		return nil, nil
	}

	TxFuncMap := make(map[string][]string)
	for _, f := range funcs {
		// map function which has transaction as param
		for _, p := range f.Params {
			if isPtrTx(p.Type()) {
				for _, b := range f.Blocks {
					for i := range b.Instrs {
						instr, ok := b.Instrs[i].(*ssa.Call)
						if !ok {
							continue
						}

						cf, ok := instr.Call.Value.(*ssa.Function)
						if !ok || !isTxExec(cf) || len(instr.Call.Args) == 0 {
							continue
						}

						query := instr.Call.Args[1].(*ssa.Const).Value.ExactString()
						if res := insertExp.FindStringSubmatch(query); len(res) < 2 {
							continue
						} else {
							tables := make([]string, 0, len(res))
							for _, r := range res[1:] {
								tables = append(tables, r)
							}
							TxFuncMap[f.Name()] = tables
						}
					}
				}
				break
			}
		}

		// count transaction call
		for _, b := range f.Blocks {
			for i := range b.Instrs {
				if instr, ok := b.Instrs[i].(*ssa.Extract); ok {
					// only target first of tuple element
					if instr.Index != 0 {
						continue
					}

					v0 := instr.Tuple.Type().(*types.Tuple).At(0)
					if !isPtrTx(v0.Type()) {
						continue
					}

					var callcnt int
					for _, ref := range *instr.Referrers() {
						call, ok := ref.(*ssa.Call)
						if !ok {
							continue
						}
						f, ok := call.Common().Value.(*ssa.Function)
						if !ok {
							continue
						}
						if isTxCommit(f) {
							break
						}
						if !isTxRollback(f) {
							callcnt++
						}
					}
					pass.Reportf(instr.Tuple.Pos(), "transaction variable declared here is used %d times until COMMIT", callcnt)
				}
			}
		}
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}

		call := n.(*ast.CallExpr)
		if !isDBBeginMethod(pass.TypesInfo, call) {
			return true
		}

		var stmts []ast.Stmt
		for i := len(stack) - 1; i >= 0; i-- {
			if b, ok := stack[i].(*ast.BlockStmt); ok {
				for j, v := range b.List {
					if v == stack[i+1] {
						stmts = b.List[j:]
						break
					}
				}
				break
			}
		}

		asg, ok := stmts[0].(*ast.AssignStmt)
		if !ok {
			return true
		}

		resp := rootIdent(asg.Lhs[0])
		if resp == nil {
			return true
		}

		for _, stmt := range stmts[1:] {
			var callexpr *ast.CallExpr

			switch stmt := stmt.(type) {
			case *ast.IfStmt: // if _, err := f(); err != {}
				if stmt.Init == nil {
					continue
				}

				ce, ok := stmt.Init.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr)
				if !ok {
					continue
				}
				callexpr = ce
			case *ast.ExprStmt: // f()
				callexpr = stmt.X.(*ast.CallExpr)
			case *ast.AssignStmt: // res, err := f()
				ce, ok := stmt.Rhs[0].(*ast.CallExpr)
				if !ok {
					continue
				}
				callexpr = ce
			default:
				continue
			}

			switch f := callexpr.Fun.(type) {
			case *ast.SelectorExpr: // tx.Exec(hoge)
				if !isPtrTxExpr(pass, f.X) {
					continue
				}

				switch f.Sel.Name {
				case "Exec":
					args := callexpr.Args
					if len(args) == 0 {
						continue // no argument
					}

					query := args[0].(*ast.BasicLit).Value
					if res := insertExp.FindStringSubmatch(query); len(res) < 2 {
						continue
					} else {
						for _, r := range res[1:] {
							pass.Reportf(f.Pos(), "INSERT %s", r)
						}
					}
				case "Commit":
				}
			case *ast.Ident: // f()
				ident, ok := callexpr.Fun.(*ast.Ident)
				if !ok {
					continue
				}

				tables, ok := TxFuncMap[ident.Name]
				if !ok {
					continue
				}
				for _, table := range tables {
					pass.Reportf(f.Pos(), "INSERT %s", table)
				}
			}
		}

		return true
	})

	return nil, nil
}

func pkgImports(pkg *types.Package, path string) bool {
	for _, imp := range pkg.Imports() {
		if imp.Path() == path {
			return true
		}
	}
	return false
}

func isPtrTxExpr(pass *analysis.Pass, e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)
	if !ok {
		return false
	}

	ptr, ok := pass.TypesInfo.ObjectOf(ident).Type().(*types.Pointer)
	if !ok {
		return false
	}

	return isPtrTx(ptr)
}

func isPtrTx(typ types.Type) bool {
	ptr, ok := typ.(*types.Pointer)
	if !ok {
		return false
	}
	if !isNamedType(ptr.Elem(), "database/sql", "Tx") {
		return false
	}

	return true
}

func isTargetMethod(typ types.Type, path, name string) bool {
	sig, _ := typ.(*types.Signature)
	recv := sig.Recv()
	if recv == nil {
		return false
	}

	ptr, ok := recv.Type().(*types.Pointer)
	if !ok {
		return false
	}
	if !isNamedType(ptr.Elem(), path, name) {
		return false
	}

	return true
}

func isDBMethod(typ types.Type) bool {
	return isTargetMethod(typ, "database/sql", "DB")
}

func isTxMethod(typ types.Type) bool {
	return isTargetMethod(typ, "database/sql", "Tx")
}

func isTxExec(f *ssa.Function) bool {
	if !isTxMethod(f.Type()) {
		return false
	}
	return f.Name() == "Exec"
}

func isTxCommit(f *ssa.Function) bool {
	if !isTxMethod(f.Type()) {
		return false
	}
	return f.Name() == "Commit"
}

func isTxRollback(f *ssa.Function) bool {
	if !isTxMethod(f.Type()) {
		return false
	}
	return f.Name() == "Rollback"
}

func isDBBeginMethod(info *types.Info, expr *ast.CallExpr) bool {
	fun, ok := expr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if fun.Sel.Name != "Begin" && fun.Sel.Name != "BeginContext" {
		return false
	}

	ptr, ok := info.ObjectOf(fun.X.(*ast.Ident)).Type().(*types.Pointer)
	if !ok {
		return false
	}
	if !isNamedType(ptr.Elem(), "database/sql", "DB") {
		return false
	}

	sig := info.Types[fun].Type.(*types.Signature)
	res := sig.Results()
	if res.Len() != 2 {
		return false
	}
	if ptr, ok := res.At(0).Type().(*types.Pointer); !ok || !isNamedType(ptr.Elem(), "database/sql", "Tx") {
		return false
	}

	errorType := types.Universe.Lookup("error").Type()
	if !types.Identical(res.At(1).Type(), errorType) {
		return false
	}

	return true
}

func rootIdent(n ast.Node) *ast.Ident {
	switch n := n.(type) {
	case *ast.SelectorExpr:
		return rootIdent(n.X)
	case *ast.Ident:
		return n
	default:
		return nil
	}
}

func isNamedType(t types.Type, path, name string) bool {
	n, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := n.Obj()
	return obj.Name() == name && obj.Pkg() != nil && obj.Pkg().Path() == path
}
