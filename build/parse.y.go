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

//line build/parse.y:1035

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
	case "=", "+=", "-=", "*=", "/=", "//=", "%=", "|=":
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
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 932

var yyAct = [...]int{
	20, 232, 229, 108, 186, 31, 7, 29, 154, 2,
	106, 146, 162, 95, 43, 9, 161, 187, 23, 105,
	215, 238, 217, 174, 40, 39, 87, 88, 89, 90,
	44, 49, 84, 36, 93, 98, 101, 54, 206, 131,
	53, 57, 83, 58, 36, 55, 51, 52, 113, 214,
	114, 39, 216, 110, 118, 119, 120, 121, 122, 123,
	124, 125, 126, 127, 128, 129, 130, 13, 132, 133,
	134, 135, 136, 137, 138, 139, 140, 56, 72, 73,
	147, 236, 48, 92, 189, 173, 85, 148, 103, 110,
	77, 94, 143, 157, 158, 116, 35, 159, 76, 27,
	164, 26, 38, 166, 210, 168, 169, 259, 33, 36,
	34, 36, 167, 111, 112, 28, 117, 109, 115, 100,
	190, 175, 86, 36, 37, 54, 97, 181, 53, 179,
	24, 32, 182, 55, 160, 245, 224, 209, 156, 25,
	244, 39, 248, 14, 15, 16, 17, 164, 151, 188,
	196, 197, 54, 191, 193, 53, 57, 202, 58, 195,
	55, 47, 47, 84, 204, 56, 208, 243, 45, 79,
	198, 211, 213, 201, 205, 78, 164, 178, 46, 207,
	205, 80, 156, 44, 220, 212, 47, 35, 242, 199,
	218, 219, 56, 38, 222, 147, 228, 225, 226, 33,
	230, 34, 148, 153, 47, 42, 227, 235, 180, 233,
	102, 234, 47, 247, 36, 37, 142, 223, 203, 194,
	47, 200, 32, 171, 237, 241, 165, 155, 240, 239,
	255, 188, 39, 221, 192, 172, 141, 249, 91, 104,
	246, 7, 177, 252, 253, 1, 230, 10, 254, 256,
	18, 231, 258, 233, 257, 234, 35, 176, 99, 27,
	96, 26, 38, 41, 50, 19, 12, 8, 33, 4,
	34, 30, 163, 152, 184, 28, 185, 81, 6, 82,
	144, 11, 145, 36, 37, 22, 0, 0, 0, 0,
	24, 32, 0, 0, 0, 0, 0, 0, 0, 25,
	0, 39, 21, 14, 15, 16, 17, 0, 250, 35,
	5, 0, 27, 0, 26, 38, 0, 0, 0, 0,
	0, 33, 0, 34, 0, 0, 0, 0, 28, 0,
	0, 6, 3, 0, 11, 0, 36, 37, 22, 0,
	0, 0, 54, 24, 32, 53, 57, 0, 58, 0,
	55, 0, 25, 0, 39, 21, 14, 15, 16, 17,
	70, 35, 0, 5, 27, 0, 26, 38, 0, 0,
	0, 0, 0, 33, 0, 34, 0, 0, 0, 0,
	28, 0, 56, 72, 73, 0, 0, 0, 36, 37,
	0, 0, 0, 0, 0, 24, 32, 0, 0, 0,
	0, 0, 0, 0, 25, 0, 39, 0, 14, 15,
	16, 17, 0, 54, 0, 107, 53, 57, 0, 58,
	0, 55, 0, 59, 251, 60, 0, 0, 0, 0,
	69, 70, 71, 0, 0, 68, 0, 0, 61, 0,
	64, 0, 0, 75, 0, 0, 65, 74, 0, 0,
	62, 63, 0, 56, 72, 73, 54, 66, 67, 53,
	57, 0, 58, 0, 55, 170, 59, 0, 60, 0,
	0, 0, 0, 69, 70, 71, 0, 0, 68, 0,
	0, 61, 0, 64, 0, 0, 75, 0, 0, 65,
	74, 0, 0, 62, 63, 0, 56, 72, 73, 54,
	66, 67, 53, 57, 0, 58, 0, 55, 0, 59,
	0, 60, 0, 0, 0, 0, 69, 70, 71, 0,
	0, 68, 0, 0, 61, 0, 64, 0, 0, 75,
	183, 0, 65, 74, 0, 0, 62, 63, 0, 56,
	72, 73, 54, 66, 67, 53, 57, 0, 58, 0,
	55, 0, 59, 0, 60, 0, 0, 0, 0, 69,
	70, 71, 0, 0, 68, 0, 0, 61, 164, 64,
	0, 0, 75, 0, 0, 65, 74, 0, 0, 62,
	63, 0, 56, 72, 73, 54, 66, 67, 53, 57,
	0, 58, 0, 55, 0, 59, 0, 60, 0, 0,
	0, 0, 69, 70, 71, 0, 0, 68, 0, 0,
	61, 0, 64, 0, 0, 75, 0, 0, 65, 74,
	0, 0, 62, 63, 0, 56, 72, 73, 54, 66,
	67, 53, 57, 0, 58, 0, 55, 0, 59, 0,
	60, 0, 0, 0, 0, 69, 70, 71, 0, 0,
	68, 0, 0, 61, 0, 64, 0, 0, 0, 0,
	0, 65, 74, 0, 0, 62, 63, 0, 56, 72,
	73, 54, 66, 67, 53, 57, 0, 58, 0, 55,
	0, 59, 0, 60, 0, 0, 0, 0, 69, 70,
	71, 0, 0, 68, 0, 0, 61, 0, 64, 0,
	0, 0, 0, 0, 65, 0, 0, 0, 62, 63,
	0, 56, 72, 73, 54, 66, 67, 53, 57, 0,
	58, 0, 55, 0, 59, 0, 60, 0, 0, 0,
	0, 69, 70, 71, 0, 0, 68, 0, 0, 61,
	0, 64, 0, 0, 0, 0, 0, 65, 0, 0,
	0, 62, 63, 0, 56, 72, 73, 54, 66, 0,
	53, 57, 0, 58, 0, 55, 0, 59, 0, 60,
	0, 0, 0, 0, 69, 70, 71, 0, 0, 0,
	0, 0, 61, 0, 64, 0, 0, 0, 0, 0,
	65, 0, 0, 0, 62, 63, 0, 56, 72, 73,
	35, 66, 149, 27, 0, 26, 38, 0, 0, 0,
	0, 0, 33, 0, 34, 0, 0, 0, 0, 28,
	0, 0, 0, 0, 0, 0, 0, 36, 37, 0,
	0, 0, 0, 0, 24, 32, 35, 0, 150, 27,
	0, 26, 38, 25, 0, 39, 0, 0, 33, 0,
	34, 0, 0, 0, 0, 28, 0, 0, 0, 0,
	0, 0, 0, 36, 37, 0, 0, 0, 0, 54,
	24, 32, 53, 57, 0, 58, 0, 55, 0, 25,
	0, 39, 0, 0, 0, 0, 69, 70, 71, 54,
	0, 0, 53, 57, 0, 58, 0, 55, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 70, 71, 56,
	72, 73, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 56,
	72, 73,
}

var yyPact = [...]int{
	-1000, -1000, 304, -1000, -1000, -1000, -35, -1000, -1000, -1000,
	192, 182, -1000, 153, 831, -1000, -1000, -1000, -16, 11,
	581, 66, 831, 164, 79, 831, 831, 831, 831, -1000,
	-1000, -1000, 233, 831, 831, 831, -1000, 199, 55, -1000,
	-1000, -43, 356, 80, 164, 831, 831, 831, 211, 831,
	831, 82, -1000, 831, 831, 831, 831, 831, 831, 831,
	831, 831, 831, 831, 831, 831, 2, 831, 831, 831,
	831, 831, 831, 831, 831, 831, 231, 203, 60, 795,
	831, 190, 218, -1000, 167, 12, 12, -1000, -1000, -1000,
	-1000, -25, 116, 538, 217, 70, 85, 217, 452, 214,
	229, 581, 52, -1000, -36, 91, -1000, -1000, -1000, 831,
	182, 211, 211, 624, 581, 195, 356, -1000, -1000, -1000,
	-1000, -1000, 121, 121, 865, 865, 865, 865, 865, 865,
	865, 831, 710, 753, 885, 33, 338, 148, 148, 667,
	495, 77, 356, -1000, 228, 210, -1000, 581, 144, 831,
	831, 152, 176, 831, -1000, 79, 831, -1000, -1000, 209,
	-1000, 146, 4, -1000, 182, 831, -1000, 117, -1000, 84,
	831, 831, -1000, -1000, -1000, -1000, -7, -37, 177, 164,
	356, -1000, 865, 831, 227, 208, -1000, -1000, 123, 12,
	12, -1000, -1000, -1000, 795, 831, 581, 581, -1000, 831,
	-1000, -1000, 581, 1, -1000, 4, 831, 44, 581, -1000,
	-1000, 581, -1000, 452, -1000, -38, -1000, -1000, 356, -1000,
	624, -1000, -1000, 77, 831, 175, 154, -1000, 581, 122,
	581, 204, -1000, -1000, 127, 624, 831, 251, -1000, -1000,
	-1000, 409, 831, 831, -1000, 831, 224, 1, -25, 624,
	-1000, 831, 581, 581, 89, -1000, -1000, -1000, 581, -1000,
}

var yyPgo = [...]int{
	0, 8, 11, 282, 280, 17, 279, 277, 4, 276,
	274, 0, 2, 83, 18, 67, 273, 91, 14, 272,
	12, 16, 7, 271, 9, 269, 267, 266, 265, 264,
	3, 15, 263, 13, 260, 258, 5, 10, 257, 1,
	251, 250, 247, 245, 242, 239,
}

var yyR1 = [...]int{
	0, 43, 37, 37, 44, 44, 38, 38, 38, 24,
	24, 24, 24, 25, 25, 41, 42, 42, 26, 26,
	26, 28, 28, 27, 27, 29, 29, 30, 32, 32,
	31, 31, 31, 31, 31, 31, 31, 31, 45, 45,
	14, 14, 14, 14, 14, 14, 14, 14, 14, 14,
	14, 14, 14, 14, 14, 4, 4, 3, 3, 2,
	2, 2, 2, 40, 40, 39, 39, 7, 7, 10,
	10, 6, 6, 9, 9, 5, 5, 5, 5, 5,
	8, 8, 8, 8, 8, 15, 15, 16, 16, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 17, 17,
	12, 12, 13, 13, 1, 1, 33, 35, 35, 34,
	34, 34, 18, 18, 36, 22, 23, 23, 23, 23,
	19, 20, 20, 21, 21,
}

var yyR2 = [...]int{
	0, 2, 5, 2, 0, 2, 0, 3, 2, 0,
	2, 2, 3, 1, 1, 5, 1, 3, 3, 6,
	1, 4, 5, 1, 4, 2, 1, 4, 0, 3,
	1, 2, 1, 3, 3, 1, 1, 1, 0, 1,
	1, 1, 1, 3, 7, 4, 4, 6, 8, 3,
	4, 4, 3, 4, 3, 0, 2, 1, 3, 1,
	3, 2, 2, 1, 3, 1, 3, 0, 2, 0,
	2, 1, 3, 1, 3, 1, 3, 2, 1, 2,
	1, 3, 5, 4, 4, 1, 3, 0, 1, 1,
	4, 2, 2, 2, 2, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 4, 3,
	3, 3, 3, 3, 3, 3, 3, 5, 1, 3,
	0, 1, 0, 2, 0, 1, 3, 1, 3, 0,
	1, 2, 1, 3, 1, 1, 3, 2, 2, 1,
	4, 1, 3, 1, 2,
}

var yyChk = [...]int{
	-1000, -43, -24, 28, -25, 59, 27, -30, -26, -31,
	-42, 30, -27, -15, 52, 53, 54, 55, -41, -28,
	-11, 51, 34, -14, 39, 48, 10, 8, 24, -22,
	-23, -36, 40, 17, 19, 5, 32, 33, 11, 50,
	59, -32, 13, -18, -14, 15, 25, 9, -15, 47,
	-29, 35, 36, 7, 4, 12, 44, 8, 10, 14,
	16, 29, 41, 42, 31, 37, 48, 49, 26, 21,
	22, 23, 45, 46, 38, 34, 32, -15, 11, 5,
	17, -7, -6, -5, -22, 7, 43, -11, -11, -11,
	-11, 5, -13, -11, -17, -33, -34, -17, -11, -35,
	-13, -11, 11, 33, -45, 62, -37, 59, -30, 37,
	9, -15, -15, -11, -11, -15, 13, 34, -11, -11,
	-11, -11, -11, -11, -11, -11, -11, -11, -11, -11,
	-11, 37, -11, -11, -11, -11, -11, -11, -11, -11,
	-11, 5, 13, 32, -4, -3, -2, -11, -22, 7,
	43, -15, -16, 13, -1, 9, 15, -22, -22, -36,
	18, -21, -20, -19, 30, 9, -1, -21, 20, -1,
	13, 9, 6, 33, 59, -31, -38, -44, -15, -14,
	13, -37, -11, 35, -10, -9, -8, -5, -22, 7,
	43, -37, 6, -1, 9, 15, -11, -11, 18, 13,
	-15, -5, -11, 9, 18, -20, 34, -18, -11, 20,
	20, -11, -33, -11, 56, 27, 59, 59, 13, -37,
	-11, 6, -1, 9, 13, -22, -22, -2, -11, -12,
	-11, -40, -39, -36, -22, -11, 37, -24, 59, -37,
	-8, -11, 13, 13, 18, 13, -1, 9, 15, -11,
	57, 15, -11, -11, -12, 6, -39, -36, -11, 18,
}

var yyDef = [...]int{
	9, -2, 0, 1, 10, 11, 0, 13, 14, 28,
	0, 0, 20, 30, 32, 35, 36, 37, 16, 23,
	85, 0, 0, 89, 67, 0, 0, 0, 0, 40,
	41, 42, 0, 122, 129, 122, 135, 139, 0, 134,
	12, 38, 0, 0, 132, 0, 0, 0, 31, 0,
	0, 0, 26, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 55,
	87, 0, 124, 71, 75, 78, 0, 91, 92, 93,
	94, 0, 0, 118, 124, 127, 0, 124, 118, 130,
	0, 118, 137, 138, 0, 39, 18, 6, 4, 0,
	0, 33, 34, 86, 17, 0, 0, 25, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107, 0, 109, 110, 111, 112, 113, 114, 115, 116,
	0, 69, 0, 43, 0, 124, 57, 59, 40, 0,
	0, 88, 0, 0, 68, 125, 0, 77, 79, 0,
	49, 0, 143, 141, 0, 125, 123, 0, 52, 0,
	0, 131, 54, 136, 27, 29, 0, 3, 0, 133,
	0, 24, 108, 0, 0, 124, 73, 80, 75, 78,
	0, 21, 45, 56, 125, 0, 61, 62, 46, 120,
	90, 72, 76, 0, 50, 144, 0, 0, 119, 51,
	53, 126, 128, 0, 9, 0, 8, 5, 0, 22,
	117, 15, 70, 125, 0, 77, 79, 58, 60, 0,
	121, 124, 63, 65, 0, 142, 0, 0, 7, 19,
	74, 81, 0, 0, 47, 120, 0, 125, 0, 140,
	2, 0, 83, 84, 0, 44, 64, 66, 82, 48,
}

var yyTok1 = [...]int{
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

var yyTok2 = [...]int{
	2, 3, 25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 39, 40, 41, 42,
	43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
	53, 54, 55, 56, 57, 58, 60, 61,
}

var yyTok3 = [...]int{
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
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
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
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
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
	yyn = yyPact[yystate]
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
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
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
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
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
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
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

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:216
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:223
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
//line build/parse.y:243
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:251
		{
			yyVAL.exprs = nil
			yyVAL.lastStmt = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:256
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
//line build/parse.y:268
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = nil
		}
	case 9:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:274
		{
			yyVAL.exprs = nil
			yyVAL.lastStmt = nil
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:279
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
//line build/parse.y:310
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:316
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
//line build/parse.y:330
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].exprs[len(yyDollar[1].exprs)-1]
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:335
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
//line build/parse.y:349
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
//line build/parse.y:364
		{
			yyDollar[1].def_header.Type = yyDollar[3].expr
			yyVAL.def_header = yyDollar[1].def_header
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:371
		{
			yyDollar[1].def_header.Function.Body = yyDollar[3].exprs
			yyDollar[1].def_header.ColonPos = yyDollar[2].pos
			yyVAL.expr = yyDollar[1].def_header
			yyVAL.lastStmt = yyDollar[3].lastStmt
		}
	case 19:
		yyDollar = yyS[yypt-6 : yypt+1]
//line build/parse.y:378
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
//line build/parse.y:388
		{
			yyVAL.expr = yyDollar[1].ifstmt
			yyVAL.lastStmt = yyDollar[1].lastStmt
		}
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:396
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
//line build/parse.y:405
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
//line build/parse.y:426
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
//line build/parse.y:443
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastStmt = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 28:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:449
		{
			yyVAL.exprs = []Expr{}
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:453
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:460
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:467
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:472
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:473
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 35:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:475
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 36:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:482
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 37:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:489
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 42:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:503
		{
			yyVAL.expr = yyDollar[1].string
		}
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:507
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 44:
		yyDollar = yyS[yypt-7 : yypt+1]
//line build/parse.y:516
		{
			load := &LoadStmt{
				Load:         yyDollar[1].pos,
				Module:       yyDollar[3].string,
				Rparen:       End{Pos: yyDollar[7].pos},
				ForceCompact: yyDollar[1].pos.Line == yyDollar[7].pos.Line,
			}
			for _, arg := range yyDollar[5].loadargs {
				load.From = append(load.From, &arg.from)
				load.To = append(load.To, &arg.to)
			}
			yyVAL.expr = load
		}
	case 45:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:530
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
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:541
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 47:
		yyDollar = yyS[yypt-6 : yypt+1]
//line build/parse.y:550
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
	case 48:
		yyDollar = yyS[yypt-8 : yypt+1]
//line build/parse.y:561
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
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:574
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 50:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:583
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
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:594
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
	case 52:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:605
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
	case 53:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:618
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[4].pos),
			}
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:627
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
	case 55:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:648
		{
			yyVAL.exprs = nil
		}
	case 56:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:652
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 57:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:658
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 58:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:662
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 60:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:669
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 61:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:673
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:677
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 63:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:682
		{
			yyVAL.loadargs = []*struct {
				from Ident
				to   Ident
			}{yyDollar[1].loadarg}
		}
	case 64:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:686
		{
			yyDollar[1].loadargs = append(yyDollar[1].loadargs, yyDollar[3].loadarg)
			yyVAL.loadargs = yyDollar[1].loadargs
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:692
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
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:709
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
	case 67:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:724
		{
			yyVAL.exprs = nil
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:728
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 69:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:733
		{
			yyVAL.exprs = nil
		}
	case 70:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:737
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:743
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:747
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 73:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:754
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 74:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:758
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 76:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:765
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:769
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 78:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:773
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, nil)
		}
	case 79:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:777
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:786
		{
			yyVAL.expr = typed(yyDollar[1].expr, yyDollar[3].expr)
		}
	case 82:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:790
		{
			yyVAL.expr = binary(typed(yyDollar[1].expr, yyDollar[3].expr), yyDollar[4].pos, yyDollar[4].tok, yyDollar[5].expr)
		}
	case 83:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:794
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, typed(yyDollar[2].expr, yyDollar[4].expr))
		}
	case 84:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:798
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, typed(yyDollar[2].expr, yyDollar[4].expr))
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:805
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
	case 87:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:820
		{
			yyVAL.expr = nil
		}
	case 90:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:828
		{
			yyVAL.expr = &LambdaExpr{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[2].exprs,
					Body:     []Expr{yyDollar[4].expr},
				},
			}
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:837
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:838
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:839
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 94:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:840
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:841
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:842
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 97:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:843
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 98:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:844
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:845
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:846
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:847
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:848
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:849
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:850
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 105:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:851
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:852
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 107:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:853
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 108:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:854
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 109:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:855
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:856
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:857
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 112:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:858
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 113:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:859
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 114:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:860
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 115:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:861
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 116:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:863
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 117:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:871
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:883
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:887
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 120:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:892
		{
			yyVAL.expr = nil
		}
	case 122:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:898
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 123:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:902
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 124:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:911
		{
			yyVAL.pos = Position{}
		}
	case 126:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:917
		{
			yyVAL.kv = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 127:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:927
		{
			yyVAL.kvs = []*KeyValueExpr{yyDollar[1].kv}
		}
	case 128:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:931
		{
			yyVAL.kvs = append(yyDollar[1].kvs, yyDollar[3].kv)
		}
	case 129:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:936
		{
			yyVAL.kvs = nil
		}
	case 130:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:940
		{
			yyVAL.kvs = yyDollar[1].kvs
		}
	case 131:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:944
		{
			yyVAL.kvs = yyDollar[1].kvs
		}
	case 133:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:951
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
	case 134:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:967
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:979
		{
			yyVAL.expr = &Ident{NamePos: yyDollar[1].pos, Name: yyDollar[1].tok}
		}
	case 136:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:985
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok + "." + yyDollar[3].tok}
		}
	case 137:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:989
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok + "."}
		}
	case 138:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:993
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: "." + yyDollar[2].tok}
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:997
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 140:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:1003
		{
			yyVAL.expr = &ForClause{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				In:   yyDollar[3].pos,
				X:    yyDollar[4].expr,
			}
		}
	case 141:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1014
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 142:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:1018
		{
			yyVAL.exprs = append(yyDollar[1].exprs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:1027
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 144:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:1031
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
		}
	}
	goto yystack /* stack new state and value */
}
