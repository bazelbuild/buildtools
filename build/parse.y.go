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

//line build/parse.y:876

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
	26, 64,
	-2, 52,
}

const yyPrivate = 57344

const yyLast = 598

var yyAct = [...]int{

	17, 195, 154, 2, 156, 147, 41, 7, 131, 118,
	23, 126, 79, 130, 19, 35, 9, 114, 85, 143,
	32, 190, 70, 71, 31, 107, 36, 75, 77, 82,
	44, 45, 197, 29, 113, 34, 172, 133, 149, 178,
	91, 88, 88, 74, 133, 133, 140, 95, 96, 97,
	98, 99, 100, 101, 102, 103, 104, 105, 106, 29,
	108, 109, 110, 111, 198, 87, 117, 139, 93, 150,
	30, 78, 81, 128, 192, 27, 119, 206, 39, 177,
	22, 212, 205, 119, 94, 129, 185, 26, 135, 28,
	127, 136, 164, 138, 133, 133, 73, 148, 29, 133,
	199, 39, 144, 152, 20, 24, 39, 37, 157, 21,
	188, 15, 31, 38, 14, 159, 39, 39, 155, 65,
	153, 165, 166, 168, 47, 64, 167, 46, 162, 161,
	47, 66, 48, 46, 49, 176, 50, 39, 48, 173,
	124, 112, 179, 181, 184, 173, 163, 173, 36, 175,
	39, 142, 134, 125, 182, 180, 189, 171, 183, 191,
	186, 187, 13, 160, 128, 194, 137, 86, 72, 196,
	173, 84, 1, 193, 119, 25, 83, 40, 80, 33,
	43, 63, 42, 69, 16, 12, 201, 8, 4, 174,
	200, 202, 132, 67, 204, 148, 203, 68, 207, 208,
	89, 90, 209, 27, 76, 123, 92, 196, 211, 7,
	145, 146, 116, 0, 0, 26, 0, 28, 27, 0,
	0, 0, 0, 22, 0, 0, 29, 0, 115, 122,
	26, 0, 28, 24, 0, 0, 6, 0, 0, 11,
	31, 29, 18, 0, 0, 0, 27, 20, 24, 0,
	151, 22, 21, 0, 15, 31, 10, 14, 26, 210,
	28, 5, 0, 0, 6, 3, 0, 11, 0, 29,
	18, 0, 0, 0, 0, 20, 24, 0, 0, 0,
	21, 0, 15, 31, 10, 14, 0, 169, 170, 5,
	47, 0, 0, 46, 49, 0, 50, 0, 48, 141,
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
	0, 27, 0, 0, 57, 61, 22, 0, 54, 55,
	0, 58, 59, 26, 0, 28, 0, 0, 0, 0,
	27, 0, 120, 0, 29, 22, 0, 0, 0, 0,
	20, 24, 26, 0, 28, 21, 0, 15, 31, 0,
	14, 0, 0, 29, 0, 0, 0, 0, 0, 20,
	24, 0, 47, 121, 21, 46, 49, 31, 50, 0,
	48, 0, 51, 0, 52, 0, 27, 0, 0, 0,
	60, 22, 0, 53, 0, 56, 0, 0, 26, 0,
	28, 0, 0, 0, 54, 55, 0, 0, 59, 29,
	0, 0, 0, 0, 0, 20, 24, 0, 47, 0,
	21, 46, 49, 31, 50, 0, 48, 0, 51, 0,
	52, 0, 0, 0, 47, 0, 60, 46, 49, 53,
	50, 56, 48, 0, 51, 0, 52, 0, 0, 0,
	54, 55, 0, 0, 0, 53, 0, 56, 0, 0,
	0, 0, 0, 0, 0, 0, 54, 55,
}
var yyPact = [...]int{

	-1000, -1000, 241, -1000, -1000, -1000, -28, -1000, -1000, -1000,
	7, 198, -1000, 92, 511, -1000, 0, 388, 511, 114,
	511, 511, 511, -1000, 163, -18, 511, 511, 511, -1000,
	-1000, -1000, -1000, -33, 162, 33, 114, 511, 511, 511,
	141, -1000, -1000, 511, 55, -1000, 511, 511, 511, 511,
	511, 511, 511, 511, 511, 511, 511, 511, -7, 511,
	511, 511, 511, 128, 6, 465, 511, 127, 144, 141,
	-1000, -1000, 465, -1000, 67, 354, 143, 354, 160, 11,
	47, 26, 286, 142, -29, 446, 31, 511, 198, 141,
	141, 422, 107, 70, -1000, -1000, -1000, -1000, 120, 120,
	126, 126, 126, 126, 126, 126, 498, 511, 544, 560,
	498, 320, 70, -1000, 157, 69, 137, 388, -1000, 77,
	511, 511, 108, 110, 511, 511, -1000, 151, 388, -1000,
	18, -1000, -1000, 198, 511, -1000, 73, -1000, 19, -1000,
	-1000, 511, 511, -1000, -1000, 148, 135, -1000, 71, 5,
	5, 97, 114, 70, -1000, -24, -1000, 498, 511, -1000,
	-1000, 68, -1000, 465, 511, 388, 388, -1000, 511, -1000,
	-1000, -1000, -1000, -1000, 3, 32, 388, -1000, -1000, 388,
	-1000, 286, 87, -1000, 31, 511, -1000, -1000, 70, 0,
	-1000, 422, -1000, -1000, 388, 64, 388, 511, 511, 70,
	-1000, 388, -1000, -1000, 213, -1000, 511, 422, 422, -1000,
	-1000, 63, -1000,
}
var yyPgo = [...]int{

	0, 11, 9, 212, 17, 5, 211, 210, 0, 1,
	43, 14, 162, 205, 204, 197, 193, 15, 192, 8,
	13, 10, 189, 3, 188, 187, 185, 184, 182, 6,
	180, 4, 16, 179, 12, 178, 176, 70, 175, 2,
	172, 171,
}
var yyR1 = [...]int{

	0, 40, 39, 39, 23, 23, 23, 23, 24, 24,
	25, 25, 25, 28, 29, 29, 27, 26, 26, 30,
	30, 31, 33, 33, 32, 32, 32, 32, 32, 32,
	41, 41, 11, 11, 11, 11, 11, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 4, 4,
	3, 3, 2, 2, 2, 2, 7, 7, 6, 6,
	5, 5, 5, 5, 12, 12, 13, 13, 15, 15,
	16, 16, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 14, 14, 9, 9, 10, 10, 1,
	1, 34, 36, 36, 35, 35, 17, 17, 37, 38,
	38, 21, 18, 19, 20, 20, 22, 22,
}
var yyR2 = [...]int{

	0, 2, 4, 1, 0, 2, 2, 3, 1, 1,
	7, 6, 1, 3, 1, 5, 4, 1, 2, 2,
	1, 4, 0, 3, 1, 2, 1, 3, 3, 1,
	0, 1, 1, 3, 4, 4, 4, 6, 8, 5,
	1, 3, 4, 4, 4, 3, 3, 3, 0, 2,
	1, 3, 1, 3, 2, 2, 0, 2, 1, 3,
	1, 3, 2, 2, 1, 3, 0, 1, 1, 3,
	0, 2, 1, 4, 2, 2, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 4, 3,
	3, 3, 5, 1, 3, 0, 1, 0, 2, 0,
	1, 3, 1, 3, 1, 2, 1, 3, 1, 1,
	2, 1, 4, 2, 1, 2, 0, 3,
}
var yyChk = [...]int{

	-1000, -40, -23, 24, -24, 48, 23, -31, -25, -32,
	43, 26, -26, -12, 44, 41, -27, -8, 29, -11,
	34, 39, 10, -21, 35, -38, 17, 5, 19, 28,
	-37, 42, 48, -33, 28, -17, -11, 15, 21, 9,
	-12, -29, -28, -30, 30, 31, 7, 4, 12, 8,
	10, 14, 16, 25, 36, 37, 27, 32, 39, 40,
	22, 33, 29, -12, 11, 5, 17, -16, -15, -12,
	-8, -8, 5, -37, -10, -8, -14, -8, -10, -34,
	-35, -10, -8, -36, -41, 51, 5, 32, 9, -12,
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
	45, -8, 6, -2, -8, -9, -8, 29, 32, 13,
	-5, -8, -39, -29, -23, 18, 13, -8, -8, -39,
	46, -9, 18,
}
var yyDef = [...]int{

	4, -2, 0, 1, 5, 6, 0, 8, 9, 22,
	0, 0, 12, 24, 26, 29, 17, 64, 0, 72,
	70, 0, 0, 32, 0, 40, 97, 97, 97, 111,
	109, 108, 7, 30, 0, 0, 106, 0, 0, 0,
	25, 18, 14, 0, 0, 20, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 48, 66, 0, 99, 68,
	74, 75, 48, 110, 0, 93, 99, 93, 0, 102,
	0, 0, 93, 104, 0, 31, 56, 0, 0, 27,
	28, 65, 0, 0, 19, 76, 77, 78, 79, 80,
	81, 82, 83, 84, 85, 86, 87, 0, 89, 90,
	91, 0, 0, 33, 0, 0, 99, -2, 50, 32,
	0, 0, 67, 0, 0, 100, 71, 0, 52, 41,
	0, 114, 116, 0, 100, 98, 0, 47, 0, 45,
	46, 0, 105, 21, 23, 0, 99, 58, 60, 0,
	0, 0, 107, 0, 13, 0, 3, 88, 0, 16,
	35, 0, 49, 100, 0, 54, 55, 36, 95, 73,
	69, 34, 42, 115, 113, 0, 94, 43, 44, 101,
	103, 0, 0, 57, 100, 0, 62, 63, 0, 0,
	4, 92, 39, 51, 53, 0, 96, 0, 0, 0,
	59, 61, 11, 15, 0, 37, 95, 117, 112, 10,
	2, 0, 38,
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
		//line build/parse.y:183
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:190
		{
			yyVAL.block = CodeBlock{
				Start:      yyDollar[2].pos,
				Statements: yyDollar[3].exprs,
				End:        End{Pos: yyDollar[4].pos},
			}
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:198
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
		//line build/parse.y:210
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:215
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
		//line build/parse.y:246
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:252
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
		//line build/parse.y:266
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:270
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 10:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:276
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
		//line build/parse.y:289
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
		//line build/parse.y:299
		{
			yyVAL.expr = yyDollar[1].ifstmt
		}
	case 13:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:306
		{
			yyVAL.ifstmt = &IfStmt{
				ElsePos: yyDollar[1].pos,
				False:   yyDollar[3].block.Statements,
			}
		}
	case 15:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:317
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
		//line build/parse.y:331
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].block.Statements,
			}
		}
	case 18:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:343
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			yyVAL.ifstmt.ElsePos = yyDollar[2].ifstmt.ElsePos
			yyVAL.ifstmt.False = yyDollar[2].ifstmt.False
		}
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:355
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 22:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:361
		{
			yyVAL.exprs = []Expr{}
		}
	case 23:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:365
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 25:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:372
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:379
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 27:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:384
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 28:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:385
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:387
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:397
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 34:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:406
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
	case 35:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:417
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
	case 36:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:428
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 37:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:437
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
	case 38:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:448
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
	case 39:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:461
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
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:478
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
	case 41:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:490
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 42:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:500
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
	case 43:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:512
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
	case 44:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:524
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
	case 45:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:536
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 46:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:546
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 47:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:556
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
	case 48:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:578
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:582
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 50:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:588
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 51:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:592
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 53:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:599
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 54:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:603
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:607
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 56:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:612
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:616
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:622
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:626
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:633
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:637
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 63:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:641
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 65:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:648
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || !tuple.Start.IsValid() {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					Comma:          Position{},
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 66:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:663
		{
			yyVAL.expr = nil
		}
	case 68:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:670
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:674
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 70:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:679
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:683
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 73:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:690
		{
			yyVAL.expr = &LambdaExpr{
				Lambda: yyDollar[1].pos,
				Var:    yyDollar[2].exprs,
				Colon:  yyDollar[3].pos,
				Expr:   yyDollar[4].expr,
			}
		}
	case 74:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:698
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:699
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 76:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:700
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:701
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:702
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:703
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:704
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:705
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:706
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:707
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:708
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 85:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:709
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:710
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:711
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 88:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:712
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:713
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:714
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:716
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 92:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:724
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 93:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:736
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 94:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:740
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 95:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:745
		{
			yyVAL.expr = nil
		}
	case 97:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:751
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 98:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:755
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 99:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:764
		{
			yyVAL.pos = Position{}
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:770
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 102:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:780
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:784
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 104:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:790
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 105:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:794
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 106:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:800
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 107:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:804
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 108:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:810
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 109:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:822
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 110:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:826
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 111:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:832
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 112:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:838
		{
			yyVAL.forc = &ForClause{
				For:  yyDollar[1].pos,
				Var:  yyDollar[2].exprs,
				In:   yyDollar[3].pos,
				Expr: yyDollar[4].expr,
			}
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:848
		{
			yyVAL.forifs = &ForClauseWithIfClausesOpt{
				For: yyDollar[1].forc,
				Ifs: yyDollar[2].ifs,
			}
		}
	case 114:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:857
		{
			yyVAL.forsifs = []*ForClauseWithIfClausesOpt{yyDollar[1].forifs}
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:860
		{
			yyVAL.forsifs = append(yyDollar[1].forsifs, yyDollar[2].forifs)
		}
	case 116:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:865
		{
			yyVAL.ifs = nil
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:869
		{
			yyVAL.ifs = append(yyDollar[1].ifs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	}
	goto yystack /* stack new state and value */
}
