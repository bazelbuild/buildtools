package warn

import (
	"bytes"
	"fmt"
)

type frame interface {
	format() string
}

type function struct {
	filename string // .bzl file where the function is defined
	name     string // original name of the function
}

// function call node (`foo(...)`)
type funCall struct {
	function
	filename  string // .bzl file where the function is called from
	nameAlias string // function name alias
	caller    string // name of the caller function
	line      int    // line on which the function is being called
}

func (fc funCall) format() string {
	return fmt.Sprintf(`  File %q, line %d, in %s
    %s(...)`, fc.filename, fc.line, fc.caller, fc.nameAlias)
}

// rule definition node (`foo = rule(...)`)
type ruleDef struct {
	function
	line int
}

func (rd ruleDef) format() string {
	return fmt.Sprintf(`  File %q, line %d
    %s = rule(...)`, rd.filename, rd.line, rd.name)
}

// alias node (`foo = bar`)
type alias struct {
	function
	filename string // .bzl file where the alias is defined at
	oldName  string
	newName  string
	line     int
}

func (a alias) format() string {
	return fmt.Sprintf(`  File %q, line %d
    %s = %s`, a.filename, a.line, a.newName, a.oldName)
}

func formatStackTrace(stackTrace []frame) string {
	buffer := bytes.Buffer{}
	buffer.WriteString("Example stack trace (statically analyzed):")
	for i := len(stackTrace) - 1; i >= 0; i-- {
		f := stackTrace[i]
		buffer.WriteRune('\n')
		buffer.WriteString(f.format())
	}
	return buffer.String()
}
