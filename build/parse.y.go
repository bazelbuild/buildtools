//line build/parse.y:13
package build

import __yyfmt__ "fmt"

//line build/parse.y:13
//line build/parse.y:18
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
	forc    *ForClause
	ifs     []*IfClause
	forifs  *ForClauseWithIfClausesOpt
	forsifs []*ForClauseWithIfClausesOpt
	string  *StringExpr
	strings []*StringExpr
	block   CodeBlock

	// supporting information
	comma    Position // position of trailing comma in list, if present
	lastRule Expr     // most recent rule, to attach line comments to
}

const _AUGM = 57346
const _AND = 57347
const _COMMENT = 57348
const _EOF = 57349
const _EQ = 57350
const _FOR = 57351
const _GE = 57352
const _IDENT = 57353
const _NUMBER = 57354
const _IF = 57355
const _ELSE = 57356
const _ELIF = 57357
const _IN = 57358
const _IS = 57359
const _LAMBDA = 57360
const _LOAD = 57361
const _LE = 57362
const _NE = 57363
const _NOT = 57364
const _OR = 57365
const _PYTHON = 57366
const _STRING = 57367
const _DEF = 57368
const _RETURN = 57369
const _INDENT = 57370
const _UNINDENT = 57371
const ShiftInstead = 57372
const _ASSERT = 57373
const _UNARY = 57374

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
	"_AUGM",
	"_AND",
	"_COMMENT",
	"_EOF",
	"_EQ",
	"_FOR",
	"_GE",
	"_IDENT",
	"_NUMBER",
	"_IF",
	"_ELSE",
	"_ELIF",
	"_IN",
	"_IS",
	"_LAMBDA",
	"_LOAD",
	"_LE",
	"_NE",
	"_NOT",
	"_OR",
	"_PYTHON",
	"_STRING",
	"_DEF",
	"_RETURN",
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

//line build/parse.y:747

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
	return &BinaryExpr{
		X:         x,
		OpStart:   pos,
		Op:        op,
		LineBreak: xend.Line < ystart.Line,
		Y:         y,
	}
}

// isSimpleExpression returns whether an expression is simple and allowed to exist in
// compact forms of sequences.
// The formal criteria are the following: an expression is considered simple if it's
// a literal (variable, string or a number), a literal with a unary operator or an empty sequence.
func isSimpleExpression(expr *Expr) bool {
	switch x := (*expr).(type) {
	case *Ident, *LiteralExpr, *StringExpr:
		return true
	case *UnaryExpr:
		return isSimpleExpression(&x.X)
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

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 814

var yyAct = [...]int{

	13, 114, 138, 2, 140, 76, 17, 7, 120, 71,
	36, 9, 119, 83, 132, 60, 33, 61, 37, 66,
	67, 68, 163, 32, 104, 168, 72, 74, 79, 39,
	40, 110, 86, 35, 156, 81, 75, 78, 86, 87,
	122, 165, 90, 91, 92, 93, 94, 95, 96, 97,
	98, 99, 100, 101, 102, 103, 169, 105, 106, 107,
	108, 122, 85, 88, 112, 113, 151, 122, 131, 129,
	66, 155, 31, 111, 122, 128, 26, 174, 20, 117,
	89, 28, 173, 73, 180, 66, 135, 125, 25, 127,
	27, 122, 63, 136, 134, 133, 118, 70, 62, 29,
	30, 170, 65, 147, 64, 141, 18, 23, 123, 160,
	19, 150, 143, 32, 144, 42, 148, 149, 41, 44,
	126, 45, 116, 43, 149, 145, 115, 42, 152, 37,
	41, 157, 159, 154, 152, 43, 152, 158, 84, 69,
	162, 82, 38, 164, 1, 24, 80, 77, 167, 166,
	34, 12, 8, 4, 152, 153, 22, 21, 121, 124,
	26, 0, 20, 0, 171, 28, 0, 172, 0, 175,
	176, 0, 25, 177, 27, 167, 179, 7, 6, 0,
	0, 11, 0, 29, 30, 16, 0, 0, 0, 0,
	18, 23, 0, 0, 19, 0, 15, 32, 10, 14,
	26, 178, 20, 5, 0, 28, 0, 0, 0, 0,
	0, 0, 25, 0, 27, 0, 0, 0, 6, 3,
	0, 11, 0, 29, 30, 16, 0, 0, 0, 0,
	18, 23, 0, 0, 19, 0, 15, 32, 10, 14,
	26, 0, 20, 5, 0, 28, 0, 0, 0, 0,
	0, 0, 25, 0, 27, 0, 0, 0, 0, 0,
	0, 0, 0, 29, 30, 0, 0, 0, 0, 0,
	18, 23, 0, 0, 19, 0, 15, 32, 42, 14,
	0, 41, 44, 139, 45, 0, 43, 130, 46, 52,
	47, 0, 0, 0, 0, 53, 57, 0, 0, 48,
	0, 51, 0, 0, 59, 0, 0, 54, 58, 0,
	0, 49, 50, 55, 56, 42, 0, 0, 41, 44,
	0, 45, 0, 43, 161, 46, 52, 47, 0, 0,
	0, 0, 53, 57, 0, 0, 48, 0, 51, 0,
	0, 59, 0, 0, 54, 58, 0, 0, 49, 50,
	55, 56, 42, 0, 0, 41, 44, 0, 45, 0,
	43, 0, 46, 52, 47, 0, 146, 0, 0, 53,
	57, 0, 0, 48, 0, 51, 0, 0, 59, 0,
	0, 54, 58, 0, 0, 49, 50, 55, 56, 42,
	0, 0, 41, 44, 0, 45, 0, 43, 0, 46,
	52, 47, 0, 0, 0, 0, 53, 57, 0, 0,
	48, 122, 51, 0, 0, 59, 0, 0, 54, 58,
	0, 0, 49, 50, 55, 56, 42, 0, 0, 41,
	44, 0, 45, 0, 43, 0, 46, 52, 47, 0,
	0, 0, 0, 53, 57, 0, 0, 48, 0, 51,
	0, 0, 59, 142, 0, 54, 58, 0, 0, 49,
	50, 55, 56, 42, 0, 0, 41, 44, 0, 45,
	0, 43, 137, 46, 52, 47, 0, 0, 0, 0,
	53, 57, 0, 0, 48, 0, 51, 0, 0, 59,
	0, 0, 54, 58, 0, 0, 49, 50, 55, 56,
	42, 0, 0, 41, 44, 0, 45, 0, 43, 109,
	46, 52, 47, 0, 0, 0, 0, 53, 57, 0,
	0, 48, 0, 51, 0, 0, 59, 0, 0, 54,
	58, 0, 0, 49, 50, 55, 56, 42, 0, 0,
	41, 44, 0, 45, 0, 43, 0, 46, 52, 47,
	0, 0, 0, 0, 53, 57, 0, 0, 48, 0,
	51, 0, 0, 59, 0, 0, 54, 58, 0, 0,
	49, 50, 55, 56, 42, 0, 0, 41, 44, 0,
	45, 0, 43, 0, 46, 52, 47, 0, 0, 0,
	0, 53, 57, 0, 0, 48, 0, 51, 0, 0,
	0, 0, 0, 54, 58, 0, 0, 49, 50, 55,
	56, 42, 0, 0, 41, 44, 0, 45, 0, 43,
	0, 46, 0, 47, 0, 0, 0, 0, 0, 57,
	0, 0, 48, 0, 51, 0, 0, 59, 0, 0,
	54, 58, 0, 0, 49, 50, 55, 56, 26, 0,
	20, 0, 0, 28, 0, 0, 0, 0, 0, 0,
	25, 0, 27, 0, 0, 0, 0, 0, 0, 0,
	0, 29, 30, 0, 0, 0, 0, 0, 18, 23,
	0, 0, 19, 0, 15, 32, 42, 14, 0, 41,
	44, 0, 45, 0, 43, 0, 46, 0, 47, 0,
	0, 0, 0, 0, 57, 0, 0, 48, 0, 51,
	0, 0, 0, 0, 0, 54, 58, 0, 0, 49,
	50, 55, 56, 42, 0, 0, 41, 44, 0, 45,
	0, 43, 0, 46, 0, 47, 0, 26, 0, 0,
	0, 57, 28, 0, 48, 0, 51, 0, 0, 25,
	0, 27, 0, 0, 0, 0, 49, 50, 0, 56,
	29, 30, 0, 42, 0, 0, 41, 44, 23, 45,
	0, 43, 0, 46, 32, 47, 0, 0, 0, 42,
	0, 57, 41, 44, 48, 45, 51, 43, 0, 46,
	0, 47, 0, 0, 0, 0, 49, 50, 0, 0,
	48, 0, 51, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 49, 50,
}
var yyPact = [...]int{

	-1000, -1000, 195, -1000, -1000, -1000, -32, -1000, -1000, -1000,
	5, 732, -2, 533, 71, -1000, 71, 87, 71, 71,
	71, -1000, -1000, 134, -19, 71, 71, 71, 732, -1000,
	-1000, -1000, -1000, -1000, -38, 133, 29, 87, 71, 50,
	-1000, 71, 71, 71, 71, 71, 71, 71, 71, 71,
	71, 71, 71, 71, 71, -9, 71, 71, 71, 71,
	533, 496, 3, 71, 71, 113, 533, -1000, -1000, 71,
	-1000, 78, 385, 99, 385, 114, 41, 55, 49, 274,
	59, -1000, -34, 643, 71, 71, 732, 459, 235, -1000,
	-1000, -1000, -1000, 123, 123, 111, 111, 111, 111, 111,
	111, 607, 607, 719, 71, 759, 775, 719, 422, 235,
	-1000, 108, 385, 348, 90, 71, 71, 105, -1000, 48,
	-1000, -1000, 732, 71, -1000, 65, -1000, 14, -1000, -1000,
	71, 71, -1000, -1000, 103, 311, 87, 235, -1000, -23,
	-1000, 719, 71, -1000, -1000, 35, -1000, 71, 682, 533,
	-1000, -1000, -1000, -5, 23, -1000, -1000, 533, -1000, 274,
	88, 235, -1000, -1000, 682, -1000, 64, 533, 71, 71,
	235, -1000, 155, -1000, 71, 570, 570, -1000, -1000, 66,
	-1000,
}
var yyPgo = [...]int{

	0, 159, 0, 1, 6, 83, 9, 10, 158, 8,
	12, 157, 156, 155, 3, 153, 152, 151, 4, 11,
	150, 5, 147, 146, 72, 145, 2, 144, 142, 141,
}
var yyR1 = [...]int{

	0, 27, 26, 26, 14, 14, 14, 14, 15, 15,
	16, 16, 16, 17, 17, 17, 28, 28, 18, 20,
	20, 19, 19, 19, 19, 29, 29, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 3, 3,
	1, 1, 21, 23, 23, 22, 22, 5, 5, 6,
	6, 7, 7, 24, 25, 25, 11, 12, 8, 9,
	10, 10, 13, 13,
}
var yyR2 = [...]int{

	0, 2, 4, 1, 0, 2, 2, 3, 1, 1,
	7, 6, 1, 4, 5, 4, 2, 1, 4, 0,
	3, 1, 2, 1, 1, 0, 1, 1, 1, 3,
	4, 4, 4, 6, 8, 5, 1, 3, 4, 4,
	4, 3, 3, 3, 2, 1, 4, 2, 2, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 4, 3, 3, 3, 5, 0, 1,
	0, 1, 3, 1, 3, 1, 2, 1, 3, 0,
	2, 1, 3, 1, 1, 2, 1, 1, 4, 2,
	1, 2, 0, 3,
}
var yyChk = [...]int{

	-1000, -27, -14, 24, -15, 48, 23, -18, -16, -19,
	43, 26, -17, -2, 44, 41, 30, -4, 35, 39,
	7, -11, -12, 36, -25, 17, 5, 19, 10, 28,
	29, -24, 42, 48, -20, 28, -7, -4, -28, 31,
	32, 7, 4, 12, 8, 10, 14, 16, 25, 37,
	38, 27, 15, 21, 33, 39, 40, 22, 34, 30,
	-2, -2, 11, 5, 17, -5, -2, -2, -2, 5,
	-24, -6, -2, -5, -2, -6, -21, -22, -6, -2,
	-23, -4, -29, 51, 5, 33, 9, -2, 13, 30,
	-2, -2, -2, -2, -2, -2, -2, -2, -2, -2,
	-2, -2, -2, -2, 33, -2, -2, -2, -2, 13,
	28, -6, -2, -2, -3, 13, 9, -6, 18, -10,
	-9, -8, 26, 9, -1, -10, 6, -10, 20, 20,
	13, 9, 48, -19, -6, -2, -4, 13, -26, 48,
	-18, -2, 31, -26, 6, -10, 18, 13, -2, -2,
	6, 18, -9, -13, -7, 6, 20, -2, -21, -2,
	6, 13, -26, 45, -2, 6, -3, -2, 30, 33,
	13, -26, -14, 18, 13, -2, -2, -26, 46, -3,
	18,
}
var yyDef = [...]int{

	4, -2, 0, 1, 5, 6, 0, 8, 9, 19,
	0, 0, 12, 21, 23, 24, 0, 45, 0, 0,
	0, 27, 28, 0, 36, 79, 79, 79, 0, 86,
	87, 84, 83, 7, 25, 0, 0, 81, 0, 0,
	17, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	22, 0, 0, 79, 68, 0, 77, 47, 48, 79,
	85, 0, 77, 70, 77, 0, 73, 0, 0, 77,
	75, 44, 0, 26, 79, 0, 0, 0, 0, 16,
	49, 50, 51, 52, 53, 54, 55, 56, 57, 58,
	59, 60, 61, 62, 0, 64, 65, 66, 0, 0,
	29, 0, 77, 69, 0, 0, 0, 0, 37, 0,
	90, 92, 0, 71, 80, 0, 43, 0, 41, 42,
	0, 76, 18, 20, 0, 0, 82, 0, 15, 0,
	3, 63, 0, 13, 31, 0, 32, 68, 46, 78,
	30, 38, 91, 89, 0, 39, 40, 72, 74, 0,
	0, 0, 14, 4, 67, 35, 0, 69, 0, 0,
	0, 11, 0, 33, 68, 93, 88, 10, 2, 0,
	34,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	48, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 3, 3,
	5, 6, 7, 8, 9, 10, 11, 12, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 51,
	14, 15, 16, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 17, 3, 18, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 19, 3, 20,
}
var yyTok2 = [...]int{

	2, 3, 21, 22, 23, 24, 25, 26, 27, 28,
	29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
	39, 40, 41, 42, 43, 44, 45, 46, 47, 49,
	50,
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
		//line build/parse.y:169
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:176
		{
			yyVAL.block = CodeBlock{
				Start:      yyDollar[2].pos,
				Statements: yyDollar[3].exprs,
				End:        End{Pos: yyDollar[4].pos},
			}
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:184
		{
			// simple_stmt is never empty
			start, _ := yyDollar[1].exprs[0].Span()
			_, end := yyDollar[1].exprs[len(yyDollar[1].exprs)-1].Span()
			yyVAL.block = CodeBlock{
				Start:      start,
				Statements: yyDollar[1].exprs,
				End:        End{Pos: end},
			}
		}
	case 4:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:196
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:201
		{
			// If this statement follows a comment block,
			// attach the comments to the statement.
			if cb, ok := yyDollar[1].lastRule.(*CommentBlock); ok {
				yyVAL.exprs = append(yyDollar[1].exprs[:len(yyDollar[1].exprs)-1], yyDollar[2].exprs...)
				yyDollar[2].exprs[0].Comment().Before = cb.After
				yyVAL.lastRule = yyDollar[2].exprs[len(yyDollar[2].exprs)-1]
				break
			}

			// Otherwise add to list.
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
			yyVAL.lastRule = yyDollar[2].exprs[len(yyDollar[2].exprs)-1]

			// Consider this input:
			//
			//	foo()
			//	# bar
			//	baz()
			//
			// If we've just parsed baz(), the # bar is attached to
			// foo() as an After comment. Make it a Before comment
			// for baz() instead.
			if x := yyDollar[1].lastRule; x != nil {
				com := x.Comment()
				// stmt is never empty
				yyDollar[2].exprs[0].Comment().Before = com.After
				com.After = nil
			}
		}
	case 6:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:232
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:238
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = yyDollar[1].lastRule
			if yyVAL.lastRule == nil {
				cb := &CommentBlock{Start: yyDollar[2].pos}
				yyVAL.exprs = append(yyVAL.exprs, cb)
				yyVAL.lastRule = cb
			}
			com := yyVAL.lastRule.Comment()
			com.After = append(com.After, Comment{Start: yyDollar[2].pos, Token: yyDollar[2].tok})
		}
	case 8:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:252
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:256
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 10:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:262
		{
			yyVAL.expr = &FuncDef{
				Start:          yyDollar[1].pos,
				Name:           yyDollar[2].tok,
				ListStart:      yyDollar[3].pos,
				Args:           yyDollar[4].exprs,
				Body:           yyDollar[7].block,
				End:            yyDollar[7].block.End,
				ForceCompact:   forceCompact(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
				ForceMultiLine: forceMultiLine(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
			}
		}
	case 11:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:275
		{
			yyVAL.expr = &ForLoop{
				Start:    yyDollar[1].pos,
				LoopVars: yyDollar[2].exprs,
				Iterable: yyDollar[4].expr,
				Body:     yyDollar[6].block,
				End:      yyDollar[6].block.End,
			}
		}
	case 12:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:285
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 13:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:291
		{
			yyVAL.expr = &IfElse{
				Start: yyDollar[1].pos,
				Conditions: []Condition{
					Condition{
						If:   yyDollar[2].expr,
						Then: yyDollar[4].block,
					},
				},
				End: yyDollar[4].block.End,
			}
		}
	case 14:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:304
		{
			block := yyDollar[1].expr.(*IfElse)
			block.Conditions = append(block.Conditions, Condition{
				If:   yyDollar[3].expr,
				Then: yyDollar[5].block,
			})
			block.End = yyDollar[5].block.End
			yyVAL.expr = block
		}
	case 15:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:314
		{
			block := yyDollar[1].expr.(*IfElse)
			block.Conditions = append(block.Conditions, Condition{
				Then: yyDollar[4].block,
			})
			block.End = yyDollar[4].block.End
			yyVAL.expr = block
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:329
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 19:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:335
		{
			yyVAL.exprs = []Expr{}
		}
	case 20:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:339
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 22:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:346
		{
			_, end := yyDollar[2].expr.Span()
			yyVAL.expr = &ReturnExpr{
				X:   yyDollar[2].expr,
				End: end,
			}
		}
	case 23:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:354
		{
			yyVAL.expr = &ReturnExpr{End: yyDollar[1].pos}
		}
	case 24:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:358
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:369
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 30:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:378
		{
			yyVAL.expr = &CallExpr{
				X:              &Ident{NamePos: yyDollar[1].pos, Name: "load"},
				ListStart:      yyDollar[2].pos,
				List:           yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceCompact:   forceCompact(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
				ForceMultiLine: forceMultiLine(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 31:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:389
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
	case 32:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:400
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 33:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:409
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
	case 34:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:420
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
	case 35:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:433
		{
			yyVAL.expr = &CallExpr{
				X:         yyDollar[1].expr,
				ListStart: yyDollar[2].pos,
				List: []Expr{
					&ListForExpr{
						Brack: "",
						Start: yyDollar[2].pos,
						X:     yyDollar[3].expr,
						For:   yyDollar[4].forsifs,
						End:   End{Pos: yyDollar[5].pos},
					},
				},
				End: End{Pos: yyDollar[5].pos},
			}
		}
	case 36:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:450
		{
			if len(yyDollar[1].strings) == 1 {
				yyVAL.expr = yyDollar[1].strings[0]
				break
			}
			yyVAL.expr = yyDollar[1].strings[0]
			for _, x := range yyDollar[1].strings[1:] {
				_, end := yyVAL.expr.Span()
				yyVAL.expr = binary(yyVAL.expr, end, "+", x)
			}
		}
	case 37:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:462
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 38:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:472
		{
			exprStart, _ := yyDollar[2].expr.Span()
			yyVAL.expr = &ListForExpr{
				Brack:          "[]",
				Start:          yyDollar[1].pos,
				X:              yyDollar[2].expr,
				For:            yyDollar[3].forsifs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: yyDollar[1].pos.Line != exprStart.Line,
			}
		}
	case 39:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:484
		{
			exprStart, _ := yyDollar[2].expr.Span()
			yyVAL.expr = &ListForExpr{
				Brack:          "()",
				Start:          yyDollar[1].pos,
				X:              yyDollar[2].expr,
				For:            yyDollar[3].forsifs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: yyDollar[1].pos.Line != exprStart.Line,
			}
		}
	case 40:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:496
		{
			exprStart, _ := yyDollar[2].expr.Span()
			yyVAL.expr = &ListForExpr{
				Brack:          "{}",
				Start:          yyDollar[1].pos,
				X:              yyDollar[2].expr,
				For:            yyDollar[3].forsifs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: yyDollar[1].pos.Line != exprStart.Line,
			}
		}
	case 41:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:508
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 42:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:518
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:528
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
					Comma:          yyDollar[2].comma,
					End:            End{Pos: yyDollar[3].pos},
					ForceCompact:   forceCompact(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
					ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
				}
			}
		}
	case 44:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:548
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:553
		{
			yyVAL.expr = &LambdaExpr{
				Lambda: yyDollar[1].pos,
				Var:    yyDollar[2].exprs,
				Colon:  yyDollar[3].pos,
				Expr:   yyDollar[4].expr,
			}
		}
	case 47:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:561
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 48:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:562
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:563
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 50:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:564
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 51:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:565
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 52:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:566
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 53:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:567
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:568
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:569
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 56:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:570
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 57:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:571
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 58:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:572
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:573
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 60:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:574
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:575
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 62:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:576
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 63:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:577
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 64:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:578
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 65:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:579
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:581
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 67:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:589
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 68:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:600
		{
			yyVAL.expr = nil
		}
	case 70:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:610
		{
			yyVAL.pos = Position{}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:616
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 73:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:626
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 74:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:630
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:636
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 76:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:640
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:646
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:650
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 79:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:655
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 80:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:659
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 81:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:665
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:669
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 83:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:675
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 84:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:687
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 85:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:691
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:697
		{
			yyVAL.expr = &Ident{NamePos: yyDollar[1].pos, Name: yyDollar[1].tok}
		}
	case 87:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:703
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 88:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:709
		{
			yyVAL.forc = &ForClause{
				For:  yyDollar[1].pos,
				Var:  yyDollar[2].exprs,
				In:   yyDollar[3].pos,
				Expr: yyDollar[4].expr,
			}
		}
	case 89:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:719
		{
			yyVAL.forifs = &ForClauseWithIfClausesOpt{
				For: yyDollar[1].forc,
				Ifs: yyDollar[2].ifs,
			}
		}
	case 90:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:728
		{
			yyVAL.forsifs = []*ForClauseWithIfClausesOpt{yyDollar[1].forifs}
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:731
		{
			yyVAL.forsifs = append(yyDollar[1].forsifs, yyDollar[2].forifs)
		}
	case 92:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:736
		{
			yyVAL.ifs = nil
		}
	case 93:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:740
		{
			yyVAL.ifs = append(yyDollar[1].ifs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	}
	goto yystack /* stack new state and value */
}
