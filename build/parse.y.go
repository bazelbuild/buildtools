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
const _STAR_STAR = 57363
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

//line build/parse.y:910

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
	-1, 117,
	26, 67,
	-2, 55,
}

const yyPrivate = 57344

const yyLast = 613

var yyAct = [...]int{

	17, 156, 195, 2, 7, 147, 154, 41, 131, 118,
	126, 23, 79, 130, 19, 35, 9, 114, 85, 205,
	213, 143, 70, 71, 32, 31, 36, 75, 77, 82,
	44, 45, 197, 107, 29, 113, 34, 133, 149, 140,
	91, 204, 88, 74, 206, 88, 139, 95, 96, 97,
	98, 99, 100, 101, 102, 103, 104, 105, 106, 29,
	108, 109, 110, 111, 93, 198, 117, 172, 87, 150,
	178, 78, 81, 128, 192, 133, 133, 119, 30, 177,
	94, 216, 129, 27, 119, 120, 185, 135, 22, 164,
	127, 136, 39, 138, 133, 26, 182, 28, 148, 133,
	208, 39, 144, 152, 73, 207, 29, 39, 157, 133,
	167, 188, 20, 24, 27, 47, 121, 21, 46, 159,
	31, 165, 166, 48, 39, 199, 26, 162, 28, 161,
	37, 168, 124, 184, 65, 176, 38, 29, 86, 173,
	64, 163, 179, 181, 24, 173, 66, 173, 36, 175,
	47, 31, 39, 46, 49, 180, 50, 183, 48, 191,
	189, 186, 187, 13, 128, 194, 142, 39, 39, 196,
	173, 153, 112, 193, 134, 119, 125, 171, 40, 160,
	137, 72, 63, 84, 69, 1, 201, 190, 25, 83,
	200, 80, 33, 43, 42, 202, 148, 203, 209, 210,
	16, 89, 90, 12, 8, 4, 211, 92, 212, 196,
	174, 214, 132, 67, 7, 68, 76, 123, 145, 27,
	146, 116, 0, 0, 22, 0, 0, 0, 0, 115,
	122, 26, 0, 28, 0, 0, 0, 6, 0, 0,
	11, 0, 29, 18, 0, 0, 0, 27, 20, 24,
	0, 151, 22, 21, 0, 15, 31, 10, 14, 26,
	215, 28, 5, 0, 0, 6, 3, 0, 11, 0,
	29, 18, 0, 0, 0, 0, 20, 24, 0, 0,
	0, 21, 27, 15, 31, 10, 14, 22, 169, 170,
	5, 0, 0, 0, 26, 0, 28, 0, 0, 0,
	0, 0, 0, 0, 0, 29, 0, 0, 0, 0,
	0, 20, 24, 0, 0, 0, 21, 0, 15, 31,
	47, 14, 0, 46, 49, 155, 50, 0, 48, 141,
	51, 0, 52, 0, 0, 0, 0, 0, 60, 0,
	0, 53, 0, 56, 0, 62, 0, 0, 57, 61,
	0, 0, 54, 55, 47, 58, 59, 46, 49, 0,
	50, 0, 48, 0, 51, 0, 52, 0, 0, 0,
	0, 0, 60, 0, 0, 53, 0, 56, 0, 62,
	158, 0, 57, 61, 0, 0, 54, 55, 47, 58,
	59, 46, 49, 0, 50, 0, 48, 0, 51, 0,
	52, 0, 0, 0, 0, 0, 60, 0, 0, 53,
	133, 56, 0, 62, 0, 0, 57, 61, 0, 0,
	54, 55, 47, 58, 59, 46, 49, 0, 50, 0,
	48, 0, 51, 0, 52, 0, 0, 0, 0, 0,
	60, 0, 0, 53, 0, 56, 0, 62, 0, 0,
	57, 61, 0, 0, 54, 55, 47, 58, 59, 46,
	49, 0, 50, 0, 48, 0, 51, 0, 52, 0,
	0, 0, 0, 0, 60, 0, 0, 53, 0, 56,
	0, 0, 0, 0, 57, 61, 0, 0, 54, 55,
	47, 58, 59, 46, 49, 0, 50, 0, 48, 0,
	51, 0, 52, 0, 0, 27, 0, 0, 60, 0,
	22, 53, 0, 56, 0, 0, 0, 26, 57, 28,
	0, 0, 54, 55, 0, 58, 59, 0, 29, 0,
	0, 0, 0, 0, 20, 24, 0, 0, 0, 21,
	0, 15, 31, 47, 14, 0, 46, 49, 0, 50,
	0, 48, 0, 51, 0, 52, 0, 0, 0, 0,
	0, 60, 0, 0, 53, 0, 56, 0, 0, 0,
	0, 57, 0, 0, 27, 54, 55, 47, 58, 22,
	46, 49, 0, 50, 0, 48, 26, 51, 28, 52,
	0, 0, 0, 0, 0, 0, 0, 29, 53, 0,
	56, 0, 0, 20, 24, 57, 0, 0, 21, 54,
	55, 31, 58,
}
var yyPact = [...]int{

	-1000, -1000, 242, -1000, -1000, -1000, -24, -1000, -1000, -1000,
	8, 109, -1000, 115, 569, -1000, 0, 418, 569, 129,
	569, 569, 569, -1000, 176, -17, 569, 569, 569, -1000,
	-1000, -1000, -1000, -33, 133, 36, 129, 569, 569, 569,
	143, -1000, -1000, 569, 51, -1000, 569, 569, 569, 569,
	569, 569, 569, 569, 569, 569, 569, 569, 1, 569,
	569, 569, 569, 159, 7, 78, 569, 119, 167, 143,
	-1000, -1000, 78, -1000, 64, 384, 165, 384, 174, 11,
	26, 19, 316, 157, -27, 500, 31, 569, 109, 143,
	143, 452, 158, 277, -1000, -1000, -1000, -1000, 111, 111,
	146, 146, 146, 146, 146, 146, 146, 569, 539, 573,
	486, 350, 277, -1000, 173, 83, 132, 418, -1000, 74,
	569, 569, 92, 118, 569, 569, -1000, 171, 418, -1000,
	49, -1000, -1000, 109, 569, -1000, 73, -1000, 50, -1000,
	-1000, 569, 569, -1000, -1000, 90, 124, -1000, 71, 6,
	6, 98, 129, 277, -1000, -1000, -1000, 146, 569, -1000,
	-1000, 68, -1000, 78, 569, 418, 418, -1000, 569, -1000,
	-1000, -1000, -1000, -1000, 3, 33, 418, -1000, -1000, 418,
	-1000, 316, 112, -1000, 31, 569, -1000, -1000, 277, 0,
	-4, 452, -1000, -1000, 418, 87, 418, 569, 569, 277,
	-1000, 418, -1000, -1000, -1000, -28, -1000, -1000, 569, 452,
	452, -1000, 214, -1000, 63, -1000, -1000,
}
var yyPgo = [...]int{

	0, 10, 9, 221, 17, 5, 220, 218, 0, 2,
	43, 14, 163, 217, 216, 215, 213, 15, 212, 8,
	13, 11, 210, 3, 205, 204, 203, 200, 194, 7,
	193, 1, 16, 192, 12, 191, 189, 78, 188, 6,
	187, 185, 183,
}
var yyR1 = [...]int{

	0, 41, 39, 39, 40, 40, 40, 23, 23, 23,
	23, 24, 24, 25, 25, 25, 28, 29, 29, 27,
	26, 26, 30, 30, 31, 33, 33, 32, 32, 32,
	32, 32, 32, 42, 42, 11, 11, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 11, 11,
	11, 4, 4, 3, 3, 2, 2, 2, 2, 7,
	7, 6, 6, 5, 5, 5, 5, 12, 12, 13,
	13, 15, 15, 16, 16, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 14, 14, 9, 9,
	10, 10, 1, 1, 34, 36, 36, 35, 35, 17,
	17, 37, 38, 38, 21, 18, 19, 20, 20, 22,
	22,
}
var yyR2 = [...]int{

	0, 2, 5, 1, 0, 3, 2, 0, 2, 2,
	3, 1, 1, 7, 6, 1, 3, 1, 5, 4,
	1, 2, 2, 1, 4, 0, 3, 1, 2, 1,
	3, 3, 1, 0, 1, 1, 3, 4, 4, 4,
	6, 8, 5, 1, 3, 4, 4, 4, 3, 3,
	3, 0, 2, 1, 3, 1, 3, 2, 2, 0,
	2, 1, 3, 1, 3, 2, 2, 1, 3, 0,
	1, 1, 3, 0, 2, 1, 4, 2, 2, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 4, 3, 3, 3, 5, 1, 3, 0, 1,
	0, 2, 0, 1, 3, 1, 3, 1, 2, 1,
	3, 1, 1, 2, 1, 4, 2, 1, 2, 0,
	3,
}
var yyChk = [...]int{

	-1000, -41, -23, 24, -24, 48, 23, -31, -25, -32,
	43, 26, -26, -12, 44, 41, -27, -8, 29, -11,
	34, 39, 10, -21, 35, -38, 17, 5, 19, 28,
	-37, 42, 48, -33, 28, -17, -11, 15, 21, 9,
	-12, -29, -28, -30, 30, 31, 7, 4, 12, 8,
	10, 14, 16, 25, 36, 37, 27, 32, 39, 40,
	22, 33, 29, -12, 11, 5, 17, -16, -15, -12,
	-8, -8, 5, -37, -10, -8, -14, -8, -10, -34,
	-35, -10, -8, -36, -42, 51, 5, 32, 9, -12,
	-12, -8, -12, 13, 29, -8, -8, -8, -8, -8,
	-8, -8, -8, -8, -8, -8, -8, 32, -8, -8,
	-8, -8, 13, 28, -4, -12, -3, -8, -2, -21,
	7, 38, -12, -13, 13, 9, -1, -4, -8, 18,
	-20, -19, -18, 26, 9, -1, -20, 6, -20, 20,
	20, 13, 9, 48, -32, -7, -6, -5, -21, 7,
	38, -12, -11, 13, -39, 48, -31, -8, 30, -39,
	6, -20, -1, 9, 15, -8, -8, 18, 13, -12,
	-12, 6, 18, -19, -22, -17, -8, 6, 20, -8,
	-34, -8, 6, -1, 9, 15, -21, -21, 13, -39,
	-40, -8, 6, -2, -8, -9, -8, 29, 32, 13,
	-5, -8, -39, -29, 45, 23, 48, 18, 13, -8,
	-8, -39, -23, 48, -9, 46, 18,
}
var yyDef = [...]int{

	7, -2, 0, 1, 8, 9, 0, 11, 12, 25,
	0, 0, 15, 27, 29, 32, 20, 67, 0, 75,
	73, 0, 0, 35, 0, 43, 100, 100, 100, 114,
	112, 111, 10, 33, 0, 0, 109, 0, 0, 0,
	28, 21, 17, 0, 0, 23, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 51, 69, 0, 102, 71,
	77, 78, 51, 113, 0, 96, 102, 96, 0, 105,
	0, 0, 96, 107, 0, 34, 59, 0, 0, 30,
	31, 68, 0, 0, 22, 79, 80, 81, 82, 83,
	84, 85, 86, 87, 88, 89, 90, 0, 92, 93,
	94, 0, 0, 36, 0, 0, 102, -2, 53, 35,
	0, 0, 70, 0, 0, 103, 74, 0, 55, 44,
	0, 117, 119, 0, 103, 101, 0, 50, 0, 48,
	49, 0, 108, 24, 26, 0, 102, 61, 63, 0,
	0, 0, 110, 0, 16, 4, 3, 91, 0, 19,
	38, 0, 52, 103, 0, 57, 58, 39, 98, 76,
	72, 37, 45, 118, 116, 0, 97, 46, 47, 104,
	106, 0, 0, 60, 103, 0, 65, 66, 0, 0,
	0, 95, 42, 54, 56, 0, 99, 0, 0, 0,
	62, 64, 14, 18, 7, 0, 6, 40, 98, 120,
	115, 13, 0, 5, 0, 2, 41,
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
		//line build/parse.y:184
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:191
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
			yyVAL.block = CodeBlock{
				Start:      yyDollar[3].pos,
				Statements: statements,
				End:        End{Pos: yyDollar[5].pos},
			}
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:214
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
		//line build/parse.y:226
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 5:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:231
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
	case 6:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:243
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:249
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:254
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
	case 9:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:285
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:291
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
	case 11:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:305
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 12:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:309
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 13:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:315
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
	case 14:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:328
		{
			yyVAL.expr = &ForLoop{
				Start:    yyDollar[1].pos,
				LoopVars: yyDollar[2].exprs,
				Iterable: yyDollar[4].expr,
				Body:     yyDollar[6].block,
				End:      yyDollar[6].block.End,
			}
		}
	case 15:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:338
		{
			yyVAL.expr = yyDollar[1].ifstmt
		}
	case 16:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:345
		{
			yyVAL.ifstmt = &IfStmt{
				ElsePos: yyDollar[1].pos,
				False:   yyDollar[3].block.Statements,
			}
		}
	case 18:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:356
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
	case 19:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:370
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].block.Statements,
			}
		}
	case 21:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:382
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			yyVAL.ifstmt.ElsePos = yyDollar[2].ifstmt.ElsePos
			yyVAL.ifstmt.False = yyDollar[2].ifstmt.False
		}
	case 24:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:394
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 25:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:400
		{
			yyVAL.exprs = []Expr{}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:404
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:411
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:418
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:423
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:424
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:426
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:436
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 37:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:445
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
	case 38:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:456
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
	case 39:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:467
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 40:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:476
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
	case 41:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:487
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
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:500
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
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:517
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
	case 44:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:529
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 45:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:538
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
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:550
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
	case 47:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:562
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
	case 48:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:574
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:583
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 50:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:592
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
	case 51:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:613
		{
			yyVAL.exprs = nil
		}
	case 52:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:617
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 53:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:623
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:627
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 56:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:634
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:638
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:642
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 59:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:647
		{
			yyVAL.exprs = nil
		}
	case 60:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:651
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:657
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 62:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:661
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 64:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:668
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 65:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:672
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 66:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:676
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 68:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:683
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || !tuple.Start.IsValid() {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 69:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:697
		{
			yyVAL.expr = nil
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:704
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:708
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 73:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:713
		{
			yyVAL.exprs = nil
		}
	case 74:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:717
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 76:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:724
		{
			yyVAL.expr = &LambdaExpr{
				Lambda: yyDollar[1].pos,
				Var:    yyDollar[2].exprs,
				Colon:  yyDollar[3].pos,
				Expr:   yyDollar[4].expr,
			}
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:732
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 78:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:733
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:734
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:735
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:736
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:737
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:738
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:739
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 85:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:740
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:741
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:742
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:743
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:744
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:745
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 91:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:746
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 92:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:747
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 93:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:748
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 94:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:750
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 95:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:758
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 96:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:770
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 97:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:774
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 98:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:779
		{
			yyVAL.expr = nil
		}
	case 100:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:785
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 101:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:789
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 102:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:798
		{
			yyVAL.pos = Position{}
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:804
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 105:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:814
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:818
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:824
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 108:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:828
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 109:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:834
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:838
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 111:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:844
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 112:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:856
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:860
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 114:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:866
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 115:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:872
		{
			yyVAL.forc = &ForClause{
				For:  yyDollar[1].pos,
				Var:  yyDollar[2].exprs,
				In:   yyDollar[3].pos,
				Expr: yyDollar[4].expr,
			}
		}
	case 116:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:882
		{
			yyVAL.forifs = &ForClauseWithIfClausesOpt{
				For: yyDollar[1].forc,
				Ifs: yyDollar[2].ifs,
			}
		}
	case 117:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:891
		{
			yyVAL.forsifs = []*ForClauseWithIfClausesOpt{yyDollar[1].forifs}
		}
	case 118:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:894
		{
			yyVAL.forsifs = append(yyDollar[1].forsifs, yyDollar[2].forifs)
		}
	case 119:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:899
		{
			yyVAL.ifs = nil
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:903
		{
			yyVAL.ifs = append(yyDollar[1].ifs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	}
	goto yystack /* stack new state and value */
}
