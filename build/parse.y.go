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
	string  *StringExpr
	strings []*StringExpr
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

//line build/parse.y:864

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

const yyLast = 606

var yyAct = [...]int{

	17, 153, 190, 2, 7, 144, 151, 41, 116, 125,
	77, 129, 19, 35, 9, 23, 128, 114, 85, 200,
	208, 202, 70, 71, 36, 140, 32, 75, 80, 83,
	31, 44, 45, 170, 107, 29, 113, 34, 131, 136,
	91, 199, 88, 74, 201, 88, 146, 95, 96, 97,
	98, 99, 100, 101, 102, 103, 104, 105, 106, 13,
	108, 109, 110, 111, 173, 193, 117, 29, 87, 93,
	131, 79, 82, 117, 40, 168, 211, 147, 63, 30,
	69, 118, 135, 131, 204, 94, 133, 39, 118, 203,
	126, 127, 28, 37, 134, 180, 160, 89, 90, 38,
	141, 149, 145, 92, 26, 73, 27, 39, 154, 39,
	39, 65, 194, 183, 150, 29, 163, 64, 164, 156,
	161, 162, 24, 66, 47, 158, 121, 46, 49, 31,
	50, 39, 48, 172, 47, 112, 123, 46, 174, 176,
	169, 179, 48, 159, 36, 171, 169, 148, 39, 175,
	138, 132, 124, 178, 177, 167, 187, 184, 157, 139,
	117, 189, 181, 182, 86, 191, 72, 84, 188, 186,
	1, 192, 185, 25, 81, 118, 78, 33, 43, 42,
	16, 196, 12, 165, 166, 195, 8, 4, 130, 67,
	197, 68, 198, 76, 205, 145, 122, 142, 143, 28,
	115, 206, 0, 207, 22, 191, 0, 209, 0, 7,
	0, 26, 0, 27, 0, 0, 0, 6, 0, 0,
	11, 0, 29, 18, 0, 0, 0, 28, 20, 24,
	0, 0, 22, 21, 0, 15, 31, 10, 14, 26,
	210, 27, 5, 0, 0, 6, 3, 0, 11, 0,
	29, 18, 0, 0, 0, 28, 20, 24, 0, 0,
	22, 21, 0, 15, 31, 10, 14, 26, 0, 27,
	5, 0, 0, 0, 0, 0, 0, 0, 29, 0,
	0, 0, 0, 0, 20, 24, 0, 0, 0, 21,
	0, 15, 31, 47, 14, 0, 46, 49, 152, 50,
	0, 48, 137, 51, 0, 52, 0, 0, 0, 0,
	0, 60, 0, 0, 53, 0, 56, 0, 62, 0,
	0, 57, 61, 0, 0, 54, 55, 47, 58, 59,
	46, 49, 0, 50, 0, 48, 0, 51, 0, 52,
	0, 0, 0, 0, 0, 60, 0, 0, 53, 0,
	56, 0, 62, 155, 0, 57, 61, 0, 0, 54,
	55, 47, 58, 59, 46, 49, 0, 50, 0, 48,
	0, 51, 0, 52, 0, 0, 0, 0, 0, 60,
	0, 0, 53, 131, 56, 0, 62, 0, 0, 57,
	61, 0, 0, 54, 55, 47, 58, 59, 46, 49,
	0, 50, 0, 48, 0, 51, 0, 52, 0, 0,
	0, 0, 0, 60, 0, 0, 53, 0, 56, 0,
	62, 0, 0, 57, 61, 0, 0, 54, 55, 47,
	58, 59, 46, 49, 0, 50, 0, 48, 0, 51,
	0, 52, 0, 0, 0, 0, 0, 60, 0, 0,
	53, 0, 56, 0, 0, 0, 0, 57, 61, 0,
	0, 54, 55, 47, 58, 59, 46, 49, 0, 50,
	0, 48, 0, 51, 0, 52, 0, 0, 28, 0,
	0, 60, 0, 22, 53, 0, 56, 0, 0, 0,
	26, 57, 27, 0, 0, 54, 55, 0, 58, 59,
	0, 29, 0, 0, 28, 0, 119, 20, 24, 22,
	0, 0, 21, 0, 15, 31, 26, 14, 27, 0,
	0, 0, 0, 0, 0, 0, 0, 29, 0, 0,
	0, 0, 0, 20, 24, 0, 47, 120, 21, 46,
	49, 31, 50, 0, 48, 0, 51, 0, 52, 0,
	0, 0, 0, 0, 60, 0, 0, 53, 0, 56,
	0, 0, 0, 0, 57, 0, 0, 28, 54, 55,
	47, 58, 22, 46, 49, 0, 50, 0, 48, 26,
	51, 27, 52, 0, 0, 0, 0, 0, 0, 0,
	29, 53, 0, 56, 0, 0, 20, 24, 57, 0,
	0, 21, 54, 55, 31, 58,
}
var yyPact = [...]int{

	-1000, -1000, 222, -1000, -1000, -1000, -22, -1000, -1000, -1000,
	9, 87, -1000, 78, 562, -1000, 1, 391, 562, 106,
	562, 562, 562, -1000, 161, -12, 562, 562, 562, -1000,
	-1000, -1000, -1000, -33, 159, 36, 106, 562, 562, 562,
	139, -1000, -1000, 562, 56, -1000, 562, 562, 562, 562,
	562, 562, 562, 562, 562, 562, 562, 562, 2, 562,
	562, 562, 562, 122, 8, 499, 562, 123, 143, 139,
	-1000, -1000, 499, -1000, 73, 357, 142, 12, 62, 19,
	289, 141, 153, 391, -23, 473, 39, 562, 87, 139,
	139, 425, 101, 250, -1000, -1000, -1000, -1000, 130, 130,
	120, 120, 120, 120, 120, 120, 120, 562, 532, 566,
	459, 323, 250, -1000, 152, 134, -1000, 391, 81, 562,
	562, 98, 105, 562, 562, -1000, 149, -1000, 57, 4,
	-1000, 87, 562, -1000, 44, -1000, -1000, 562, 562, -1000,
	-1000, -1000, 148, 132, -1000, 80, 7, 7, 100, 106,
	250, -1000, -1000, -1000, 120, 562, -1000, -1000, -1000, 499,
	562, 391, 391, -1000, 562, -1000, -1000, -1000, -1000, 4,
	562, 33, 391, -1000, 391, -1000, 289, 99, -1000, 39,
	562, -1000, -1000, 250, 1, -4, -27, 425, -1000, 391,
	71, 391, 425, 562, 250, -1000, 391, -1000, -1000, -1000,
	-28, -1000, -1000, -1000, 562, 425, -1000, 194, -1000, 58,
	-1000, -1000,
}
var yyPgo = [...]int{

	0, 9, 8, 200, 17, 5, 198, 197, 0, 2,
	43, 12, 59, 196, 193, 191, 189, 13, 188, 11,
	16, 15, 3, 187, 186, 182, 180, 179, 7, 178,
	1, 14, 177, 10, 176, 174, 79, 173, 6, 172,
	170, 169, 167,
}
var yyR1 = [...]int{

	0, 40, 38, 38, 41, 41, 39, 39, 39, 22,
	22, 22, 22, 23, 23, 24, 24, 24, 27, 28,
	28, 26, 25, 25, 29, 29, 30, 32, 32, 31,
	31, 31, 31, 31, 31, 42, 42, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 11, 11,
	11, 4, 4, 3, 3, 2, 2, 2, 2, 7,
	7, 6, 6, 5, 5, 5, 5, 12, 12, 13,
	13, 15, 15, 16, 16, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 14, 14, 9, 9,
	10, 10, 1, 1, 33, 35, 35, 34, 34, 17,
	17, 36, 37, 37, 21, 18, 19, 19, 20, 20,
}
var yyR2 = [...]int{

	0, 2, 5, 2, 0, 2, 0, 3, 2, 0,
	2, 2, 3, 1, 1, 7, 6, 1, 3, 1,
	5, 4, 1, 2, 2, 1, 4, 0, 3, 1,
	2, 1, 3, 3, 1, 0, 1, 1, 3, 4,
	4, 4, 6, 8, 1, 3, 4, 4, 3, 3,
	3, 0, 2, 1, 3, 1, 3, 2, 2, 0,
	2, 1, 3, 1, 3, 2, 2, 1, 3, 0,
	1, 1, 3, 0, 2, 1, 4, 2, 2, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 4, 3, 3, 3, 5, 1, 3, 0, 1,
	0, 2, 0, 1, 3, 1, 3, 1, 2, 1,
	3, 1, 1, 2, 1, 4, 1, 3, 1, 2,
}
var yyChk = [...]int{

	-1000, -40, -22, 24, -23, 48, 23, -30, -24, -31,
	43, 26, -25, -12, 44, 41, -26, -8, 29, -11,
	34, 39, 10, -21, 35, -37, 17, 19, 5, 28,
	-36, 42, 48, -32, 28, -17, -11, 15, 21, 9,
	-12, -28, -27, -29, 30, 31, 7, 4, 12, 8,
	10, 14, 16, 25, 36, 37, 27, 32, 39, 40,
	22, 33, 29, -12, 11, 5, 17, -16, -15, -12,
	-8, -8, 5, -36, -10, -8, -14, -33, -34, -10,
	-8, -35, -10, -8, -42, 51, 5, 32, 9, -12,
	-12, -8, -12, 13, 29, -8, -8, -8, -8, -8,
	-8, -8, -8, -8, -8, -8, -8, 32, -8, -8,
	-8, -8, 13, 28, -4, -3, -2, -8, -21, 7,
	38, -12, -13, 13, 9, -1, -4, 18, -20, -19,
	-18, 26, 9, -1, -20, 20, 20, 13, 9, 6,
	48, -31, -7, -6, -5, -21, 7, 38, -12, -11,
	13, -38, 48, -30, -8, 30, -38, 6, -1, 9,
	15, -8, -8, 18, 13, -12, -12, 6, 18, -19,
	29, -17, -8, 20, -8, -33, -8, 6, -1, 9,
	15, -21, -21, 13, -38, -39, -41, -8, -2, -8,
	-9, -8, -8, 32, 13, -5, -8, -38, -28, 45,
	23, 48, 48, 18, 13, -8, -38, -22, 48, -9,
	46, 18,
}
var yyDef = [...]int{

	9, -2, 0, 1, 10, 11, 0, 13, 14, 27,
	0, 0, 17, 29, 31, 34, 22, 67, 0, 75,
	73, 0, 0, 37, 0, 44, 100, 100, 100, 114,
	112, 111, 12, 35, 0, 0, 109, 0, 0, 0,
	30, 23, 19, 0, 0, 25, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 51, 69, 0, 102, 71,
	77, 78, 51, 113, 0, 96, 102, 105, 0, 0,
	96, 107, 0, 96, 0, 36, 59, 0, 0, 32,
	33, 68, 0, 0, 24, 79, 80, 81, 82, 83,
	84, 85, 86, 87, 88, 89, 90, 0, 92, 93,
	94, 0, 0, 38, 0, 102, 53, 55, 37, 0,
	0, 70, 0, 0, 103, 74, 0, 45, 0, 118,
	116, 0, 103, 101, 0, 48, 49, 0, 108, 50,
	26, 28, 0, 102, 61, 63, 0, 0, 0, 110,
	0, 18, 6, 4, 91, 0, 21, 40, 52, 103,
	0, 57, 58, 41, 98, 76, 72, 39, 46, 119,
	0, 0, 97, 47, 104, 106, 0, 0, 60, 103,
	0, 65, 66, 0, 0, 0, 3, 95, 54, 56,
	0, 99, 117, 0, 0, 62, 64, 16, 20, 9,
	0, 8, 5, 42, 98, 115, 15, 0, 7, 0,
	2, 43,
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
		//line build/parse.y:178
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:185
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
		}
	case 3:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:204
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:212
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:217
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
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:229
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 9:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:235
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:240
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
	case 11:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:271
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:277
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
	case 13:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:291
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:295
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 15:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:301
		{
			yyVAL.expr = &DefStmt{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[4].exprs,
					Body:     yyDollar[7].exprs,
				},
				Name:           yyDollar[2].tok,
				ForceCompact:   forceCompact(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
				ForceMultiLine: forceMultiLine(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
			}
		}
	case 16:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:314
		{
			yyVAL.expr = &ForStmt{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				X:    yyDollar[4].expr,
				Body: yyDollar[6].exprs,
			}
		}
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:323
		{
			yyVAL.expr = yyDollar[1].ifstmt
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:330
		{
			yyVAL.ifstmt = &IfStmt{
				ElsePos: yyDollar[1].pos,
				False:   yyDollar[3].exprs,
			}
		}
	case 20:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:341
		{
			inner := yyDollar[5].ifstmt
			inner.If = yyDollar[1].pos
			inner.Cond = yyDollar[2].expr
			inner.True = yyDollar[4].exprs
			yyVAL.ifstmt = &IfStmt{
				ElsePos: yyDollar[1].pos,
				False:   []Expr{inner},
			}
		}
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:355
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].exprs,
			}
		}
	case 23:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:367
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			yyVAL.ifstmt.ElsePos = yyDollar[2].ifstmt.ElsePos
			yyVAL.ifstmt.False = yyDollar[2].ifstmt.False
		}
	case 26:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:379
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 27:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:385
		{
			yyVAL.exprs = []Expr{}
		}
	case 28:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:389
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 30:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:396
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 31:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:403
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:408
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:409
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 34:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:411
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 38:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:421
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 39:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:430
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
	case 40:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:441
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
	case 41:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:452
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 42:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:461
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
	case 43:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:472
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
	case 44:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:485
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
	case 45:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:497
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:506
		{
			exprStart, _ := yyDollar[2].expr.Span()
			yyVAL.expr = &Comprehension{
				Curly:          false,
				Lbrack:         yyDollar[1].pos,
				Body:           yyDollar[2].expr,
				Clauses:        yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: yyDollar[1].pos.Line != exprStart.Line,
			}
		}
	case 47:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:518
		{
			exprStart, _ := yyDollar[2].expr.Span()
			yyVAL.expr = &Comprehension{
				Curly:          true,
				Lbrack:         yyDollar[1].pos,
				Body:           yyDollar[2].expr,
				Clauses:        yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: yyDollar[1].pos.Line != exprStart.Line,
			}
		}
	case 48:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:530
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
		//line build/parse.y:539
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
		//line build/parse.y:548
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
		//line build/parse.y:569
		{
			yyVAL.exprs = nil
		}
	case 52:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:573
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 53:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:579
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:583
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 56:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:590
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:594
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:598
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 59:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:603
		{
			yyVAL.exprs = nil
		}
	case 60:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:607
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:613
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 62:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:617
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 64:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:624
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 65:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:628
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 66:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:632
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 68:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:639
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || tuple.Start.IsValid() {
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
		//line build/parse.y:653
		{
			yyVAL.expr = nil
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:660
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:664
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 73:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:669
		{
			yyVAL.exprs = nil
		}
	case 74:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:673
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 76:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:680
		{
			yyVAL.expr = &LambdaExpr{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[2].exprs,
					Body:     []Expr{yyDollar[4].expr},
				},
			}
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:689
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 78:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:690
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:691
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:692
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:693
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:694
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:695
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:696
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 85:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:697
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:698
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:699
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:700
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:701
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:702
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 91:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:703
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 92:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:704
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 93:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:705
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 94:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:707
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 95:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:715
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
		//line build/parse.y:727
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 97:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:731
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 98:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:736
		{
			yyVAL.expr = nil
		}
	case 100:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:742
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 101:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:746
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 102:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:755
		{
			yyVAL.pos = Position{}
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:761
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 105:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:771
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:775
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:781
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 108:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:785
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:792
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || tuple.Start.IsValid() {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 111:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:807
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
		//line build/parse.y:819
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:823
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 114:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:829
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 115:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:835
		{
			yyVAL.expr = &ForClause{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				In:   yyDollar[3].pos,
				X:    yyDollar[4].expr,
			}
		}
	case 116:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:845
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:848
		{
			yyVAL.exprs = append(yyDollar[1].exprs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:857
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 119:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:860
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
		}
	}
	goto yystack /* stack new state and value */
}
