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
	"'|'",
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

//line build/parse.y:901

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

const yyLast = 616

var yyAct = [...]int{

	16, 192, 22, 152, 29, 189, 7, 2, 143, 124,
	34, 128, 75, 115, 9, 127, 83, 150, 213, 202,
	204, 68, 69, 18, 139, 31, 73, 78, 81, 71,
	30, 105, 41, 42, 169, 35, 28, 112, 33, 89,
	201, 145, 130, 203, 93, 94, 95, 96, 97, 98,
	99, 100, 101, 102, 103, 104, 28, 106, 107, 108,
	109, 110, 27, 28, 116, 27, 117, 21, 58, 30,
	13, 86, 167, 146, 25, 125, 26, 25, 135, 26,
	86, 130, 219, 91, 132, 39, 28, 144, 61, 28,
	67, 133, 19, 23, 134, 196, 23, 20, 140, 30,
	92, 14, 30, 126, 85, 151, 153, 87, 88, 44,
	148, 90, 43, 197, 209, 72, 27, 45, 172, 160,
	161, 21, 206, 179, 157, 130, 58, 205, 25, 155,
	26, 159, 171, 163, 122, 120, 208, 173, 175, 168,
	28, 170, 77, 80, 178, 168, 19, 23, 180, 181,
	174, 20, 177, 30, 35, 186, 147, 38, 166, 116,
	188, 117, 38, 158, 190, 38, 162, 183, 36, 194,
	195, 193, 187, 137, 38, 37, 131, 63, 182, 123,
	199, 144, 44, 62, 84, 43, 46, 198, 47, 64,
	45, 38, 38, 164, 165, 149, 111, 210, 215, 58,
	200, 207, 176, 156, 138, 70, 82, 190, 185, 212,
	216, 194, 214, 193, 217, 211, 7, 27, 1, 191,
	184, 24, 21, 79, 76, 32, 40, 15, 12, 25,
	8, 26, 4, 129, 65, 66, 6, 74, 121, 11,
	141, 28, 17, 142, 113, 27, 114, 19, 23, 0,
	21, 0, 20, 0, 30, 10, 14, 25, 218, 26,
	5, 0, 0, 0, 6, 3, 0, 11, 0, 28,
	17, 0, 0, 0, 0, 19, 23, 0, 0, 0,
	20, 0, 30, 10, 14, 0, 44, 0, 5, 43,
	46, 0, 47, 0, 45, 136, 48, 0, 49, 0,
	0, 0, 0, 58, 0, 57, 0, 0, 50, 0,
	53, 0, 60, 0, 0, 54, 59, 0, 0, 51,
	52, 44, 55, 56, 43, 46, 0, 47, 0, 45,
	0, 48, 0, 49, 0, 0, 0, 0, 58, 0,
	57, 0, 0, 50, 0, 53, 0, 60, 154, 0,
	54, 59, 0, 0, 51, 52, 44, 55, 56, 43,
	46, 0, 47, 0, 45, 0, 48, 0, 49, 0,
	0, 0, 0, 58, 0, 57, 0, 0, 50, 130,
	53, 0, 60, 0, 0, 54, 59, 0, 0, 51,
	52, 44, 55, 56, 43, 46, 0, 47, 0, 45,
	0, 48, 0, 49, 0, 0, 0, 0, 58, 0,
	57, 0, 0, 50, 0, 53, 0, 60, 0, 0,
	54, 59, 0, 0, 51, 52, 44, 55, 56, 43,
	46, 0, 47, 0, 45, 0, 48, 0, 49, 0,
	0, 0, 0, 58, 0, 57, 0, 0, 50, 0,
	53, 0, 0, 0, 0, 54, 59, 0, 0, 51,
	52, 44, 55, 56, 43, 46, 0, 47, 0, 45,
	0, 48, 0, 49, 0, 0, 0, 0, 58, 0,
	57, 0, 0, 50, 0, 53, 0, 0, 0, 0,
	54, 0, 0, 0, 51, 52, 44, 55, 56, 43,
	46, 0, 47, 0, 45, 0, 48, 0, 49, 0,
	0, 0, 0, 58, 0, 57, 0, 0, 50, 0,
	53, 0, 0, 0, 0, 54, 0, 0, 0, 51,
	52, 44, 55, 0, 43, 46, 0, 47, 0, 45,
	0, 48, 27, 49, 0, 0, 0, 21, 58, 0,
	0, 0, 0, 50, 25, 53, 26, 0, 0, 0,
	54, 0, 0, 0, 51, 52, 28, 55, 0, 0,
	0, 0, 19, 23, 0, 0, 0, 20, 27, 30,
	118, 14, 0, 21, 0, 0, 0, 0, 0, 0,
	25, 0, 26, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 28, 0, 0, 0, 0, 0, 19, 23,
	0, 0, 119, 20, 0, 30,
}
var yyPact = [...]int{

	-1000, -1000, 240, -1000, -1000, -1000, -23, -1000, -1000, -1000,
	9, 60, -1000, 153, 111, 1, 387, 111, 172, 111,
	111, 111, -1000, 200, -12, 111, 111, 111, -1000, -1000,
	-1000, -1000, -35, 179, 71, 172, 111, 111, 111, 156,
	111, 70, -1000, 111, 111, 111, 111, 111, 111, 111,
	111, 111, 111, 111, 111, -2, 111, 111, 111, 111,
	111, 183, 8, 573, 111, 121, 170, 156, 47, 47,
	-12, -1000, 85, 352, 167, 15, 74, 58, 282, 164,
	198, 387, -24, 537, 34, 111, 60, 156, 156, 422,
	182, 57, -1000, 47, 47, 47, 105, 105, 178, 178,
	178, 178, 178, 178, 178, 111, 492, 527, 387, 457,
	317, 57, -1000, 197, 154, -1000, 387, 116, 111, 111,
	148, 120, 111, 111, -1000, 149, -1000, 54, 4, -1000,
	60, 111, -1000, 98, -1000, -1000, 111, 111, -1000, -1000,
	-1000, 196, 135, -1000, 108, 7, 7, 165, 172, 57,
	-1000, -1000, -1000, 178, 111, -1000, -1000, -1000, 573, 111,
	387, 387, -1000, 111, -1000, -1000, 27, -1000, 4, 111,
	62, 387, -1000, 387, -1000, 282, 100, -1000, 34, 111,
	-1000, -1000, 57, -1000, -5, -28, 422, -1000, 387, 109,
	387, 127, -1000, -1000, 99, 422, 111, 57, -1000, 387,
	-1000, -1000, -30, -1000, -1000, -1000, 111, 192, 27, -12,
	422, -1000, 212, -1000, 64, -1000, -1000, -1000, -1000, -1000,
}
var yyPgo = [...]int{

	0, 9, 13, 246, 244, 8, 243, 240, 0, 5,
	115, 23, 70, 238, 237, 235, 234, 10, 233, 11,
	15, 2, 7, 232, 230, 228, 227, 226, 3, 14,
	225, 12, 224, 223, 4, 221, 17, 220, 1, 219,
	218, 208, 206,
}
var yyR1 = [...]int{

	0, 40, 36, 36, 41, 41, 37, 37, 37, 22,
	22, 22, 22, 23, 23, 24, 24, 24, 26, 26,
	25, 25, 27, 27, 28, 30, 30, 29, 29, 29,
	29, 29, 42, 42, 11, 11, 11, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 4, 4,
	3, 3, 2, 2, 2, 2, 39, 39, 38, 38,
	7, 7, 6, 6, 5, 5, 5, 5, 12, 12,
	13, 13, 15, 15, 16, 16, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 14, 14,
	9, 9, 10, 10, 1, 1, 31, 33, 33, 32,
	32, 17, 17, 34, 35, 35, 21, 18, 19, 19,
	20, 20,
}
var yyR2 = [...]int{

	0, 2, 5, 2, 0, 2, 0, 3, 2, 0,
	2, 2, 3, 1, 1, 7, 6, 1, 4, 5,
	1, 4, 2, 1, 4, 0, 3, 1, 2, 1,
	3, 3, 0, 1, 1, 3, 7, 4, 4, 6,
	8, 1, 3, 4, 4, 3, 3, 3, 0, 2,
	1, 3, 1, 3, 2, 2, 1, 3, 1, 3,
	0, 2, 1, 3, 1, 3, 2, 2, 1, 3,
	0, 1, 1, 3, 0, 2, 1, 4, 2, 2,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 4, 3, 3, 3, 3, 5, 1, 3,
	0, 1, 0, 2, 0, 1, 3, 1, 3, 1,
	2, 1, 3, 1, 1, 2, 1, 4, 1, 3,
	1, 2,
}
var yyChk = [...]int{

	-1000, -40, -22, 25, -23, 48, 24, -28, -24, -29,
	43, 27, -25, -12, 44, -26, -8, 30, -11, 35,
	40, 10, -21, 36, -35, 17, 19, 5, 29, -34,
	42, 48, -30, 29, -17, -11, 15, 22, 9, -12,
	-27, 31, 32, 7, 4, 12, 8, 10, 14, 16,
	26, 37, 38, 28, 33, 40, 41, 23, 21, 34,
	30, -12, 11, 5, 17, -16, -15, -12, -8, -8,
	5, -34, -10, -8, -14, -31, -32, -10, -8, -33,
	-10, -8, -42, 51, 5, 33, 9, -12, -12, -8,
	-12, 13, 30, -8, -8, -8, -8, -8, -8, -8,
	-8, -8, -8, -8, -8, 33, -8, -8, -8, -8,
	-8, 13, 29, -4, -3, -2, -8, -21, 7, 39,
	-12, -13, 13, 9, -1, -34, 18, -20, -19, -18,
	27, 9, -1, -20, 20, 20, 13, 9, 6, 48,
	-29, -7, -6, -5, -21, 7, 39, -12, -11, 13,
	-36, 48, -28, -8, 31, -36, 6, -1, 9, 15,
	-8, -8, 18, 13, -12, -12, 9, 18, -19, 30,
	-17, -8, 20, -8, -31, -8, 6, -1, 9, 15,
	-21, -21, 13, -36, -37, -41, -8, -2, -8, -9,
	-8, -39, -38, -34, -21, -8, 33, 13, -5, -8,
	-36, 45, 24, 48, 48, 18, 13, -1, 9, 15,
	-8, -36, -22, 48, -9, 6, -38, -34, 46, 18,
}
var yyDef = [...]int{

	9, -2, 0, 1, 10, 11, 0, 13, 14, 25,
	0, 0, 17, 27, 29, 20, 68, 0, 76, 74,
	0, 0, 34, 0, 41, 102, 102, 102, 116, 114,
	113, 12, 32, 0, 0, 111, 0, 0, 0, 28,
	0, 0, 23, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 48, 70, 0, 104, 72, 78, 79,
	0, 115, 0, 98, 104, 107, 0, 0, 98, 109,
	0, 98, 0, 33, 60, 0, 0, 30, 31, 69,
	0, 0, 22, 80, 81, 82, 83, 84, 85, 86,
	87, 88, 89, 90, 91, 0, 93, 94, 95, 96,
	0, 0, 35, 0, 104, 50, 52, 34, 0, 0,
	71, 0, 0, 105, 75, 0, 42, 0, 120, 118,
	0, 105, 103, 0, 45, 46, 0, 110, 47, 24,
	26, 0, 104, 62, 64, 0, 0, 0, 112, 0,
	21, 6, 4, 92, 0, 18, 37, 49, 105, 0,
	54, 55, 38, 100, 77, 73, 0, 43, 121, 0,
	0, 99, 44, 106, 108, 0, 0, 61, 105, 0,
	66, 67, 0, 19, 0, 3, 97, 51, 53, 0,
	101, 104, 56, 58, 0, 119, 0, 0, 63, 65,
	16, 9, 0, 8, 5, 39, 100, 0, 105, 0,
	117, 15, 0, 7, 0, 36, 57, 59, 2, 40,
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
	3, 3, 3, 19, 21, 20,
}
var yyTok2 = [...]int{

	2, 3, 22, 23, 24, 25, 26, 27, 28, 29,
	30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
	40, 41, 42, 43, 44, 45, 46, 47, 49, 50,
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
		//line build/parse.y:180
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:187
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
		//line build/parse.y:206
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:214
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:219
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
		//line build/parse.y:231
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 9:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:237
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:242
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
		//line build/parse.y:273
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:279
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
		//line build/parse.y:293
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:297
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 15:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:303
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
		//line build/parse.y:316
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
		//line build/parse.y:325
		{
			yyVAL.expr = yyDollar[1].ifstmt
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:332
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].exprs,
			}
		}
	case 19:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:340
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
		//line build/parse.y:360
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
		//line build/parse.y:376
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 25:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:382
		{
			yyVAL.exprs = []Expr{}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:386
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:393
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:400
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:405
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:406
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 35:
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
	case 36:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:423
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
	case 37:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:437
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
	case 38:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:448
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 39:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:457
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
	case 40:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:468
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
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:481
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
	case 42:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:493
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 43:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:502
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
	case 44:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:514
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
	case 45:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:526
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 46:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:535
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 47:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:544
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
	case 48:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:565
		{
			yyVAL.exprs = nil
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:569
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 50:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:575
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 51:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:579
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 53:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:586
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 54:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:590
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:594
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 56:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:599
		{
			yyVAL.loadargs = []*struct {
				from Ident
				to   Ident
			}{yyDollar[1].loadarg}
		}
	case 57:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:603
		{
			yyDollar[1].loadargs = append(yyDollar[1].loadargs, yyDollar[3].loadarg)
			yyVAL.loadargs = yyDollar[1].loadargs
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:609
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
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:622
		{
			ident := yyDollar[1].expr.(*LiteralExpr)
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[3].string.Value,
					NamePos: yyDollar[3].string.Start,
				},
				to: Ident{
					Name:    ident.Token,
					NamePos: ident.Start,
				},
			}
		}
	case 60:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:637
		{
			yyVAL.exprs = nil
		}
	case 61:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:641
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:647
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 63:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:651
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 65:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:658
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 66:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:662
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 67:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:666
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:673
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
	case 70:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:688
		{
			yyVAL.expr = nil
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:695
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 73:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:699
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 74:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:704
		{
			yyVAL.exprs = nil
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:708
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 77:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:715
		{
			yyVAL.expr = &LambdaExpr{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[2].exprs,
					Body:     []Expr{yyDollar[4].expr},
				},
			}
		}
	case 78:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:724
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 79:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:725
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:726
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:727
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:728
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:729
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:730
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 85:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:731
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:732
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:733
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:734
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:735
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:736
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:737
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 92:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:738
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 93:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:739
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 94:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:740
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:741
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:743
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 97:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:751
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:763
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:767
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 100:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:772
		{
			yyVAL.expr = nil
		}
	case 102:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:778
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 103:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:782
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 104:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:791
		{
			yyVAL.pos = Position{}
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:797
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:807
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:811
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 109:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:817
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 110:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:821
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 112:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:828
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
	case 113:
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
	case 114:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:856
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:860
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 116:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:866
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 117:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:872
		{
			yyVAL.expr = &ForClause{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				In:   yyDollar[3].pos,
				X:    yyDollar[4].expr,
			}
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:882
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:885
		{
			yyVAL.exprs = append(yyDollar[1].exprs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	case 120:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:894
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 121:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:897
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
		}
	}
	goto yystack /* stack new state and value */
}
