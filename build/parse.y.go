//line build/parse.y:29
package build

import __yyfmt__ "fmt"

//line build/parse.y:29

//line build/parse.y:34
type yySymType struct {
	yys int
	// input tokens
	tok    string   // raw input syntax
	str    string   // decoding of quoted string
	pos    Position // position of token
	triple bool     // was string triple quoted?

	// partial syntax trees
	expr    Expr
	exprs   []Expr
	kv      *KeyValueExpr
	kvs     []*KeyValueExpr
	string  *StringExpr
	ifstmt  *IfStmt
	loadarg *struct {
		from Ident
		to   Ident
	}
	loadargs []*struct {
		from Ident
		to   Ident
	}
	def_header *DefStmt // partially filled in def statement, without the body

	// supporting information
	comma    Position // position of trailing comma in list, if present
	lastStmt Expr     // most recent rule, to attach line comments to
}

const _AUGM = 57346
const _AND = 57347
const _COMMENT = 57348
const _EOF = 57349
const _EQ = 57350
const _FOR = 57351
const _GE = 57352
const _IDENT = 57353
const _INT = 57354
const _IF = 57355
const _ELSE = 57356
const _ELIF = 57357
const _IN = 57358
const _IS = 57359
const _LAMBDA = 57360
const _LOAD = 57361
const _LE = 57362
const _NE = 57363
const _STAR_STAR = 57364
const _INT_DIV = 57365
const _BIT_LSH = 57366
const _BIT_RSH = 57367
const _ARROW = 57368
const _NOT = 57369
const _OR = 57370
const _STRING = 57371
const _DEF = 57372
const _RETURN = 57373
const _PASS = 57374
const _BREAK = 57375
const _CONTINUE = 57376
const _INDENT = 57377
const _UNINDENT = 57378
const ShiftInstead = 57379
const _ASSERT = 57380
const _UNARY = 57381

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"'%'",
	"'('",
	"')'",
	"'*'",
	"'+'",
	"','",
	"'-'",
	"'.'",
	"'/'",
	"':'",
	"'<'",
	"'='",
	"'>'",
	"'['",
	"']'",
	"'{'",
	"'}'",
	"'|'",
	"'&'",
	"'^'",
	"'~'",
	"_AUGM",
	"_AND",
	"_COMMENT",
	"_EOF",
	"_EQ",
	"_FOR",
	"_GE",
	"_IDENT",
	"_INT",
	"_IF",
	"_ELSE",
	"_ELIF",
	"_IN",
	"_IS",
	"_LAMBDA",
	"_LOAD",
	"_LE",
	"_NE",
	"_STAR_STAR",
	"_INT_DIV",
	"_BIT_LSH",
	"_BIT_RSH",
	"_ARROW",
	"_NOT",
	"_OR",
	"_STRING",
	"_DEF",
	"_RETURN",
	"_PASS",
	"_BREAK",
	"_CONTINUE",
	"_INDENT",
	"_UNINDENT",
	"ShiftInstead",
	"'\\n'",
	"_ASSERT",
	"_UNARY",
	"';'",
}

var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line build/parse.y:1059

// Go helper code.

// unary returns a unary expression with the given
// position, operator, and subexpression.
func unary(pos Position, op string, x Expr) Expr {
	return &UnaryExpr{
		OpStart: pos,
		Op:      op,
		X:       x,
	}
}

// binary returns a binary expression with the given
// operands, position, and operator.
func binary(x Expr, pos Position, op string, y Expr) Expr {
	_, xend := x.Span()
	ystart, _ := y.Span()

	switch op {
	case "=", "+=", "-=", "*=", "/=", "//=", "%=", "&=", "|=", "^=", "<<=", ">>=":
		return &AssignExpr{
			LHS:       x,
			OpPos:     pos,
			Op:        op,
			LineBreak: xend.Line < ystart.Line,
			RHS:       y,
		}
	}

	return &BinaryExpr{
		X:         x,
		OpStart:   pos,
		Op:        op,
		LineBreak: xend.Line < ystart.Line,
		Y:         y,
	}
}

// typed returns a TypedIdent expression
func typed(x, y Expr) *TypedIdent {
	return &TypedIdent{
		Ident: x.(*Ident),
		Type:  y,
	}
}

// isSimpleExpression returns whether an expression is simple and allowed to exist in
// compact forms of sequences.
// The formal criteria are the following: an expression is considered simple if it's
// a literal (variable, string or a number), a literal with a unary operator or an empty sequence.
func isSimpleExpression(expr *Expr) bool {
	switch x := (*expr).(type) {
	case *LiteralExpr, *StringExpr, *Ident:
		return true
	case *UnaryExpr:
		_, literal := x.X.(*LiteralExpr)
		_, ident := x.X.(*Ident)
		return literal || ident
	case *ListExpr:
		return len(x.List) == 0
	case *TupleExpr:
		return len(x.List) == 0
	case *DictExpr:
		return len(x.List) == 0
	case *SetExpr:
		return len(x.List) == 0
	default:
		return false
	}
}

// forceCompact returns the setting for the ForceCompact field for a call or tuple.
//
// NOTE 1: The field is called ForceCompact, not ForceSingleLine,
// because it only affects the formatting associated with the call or tuple syntax,
// not the formatting of the arguments. For example:
//
//	call([
//		1,
//		2,
//		3,
//	])
//
// is still a compact call even though it runs on multiple lines.
//
// In contrast the multiline form puts a linebreak after the (.
//
//	call(
//		[
//			1,
//			2,
//			3,
//		],
//	)
//
// NOTE 2: Because of NOTE 1, we cannot use start and end on the
// same line as a signal for compact mode: the formatting of an
// embedded list might move the end to a different line, which would
// then look different on rereading and cause buildifier not to be
// idempotent. Instead, we have to look at properties guaranteed
// to be preserved by the reformatting, namely that the opening
// paren and the first expression are on the same line and that
// each subsequent expression begins on the same line as the last
// one ended (no line breaks after comma).
func forceCompact(start Position, list []Expr, end Position) bool {
	if len(list) <= 1 {
		// The call or tuple will probably be compact anyway; don't force it.
		return false
	}

	// If there are any named arguments or non-string, non-literal
	// arguments, cannot force compact mode.
	line := start.Line
	for _, x := range list {
		start, end := x.Span()
		if start.Line != line {
			return false
		}
		line = end.Line
		if !isSimpleExpression(&x) {
			return false
		}
	}
	return end.Line == line
}

// forceMultiLine returns the setting for the ForceMultiLine field.
func forceMultiLine(start Position, list []Expr, end Position) bool {
	if len(list) > 1 {
		// The call will be multiline anyway, because it has multiple elements. Don't force it.
		return false
	}

	if len(list) == 0 {
		// Empty list: use position of brackets.
		return start.Line != end.Line
	}

	// Single-element list.
	// Check whether opening bracket is on different line than beginning of
	// element, or closing bracket is on different line than end of element.
	elemStart, elemEnd := list[0].Span()
	return start.Line != elemStart.Line || end.Line != elemEnd.Line
}

// forceMultiLineComprehension returns the setting for the ForceMultiLine field for a comprehension.
func forceMultiLineComprehension(start Position, expr Expr, clauses []Expr, end Position) bool {
	// Return true if there's at least one line break between start, expr, each clause, and end
	exprStart, exprEnd := expr.Span()
	if start.Line != exprStart.Line {
		return true
	}
	previousEnd := exprEnd
	for _, clause := range clauses {
		clauseStart, clauseEnd := clause.Span()
		if previousEnd.Line != clauseStart.Line {
			return true
		}
		previousEnd = clauseEnd
	}
	return previousEnd.Line != end.Line
}

// extractTrailingComments extracts trailing comments of an indented block starting with the first
// comment line with indentation less than the block indentation.
// The comments can either belong to CommentBlock statements or to the last non-comment statement
// as After-comments.
func extractTrailingComments(stmt Expr) []Expr {
	body := getLastBody(stmt)
	var comments []Expr
	if body != nil && len(*body) > 0 {
		// Get the current indentation level
		start, _ := (*body)[0].Span()
		indentation := start.LineRune

		// Find the last non-comment statement
		lastNonCommentIndex := -1
		for i, stmt := range *body {
			if _, ok := stmt.(*CommentBlock); !ok {
				lastNonCommentIndex = i
			}
		}
		if lastNonCommentIndex == -1 {
			return comments
		}

		// Iterate over the trailing comments, find the first comment line that's not indented enough,
		// dedent it and all the following comments.
		for i := lastNonCommentIndex; i < len(*body); i++ {
			stmt := (*body)[i]
			if comment := extractDedentedComment(stmt, indentation); comment != nil {
				// This comment and all the following CommentBlock statements are to be extracted.
				comments = append(comments, comment)
				comments = append(comments, (*body)[i+1:]...)
				*body = (*body)[:i+1]
				// If the current statement is a CommentBlock statement without any comment lines
				// it should be removed too.
				if i > lastNonCommentIndex && len(stmt.Comment().After) == 0 {
					*body = (*body)[:i]
				}
			}
		}
	}
	return comments
}

// extractDedentedComment extract the first comment line from `stmt` which indentation is smaller
// than `indentation`, and all following comment lines, and returns them in a newly created
// CommentBlock statement.
func extractDedentedComment(stmt Expr, indentation int) Expr {
	for i, line := range stmt.Comment().After {
		// line.Start.LineRune == 0 can't exist in parsed files, it indicates that the comment line
		// has been added by an AST modification. Don't take such lines into account.
		if line.Start.LineRune > 0 && line.Start.LineRune < indentation {
			// This and all the following lines should be dedented
			cb := &CommentBlock{
				Start:    line.Start,
				Comments: Comments{After: stmt.Comment().After[i:]},
			}
			stmt.Comment().After = stmt.Comment().After[:i]
			return cb
		}
	}
	return nil
}

// getLastBody returns the last body of a block statement (the only body for For- and DefStmt
// objects, the last in a if-elif-else chain
func getLastBody(stmt Expr) *[]Expr {
	switch block := stmt.(type) {
	case *DefStmt:
		return &block.Body
	case *ForStmt:
		return &block.Body
	case *IfStmt:
		if len(block.False) == 0 {
			return &block.True
		} else if len(block.False) == 1 {
			if next, ok := block.False[0].(*IfStmt); ok {
				// Recursively find the last block of the chain
				return getLastBody(next)
			}
		}
		return &block.False
	}
	return nil
}

//line yacctab:1
var yyExca = [...]int16{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 81,
	6, 56,
	-2, 129,
	-1, 172,
	20, 126,
	-2, 127,
}

const yyPrivate = 57344

const yyLast = 1117

var yyAct = [...]int16{
	21, 32, 149, 255, 239, 150, 108, 110, 190, 199,
	7, 97, 2, 163, 25, 155, 191, 43, 162, 107,
	9, 222, 245, 224, 177, 40, 44, 22, 89, 90,
	91, 92, 39, 51, 134, 95, 100, 103, 213, 56,
	53, 54, 55, 85, 176, 39, 204, 57, 94, 115,
	221, 116, 117, 223, 112, 193, 121, 122, 123, 124,
	125, 126, 127, 128, 129, 130, 131, 132, 133, 22,
	135, 136, 137, 138, 139, 140, 141, 142, 143, 58,
	22, 45, 243, 112, 15, 102, 36, 39, 105, 96,
	22, 194, 38, 87, 13, 146, 160, 78, 34, 167,
	35, 119, 166, 165, 217, 171, 266, 174, 86, 49,
	216, 111, 169, 22, 37, 170, 168, 48, 22, 79,
	165, 33, 120, 46, 15, 99, 185, 182, 178, 88,
	268, 39, 211, 47, 253, 186, 48, 81, 161, 252,
	236, 113, 114, 80, 165, 205, 232, 118, 157, 82,
	200, 197, 195, 56, 198, 157, 55, 59, 209, 60,
	56, 57, 210, 55, 59, 250, 60, 215, 57, 158,
	159, 72, 215, 208, 218, 220, 212, 152, 249, 206,
	44, 48, 212, 214, 154, 225, 219, 50, 228, 15,
	42, 227, 48, 58, 74, 75, 184, 104, 48, 200,
	58, 15, 145, 237, 238, 230, 181, 240, 235, 265,
	48, 151, 231, 204, 242, 172, 241, 156, 269, 229,
	196, 175, 144, 93, 106, 180, 192, 15, 1, 10,
	19, 201, 246, 248, 244, 254, 179, 251, 86, 101,
	247, 98, 41, 256, 258, 52, 20, 12, 8, 207,
	261, 262, 7, 4, 240, 31, 56, 264, 263, 55,
	59, 267, 60, 164, 57, 153, 15, 256, 271, 270,
	188, 189, 83, 84, 147, 233, 234, 148, 226, 0,
	201, 0, 0, 0, 0, 0, 0, 36, 0, 0,
	29, 0, 28, 38, 0, 0, 58, 74, 75, 34,
	0, 35, 0, 0, 0, 0, 30, 15, 0, 6,
	0, 0, 11, 192, 22, 37, 24, 0, 0, 0,
	0, 26, 33, 257, 0, 0, 15, 0, 0, 0,
	27, 0, 39, 23, 14, 16, 17, 18, 0, 259,
	36, 5, 0, 29, 0, 28, 38, 257, 0, 0,
	0, 0, 34, 0, 35, 0, 0, 0, 0, 30,
	0, 0, 6, 3, 0, 11, 0, 22, 37, 24,
	0, 0, 0, 0, 26, 33, 0, 0, 0, 0,
	0, 0, 0, 27, 0, 39, 23, 14, 16, 17,
	18, 0, 36, 0, 5, 29, 0, 28, 38, 0,
	0, 0, 0, 0, 34, 0, 35, 0, 0, 0,
	0, 30, 0, 0, 0, 0, 0, 0, 0, 22,
	37, 0, 0, 0, 0, 0, 26, 33, 0, 0,
	0, 0, 0, 0, 0, 27, 0, 39, 0, 14,
	16, 17, 18, 0, 56, 0, 109, 55, 59, 0,
	60, 0, 57, 0, 61, 260, 62, 0, 0, 0,
	0, 71, 72, 73, 0, 0, 70, 0, 0, 63,
	0, 66, 0, 0, 77, 0, 0, 67, 76, 0,
	0, 64, 65, 0, 58, 74, 75, 56, 68, 69,
	55, 59, 0, 60, 0, 57, 173, 61, 0, 62,
	0, 0, 0, 0, 71, 72, 73, 0, 0, 70,
	0, 0, 63, 0, 66, 0, 0, 77, 0, 0,
	67, 76, 0, 0, 64, 65, 0, 58, 74, 75,
	56, 68, 69, 55, 59, 0, 60, 0, 57, 0,
	61, 0, 62, 0, 0, 0, 0, 71, 72, 73,
	0, 0, 70, 0, 0, 63, 0, 66, 0, 0,
	77, 187, 0, 67, 76, 0, 0, 64, 65, 0,
	58, 74, 75, 56, 68, 69, 55, 59, 0, 60,
	0, 57, 0, 61, 183, 62, 0, 0, 0, 0,
	71, 72, 73, 0, 0, 70, 0, 0, 63, 0,
	66, 0, 0, 77, 0, 0, 67, 76, 0, 0,
	64, 65, 0, 58, 74, 75, 56, 68, 69, 55,
	59, 0, 60, 0, 57, 0, 61, 0, 62, 0,
	0, 0, 0, 71, 72, 73, 0, 0, 70, 0,
	0, 63, 165, 66, 0, 0, 77, 0, 0, 67,
	76, 0, 0, 64, 65, 0, 58, 74, 75, 56,
	68, 69, 55, 59, 0, 60, 0, 57, 0, 61,
	0, 62, 0, 0, 0, 0, 71, 72, 73, 0,
	0, 70, 0, 0, 63, 0, 66, 0, 0, 77,
	0, 0, 67, 76, 0, 0, 64, 65, 0, 58,
	74, 75, 36, 68, 69, 29, 0, 28, 38, 0,
	0, 0, 0, 0, 34, 0, 35, 0, 0, 0,
	0, 30, 0, 0, 0, 0, 0, 0, 0, 22,
	37, 0, 0, 0, 0, 0, 26, 33, 0, 0,
	0, 0, 0, 0, 0, 27, 0, 39, 0, 14,
	16, 17, 18, 56, 0, 0, 55, 59, 0, 60,
	0, 57, 0, 61, 0, 62, 0, 0, 0, 0,
	71, 72, 73, 0, 0, 70, 0, 0, 63, 0,
	66, 0, 0, 0, 0, 0, 67, 76, 0, 0,
	64, 65, 0, 58, 74, 75, 56, 68, 69, 55,
	59, 0, 60, 0, 57, 0, 61, 0, 62, 0,
	0, 0, 0, 71, 72, 73, 0, 0, 70, 0,
	0, 63, 0, 66, 0, 0, 0, 0, 0, 67,
	0, 0, 0, 64, 65, 0, 58, 74, 75, 56,
	68, 69, 55, 59, 0, 60, 0, 57, 0, 61,
	0, 62, 0, 0, 0, 0, 71, 72, 73, 0,
	0, 70, 0, 0, 63, 0, 66, 0, 0, 0,
	0, 0, 67, 0, 0, 0, 64, 65, 0, 58,
	74, 75, 56, 68, 0, 55, 59, 0, 60, 0,
	57, 0, 61, 0, 62, 0, 0, 0, 0, 71,
	72, 73, 0, 0, 0, 0, 0, 63, 0, 66,
	0, 0, 0, 0, 0, 67, 0, 0, 0, 64,
	65, 0, 58, 74, 75, 36, 68, 202, 29, 204,
	28, 38, 0, 0, 0, 0, 0, 34, 0, 35,
	0, 0, 0, 0, 30, 0, 0, 0, 0, 0,
	0, 0, 22, 37, 0, 0, 0, 0, 56, 26,
	33, 55, 59, 203, 60, 0, 57, 0, 27, 36,
	39, 202, 29, 0, 28, 38, 72, 73, 0, 0,
	0, 34, 0, 35, 0, 0, 0, 0, 30, 0,
	0, 0, 0, 0, 0, 0, 22, 37, 58, 74,
	75, 0, 0, 26, 33, 36, 0, 203, 29, 204,
	28, 38, 27, 0, 39, 0, 0, 34, 0, 35,
	0, 0, 0, 0, 30, 0, 0, 0, 0, 0,
	0, 0, 22, 37, 0, 0, 0, 0, 0, 26,
	33, 36, 0, 0, 29, 0, 28, 38, 27, 0,
	39, 0, 0, 34, 0, 35, 0, 0, 0, 0,
	30, 0, 0, 0, 0, 0, 0, 0, 22, 37,
	0, 0, 0, 0, 56, 26, 33, 55, 59, 0,
	60, 0, 57, 0, 27, 0, 39, 0, 0, 0,
	0, 71, 72, 73, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 58, 74, 75,
}

var yyPact = [...]int16{
	-1000, -1000, 335, -1000, -1000, -1000, -34, -1000, -1000, -1000,
	177, 81, -1000, 108, 1036, 174, -1000, -1000, -1000, -14,
	5, 655, -1000, 65, 1036, 132, 86, 1036, 1036, 1036,
	1036, -1000, -1000, 218, 1036, 1036, 1036, 186, 55, -1000,
	-1000, -43, 387, 74, 132, -1000, 1036, 1036, 1036, 201,
	1036, 1036, 1036, 88, -1000, 1036, 1036, 1036, 1036, 1036,
	1036, 1036, 1036, 1036, 1036, 1036, 1036, 1036, -3, 1036,
	1036, 1036, 1036, 1036, 1036, 1036, 1036, 1036, 217, 189,
	63, 202, 1036, 171, 208, -1000, 140, 58, 58, -1000,
	-1000, -1000, -1000, 202, 120, 612, 202, 73, 92, 206,
	483, 202, 215, 655, 11, -1000, -35, 697, -1000, -1000,
	-1000, 1036, 81, 201, 201, 749, 569, 655, 183, 387,
	-1000, -1000, -1000, -1000, -1000, 35, 35, 1070, 1070, 1070,
	1070, 1070, 1070, 1070, 1036, 835, 878, 954, 252, 149,
	156, 156, 792, 526, 48, 387, -1000, 214, 202, 964,
	204, -1000, 127, 166, 1036, -1000, 86, 1036, -1000, -1000,
	-18, -1000, 114, 4, -1000, 81, 1000, -1000, 90, -1000,
	84, 1000, -1000, 1036, 1000, -1000, -1000, -1000, -1000, -6,
	-36, 172, 132, 1036, 387, -1000, 1070, 1036, 213, 203,
	-1000, -1000, 133, 58, 58, -1000, -1000, -1000, 920, -1000,
	655, 125, 1036, 1036, -1000, -1000, 1036, -1000, -1000, 655,
	202, -1000, 4, 1036, 45, 655, -1000, -1000, 655, -1000,
	483, -1000, -37, -1000, -1000, 387, 201, -1000, 749, -1000,
	-1000, 48, 1036, 165, 152, -1000, 1036, 655, 655, 121,
	655, 37, 749, 1036, 282, -1000, -1000, -1000, 440, 1036,
	1036, 655, -1000, 1036, 200, -1000, -1000, 91, 749, -1000,
	1036, 655, 655, 112, 212, -5, -18, 655, -1000, -1000,
	-1000, -1000,
}

var yyPgo = [...]int16{
	0, 15, 5, 2, 9, 277, 274, 16, 273, 272,
	8, 271, 270, 0, 4, 48, 14, 94, 265, 89,
	17, 263, 13, 18, 81, 255, 12, 253, 248, 247,
	246, 245, 7, 20, 242, 11, 241, 239, 1, 6,
	236, 3, 235, 230, 229, 228, 225, 224,
}

var yyR1 = [...]int8{
	0, 45, 39, 39, 46, 46, 40, 40, 40, 26,
	26, 26, 26, 27, 27, 43, 44, 44, 28, 28,
	28, 30, 30, 29, 29, 31, 31, 32, 34, 34,
	33, 33, 33, 33, 33, 33, 33, 33, 33, 47,
	47, 16, 16, 16, 16, 16, 16, 16, 16, 16,
	16, 16, 16, 16, 16, 16, 6, 6, 5, 5,
	4, 4, 4, 4, 42, 42, 41, 41, 9, 9,
	12, 12, 8, 8, 11, 11, 7, 7, 7, 7,
	7, 10, 10, 10, 10, 10, 17, 17, 18, 18,
	13, 13, 13, 13, 13, 13, 13, 13, 13, 13,
	13, 13, 13, 13, 13, 13, 13, 13, 13, 13,
	13, 13, 13, 13, 13, 13, 13, 13, 13, 19,
	19, 14, 14, 15, 15, 1, 1, 2, 2, 3,
	3, 35, 37, 37, 36, 36, 36, 20, 20, 38,
	24, 25, 25, 25, 25, 21, 22, 22, 23, 23,
}

var yyR2 = [...]int8{
	0, 2, 5, 2, 0, 2, 0, 3, 2, 0,
	2, 2, 3, 1, 1, 5, 1, 3, 3, 6,
	1, 4, 5, 1, 4, 2, 1, 4, 0, 3,
	1, 2, 1, 3, 5, 3, 1, 1, 1, 0,
	1, 1, 1, 1, 3, 8, 4, 4, 6, 8,
	3, 4, 4, 3, 4, 3, 0, 2, 2, 3,
	1, 3, 2, 2, 1, 3, 1, 3, 0, 2,
	0, 2, 1, 3, 1, 3, 1, 3, 2, 1,
	2, 1, 3, 5, 4, 4, 1, 3, 0, 1,
	1, 4, 2, 2, 2, 2, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 4,
	3, 3, 3, 3, 3, 3, 3, 3, 5, 1,
	3, 0, 1, 0, 2, 0, 1, 1, 2, 0,
	1, 3, 1, 3, 0, 1, 2, 1, 3, 1,
	1, 3, 2, 2, 1, 4, 1, 3, 1, 2,
}

var yyChk = [...]int16{
	-1000, -45, -26, 28, -27, 59, 27, -32, -28, -33,
	-44, 30, -29, -17, 52, -24, 53, 54, 55, -43,
	-30, -13, 32, 51, 34, -16, 39, 48, 10, 8,
	24, -25, -38, 40, 17, 19, 5, 33, 11, 50,
	59, -34, 13, -20, -16, -24, 15, 25, 9, -17,
	13, 47, -31, 35, 36, 7, 4, 12, 44, 8,
	10, 14, 16, 29, 41, 42, 31, 37, 48, 49,
	26, 21, 22, 23, 45, 46, 38, 34, 32, -17,
	11, 5, 17, -9, -8, -7, -24, 7, 43, -13,
	-13, -13, -13, 5, -15, -13, -19, -35, -36, -19,
	-13, -37, -15, -13, 11, 33, -47, 62, -39, 59,
	-32, 37, 9, -17, -17, -13, -13, -13, -17, 13,
	34, -13, -13, -13, -13, -13, -13, -13, -13, -13,
	-13, -13, -13, -13, 37, -13, -13, -13, -13, -13,
	-13, -13, -13, -13, 5, 13, 32, -6, -5, -3,
	-2, 9, -17, -18, 13, -1, 9, 15, -24, -24,
	-3, 18, -23, -22, -21, 30, -2, -3, -23, 20,
	-1, -2, 9, 13, -2, 6, 33, 59, -33, -40,
	-46, -17, -16, 15, 13, -39, -13, 35, -12, -11,
	-10, -7, -24, 7, 43, -39, 6, -3, -2, -4,
	-13, -24, 7, 43, 9, 18, 13, -17, -7, -13,
	-38, 18, -22, 34, -20, -13, 20, 20, -13, -35,
	-13, 56, 27, 59, 59, 13, -17, -39, -13, 6,
	-1, 9, 13, -24, -24, -4, 15, -13, -13, -14,
	-13, -2, -13, 37, -26, 59, -39, -10, -13, 13,
	13, -13, 18, 13, -42, -41, -38, -24, -13, 57,
	15, -13, -13, -14, -3, 9, 15, -13, 18, 6,
	-41, -38,
}

var yyDef = [...]int16{
	9, -2, 0, 1, 10, 11, 0, 13, 14, 28,
	0, 0, 20, 30, 32, 41, 36, 37, 38, 16,
	23, 86, 140, 0, 0, 90, 68, 0, 0, 0,
	0, 42, 43, 0, 123, 134, 123, 144, 0, 139,
	12, 39, 0, 0, 137, 41, 0, 0, 0, 31,
	0, 0, 0, 0, 26, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, -2, 88, 0, 125, 72, 76, 79, 0, 92,
	93, 94, 95, 129, 0, 119, 129, 132, 0, 125,
	119, 135, 0, 119, 142, 143, 0, 40, 18, 6,
	4, 0, 0, 33, 35, 87, 0, 17, 0, 0,
	25, 96, 97, 98, 99, 100, 101, 102, 103, 104,
	105, 106, 107, 108, 0, 110, 111, 112, 113, 114,
	115, 116, 117, 0, 70, 0, 44, 0, 129, 0,
	130, 127, 89, 0, 0, 69, 126, 0, 78, 80,
	0, 50, 0, 148, 146, 0, 130, 124, 0, 53,
	0, 0, -2, 0, 136, 55, 141, 27, 29, 0,
	3, 0, 138, 0, 0, 24, 109, 0, 0, 125,
	74, 81, 76, 79, 0, 21, 46, 57, 130, 58,
	60, 41, 0, 0, 128, 47, 121, 91, 73, 77,
	0, 51, 149, 0, 0, 120, 52, 54, 131, 133,
	0, 9, 0, 8, 5, 0, 34, 22, 118, 15,
	71, 126, 0, 78, 80, 59, 0, 62, 63, 0,
	122, 0, 147, 0, 0, 7, 19, 75, 82, 0,
	0, 61, 48, 121, 129, 64, 66, 0, 145, 2,
	0, 84, 85, 0, 0, 127, 0, 83, 49, 45,
	65, 67,
}

var yyTok1 = [...]int8{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	59, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 22, 3,
	5, 6, 7, 8, 9, 10, 11, 12, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 62,
	14, 15, 16, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 17, 3, 18, 23, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 19, 21, 20, 24,
}

var yyTok2 = [...]int8{
	2, 3, 25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 39, 40, 41, 42,
	43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
	53, 54, 55, 56, 57, 58, 60, 61,
}

var yyTok3 = [...]int8{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := int(yyPact[state])
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && int(yyChk[int(yyAct[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || int(yyExca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := int(yyExca[i])
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = int(yyTok1[0])
		goto out
	}
	if char < len(yyTok1) {
		token = int(yyTok1[char])
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = int(yyTok2[char-yyPrivate])
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = int(yyTok3[i+0])
		if token == char {
			token = int(yyTok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int(yyTok2[1]) /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = int(yyPact[yystate])
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = int(yyAct[yyn])
	if int(yyChk[yyn]) == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = int(yyDef[yystate])
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && int(yyExca[xi+1]) == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = int(yyExca[xi+0])
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = int(yyExca[xi+1])
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = int(yyPact[yyS[yyp].yys]) + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = int(yyAct[yyn]) /* simulate a shift of "error" */
					if int(yyChk[yystate]) == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= int(yyR2[yyn])
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = int(yyR1[yyn])
	yyg := int(yyPgo[yyn])
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = int(yyAct[yyg])
	} else {
		yystate = int(yyAct[yyj])
		if int(yyChk[yystate]) != -yyn {
			yystate = int(yyAct[yyg])
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:218
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:225
		{
			statements := yyDollar[4].exprs
			if yyDollar[2].exprs != nil {
				// $2 can only contain *CommentBlock objects, each of them contains a non-empty After slice
				cb := yyDollar[2].exprs[len(yyDollar[2].exprs)-1].(*CommentBlock)
				// $4 can't be empty and can't start with a comment
				stmt := yyDollar[4].exprs[0]
				start, _ := stmt.Span()
				if start.Line-cb.After[len(cb.After)-1].Start.Line == 1 {
					// The first statement of $4 starts on the next line after the last comment of $2.
					// Attach the last comment to the first statement
					stmt.Comment().Before = cb.After
					yyDollar[2].exprs = yyDollar[2].exprs[:len(yyDollar[2].exprs)-1]
				}
				statements = append(yyDollar[2].exprs, yyDollar[4].exprs...)
			}
			yyVAL.exprs = statements
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 3:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:245
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:253
		{
			yyVAL.exprs = nil
			yyVAL.lastStmt = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:258
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].lastStmt
			if yyVAL.lastStmt == nil {
				cb := &CommentBlock{Start: yyDollar[2].pos}
				yyVAL.exprs = append(yyVAL.exprs, cb)
				yyVAL.lastStmt = cb
			}
			com := yyVAL.lastStmt.Comment()
			com.After = append(com.After, Comment{Start: yyDollar[2].pos, Token: yyDollar[2].tok})
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:270
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = nil
		}
	case 9:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:276
		{
			yyVAL.exprs = nil
			yyVAL.lastStmt = nil
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:281
		{
			// If this statement follows a comment block,
			// attach the comments to the statement.
			if cb, ok := yyDollar[1].lastStmt.(*CommentBlock); ok {
				yyVAL.exprs = append(yyDollar[1].exprs[:len(yyDollar[1].exprs)-1], yyDollar[2].exprs...)
				yyDollar[2].exprs[0].Comment().Before = cb.After
				yyVAL.lastStmt = yyDollar[2].lastStmt
				break
			}

			// Otherwise add to list.
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
			yyVAL.lastStmt = yyDollar[2].lastStmt

			// Consider this input:
			//
			//	foo()
			//	# bar
			//	baz()
			//
			// If we've just parsed baz(), the # bar is attached to
			// foo() as an After comment. Make it a Before comment
			// for baz() instead.
			if x := yyDollar[1].lastStmt; x != nil {
				com := x.Comment()
				// stmt is never empty
				yyDollar[2].exprs[0].Comment().Before = com.After
				com.After = nil
			}
		}
	case 11:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:312
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:318
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].lastStmt
			if yyVAL.lastStmt == nil {
				cb := &CommentBlock{Start: yyDollar[2].pos}
				yyVAL.exprs = append(yyVAL.exprs, cb)
				yyVAL.lastStmt = cb
			}
			com := yyVAL.lastStmt.Comment()
			com.After = append(com.After, Comment{Start: yyDollar[2].pos, Token: yyDollar[2].tok})
		}
	case 13:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:332
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].exprs[len(yyDollar[1].exprs)-1]
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:337
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
			yyVAL.lastStmt = yyDollar[1].expr
			if cbs := extractTrailingComments(yyDollar[1].expr); len(cbs) > 0 {
				yyVAL.exprs = append(yyVAL.exprs, cbs...)
				yyVAL.lastStmt = cbs[len(cbs)-1]
				if yyDollar[1].lastStmt == nil {
					yyVAL.lastStmt = nil
				}
			}
		}
	case 15:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:351
		{
			yyVAL.def_header = &DefStmt{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[4].exprs,
				},
				Name:           yyDollar[2].tok,
				ForceCompact:   forceCompact(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
				ForceMultiLine: forceMultiLine(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
			}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:366
		{
			yyDollar[1].def_header.Type = yyDollar[3].expr
			yyVAL.def_header = yyDollar[1].def_header
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:373
		{
			yyDollar[1].def_header.Function.Body = yyDollar[3].exprs
			yyDollar[1].def_header.ColonPos = yyDollar[2].pos
			yyVAL.expr = yyDollar[1].def_header
			yyVAL.lastStmt = yyDollar[3].lastStmt
		}
	case 19:
		yyDollar = yyS[yypt-6 : yypt+1]
//line build/parse.y:380
		{
			yyVAL.expr = &ForStmt{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				X:    yyDollar[4].expr,
				Body: yyDollar[6].exprs,
			}
			yyVAL.lastStmt = yyDollar[6].lastStmt
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:390
		{
			yyVAL.expr = yyDollar[1].ifstmt
			yyVAL.lastStmt = yyDollar[1].lastStmt
		}
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:398
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].exprs,
			}
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 22:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:407
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = End{Pos: yyDollar[2].pos}
			inner.False = []Expr{
				&IfStmt{
					If:   yyDollar[2].pos,
					Cond: yyDollar[3].expr,
					True: yyDollar[5].exprs,
				},
			}
			yyVAL.lastStmt = yyDollar[5].lastStmt
		}
	case 24:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:428
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = End{Pos: yyDollar[2].pos}
			inner.False = yyDollar[4].exprs
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:445
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastStmt = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 28:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:451
		{
			yyVAL.exprs = []Expr{}
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:455
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:462
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:469
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:474
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 34:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:475
		{
			yyVAL.expr = binary(typed(yyDollar[1].expr, yyDollar[3].expr), yyDollar[4].pos, yyDollar[4].tok, yyDollar[5].expr)
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:476
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 36:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:478
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 37:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:485
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:492
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:506
		{
			yyVAL.expr = yyDollar[1].string
		}
	case 44:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:510
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 45:
		yyDollar = yyS[yypt-8 : yypt+1]
//line build/parse.y:519
		{
			load := &LoadStmt{
				Load:         yyDollar[1].pos,
				Module:       yyDollar[4].string,
				Rparen:       End{Pos: yyDollar[8].pos},
				ForceCompact: yyDollar[2].pos.Line == yyDollar[8].pos.Line,
			}
			for _, arg := range yyDollar[6].loadargs {
				load.From = append(load.From, &arg.from)
				load.To = append(load.To, &arg.to)
			}
			yyVAL.expr = load
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:533
		{
			yyVAL.expr = &CallExpr{
				X:              yyDollar[1].expr,
				ListStart:      yyDollar[2].pos,
				List:           yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceCompact:   forceCompact(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
				ForceMultiLine: forceMultiLine(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 47:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:544
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 48:
		yyDollar = yyS[yypt-6 : yypt+1]
//line build/parse.y:553
		{
			yyVAL.expr = &SliceExpr{
				X:          yyDollar[1].expr,
				SliceStart: yyDollar[2].pos,
				From:       yyDollar[3].expr,
				FirstColon: yyDollar[4].pos,
				To:         yyDollar[5].expr,
				End:        yyDollar[6].pos,
			}
		}
	case 49:
		yyDollar = yyS[yypt-8 : yypt+1]
//line build/parse.y:564
		{
			yyVAL.expr = &SliceExpr{
				X:           yyDollar[1].expr,
				SliceStart:  yyDollar[2].pos,
				From:        yyDollar[3].expr,
				FirstColon:  yyDollar[4].pos,
				To:          yyDollar[5].expr,
				SecondColon: yyDollar[6].pos,
				Step:        yyDollar[7].expr,
				End:         yyDollar[8].pos,
			}
		}
	case 50:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:577
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:586
		{
			yyVAL.expr = &Comprehension{
				Curly:          false,
				Lbrack:         yyDollar[1].pos,
				Body:           yyDollar[2].expr,
				Clauses:        yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLineComprehension(yyDollar[1].pos, yyDollar[2].expr, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 52:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:597
		{
			yyVAL.expr = &Comprehension{
				Curly:          true,
				Lbrack:         yyDollar[1].pos,
				Body:           yyDollar[2].kv,
				Clauses:        yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLineComprehension(yyDollar[1].pos, yyDollar[2].kv, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 53:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:608
		{
			exprValues := make([]Expr, 0, len(yyDollar[2].kvs))
			for _, kv := range yyDollar[2].kvs {
				exprValues = append(exprValues, Expr(kv))
			}
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].kvs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, exprValues, yyDollar[3].pos),
			}
		}
	case 54:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:621
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[4].pos),
			}
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:630
		{
			if len(yyDollar[2].exprs) == 1 && yyDollar[2].comma.Line == 0 {
				// Just a parenthesized expression, not a tuple.
				yyVAL.expr = &ParenExpr{
					Start:          yyDollar[1].pos,
					X:              yyDollar[2].exprs[0],
					End:            End{Pos: yyDollar[3].pos},
					ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
				}
			} else {
				yyVAL.expr = &TupleExpr{
					Start:          yyDollar[1].pos,
					List:           yyDollar[2].exprs,
					End:            End{Pos: yyDollar[3].pos},
					ForceCompact:   forceCompact(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
					ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
				}
			}
		}
	case 56:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:651
		{
			yyVAL.exprs = nil
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:655
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:661
		{
			yyVAL.exprs = []Expr{yyDollar[2].expr}
		}
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:665
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:672
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:676
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 63:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:680
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:685
		{
			yyVAL.loadargs = []*struct {
				from Ident
				to   Ident
			}{yyDollar[1].loadarg}
		}
	case 65:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:689
		{
			yyDollar[1].loadargs = append(yyDollar[1].loadargs, yyDollar[3].loadarg)
			yyVAL.loadargs = yyDollar[1].loadargs
		}
	case 66:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:695
		{
			start := yyDollar[1].string.Start.add("'")
			if yyDollar[1].string.TripleQuote {
				start = start.add("''")
			}
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[1].string.Value,
					NamePos: start,
				},
				to: Ident{
					Name:    yyDollar[1].string.Value,
					NamePos: start,
				},
			}
		}
	case 67:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:712
		{
			start := yyDollar[3].string.Start.add("'")
			if yyDollar[3].string.TripleQuote {
				start = start.add("''")
			}
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[3].string.Value,
					NamePos: start,
				},
				to: *yyDollar[1].expr.(*Ident),
			}
		}
	case 68:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:727
		{
			yyVAL.exprs = nil
		}
	case 69:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:731
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 70:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:736
		{
			yyVAL.exprs = nil
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:740
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:746
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 73:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:750
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:757
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 75:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:761
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:768
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 78:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:772
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 79:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:776
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, nil)
		}
	case 80:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:780
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:789
		{
			yyVAL.expr = typed(yyDollar[1].expr, yyDollar[3].expr)
		}
	case 83:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:793
		{
			yyVAL.expr = binary(typed(yyDollar[1].expr, yyDollar[3].expr), yyDollar[4].pos, yyDollar[4].tok, yyDollar[5].expr)
		}
	case 84:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:797
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, typed(yyDollar[2].expr, yyDollar[4].expr))
		}
	case 85:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:801
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, typed(yyDollar[2].expr, yyDollar[4].expr))
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:808
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || !tuple.NoBrackets {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					NoBrackets:     true,
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 88:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:823
		{
			yyVAL.expr = nil
		}
	case 91:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:831
		{
			yyVAL.expr = &LambdaExpr{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[2].exprs,
					Body:     []Expr{yyDollar[4].expr},
				},
			}
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:840
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:841
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 94:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:842
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 95:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:843
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:844
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 97:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:845
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 98:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:846
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:847
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:848
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:849
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:850
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:851
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:852
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 105:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:853
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:854
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 107:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:855
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:856
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 109:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:857
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:858
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:859
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 112:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:860
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 113:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:861
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 114:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:862
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 115:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:863
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 116:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:864
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:866
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 118:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:874
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 119:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:886
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:890
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 121:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:895
		{
			yyVAL.expr = nil
		}
	case 123:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:901
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 124:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:905
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 125:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:915
		{
			yyVAL.pos = Position{}
		}
	case 128:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:926
		{
			yyVAL.pos = yyDollar[1].pos
		}
	case 129:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:934
		{
			yyVAL.pos = Position{}
		}
	case 131:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:941
		{
			yyVAL.kv = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 132:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:951
		{
			yyVAL.kvs = []*KeyValueExpr{yyDollar[1].kv}
		}
	case 133:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:955
		{
			yyVAL.kvs = append(yyDollar[1].kvs, yyDollar[3].kv)
		}
	case 134:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:960
		{
			yyVAL.kvs = nil
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:964
		{
			yyVAL.kvs = yyDollar[1].kvs
		}
	case 136:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:968
		{
			yyVAL.kvs = yyDollar[1].kvs
		}
	case 138:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:975
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || !tuple.NoBrackets {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					NoBrackets:     true,
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:991
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1003
		{
			yyVAL.expr = &Ident{NamePos: yyDollar[1].pos, Name: yyDollar[1].tok}
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:1009
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok + "." + yyDollar[3].tok}
		}
	case 142:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:1013
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok + "."}
		}
	case 143:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:1017
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: "." + yyDollar[2].tok}
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1021
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 145:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:1027
		{
			yyVAL.expr = &ForClause{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				In:   yyDollar[3].pos,
				X:    yyDollar[4].expr,
			}
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1038
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:1042
		{
			yyVAL.exprs = append(yyDollar[1].exprs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1051
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 149:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:1055
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
		}
	}
	goto yystack /* stack new state and value */
}
