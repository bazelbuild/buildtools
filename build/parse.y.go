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

//line build/parse.y:859

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

const yyLast = 596

var yyAct = [...]int{

	17, 151, 188, 2, 7, 142, 149, 114, 123, 75,
	35, 127, 19, 9, 126, 23, 112, 83, 205, 197,
	199, 138, 68, 69, 36, 32, 31, 73, 78, 81,
	105, 42, 43, 168, 29, 111, 144, 34, 129, 86,
	89, 196, 86, 134, 198, 93, 94, 95, 96, 97,
	98, 99, 100, 101, 102, 103, 104, 29, 106, 107,
	108, 109, 191, 13, 115, 85, 72, 145, 171, 91,
	30, 115, 166, 133, 129, 208, 201, 125, 40, 116,
	129, 200, 61, 131, 67, 92, 116, 124, 178, 158,
	132, 192, 39, 162, 77, 80, 71, 139, 37, 147,
	143, 87, 88, 63, 38, 90, 152, 121, 177, 62,
	39, 28, 39, 157, 181, 64, 39, 154, 159, 160,
	148, 161, 156, 26, 45, 27, 39, 44, 119, 175,
	110, 170, 46, 39, 29, 136, 172, 174, 167, 130,
	169, 24, 36, 122, 167, 165, 173, 155, 31, 146,
	176, 137, 84, 70, 185, 182, 82, 184, 115, 187,
	179, 180, 1, 189, 45, 186, 183, 44, 47, 190,
	48, 25, 46, 116, 79, 76, 33, 41, 16, 194,
	12, 8, 4, 193, 128, 163, 164, 65, 195, 28,
	66, 74, 202, 143, 22, 120, 140, 141, 113, 203,
	204, 26, 189, 27, 206, 0, 7, 6, 0, 0,
	11, 0, 29, 18, 0, 0, 0, 28, 20, 24,
	0, 0, 22, 21, 0, 15, 31, 10, 14, 26,
	207, 27, 5, 0, 0, 6, 3, 0, 11, 0,
	29, 18, 0, 0, 0, 28, 20, 24, 0, 0,
	22, 21, 0, 15, 31, 10, 14, 26, 0, 27,
	5, 0, 0, 0, 0, 0, 0, 0, 29, 0,
	0, 0, 0, 0, 20, 24, 0, 0, 0, 21,
	0, 15, 31, 45, 14, 0, 44, 47, 150, 48,
	0, 46, 135, 49, 0, 50, 0, 0, 0, 0,
	0, 58, 0, 0, 51, 0, 54, 0, 60, 0,
	0, 55, 59, 0, 0, 52, 53, 45, 56, 57,
	44, 47, 0, 48, 0, 46, 0, 49, 0, 50,
	0, 0, 0, 0, 0, 58, 0, 0, 51, 0,
	54, 0, 60, 153, 0, 55, 59, 0, 0, 52,
	53, 45, 56, 57, 44, 47, 0, 48, 0, 46,
	0, 49, 0, 50, 0, 0, 0, 0, 0, 58,
	0, 0, 51, 129, 54, 0, 60, 0, 0, 55,
	59, 0, 0, 52, 53, 45, 56, 57, 44, 47,
	0, 48, 0, 46, 0, 49, 0, 50, 0, 0,
	0, 0, 0, 58, 0, 0, 51, 0, 54, 0,
	60, 0, 0, 55, 59, 0, 0, 52, 53, 45,
	56, 57, 44, 47, 0, 48, 0, 46, 0, 49,
	0, 50, 0, 0, 0, 0, 0, 58, 0, 0,
	51, 0, 54, 0, 0, 0, 0, 55, 59, 0,
	0, 52, 53, 45, 56, 57, 44, 47, 0, 48,
	0, 46, 0, 49, 0, 50, 0, 0, 28, 0,
	0, 58, 0, 22, 51, 0, 54, 0, 0, 0,
	26, 55, 27, 0, 0, 52, 53, 0, 56, 57,
	0, 29, 0, 0, 28, 0, 117, 20, 24, 22,
	0, 0, 21, 0, 15, 31, 26, 14, 27, 0,
	0, 0, 0, 0, 0, 0, 0, 29, 0, 0,
	0, 0, 0, 20, 24, 0, 45, 118, 21, 44,
	47, 31, 48, 0, 46, 0, 49, 0, 50, 0,
	0, 0, 0, 0, 58, 0, 0, 51, 0, 54,
	0, 0, 0, 0, 55, 0, 0, 28, 52, 53,
	45, 56, 22, 44, 47, 0, 48, 0, 46, 26,
	49, 27, 50, 0, 0, 0, 0, 0, 0, 0,
	29, 51, 0, 54, 0, 0, 20, 24, 55, 0,
	0, 21, 52, 53, 31, 56,
}
var yyPact = [...]int{

	-1000, -1000, 212, -1000, -1000, -1000, -23, -1000, -1000, -1000,
	9, 106, -1000, 83, 552, -1000, 1, 381, 552, 98,
	552, 552, 552, -1000, 148, -16, 552, 552, 552, -1000,
	-1000, -1000, -1000, -34, 147, 33, 98, 552, 552, 552,
	124, 552, 56, -1000, 552, 552, 552, 552, 552, 552,
	552, 552, 552, 552, 552, 552, -2, 552, 552, 552,
	552, 117, 7, 489, 552, 94, 134, 124, -1000, -1000,
	489, -1000, 59, 347, 130, 12, 53, 23, 279, 126,
	145, 381, -27, 463, 29, 552, 106, 124, 124, 415,
	107, 240, -1000, -1000, -1000, -1000, 120, 120, 160, 160,
	160, 160, 160, 160, 160, 552, 522, 556, 449, 313,
	240, -1000, 141, 104, -1000, 381, 74, 552, 552, 103,
	80, 552, 552, -1000, 139, -1000, 54, 4, -1000, 106,
	552, -1000, 48, -1000, -1000, 552, 552, -1000, -1000, -1000,
	123, 99, -1000, 73, 6, 6, 101, 98, 240, -1000,
	-1000, -1000, 160, 552, -1000, -1000, -1000, 489, 552, 381,
	381, -1000, 552, -1000, -1000, -1000, -1000, 4, 552, 30,
	381, -1000, 381, -1000, 279, 78, -1000, 29, 552, -1000,
	-1000, 240, -1000, -4, -28, 415, -1000, 381, 63, 381,
	415, 552, 240, -1000, 381, -1000, -1000, -30, -1000, -1000,
	-1000, 552, 415, -1000, 184, -1000, 57, -1000, -1000,
}
var yyPgo = [...]int{

	0, 8, 7, 198, 16, 5, 197, 196, 0, 2,
	66, 12, 63, 195, 191, 190, 187, 10, 184, 11,
	14, 15, 3, 182, 181, 180, 178, 177, 1, 13,
	176, 9, 175, 174, 70, 171, 6, 166, 162, 157,
	156,
}
var yyR1 = [...]int{

	0, 38, 36, 36, 39, 39, 37, 37, 37, 22,
	22, 22, 22, 23, 23, 24, 24, 24, 26, 26,
	25, 25, 27, 27, 28, 30, 30, 29, 29, 29,
	29, 29, 29, 40, 40, 11, 11, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 11, 4,
	4, 3, 3, 2, 2, 2, 2, 7, 7, 6,
	6, 5, 5, 5, 5, 12, 12, 13, 13, 15,
	15, 16, 16, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 14, 14, 9, 9, 10, 10,
	1, 1, 31, 33, 33, 32, 32, 17, 17, 34,
	35, 35, 21, 18, 19, 19, 20, 20,
}
var yyR2 = [...]int{

	0, 2, 5, 2, 0, 2, 0, 3, 2, 0,
	2, 2, 3, 1, 1, 7, 6, 1, 4, 5,
	1, 4, 2, 1, 4, 0, 3, 1, 2, 1,
	3, 3, 1, 0, 1, 1, 3, 4, 4, 4,
	6, 8, 1, 3, 4, 4, 3, 3, 3, 0,
	2, 1, 3, 1, 3, 2, 2, 0, 2, 1,
	3, 1, 3, 2, 2, 1, 3, 0, 1, 1,
	3, 0, 2, 1, 4, 2, 2, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 4,
	3, 3, 3, 5, 1, 3, 0, 1, 0, 2,
	0, 1, 3, 1, 3, 1, 2, 1, 3, 1,
	1, 2, 1, 4, 1, 3, 1, 2,
}
var yyChk = [...]int{

	-1000, -38, -22, 24, -23, 48, 23, -28, -24, -29,
	43, 26, -25, -12, 44, 41, -26, -8, 29, -11,
	34, 39, 10, -21, 35, -35, 17, 19, 5, 28,
	-34, 42, 48, -30, 28, -17, -11, 15, 21, 9,
	-12, -27, 30, 31, 7, 4, 12, 8, 10, 14,
	16, 25, 36, 37, 27, 32, 39, 40, 22, 33,
	29, -12, 11, 5, 17, -16, -15, -12, -8, -8,
	5, -34, -10, -8, -14, -31, -32, -10, -8, -33,
	-10, -8, -40, 51, 5, 32, 9, -12, -12, -8,
	-12, 13, 29, -8, -8, -8, -8, -8, -8, -8,
	-8, -8, -8, -8, -8, 32, -8, -8, -8, -8,
	13, 28, -4, -3, -2, -8, -21, 7, 38, -12,
	-13, 13, 9, -1, -4, 18, -20, -19, -18, 26,
	9, -1, -20, 20, 20, 13, 9, 6, 48, -29,
	-7, -6, -5, -21, 7, 38, -12, -11, 13, -36,
	48, -28, -8, 30, -36, 6, -1, 9, 15, -8,
	-8, 18, 13, -12, -12, 6, 18, -19, 29, -17,
	-8, 20, -8, -31, -8, 6, -1, 9, 15, -21,
	-21, 13, -36, -37, -39, -8, -2, -8, -9, -8,
	-8, 32, 13, -5, -8, -36, 45, 23, 48, 48,
	18, 13, -8, -36, -22, 48, -9, 46, 18,
}
var yyDef = [...]int{

	9, -2, 0, 1, 10, 11, 0, 13, 14, 25,
	0, 0, 17, 27, 29, 32, 20, 65, 0, 73,
	71, 0, 0, 35, 0, 42, 98, 98, 98, 112,
	110, 109, 12, 33, 0, 0, 107, 0, 0, 0,
	28, 0, 0, 23, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 49, 67, 0, 100, 69, 75, 76,
	49, 111, 0, 94, 100, 103, 0, 0, 94, 105,
	0, 94, 0, 34, 57, 0, 0, 30, 31, 66,
	0, 0, 22, 77, 78, 79, 80, 81, 82, 83,
	84, 85, 86, 87, 88, 0, 90, 91, 92, 0,
	0, 36, 0, 100, 51, 53, 35, 0, 0, 68,
	0, 0, 101, 72, 0, 43, 0, 116, 114, 0,
	101, 99, 0, 46, 47, 0, 106, 48, 24, 26,
	0, 100, 59, 61, 0, 0, 0, 108, 0, 21,
	6, 4, 89, 0, 18, 38, 50, 101, 0, 55,
	56, 39, 96, 74, 70, 37, 44, 117, 0, 0,
	95, 45, 102, 104, 0, 0, 58, 101, 0, 63,
	64, 0, 19, 0, 3, 93, 52, 54, 0, 97,
	115, 0, 0, 60, 62, 16, 9, 0, 8, 5,
	40, 96, 113, 15, 0, 7, 0, 2, 41,
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
		//line build/parse.y:176
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:183
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
		//line build/parse.y:202
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:210
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:215
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
		//line build/parse.y:227
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 9:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:233
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:238
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
		//line build/parse.y:269
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:275
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
		//line build/parse.y:289
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:293
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 15:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:299
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
		//line build/parse.y:312
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
		//line build/parse.y:321
		{
			yyVAL.expr = yyDollar[1].ifstmt
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:328
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].exprs,
			}
		}
	case 19:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:336
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = yyDollar[2].pos
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
		//line build/parse.y:356
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = yyDollar[2].pos
			inner.False = yyDollar[4].exprs
		}
	case 24:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:372
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 25:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:378
		{
			yyVAL.exprs = []Expr{}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:382
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:389
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:396
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:401
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:402
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:404
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:414
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
		//line build/parse.y:423
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
		//line build/parse.y:434
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
		//line build/parse.y:445
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
		//line build/parse.y:454
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
		//line build/parse.y:465
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
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:490
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 44:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:499
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
	case 45:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:511
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
	case 46:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:523
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 47:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:532
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 48:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:541
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
	case 49:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:562
		{
			yyVAL.exprs = nil
		}
	case 50:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:566
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 51:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:572
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 52:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:576
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:583
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:587
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 56:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:591
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 57:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:596
		{
			yyVAL.exprs = nil
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:600
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 59:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:606
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 60:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:610
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 62:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:617
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 63:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:621
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 64:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:625
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:632
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
	case 67:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:647
		{
			yyVAL.expr = nil
		}
	case 69:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:654
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 70:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:658
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 71:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:663
		{
			yyVAL.exprs = nil
		}
	case 72:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:667
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 74:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:674
		{
			yyVAL.expr = &LambdaExpr{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[2].exprs,
					Body:     []Expr{yyDollar[4].expr},
				},
			}
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:683
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 76:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:684
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:685
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:686
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:687
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:688
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:689
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:690
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:691
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:692
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 85:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:693
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:694
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:695
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:696
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 89:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:697
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:698
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:699
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 92:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:701
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 93:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:709
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 94:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:721
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:725
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 96:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:730
		{
			yyVAL.expr = nil
		}
	case 98:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:736
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 99:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:740
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 100:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:749
		{
			yyVAL.pos = Position{}
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:755
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 103:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:765
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:769
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 105:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:775
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 106:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:779
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:786
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
	case 109:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:802
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 110:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:814
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 111:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:818
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 112:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:824
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 113:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:830
		{
			yyVAL.expr = &ForClause{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				In:   yyDollar[3].pos,
				X:    yyDollar[4].expr,
			}
		}
	case 114:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:840
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 115:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:843
		{
			yyVAL.exprs = append(yyDollar[1].exprs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	case 116:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:852
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 117:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:855
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
		}
	}
	goto yystack /* stack new state and value */
}
