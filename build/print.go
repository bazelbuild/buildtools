/*
Copyright 2016 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/
// Printing of syntax trees.

package build

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bazelbuild/buildtools/tables"
)

const nestedIndentation = 4 // Indentation of nested blocks
const listIndentation = 4   // Indentation of multiline expressions

// Format returns the formatted form of the given BUILD file.
func Format(f *File) []byte {
	pr := &printer{}
	pr.file(f)
	return pr.Bytes()
}

// FormatString returns the string form of the given expression.
func FormatString(x Expr) string {
	pr := &printer{}
	switch x := x.(type) {
	case *File:
		pr.file(x)
	default:
		pr.expr(x, precLow)
	}
	return pr.String()
}

// A printer collects the state during printing of a file or expression.
type printer struct {
	bytes.Buffer           // output buffer
	comment      []Comment // pending end-of-line comments
	margin       int       // left margin (indent), a number of spaces
	depth        int       // nesting depth inside ( ) [ ] { }
	needsNewLine bool      // true if the next statement needs a new line before it
}

// printf prints to the buffer.
func (p *printer) printf(format string, args ...interface{}) {
	fmt.Fprintf(p, format, args...)
}

// indent returns the position on the current line, in bytes, 0-indexed.
func (p *printer) indent() int {
	b := p.Bytes()
	n := 0
	for n < len(b) && b[len(b)-1-n] != '\n' {
		n++
	}
	return n
}

// newline ends the current line, flushing end-of-line comments.
// It must only be called when printing a newline is known to be safe:
// when not inside an expression or when p.depth > 0.
// To break a line inside an expression that might not be enclosed
// in brackets of some kind, use breakline instead.
func (p *printer) newline() {
	p.needsNewLine = false
	if len(p.comment) > 0 {
		p.printf("  ")
		for i, com := range p.comment {
			if i > 0 {
				p.trim()
				p.printf("\n%*s", p.margin, "")
			}
			p.printf("%s", strings.TrimSpace(com.Token))
		}
		p.comment = p.comment[:0]
	}

	p.trim()
	p.printf("\n%*s", p.margin, "")
}

// softNewline postpones a call to newline to the next call of p.newlineIfNeeded()
// If softNewline is called several times, just one newline is printed.
// Usecase: if there are several nested blocks ending at the same time, for instance
//
//     if True:
//         for a in b:
//             pass
//     foo()
//
// the last statement (`pass`) doesn't end with a newline, each block ends with a lazy newline
// which actually gets printed only once when right before the next statement (`foo()`) is printed.
func (p *printer) softNewline() {
	p.needsNewLine = true
}

// newlineIfNeeded calls newline if softNewline() has previously been called
func (p *printer) newlineIfNeeded() {
	if p.needsNewLine == true {
		p.newline()
	}
}

// breakline breaks the current line, inserting a continuation \ if needed.
// If no continuation \ is needed, breakline flushes end-of-line comments.
func (p *printer) breakline() {
	if p.depth == 0 {
		// Cannot have both final \ and comments.
		p.printf(" \\\n%*s", p.margin, "")
		return
	}

	// Safe to use newline.
	p.newline()
}

// trim removes trailing spaces from the current line.
func (p *printer) trim() {
	// Remove trailing space from line we're about to end.
	b := p.Bytes()
	n := len(b)
	for n > 0 && b[n-1] == ' ' {
		n--
	}
	p.Truncate(n)
}

// file formats the given file into the print buffer.
func (p *printer) file(f *File) {
	for _, com := range f.Before {
		p.printf("%s", strings.TrimSpace(com.Token))
		p.newline()
	}

	p.statements(f.Stmt)

	for _, com := range f.After {
		p.printf("%s", strings.TrimSpace(com.Token))
		p.newline()
	}

	p.newlineIfNeeded()
}

func (p *printer) statements(stmts []Expr) {
	for i, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *CommentBlock:
			// comments already handled

		case *PythonBlock:
			for _, com := range stmt.Before {
				p.printf("%s", strings.TrimSpace(com.Token))
				p.newline()
			}
			p.printf("%s", stmt.Token)

		default:
			p.expr(stmt, precLow)
		}

		// A CommentBlock is an empty statement without a body,
		// it doesn't need an line break after the body
		if _, ok := stmt.(*CommentBlock); !ok {
			p.softNewline()
		}

		for _, com := range stmt.Comment().After {
			p.newlineIfNeeded()
			p.printf("%s", strings.TrimSpace(com.Token))
			p.softNewline()
		}

		// Print an empty line break after the statement unless it's the last statement in the sequence.
		// In that case a line break should be printed when the block or the file ends.
		if i < len(stmts)-1 {
			p.newline()
		}

		if i+1 < len(stmts) && !compactStmt(stmt, stmts[i+1], p.margin == 0) {
			p.newline()
		}
	}
}

// compactStmt reports whether the pair of statements s1, s2
// should be printed without an intervening blank line.
// We omit the blank line when both are subinclude statements
// and the second one has no leading comments.
func compactStmt(s1, s2 Expr, isTopLevel bool) bool {
	if len(s2.Comment().Before) > 0 {
		return false
	}
	if isCall(s1, "load") && isCall(s2, "load") {
		// Load statements
		return true
	} else if tables.FormattingMode == tables.BuildMode && isTopLevel {
		// Top-level statements in a BUILD file
		return false
	} else if isFunctionDefinition(s1) || isFunctionDefinition(s2) {
		// On of the statements is a function definition
		return false
	} else {
		// Depend on how the statements have been printed in the original file
		_, end := s1.Span()
		start, _ := s2.Span()
		return start.Line-end.Line <= 1
	}
}

// isCall reports whether x is a call to a function with the given name.
func isCall(x Expr, name string) bool {
	c, ok := x.(*CallExpr)
	if !ok {
		return false
	}
	nam, ok := c.X.(*LiteralExpr)
	if !ok {
		return false
	}
	return nam.Token == name
}

// isCodeBlock checks if the statement is a code block (def, if, for, etc.)
func isCodeBlock(x Expr) bool {
	switch x.(type) {
	case *DefStmt:
		return true
	case *ForStmt:
		return true
	case *IfStmt:
		return true
	default:
		return false
	}
}

// isFunctionDefinition checks if the statement is a def code block
func isFunctionDefinition(x Expr) bool {
	switch x.(type) {
	case *DefStmt:
		return true
	default:
		return false
	}
}

// Expression formatting.

// The expression formatter must introduce parentheses to force the
// meaning described by the parse tree. We preserve parentheses in the
// input, so extra parentheses are only needed if we have edited the tree.
//
// For example consider these expressions:
//	(1) "x" "y" % foo
//	(2) "x" + "y" % foo
//	(3) "x" + ("y" % foo)
//	(4) ("x" + "y") % foo
// When we parse (1), we represent the concatenation as an addition.
// However, if we print the addition back out without additional parens,
// as in (2), it has the same meaning as (3), which is not the original
// meaning. To preserve the original meaning we must add parens as in (4).
//
// To allow arbitrary rewrites to be formatted properly, we track full
// operator precedence while printing instead of just handling this one
// case of string concatenation.
//
// The precedences are assigned values low to high. A larger number
// binds tighter than a smaller number. All binary operators bind
// left-to-right.
const (
	precLow = iota
	precAssign
	precComma
	precColon
	precIfElse
	precOr
	precAnd
	precCmp
	precAdd
	precMultiply
	precUnary
	precSuffix
)

// opPrec gives the precedence for operators found in a BinaryExpr.
var opPrec = map[string]int{
	"=":      precAssign,
	"+=":     precAssign,
	"-=":     precAssign,
	"*=":     precAssign,
	"/=":     precAssign,
	"//=":    precAssign,
	"%=":     precAssign,
	"or":     precOr,
	"and":    precAnd,
	"in":     precCmp,
	"not in": precCmp,
	"<":      precCmp,
	">":      precCmp,
	"==":     precCmp,
	"!=":     precCmp,
	"<=":     precCmp,
	">=":     precCmp,
	"+":      precAdd,
	"-":      precAdd,
	"*":      precMultiply,
	"/":      precMultiply,
	"//":     precMultiply,
	"%":      precMultiply,
}

// expr prints the expression v to the print buffer.
// The value outerPrec gives the precedence of the operator
// outside expr. If that operator binds tighter than v's operator,
// expr must introduce parentheses to preserve the meaning
// of the parse tree (see above).
func (p *printer) expr(v Expr, outerPrec int) {
	// Emit line-comments preceding this expression.
	// If we are in the middle of an expression but not inside ( ) [ ] { }
	// then we cannot just break the line: we'd have to end it with a \.
	// However, even then we can't emit line comments since that would
	// end the expression. This is only a concern if we have rewritten
	// the parse tree. If comments were okay before this expression in
	// the original input they're still okay now, in the absense of rewrites.
	//
	// TODO(bazel-team): Check whether it is valid to emit comments right now,
	// and if not, insert them earlier in the output instead, at the most
	// recent \n not following a \ line.
	p.newlineIfNeeded()

	if before := v.Comment().Before; len(before) > 0 {
		// Want to print a line comment.
		// Line comments must be at the current margin.
		p.trim()
		if p.indent() > 0 {
			// There's other text on the line. Start a new line.
			p.printf("\n")
		}
		// Re-indent to margin.
		p.printf("%*s", p.margin, "")
		for _, com := range before {
			p.printf("%s", strings.TrimSpace(com.Token))
			p.newline()
		}
	}

	// Do we introduce parentheses?
	// The result depends on the kind of expression.
	// Each expression type that might need parentheses
	// calls addParen with its own precedence.
	// If parentheses are necessary, addParen prints the
	// opening parenthesis and sets parenthesized so that
	// the code after the switch can print the closing one.
	parenthesized := false
	addParen := func(prec int) {
		if prec < outerPrec {
			p.printf("(")
			p.depth++
			parenthesized = true
		}
	}

	switch v := v.(type) {
	default:
		panic(fmt.Errorf("printer: unexpected type %T", v))

	case *LiteralExpr:
		p.printf("%s", v.Token)

	case *StringExpr:
		// If the Token is a correct quoting of Value, use it.
		// This preserves the specific escaping choices that
		// BUILD authors have made, and it also works around
		// b/7272572.
		if strings.HasPrefix(v.Token, `"`) {
			s, triple, err := unquote(v.Token)
			if s == v.Value && triple == v.TripleQuote && err == nil {
				p.printf("%s", v.Token)
				break
			}
		}

		p.printf("%s", quote(v.Value, v.TripleQuote))

	case *DotExpr:
		addParen(precSuffix)
		p.expr(v.X, precSuffix)
		p.printf(".%s", v.Name)

	case *IndexExpr:
		addParen(precSuffix)
		p.expr(v.X, precSuffix)
		p.printf("[")
		p.expr(v.Y, precLow)
		p.printf("]")

	case *KeyValueExpr:
		p.expr(v.Key, precLow)
		p.printf(": ")
		p.expr(v.Value, precLow)

	case *SliceExpr:
		addParen(precSuffix)
		p.expr(v.X, precSuffix)
		p.printf("[")
		if v.From != nil {
			p.expr(v.From, precLow)
		}
		p.printf(":")
		if v.To != nil {
			p.expr(v.To, precLow)
		}
		if v.SecondColon.Byte != 0 {
			p.printf(":")
			if v.Step != nil {
				p.expr(v.Step, precLow)
			}
		}
		p.printf("]")

	case *UnaryExpr:
		addParen(precUnary)
		if v.Op == "not" {
			p.printf("not ") // Requires a space after it.
		} else {
			p.printf("%s", v.Op)
		}
		p.expr(v.X, precUnary)

	case *LambdaExpr:
		addParen(precColon)
		p.printf("lambda ")
		for i, param := range v.Params {
			if i > 0 {
				p.printf(", ")
			}
			p.expr(param, precLow)
		}
		p.printf(": ")
		p.expr(v.Body[0], precLow) // lambdas should have exactly one statement

	case *BinaryExpr:
		// Precedence: use the precedence of the operator.
		// Since all binary expressions format left-to-right,
		// it is okay for the left side to reuse the same operator
		// without parentheses, so we use prec for v.X.
		// For the same reason, the right side cannot reuse the same
		// operator, or else a parse tree for a + (b + c), where the ( ) are
		// not present in the source, will format as a + b + c, which
		// means (a + b) + c. Treat the right expression as appearing
		// in a context one precedence level higher: use prec+1 for v.Y.
		//
		// Line breaks: if we are to break the line immediately after
		// the operator, introduce a margin at the current column,
		// so that the second operand lines up with the first one and
		// also so that neither operand can use space to the left.
		// If the operator is an =, indent the right side another 4 spaces.
		prec := opPrec[v.Op]
		addParen(prec)
		m := p.margin
		if v.LineBreak {
			p.margin = p.indent()
			if v.Op == "=" {
				p.margin += listIndentation
			}
		}

		p.expr(v.X, prec)
		p.printf(" %s", v.Op)
		if v.LineBreak {
			p.breakline()
		} else {
			p.printf(" ")
		}
		p.expr(v.Y, prec+1)
		p.margin = m

	case *ParenExpr:
		p.seq("()", &v.Start, &[]Expr{v.X}, &v.End, modeParen, false, v.ForceMultiLine)

	case *CallExpr:
		addParen(precSuffix)
		p.expr(v.X, precSuffix)
		p.seq("()", &v.ListStart, &v.List, &v.End, modeCall, v.ForceCompact, v.ForceMultiLine)

	case *ListExpr:
		p.seq("[]", &v.Start, &v.List, &v.End, modeList, false, v.ForceMultiLine)

	case *SetExpr:
		p.seq("{}", &v.Start, &v.List, &v.End, modeList, false, v.ForceMultiLine)

	case *TupleExpr:
		mode := modeTuple
		if !v.Start.IsValid() {
			mode = modeSeq
		}
		p.seq("()", &v.Start, &v.List, &v.End, mode, v.ForceCompact, v.ForceMultiLine)

	case *DictExpr:
		var list []Expr
		for _, x := range v.List {
			list = append(list, x)
		}
		p.seq("{}", &v.Start, &list, &v.End, modeDict, false, v.ForceMultiLine)

	case *Comprehension:
		p.listFor(v)

	case *ConditionalExpr:
		addParen(precSuffix)
		p.expr(v.Then, precIfElse)
		p.printf(" if ")
		p.expr(v.Test, precIfElse)
		p.printf(" else ")
		p.expr(v.Else, precIfElse)

	case *ReturnStmt:
		p.printf("return")
		if v.Result != nil {
			p.printf(" ")
			p.expr(v.Result, precSuffix)
		}

	case *DefStmt:
		p.printf("def ")
		p.printf(v.Name)
		p.seq("()", &v.StartPos, &v.Params, nil, modeCall, v.ForceCompact, v.ForceMultiLine)
		p.printf(":")
		p.margin += nestedIndentation
		p.newline()
		p.statements(v.Body)
		p.margin -= nestedIndentation

	case *ForStmt:
		p.printf("for ")
		p.expr(v.Vars, precLow)
		p.printf(" in ")
		p.expr(v.X, precLow)
		p.printf(":")
		p.margin += nestedIndentation
		p.newline()
		p.statements(v.Body)
		p.margin -= nestedIndentation

	case *IfStmt:
		block := v
		isFirst := true
		needsEmptyLine := false
		for {
			p.newlineIfNeeded()
			if !isFirst {
				if needsEmptyLine {
					p.newline()
				}
				p.printf("el")
			}
			p.printf("if ")
			p.expr(block.Cond, precLow)
			p.printf(":")
			p.margin += nestedIndentation
			p.newline()
			p.statements(block.True)
			p.margin -= nestedIndentation

			isFirst = false
			_, end := block.True[len(block.True)-1].Span()
			needsEmptyLine = block.ElsePos.Line-end.Line > 1

			if len(block.False) == 1 {
				next, ok := block.False[0].(*IfStmt)
				if ok {
					block = next
					continue
				}
			}
			break
		}

		if len(block.False) > 0 {
			p.newlineIfNeeded()
			if needsEmptyLine {
				p.newline()
			}
			p.printf("else:")
			p.margin += nestedIndentation
			p.newline()
			p.statements(block.False)
			p.margin -= nestedIndentation
		}
	}

	// Add closing parenthesis if needed.
	if parenthesized {
		p.depth--
		p.printf(")")
	}

	// Queue end-of-line comments for printing when we
	// reach the end of the line.
	p.comment = append(p.comment, v.Comment().Suffix...)
}

// A seqMode describes a formatting mode for a sequence of values,
// like a list or call arguments.
type seqMode int

const (
	_ seqMode = iota

	modeCall  // f(x)
	modeList  // [x]
	modeTuple // (x,)
	modeParen // (x)
	modeDict  // {x:y}
	modeSeq   // x, y
)

// useCompactMode reports whether a sequence should be formatted in a compact mode
func useCompactMode(start *Position, list *[]Expr, end *End, forceCompact, forceMultiLine bool) bool {
	// If there are line comments, use multiline
	// so we can print the comments before the closing bracket.
	for _, x := range *list {
		if len(x.Comment().Before) > 0 {
			return false
		}
	}
	if end != nil && len(end.Before) > 0 {
		return false
	}

	// Implicit tuples are always compact
	if !start.IsValid() {
		return true
	}

	// In Default printing mode try to keep the original printing style
	if tables.FormattingMode == tables.DefaultMode {
		// If every element (including the brackets) ends on the same line where the next element starts,
		// use the compact mode, otherwise use multiline mode
		previousEnd := start.Line
		for _, x := range *list {
			start, end := x.Span()
			if start.Line != previousEnd {
				return false
			}
			previousEnd = end.Line
		}
		if end != nil && previousEnd != end.Pos.Line {
			return false
		}
		return true
	}
	// In Build mode, use the forceMultiline and forceCompact values
	if forceMultiLine {
		return false
	}
	if forceCompact {
		return true
	}
	// If neither of the flags are set, use compact mode only for empty or 1-element sequences
	return len(*list) <= 1
}

// seq formats a list of values inside a given bracket pair (brack = "()", "[]", "{}").
// The end node holds any trailing comments to be printed just before the
// closing bracket.
// The mode parameter specifies the sequence mode (see above).
// If multiLine is true, seq avoids the compact form even
// for 0- and 1-element sequences.
func (p *printer) seq(brack string, start *Position, list *[]Expr, end *End, mode seqMode, forceCompact, forceMultiLine bool) {
	if mode != modeSeq {
		p.printf("%s", brack[:1])
	}
	p.depth++
	defer func() {
		p.depth--
		if mode != modeSeq {
			p.printf("%s", brack[1:])
		}
	}()

	if useCompactMode(start, list, end, forceCompact, forceMultiLine) {
		for i, x := range *list {
			if i > 0 {
				p.printf(", ")
			}
			p.expr(x, precLow)
		}
		// Single-element tuple must end with comma, to mark it as a tuple.
		if len(*list) == 1 && mode == modeTuple {
			p.printf(",")
		}
		return
	}
	// Multi-line form.
	p.margin += listIndentation
	for i, x := range *list {
		// If we are about to break the line before the first
		// element and there are trailing end-of-line comments
		// waiting to be printed, delay them and print them as
		// whole-line comments preceding that element.
		// Do this by printing a newline ourselves and positioning
		// so that the end-of-line comment, with the two spaces added,
		// will line up with the current margin.
		if i == 0 && len(p.comment) > 0 {
			p.printf("\n%*s", p.margin-2, "")
		}

		p.newline()
		p.expr(x, precLow)
		if mode != modeParen || i+1 < len(*list) {
			p.printf(",")
		}
	}
	// Final comments.
	if end != nil {
		for _, com := range end.Before {
			p.newline()
			p.printf("%s", strings.TrimSpace(com.Token))
		}
	}
	p.margin -= listIndentation
	p.newline()
}

// listFor formats a ListForExpr (list comprehension).
// The single-line form is:
//	[x for y in z if c]
//
// and the multi-line form is:
//	[
//	    x
//	    for y in z
//	    if c
//	]
//
func (p *printer) listFor(v *Comprehension) {
	multiLine := v.ForceMultiLine || len(v.End.Before) > 0

	// space breaks the line in multiline mode
	// or else prints a space.
	space := func() {
		if multiLine {
			p.breakline()
		} else {
			p.printf(" ")
		}
	}

	open, close := "[", "]"
	if v.Curly {
		open, close = "{", "}"
	}
	p.depth++
	p.printf("%s", open)

	if multiLine {
		p.margin += listIndentation
		p.newline()
	}

	p.expr(v.Body, precLow)

	for _, c := range v.Clauses {
		space()
		switch clause := c.(type) {
		case *ForClause:
			p.printf("for ")
			p.expr(clause.Vars, precLow)
			p.printf(" in ")
			p.expr(clause.X, precLow)
		case *IfClause:
			p.printf("if ")
			p.expr(clause.Cond, precLow)
		}
		p.comment = append(p.comment, c.Comment().Suffix...)
	}

	if multiLine {
		for _, com := range v.End.Before {
			p.newline()
			p.printf("%s", strings.TrimSpace(com.Token))
		}
		p.margin -= listIndentation
		p.newline()
	}

	p.printf("%s", close)
	p.depth--
}

func (p *printer) isTopLevel() bool {
	return p.margin == 0
}
