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

type ruleDef struct {
	filename string
	name     string
	line     int
}

func (rd ruleDef) format() string {
	return fmt.Sprintf(`  File %q, line %d
    %s = rule(...)`, rd.filename, rd.line, rd.name)
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
