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
	ifstmt  *IfStmt

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
const _IF = 57354
const _ELSE = 57355
const _ELIF = 57356
const _IN = 57357
const _IS = 57358
const _LAMBDA = 57359
const _LOAD = 57360
const _LE = 57361
const _NE = 57362
const _NOT = 57363
const _OR = 57364
const _PYTHON = 57365
const _STRING = 57366
const _DEF = 57367
const _RETURN = 57368
const _INDENT = 57369
const _UNINDENT = 57370
const ShiftInstead = 57371
const _ASSERT = 57372
const _UNARY = 57373

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

//line build/parse.y:756

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
	case *LiteralExpr, *StringExpr:
		return true
	case *UnaryExpr:
		_, ok := x.X.(*LiteralExpr)
		return ok
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

const yyLast = 761

var yyAct = [...]int{

	13, 114, 141, 2, 143, 17, 77, 7, 57, 121,
	35, 9, 84, 120, 133, 56, 32, 36, 165, 66,
	67, 68, 69, 31, 60, 61, 73, 75, 80, 87,
	102, 169, 110, 34, 82, 72, 87, 123, 88, 89,
	90, 91, 92, 93, 94, 95, 96, 97, 98, 99,
	100, 101, 170, 103, 104, 105, 106, 108, 130, 86,
	107, 166, 76, 79, 112, 113, 157, 152, 156, 30,
	129, 66, 123, 109, 176, 123, 74, 182, 26, 175,
	20, 123, 119, 28, 171, 147, 66, 136, 123, 126,
	25, 128, 27, 137, 71, 65, 134, 132, 63, 111,
	124, 29, 85, 138, 62, 116, 118, 18, 23, 115,
	64, 19, 161, 151, 31, 144, 148, 149, 127, 38,
	150, 135, 37, 26, 70, 149, 145, 39, 28, 36,
	153, 83, 158, 160, 155, 25, 153, 27, 153, 159,
	163, 1, 24, 164, 81, 78, 29, 33, 168, 167,
	59, 58, 16, 23, 38, 153, 12, 37, 40, 31,
	41, 8, 39, 4, 154, 172, 22, 122, 125, 174,
	177, 178, 0, 173, 179, 0, 0, 168, 181, 7,
	26, 0, 20, 0, 0, 28, 0, 0, 0, 0,
	0, 0, 25, 0, 27, 0, 0, 0, 6, 0,
	0, 11, 0, 29, 21, 0, 0, 0, 0, 18,
	23, 0, 0, 19, 0, 15, 31, 10, 14, 26,
	180, 20, 5, 0, 28, 0, 0, 0, 0, 0,
	0, 25, 0, 27, 0, 0, 0, 6, 3, 26,
	11, 20, 29, 21, 28, 0, 0, 0, 18, 23,
	0, 25, 19, 27, 15, 31, 10, 14, 0, 0,
	0, 5, 29, 0, 0, 0, 0, 0, 18, 23,
	0, 0, 19, 0, 15, 31, 38, 14, 0, 37,
	40, 142, 41, 0, 39, 131, 42, 48, 43, 0,
	0, 0, 0, 49, 53, 0, 0, 44, 0, 47,
	0, 55, 0, 0, 50, 54, 0, 0, 45, 46,
	51, 52, 38, 0, 0, 37, 40, 0, 41, 0,
	39, 162, 42, 48, 43, 0, 0, 0, 0, 49,
	53, 0, 0, 44, 0, 47, 0, 55, 0, 0,
	50, 54, 0, 0, 45, 46, 51, 52, 38, 0,
	0, 37, 40, 0, 41, 0, 39, 0, 42, 48,
	43, 0, 146, 0, 0, 49, 53, 0, 0, 44,
	0, 47, 0, 55, 0, 0, 50, 54, 0, 0,
	45, 46, 51, 52, 38, 0, 0, 37, 40, 0,
	41, 0, 39, 0, 42, 48, 43, 0, 0, 0,
	0, 49, 53, 0, 0, 44, 123, 47, 0, 55,
	0, 0, 50, 54, 0, 0, 45, 46, 51, 52,
	38, 0, 0, 37, 40, 0, 41, 0, 39, 140,
	42, 48, 43, 0, 0, 0, 0, 49, 53, 0,
	0, 44, 0, 47, 0, 55, 0, 0, 50, 54,
	0, 0, 45, 46, 51, 52, 38, 0, 0, 37,
	40, 0, 41, 0, 39, 0, 42, 48, 43, 0,
	0, 0, 0, 49, 53, 0, 0, 44, 0, 47,
	0, 55, 139, 0, 50, 54, 0, 0, 45, 46,
	51, 52, 38, 0, 0, 37, 40, 0, 41, 0,
	39, 117, 42, 48, 43, 0, 0, 0, 0, 49,
	53, 0, 0, 44, 0, 47, 0, 55, 0, 0,
	50, 54, 0, 0, 45, 46, 51, 52, 38, 0,
	0, 37, 40, 0, 41, 0, 39, 0, 42, 48,
	43, 0, 0, 0, 0, 49, 53, 0, 0, 44,
	0, 47, 0, 55, 0, 0, 50, 54, 0, 0,
	45, 46, 51, 52, 38, 0, 0, 37, 40, 0,
	41, 0, 39, 0, 42, 48, 43, 0, 0, 0,
	0, 49, 53, 0, 0, 44, 0, 47, 0, 0,
	0, 0, 50, 54, 0, 0, 45, 46, 51, 52,
	38, 0, 0, 37, 40, 0, 41, 0, 39, 0,
	42, 0, 43, 0, 0, 0, 0, 0, 53, 0,
	0, 44, 0, 47, 0, 55, 0, 0, 50, 54,
	0, 0, 45, 46, 51, 52, 38, 0, 0, 37,
	40, 0, 41, 0, 39, 0, 42, 0, 43, 0,
	0, 0, 0, 0, 53, 0, 0, 44, 0, 47,
	0, 26, 0, 20, 50, 54, 28, 0, 45, 46,
	51, 52, 0, 25, 0, 27, 0, 38, 0, 0,
	37, 40, 0, 41, 29, 39, 0, 42, 0, 43,
	18, 23, 0, 0, 19, 53, 15, 31, 44, 14,
	47, 0, 0, 0, 0, 0, 0, 0, 0, 45,
	46, 38, 52, 0, 37, 40, 0, 41, 0, 39,
	0, 42, 0, 43, 0, 0, 0, 38, 0, 53,
	37, 40, 44, 41, 47, 39, 0, 42, 0, 43,
	0, 0, 0, 45, 46, 0, 0, 0, 44, 0,
	47, 0, 0, 0, 0, 0, 0, 0, 0, 45,
	46,
}
var yyPact = [...]int{

	-1000, -1000, 214, -1000, -1000, -1000, -31, -1000, -1000, -1000,
	5, 118, -1000, 524, 73, -1000, -6, 93, 73, 73,
	73, 73, -1000, 119, -18, 73, 73, 73, 118, -1000,
	-1000, -1000, -1000, -38, 97, 27, 93, 73, 73, 73,
	73, 73, 73, 73, 73, 73, 73, 73, 73, 73,
	73, -2, 73, 73, 73, 73, 524, -1000, -1000, 73,
	44, -1000, 4, 73, 73, 96, 524, -1000, -1000, 488,
	73, -1000, 64, 380, 91, 380, 112, 11, 50, 38,
	272, 88, -1000, -33, 656, 73, 73, 118, -1000, -1000,
	-1000, 115, 115, 150, 150, 150, 150, 150, 150, 596,
	596, 673, 73, 707, 723, 673, 452, 416, 234, -1000,
	-1000, 109, 380, 344, 72, 73, 73, 234, 107, -1000,
	49, -1000, -1000, 118, 73, -1000, 62, -1000, 46, -1000,
	-1000, 73, 73, -1000, -1000, 106, 308, 93, 673, 73,
	234, -1000, -26, -1000, -1000, 55, -1000, 73, 632, 524,
	-1000, -1000, -1000, -1000, 2, 20, -1000, -1000, 524, -1000,
	272, 71, 234, 632, -6, -1000, -1000, 61, 524, 73,
	73, 234, -1000, -1000, 175, -1000, 73, 560, 560, -1000,
	-1000, 59, -1000,
}
var yyPgo = [...]int{

	0, 168, 0, 1, 5, 76, 35, 10, 167, 9,
	13, 166, 164, 3, 163, 161, 156, 152, 151, 8,
	150, 4, 11, 147, 6, 145, 144, 69, 142, 2,
	141, 131,
}
var yyR1 = [...]int{

	0, 30, 29, 29, 13, 13, 13, 13, 14, 14,
	15, 15, 15, 18, 19, 19, 17, 16, 16, 20,
	20, 21, 23, 23, 22, 22, 22, 22, 31, 31,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	3, 3, 1, 1, 24, 26, 26, 25, 25, 5,
	5, 6, 6, 7, 7, 27, 28, 28, 11, 8,
	9, 10, 10, 12, 12,
}
var yyR2 = [...]int{

	0, 2, 4, 1, 0, 2, 2, 3, 1, 1,
	7, 6, 1, 3, 1, 5, 4, 1, 2, 2,
	1, 4, 0, 3, 1, 2, 1, 1, 0, 1,
	1, 3, 4, 4, 4, 6, 8, 5, 1, 3,
	4, 4, 4, 3, 3, 3, 2, 1, 4, 2,
	2, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 4, 3, 3, 3, 5,
	0, 1, 0, 1, 3, 1, 3, 1, 2, 1,
	3, 0, 2, 1, 3, 1, 1, 2, 1, 4,
	2, 1, 2, 0, 3,
}
var yyChk = [...]int{

	-1000, -30, -13, 24, -14, 47, 23, -21, -15, -22,
	42, 26, -16, -2, 43, 40, -17, -4, 34, 38,
	7, 29, -11, 35, -28, 17, 5, 19, 10, 28,
	-27, 41, 47, -23, 28, -7, -4, 7, 4, 12,
	8, 10, 14, 16, 25, 36, 37, 27, 15, 21,
	32, 38, 39, 22, 33, 29, -2, -19, -18, -20,
	30, 31, 11, 5, 17, -5, -2, -2, -2, -2,
	5, -27, -6, -2, -5, -2, -6, -24, -25, -6,
	-2, -26, -4, -31, 50, 5, 32, 9, -2, -2,
	-2, -2, -2, -2, -2, -2, -2, -2, -2, -2,
	-2, -2, 32, -2, -2, -2, -2, -2, 13, 29,
	28, -6, -2, -2, -3, 13, 9, 13, -6, 18,
	-10, -9, -8, 26, 9, -1, -10, 6, -10, 20,
	20, 13, 9, 47, -22, -6, -2, -4, -2, 30,
	13, -29, 47, -21, 6, -10, 18, 13, -2, -2,
	-29, 6, 18, -9, -12, -7, 6, 20, -2, -24,
	-2, 6, 13, -2, -29, 44, 6, -3, -2, 29,
	32, 13, -29, -19, -13, 18, 13, -2, -2, -29,
	45, -3, 18,
}
var yyDef = [...]int{

	4, -2, 0, 1, 5, 6, 0, 8, 9, 22,
	0, 0, 12, 24, 26, 27, 17, 47, 0, 0,
	0, 0, 30, 0, 38, 81, 81, 81, 0, 88,
	86, 85, 7, 28, 0, 0, 83, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 25, 18, 14, 0,
	0, 20, 0, 81, 70, 0, 79, 49, 50, 0,
	81, 87, 0, 79, 72, 79, 0, 75, 0, 0,
	79, 77, 46, 0, 29, 81, 0, 0, 51, 52,
	53, 54, 55, 56, 57, 58, 59, 60, 61, 62,
	63, 64, 0, 66, 67, 68, 0, 0, 0, 19,
	31, 0, 79, 71, 0, 0, 0, 0, 0, 39,
	0, 91, 93, 0, 73, 82, 0, 45, 0, 43,
	44, 0, 78, 21, 23, 0, 0, 84, 65, 0,
	0, 13, 0, 3, 33, 0, 34, 70, 48, 80,
	16, 32, 40, 92, 90, 0, 41, 42, 74, 76,
	0, 0, 0, 69, 0, 4, 37, 0, 71, 0,
	0, 0, 11, 15, 0, 35, 70, 94, 89, 10,
	2, 0, 36,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	47, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 3, 3,
	5, 6, 7, 8, 9, 10, 11, 12, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 50,
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
	39, 40, 41, 42, 43, 44, 45, 46, 48, 49,
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
		//line build/parse.y:172
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:179
		{
			yyVAL.block = CodeBlock{
				Start:      yyDollar[2].pos,
				Statements: yyDollar[3].exprs,
				End:        End{Pos: yyDollar[4].pos},
			}
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:187
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
		//line build/parse.y:199
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:204
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
		//line build/parse.y:235
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:241
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
		//line build/parse.y:255
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:259
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 10:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:265
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
		//line build/parse.y:278
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
		//line build/parse.y:288
		{
			yyVAL.expr = yyDollar[1].ifstmt
		}
	case 13:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:295
		{
			yyVAL.ifstmt = &IfStmt{
				ElsePos: yyDollar[1].pos,
				False:   yyDollar[3].block.Statements,
			}
		}
	case 15:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:306
		{
			inner := yyDollar[5].ifstmt
			inner.If = yyDollar[1].pos
			inner.Cond = yyDollar[2].expr
			inner.True = yyDollar[4].block.Statements
			yyVAL.ifstmt = &IfStmt{
				ElsePos: yyDollar[1].pos,
				False:   []Expr{inner},
			}
		}
	case 16:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:320
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].block.Statements,
			}
		}
	case 18:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:332
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			yyVAL.ifstmt.ElsePos = yyDollar[2].ifstmt.ElsePos
			yyVAL.ifstmt.False = yyDollar[2].ifstmt.False
		}
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:344
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 22:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:350
		{
			yyVAL.exprs = []Expr{}
		}
	case 23:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:354
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 25:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:361
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:368
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:374
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:384
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 32:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:393
		{
			yyVAL.expr = &CallExpr{
				X:              &LiteralExpr{Start: yyDollar[1].pos, Token: "load"},
				ListStart:      yyDollar[2].pos,
				List:           yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceCompact:   forceCompact(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
				ForceMultiLine: forceMultiLine(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 33:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:404
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
	case 34:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:415
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 35:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:424
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
	case 36:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:435
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
	case 37:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:448
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
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:465
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
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:477
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 40:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:487
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
	case 41:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:499
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
	case 42:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:511
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
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:523
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 44:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:533
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 45:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:543
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
	case 46:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:563
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 48:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:568
		{
			yyVAL.expr = &LambdaExpr{
				Lambda: yyDollar[1].pos,
				Var:    yyDollar[2].exprs,
				Colon:  yyDollar[3].pos,
				Expr:   yyDollar[4].expr,
			}
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:576
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 50:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:577
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 51:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:578
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 52:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:579
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 53:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:580
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:581
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:582
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 56:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:583
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 57:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:584
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 58:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:585
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:586
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 60:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:587
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:588
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 62:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:589
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 63:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:590
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 64:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:591
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 65:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:592
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:593
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 67:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:594
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 68:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:596
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 69:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:604
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 70:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:615
		{
			yyVAL.expr = nil
		}
	case 72:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:625
		{
			yyVAL.pos = Position{}
		}
	case 74:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:631
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:641
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 76:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:645
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:651
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 78:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:655
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 79:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:661
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:665
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 81:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:670
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 82:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:674
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 83:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:680
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:684
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:690
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:702
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 87:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:706
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 88:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:712
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 89:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:718
		{
			yyVAL.forc = &ForClause{
				For:  yyDollar[1].pos,
				Var:  yyDollar[2].exprs,
				In:   yyDollar[3].pos,
				Expr: yyDollar[4].expr,
			}
		}
	case 90:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:728
		{
			yyVAL.forifs = &ForClauseWithIfClausesOpt{
				For: yyDollar[1].forc,
				Ifs: yyDollar[2].ifs,
			}
		}
	case 91:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:737
		{
			yyVAL.forsifs = []*ForClauseWithIfClausesOpt{yyDollar[1].forifs}
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:740
		{
			yyVAL.forsifs = append(yyDollar[1].forsifs, yyDollar[2].forifs)
		}
	case 93:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:745
		{
			yyVAL.ifs = nil
		}
	case 94:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:749
		{
			yyVAL.ifs = append(yyDollar[1].ifs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	}
	goto yystack /* stack new state and value */
}
