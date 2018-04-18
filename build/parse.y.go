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
	loadarg *struct {
		from Ident
		to   Ident
	}
	loadargs []*struct {
		from Ident
		to   Ident
	}

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
const _STAR_STAR = 57364
const _NOT = 57365
const _OR = 57366
const _PYTHON = 57367
const _STRING = 57368
const _DEF = 57369
const _RETURN = 57370
const _INDENT = 57371
const _UNINDENT = 57372
const ShiftInstead = 57373
const _ASSERT = 57374
const _UNARY = 57375

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

//line build/parse.y:911

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

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 624

var yyAct = [...]int{

	17, 195, 23, 155, 32, 192, 7, 2, 146, 127,
	37, 131, 78, 118, 9, 19, 130, 153, 86, 216,
	205, 207, 71, 72, 142, 34, 30, 38, 76, 81,
	84, 74, 33, 108, 89, 44, 45, 172, 30, 115,
	89, 33, 92, 204, 148, 75, 206, 96, 97, 98,
	99, 100, 101, 102, 103, 104, 105, 106, 107, 199,
	109, 110, 111, 112, 113, 88, 30, 119, 94, 120,
	13, 36, 175, 133, 80, 83, 138, 149, 128, 133,
	61, 137, 222, 170, 129, 42, 95, 135, 209, 64,
	147, 70, 133, 208, 41, 136, 47, 212, 182, 46,
	49, 143, 50, 165, 48, 151, 51, 200, 52, 156,
	90, 91, 162, 61, 93, 60, 166, 41, 53, 125,
	56, 185, 163, 164, 41, 41, 57, 160, 152, 114,
	54, 55, 158, 58, 59, 174, 211, 41, 123, 181,
	176, 178, 171, 39, 173, 169, 161, 41, 171, 38,
	40, 183, 184, 177, 29, 180, 140, 134, 189, 150,
	126, 218, 119, 191, 120, 179, 27, 193, 28, 159,
	186, 141, 197, 198, 196, 190, 66, 47, 30, 31,
	46, 87, 65, 202, 147, 48, 25, 73, 67, 85,
	201, 188, 1, 33, 61, 194, 167, 168, 187, 26,
	213, 82, 79, 203, 210, 35, 43, 16, 12, 8,
	193, 4, 215, 219, 197, 217, 196, 220, 214, 7,
	29, 24, 132, 47, 68, 22, 46, 49, 69, 50,
	77, 48, 27, 124, 28, 144, 145, 116, 117, 6,
	61, 0, 11, 0, 30, 31, 18, 0, 0, 0,
	29, 20, 25, 0, 0, 22, 21, 0, 15, 33,
	10, 14, 27, 221, 28, 5, 0, 0, 0, 6,
	3, 0, 11, 29, 30, 31, 18, 0, 22, 0,
	0, 20, 25, 0, 0, 27, 21, 28, 15, 33,
	10, 14, 0, 0, 0, 5, 0, 30, 31, 0,
	0, 0, 0, 0, 20, 25, 0, 0, 0, 21,
	0, 15, 33, 47, 14, 0, 46, 49, 154, 50,
	0, 48, 139, 51, 0, 52, 0, 0, 0, 0,
	61, 0, 60, 0, 0, 53, 0, 56, 0, 0,
	63, 0, 0, 57, 62, 0, 0, 54, 55, 47,
	58, 59, 46, 49, 0, 50, 0, 48, 0, 51,
	0, 52, 0, 0, 0, 0, 61, 0, 60, 0,
	0, 53, 0, 56, 0, 0, 63, 157, 0, 57,
	62, 0, 0, 54, 55, 47, 58, 59, 46, 49,
	0, 50, 0, 48, 0, 51, 0, 52, 0, 0,
	0, 0, 61, 0, 60, 0, 0, 53, 133, 56,
	0, 0, 63, 0, 0, 57, 62, 0, 0, 54,
	55, 47, 58, 59, 46, 49, 0, 50, 0, 48,
	0, 51, 0, 52, 0, 0, 0, 0, 61, 0,
	60, 0, 0, 53, 0, 56, 0, 0, 63, 0,
	0, 57, 62, 0, 0, 54, 55, 47, 58, 59,
	46, 49, 0, 50, 0, 48, 0, 51, 0, 52,
	0, 0, 29, 0, 61, 0, 60, 22, 0, 53,
	0, 56, 0, 0, 27, 0, 28, 57, 62, 0,
	0, 54, 55, 0, 58, 59, 30, 31, 0, 0,
	29, 0, 121, 20, 25, 22, 0, 0, 21, 0,
	15, 33, 27, 14, 28, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 30, 31, 0, 0, 0, 0,
	0, 20, 25, 0, 47, 122, 21, 46, 49, 33,
	50, 0, 48, 0, 51, 0, 52, 0, 0, 0,
	0, 61, 0, 60, 0, 0, 53, 0, 56, 0,
	0, 0, 0, 0, 57, 0, 0, 0, 54, 55,
	47, 58, 0, 46, 49, 0, 50, 0, 48, 0,
	51, 0, 52, 0, 29, 0, 0, 61, 0, 22,
	0, 0, 53, 0, 56, 0, 27, 0, 28, 0,
	57, 0, 0, 0, 54, 55, 0, 58, 30, 31,
	0, 0, 0, 0, 0, 20, 25, 0, 0, 0,
	21, 0, 0, 33,
}
var yyPact = [...]int{

	-1000, -1000, 245, -1000, -1000, -1000, -25, -1000, -1000, -1000,
	42, 149, -1000, 128, 579, -1000, 3, 417, 579, 171,
	579, 579, 579, -1000, -1000, 182, -12, 579, 579, 579,
	-1000, -1000, -1000, -1000, -1000, -35, 176, 31, 171, 579,
	579, 579, 138, 579, 55, -1000, 579, 579, 579, 579,
	579, 579, 579, 579, 579, 579, 579, 579, -1, 579,
	579, 579, 579, 579, 116, 10, 495, 579, 106, 151,
	138, 59, 59, -12, -1000, 66, 381, 148, 46, 61,
	56, 309, 147, 165, 417, -26, 467, 37, 579, 149,
	138, 138, 453, 115, 268, -1000, 59, 59, 59, 173,
	173, 219, 219, 219, 219, 219, 219, 219, 579, 530,
	566, 417, 92, 345, 268, -1000, 163, 137, -1000, 417,
	97, 579, 579, 85, 103, 579, 579, -1000, 136, -1000,
	65, 6, -1000, 149, 579, -1000, 52, -1000, -1000, 579,
	579, -1000, -1000, -1000, 159, 130, -1000, 83, 9, 9,
	108, 171, 268, -1000, -1000, -1000, 219, 579, -1000, -1000,
	-1000, 495, 579, 417, 417, -1000, 579, -1000, -1000, -3,
	-1000, 6, 579, 25, 417, -1000, 417, -1000, 309, 94,
	-1000, 37, 579, -1000, -1000, 268, -1000, -4, -29, 453,
	-1000, 417, 75, 417, 127, -1000, -1000, 82, 453, 579,
	268, -1000, 417, -1000, -1000, -31, -1000, -1000, -1000, 579,
	155, -3, -12, 453, -1000, 215, -1000, 64, -1000, -1000,
	-1000, -1000, -1000,
}
var yyPgo = [...]int{

	0, 9, 13, 238, 237, 8, 236, 235, 0, 5,
	45, 15, 70, 233, 230, 228, 224, 10, 222, 11,
	16, 2, 221, 7, 211, 209, 208, 207, 206, 3,
	14, 205, 12, 202, 201, 4, 199, 17, 198, 1,
	195, 192, 191, 189,
}
var yyR1 = [...]int{

	0, 41, 37, 37, 42, 42, 38, 38, 38, 23,
	23, 23, 23, 24, 24, 25, 25, 25, 27, 27,
	26, 26, 28, 28, 29, 31, 31, 30, 30, 30,
	30, 30, 30, 43, 43, 11, 11, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 11, 11,
	4, 4, 3, 3, 2, 2, 2, 2, 40, 40,
	39, 39, 7, 7, 6, 6, 5, 5, 5, 5,
	12, 12, 13, 13, 15, 15, 16, 16, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	14, 14, 9, 9, 10, 10, 1, 1, 32, 34,
	34, 33, 33, 17, 17, 35, 36, 36, 21, 22,
	18, 19, 19, 20, 20,
}
var yyR2 = [...]int{

	0, 2, 5, 2, 0, 2, 0, 3, 2, 0,
	2, 2, 3, 1, 1, 7, 6, 1, 4, 5,
	1, 4, 2, 1, 4, 0, 3, 1, 2, 1,
	3, 3, 1, 0, 1, 1, 1, 3, 7, 4,
	4, 6, 8, 1, 3, 4, 4, 3, 3, 3,
	0, 2, 1, 3, 1, 3, 2, 2, 1, 3,
	1, 3, 0, 2, 1, 3, 1, 3, 2, 2,
	1, 3, 0, 1, 1, 3, 0, 2, 1, 4,
	2, 2, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 4, 3, 3, 3, 3, 5,
	1, 3, 0, 1, 0, 2, 0, 1, 3, 1,
	3, 1, 2, 1, 3, 1, 1, 2, 1, 1,
	4, 1, 3, 1, 2,
}
var yyChk = [...]int{

	-1000, -41, -23, 25, -24, 50, 24, -29, -25, -30,
	45, 27, -26, -12, 46, 43, -27, -8, 31, -11,
	36, 41, 10, -21, -22, 37, -36, 17, 19, 5,
	29, 30, -35, 44, 50, -31, 29, -17, -11, 15,
	22, 9, -12, -28, 32, 33, 7, 4, 12, 8,
	10, 14, 16, 26, 38, 39, 28, 34, 41, 42,
	23, 21, 35, 31, -12, 11, 5, 17, -16, -15,
	-12, -8, -8, 5, -35, -10, -8, -14, -32, -33,
	-10, -8, -34, -10, -8, -43, 53, 5, 34, 9,
	-12, -12, -8, -12, 13, 31, -8, -8, -8, -8,
	-8, -8, -8, -8, -8, -8, -8, -8, 34, -8,
	-8, -8, -8, -8, 13, 29, -4, -3, -2, -8,
	-21, 7, 40, -12, -13, 13, 9, -1, -35, 18,
	-20, -19, -18, 27, 9, -1, -20, 20, 20, 13,
	9, 6, 50, -30, -7, -6, -5, -21, 7, 40,
	-12, -11, 13, -37, 50, -29, -8, 32, -37, 6,
	-1, 9, 15, -8, -8, 18, 13, -12, -12, 9,
	18, -19, 31, -17, -8, 20, -8, -32, -8, 6,
	-1, 9, 15, -21, -21, 13, -37, -38, -42, -8,
	-2, -8, -9, -8, -40, -39, -35, -21, -8, 34,
	13, -5, -8, -37, 47, 24, 50, 50, 18, 13,
	-1, 9, 15, -8, -37, -23, 50, -9, 6, -39,
	-35, 48, 18,
}
var yyDef = [...]int{

	9, -2, 0, 1, 10, 11, 0, 13, 14, 25,
	0, 0, 17, 27, 29, 32, 20, 70, 0, 78,
	76, 0, 0, 35, 36, 0, 43, 104, 104, 104,
	118, 119, 116, 115, 12, 33, 0, 0, 113, 0,
	0, 0, 28, 0, 0, 23, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 50, 72, 0, 106,
	74, 80, 81, 0, 117, 0, 100, 106, 109, 0,
	0, 100, 111, 0, 100, 0, 34, 62, 0, 0,
	30, 31, 71, 0, 0, 22, 82, 83, 84, 85,
	86, 87, 88, 89, 90, 91, 92, 93, 0, 95,
	96, 97, 98, 0, 0, 37, 0, 106, 52, 54,
	35, 0, 0, 73, 0, 0, 107, 77, 0, 44,
	0, 123, 121, 0, 107, 105, 0, 47, 48, 0,
	112, 49, 24, 26, 0, 106, 64, 66, 0, 0,
	0, 114, 0, 21, 6, 4, 94, 0, 18, 39,
	51, 107, 0, 56, 57, 40, 102, 79, 75, 0,
	45, 124, 0, 0, 101, 46, 108, 110, 0, 0,
	63, 107, 0, 68, 69, 0, 19, 0, 3, 99,
	53, 55, 0, 103, 106, 58, 60, 0, 122, 0,
	0, 65, 67, 16, 9, 0, 8, 5, 41, 102,
	0, 107, 0, 120, 15, 0, 7, 0, 38, 59,
	61, 2, 42,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	50, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 3, 3,
	5, 6, 7, 8, 9, 10, 11, 12, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 53,
	14, 15, 16, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 17, 3, 18, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 19, 21, 20,
}
var yyTok2 = [...]int{

	2, 3, 22, 23, 24, 25, 26, 27, 28, 29,
	30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
	40, 41, 42, 43, 44, 45, 46, 47, 48, 49,
	51, 52,
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
		//line build/parse.y:183
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:190
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
		//line build/parse.y:209
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:217
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:222
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
		//line build/parse.y:234
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 9:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:240
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:245
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
		//line build/parse.y:276
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:282
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
		//line build/parse.y:296
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:300
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 15:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:306
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
		//line build/parse.y:319
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
		//line build/parse.y:328
		{
			yyVAL.expr = yyDollar[1].ifstmt
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:335
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].exprs,
			}
		}
	case 19:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:343
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
		}
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:363
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = End{Pos: yyDollar[2].pos}
			inner.False = yyDollar[4].exprs
		}
	case 24:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:379
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 25:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:385
		{
			yyVAL.exprs = []Expr{}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:389
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:396
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:403
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:408
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:409
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:411
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 37:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:422
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 38:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:431
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
	case 39:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:445
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
	case 40:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:456
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 41:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:465
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
	case 42:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:476
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
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:489
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
		//line build/parse.y:501
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
		//line build/parse.y:510
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
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:522
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
	case 47:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:534
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 48:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:543
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:552
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
	case 50:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:573
		{
			yyVAL.exprs = nil
		}
	case 51:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:577
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 52:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:583
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 53:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:587
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:594
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 56:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:598
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:602
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:607
		{
			yyVAL.loadargs = []*struct {
				from Ident
				to   Ident
			}{yyDollar[1].loadarg}
		}
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:611
		{
			yyDollar[1].loadargs = append(yyDollar[1].loadargs, yyDollar[3].loadarg)
			yyVAL.loadargs = yyDollar[1].loadargs
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:617
		{
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[1].string.Value,
					NamePos: yyDollar[1].string.Start,
				},
				to: Ident{
					Name:    yyDollar[1].string.Value,
					NamePos: yyDollar[1].string.Start,
				},
			}
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:630
		{
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[3].string.Value,
					NamePos: yyDollar[3].string.Start,
				},
				to: *yyDollar[1].expr.(*Ident),
			}
		}
	case 62:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:641
		{
			yyVAL.exprs = nil
		}
	case 63:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:645
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:651
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 65:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:655
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 67:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:662
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:666
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 69:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:670
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 71:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:677
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
	case 72:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:692
		{
			yyVAL.expr = nil
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:699
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 75:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:703
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 76:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:708
		{
			yyVAL.exprs = nil
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:712
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 79:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:719
		{
			yyVAL.expr = &LambdaExpr{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[2].exprs,
					Body:     []Expr{yyDollar[4].expr},
				},
			}
		}
	case 80:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:728
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 81:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:729
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:730
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:731
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:732
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 85:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:733
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:734
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:735
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:736
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:737
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:738
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:739
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 92:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:740
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 93:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:741
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 94:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:742
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:743
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:744
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 97:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:745
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 98:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:747
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 99:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:755
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 100:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:767
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:771
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 102:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:776
		{
			yyVAL.expr = nil
		}
	case 104:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:782
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 105:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:786
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 106:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:795
		{
			yyVAL.pos = Position{}
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:801
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 109:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:811
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:815
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 111:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:821
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 112:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:825
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 114:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:832
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
	case 115:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:848
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 116:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:860
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 117:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:864
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:870
		{
			yyVAL.expr = &Ident{NamePos: yyDollar[1].pos, Name: yyDollar[1].tok}
		}
	case 119:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:876
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 120:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:882
		{
			yyVAL.expr = &ForClause{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				In:   yyDollar[3].pos,
				X:    yyDollar[4].expr,
			}
		}
	case 121:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:892
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:895
		{
			yyVAL.exprs = append(yyDollar[1].exprs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	case 123:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:904
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 124:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:907
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
		}
	}
	goto yystack /* stack new state and value */
}
