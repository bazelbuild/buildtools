package warn

import (
	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/bzlenv"
)

// Type describes an expression type in Starlark.
type Type int

// List of known types
const (
	Unknown Type = iota
	Bool
	Ctx
	CtxActions
	CtxActionsArgs
	Depset
	Dict
	Int
	None
	String
)

func (t Type) String() string {
	return [...]string{
		"unknown",
		"bool",
		"ctx",
		"ctx.actions",
		"ctx.actions.args",
		"depset",
		"dict",
		"int",
		"none",
		"string",
	}[t]
}

func detectTypes(f *build.File) map[build.Expr]Type {
	variables := make(map[int]Type)
	result := make(map[build.Expr]Type)

	var walk func(e *build.Expr, env *bzlenv.Environment)
	walk = func(e *build.Expr, env *bzlenv.Environment) {
		// Postorder: determining types of subnodes may help with this node's type
		walkOnce(*e, env, walk)

		nodeType := Unknown
		defer func() {
			if nodeType != Unknown {
				result[*e] = nodeType
			}
		}()

		switch node := (*e).(type) {
		case *build.StringExpr:
			nodeType = String
		case *build.DictExpr:
			nodeType = Dict
		case *build.LiteralExpr:
			nodeType = Int
		case *build.Comprehension:
			if node.Curly {
				nodeType = Dict
			}
		case *build.CallExpr:
			if ident, ok := (node.X).(*build.Ident); ok {
				switch ident.Name {
				case "depset":
					nodeType = Depset
				case "dict":
					nodeType = Dict
				}
			} else if dot, ok := (node.X).(*build.DotExpr); ok {
				if result[dot.X] == CtxActions && dot.Name == "args" {
					nodeType = CtxActionsArgs
				}
			}
		case *build.ParenExpr:
			nodeType = result[node.X]
		case *build.Ident:
			switch node.Name {
			case "True", "False":
				nodeType = Bool
				return
			case "None":
				nodeType = None
				return
			case "ctx":
				binding := env.Get(node.Name)
				if binding != nil && binding.Kind == bzlenv.Parameter {
					nodeType = Ctx
					return
				}
			}
			binding := env.Get(node.Name)
			if binding != nil {
				if t, ok := variables[binding.ID]; ok {
					nodeType = t
				}
			}
		case *build.DotExpr:
			if result[node.X] == Ctx && node.Name == "actions" {
				nodeType = CtxActions
			}
		case *build.BinaryExpr:
			switch node.Op {
			case "=", "+=", "-=", "*=", "/=", "//=", "%=", "|=":
				// Assignments
				t, ok := result[node.Y]
				if !ok {
					return
				}
				if node.Op == "%=" && t != String {
					// If the right hand side is not a string, the left hand side can still be a string
					return
				}
				ident, ok := (node.X).(*build.Ident)
				if !ok {
					return
				}
				binding := env.Get(ident.Name)
				if binding == nil {
					return
				}
				variables[binding.ID] = t

			case ">", ">=", "<", "<=", "==", "!=", "in", "not in":
				// Boolean
				nodeType = Bool

			case "+", "-", "*", "/", "//", "%", "|":
				// We assume these operators can only applied to expressions of the same type and
				// preserve the type
				if t, ok := result[node.X]; ok {
					nodeType = t
				} else if t, ok := result[node.Y]; ok {
					if node.Op != "%" || t == String {
						// The percent operator is special because it can be applied to to arguments of
						// different types (`"%s\n" % foo`), and we can't assume that the expression has
						// type X if the right-hand side has the type X.
						nodeType = t
					}
				}
			}
		}
	}
	var expr build.Expr = f
	walk(&expr, bzlenv.NewEnvironment())

	return result
}

// walkOnce is a wrapper for bzlint.WalkOnceWithEnvironment which skips the left hand side for
// named parameters of functions. E.g. for `foo(x, y = z)` it visits `foo`, `x`, and `z`.
// In the following example `x` in the last line shouldn't be recognised as int, but 'y' should:
//
//    x = 3
//    y = 5
//    foo(x = y)
func walkOnce(node build.Expr, env *bzlenv.Environment, fct func(e *build.Expr, env *bzlenv.Environment)) {
	switch expr := node.(type) {
	case *build.CallExpr:
		fct(&expr.X, env)
		for _, param := range expr.List {
			if binary, ok := param.(*build.BinaryExpr); ok && binary.Op == "=" {
				fct(&binary.Y, env)
			} else {
				fct(&param, env)
			}
		}
	default:
		bzlenv.WalkOnceWithEnvironment(expr, env, fct)
	}
}
