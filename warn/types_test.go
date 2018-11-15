package warn

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/testutils"
)

func checkTypes(t *testing.T, input, output string) {
	input = strings.TrimLeft(input, "\n")
	f, err := build.Parse("test.bzl", []byte(input))
	if err != nil {
		t.Fatalf("%v", err)
	}
	types := detectTypes(f)

	var edit func(expr build.Expr, stack []build.Expr) build.Expr
	edit = func(expr build.Expr, stack []build.Expr) build.Expr {
		t, ok := types[expr]
		if !ok {
			return nil
		}
		// Traverse the node's children before modifying this node.
		build.EditChildren(expr, edit)
		start, _ := expr.Span()
		return &build.Ident{
			Name:    fmt.Sprintf("%s:<%s>", t, build.FormatString(expr)),
			NamePos: start,
		}
	}

	build.Edit(f, edit)

	want := []byte(strings.TrimLeft(output, "\n"))
	have := build.Format(f)
	if !bytes.Equal(have, want) {
		t.Errorf("detected types incorrectly: diff shows -expected, +ours")
		testutils.Tdiff(t, want, have)
	}
}

func TestTypes(t *testing.T) {
	checkTypes(t, `
s = "string"
s2 = s
d = {}
d2 = {foo: bar}
d3 = dict(**foo)
d4 = {k: v for k, v in foo}
dep = depset(items=[s, d])
foo = bar
`, `
s = string:<"string">
s2 = string:<s>
d = dict:<{}>
d2 = dict:<{foo: bar}>
d3 = dict:<dict(**foo)>
d4 = dict:<{k: v for k, v in foo}>
dep = depset:<depset(items = [
    string:<s>,
    dict:<d>,
])>
foo = bar
`)
}

func TestScopesFunction(t *testing.T) {
	checkTypes(t, `
s = "string"

def f():
    s1 = s

def g():
    s2 = s1
`, `
s = string:<"string">

def f():
    s1 = string:<s>

def g():
    s2 = s1
`)
}

func TestScopesParameters(t *testing.T) {
	checkTypes(t, `
x = 3
y = 4
z = 5

foo(y = "bar")
foo(x, y = bar(z = z), t + z)


def f(z = "bar"):
    return z

bar(x, y, z)
`, `
x = int:<3>
y = int:<4>
z = int:<5>

foo(y = string:<"bar">)
foo(int:<x>, y = bar(z = int:<z>), int:<t + int:<z>>)

def f(z = string:<"bar">):
    return string:<z>

bar(int:<x>, int:<y>, int:<z>)
`)
}

func TestBinaryOperators(t *testing.T) {
	checkTypes(t, `
i = 1
d = {}
s = depset()

i - foo
foo - i

d + bar
bar + d

s | baz
baz | s
`, `
i = int:<1>
d = dict:<{}>
s = depset:<depset()>

int:<int:<i> - foo>
int:<foo - int:<i>>

dict:<dict:<d> + bar>
dict:<bar + dict:<d>>

depset:<depset:<s> | baz>
depset:<baz | depset:<s>>
`)
}

func TestPercentOperator(t *testing.T) {
	checkTypes(t, `
n = 3
s = "foo"

foo % n
foo % s
foo % bar

n % foo
s % foo

s %= foo
n %= foo

baz = unknown
baz %= s
baz

boq = unknown
boq %= n
boq
`, `
n = int:<3>
s = string:<"foo">

foo % int:<n>
string:<foo % string:<s>>
foo % bar

int:<int:<n> % foo>
string:<string:<s> % foo>

string:<s> %= foo
int:<n> %= foo

baz = unknown
baz %= string:<s>
string:<baz>

boq = unknown
boq %= int:<n>
boq
`)
}

func TestContext(t *testing.T) {
	checkTypes(t, `
def foobar(ctx, foo, bar):
    ctx
    ctx.actions
    ctx.actions.args()

    actions = ctx.actions
    not_args = actions.args
    args = actions.args()
    args
`, `
def foobar(ctx:<ctx>, foo, bar):
    ctx:<ctx>
    ctx.actions:<ctx:<ctx>.actions>
    ctx.actions.args:<ctx.actions:<ctx:<ctx>.actions>.args()>

    actions = ctx.actions:<ctx:<ctx>.actions>
    not_args = ctx.actions:<actions>.args
    args = ctx.actions.args:<ctx.actions:<actions>.args()>
    ctx.actions.args:<args>
`)
}

func TestContextFalse(t *testing.T) {
	checkTypes(t, `
def foobar(foo, bar):
    ctx
    ctx.actions
    ctx.actions.args()
`, `
def foobar(foo, bar):
    ctx
    ctx.actions
    ctx.actions.args()
`)
}
