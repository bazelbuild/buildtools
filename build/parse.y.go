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

const _ADDEQ = 57346
const _AND = 57347
const _COMMENT = 57348
const _EOF = 57349
const _EQ = 57350
const _FOR = 57351
const _GE = 57352
const _IDENT = 57353
const _IF = 57354
const _ELSE = 57355
const _IN = 57356
const _IS = 57357
const _LAMBDA = 57358
const _LE = 57359
const _NE = 57360
const _NOT = 57361
const _OR = 57362
const _PYTHON = 57363
const _STRING = 57364
const _DEF = 57365
const _RETURN = 57366
const _INDENT = 57367
const _UNINDENT = 57368
const ShiftInstead = 57369
const _ASSERT = 57370
const _UNARY = 57371

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
	"_ADDEQ",
	"_AND",
	"_COMMENT",
	"_EOF",
	"_EQ",
	"_FOR",
	"_GE",
	"_IDENT",
	"_IF",
	"_ELSE",
	"_IN",
	"_IS",
	"_LAMBDA",
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

//line build/parse.y:672

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

const yyNprod = 84
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 571

var yyAct = [...]int{

	11, 2, 95, 7, 63, 99, 58, 100, 9, 70,
	112, 27, 151, 49, 26, 86, 54, 55, 56, 138,
	91, 59, 61, 66, 29, 135, 14, 102, 62, 65,
	109, 72, 73, 74, 75, 76, 77, 78, 79, 80,
	81, 82, 83, 84, 85, 102, 87, 88, 89, 90,
	68, 139, 93, 94, 21, 129, 17, 123, 92, 23,
	128, 102, 152, 108, 25, 102, 20, 105, 22, 107,
	143, 98, 54, 140, 60, 142, 141, 24, 114, 113,
	102, 120, 15, 51, 57, 16, 111, 115, 26, 50,
	53, 21, 103, 17, 133, 52, 23, 121, 122, 118,
	117, 106, 97, 20, 122, 22, 96, 124, 71, 69,
	1, 130, 132, 124, 24, 124, 131, 134, 147, 15,
	19, 137, 16, 136, 13, 26, 124, 12, 21, 127,
	67, 148, 64, 23, 28, 31, 8, 4, 30, 144,
	20, 146, 22, 32, 137, 149, 150, 21, 125, 17,
	18, 24, 23, 153, 101, 126, 104, 0, 0, 20,
	0, 22, 26, 0, 0, 6, 145, 0, 0, 0,
	24, 0, 21, 0, 17, 15, 0, 23, 16, 0,
	13, 26, 10, 12, 20, 154, 22, 5, 0, 0,
	6, 3, 0, 31, 0, 24, 30, 33, 0, 34,
	15, 32, 0, 16, 0, 13, 26, 10, 12, 0,
	31, 0, 5, 30, 33, 0, 34, 0, 32, 110,
	35, 41, 36, 0, 0, 0, 0, 42, 46, 0,
	0, 37, 0, 40, 0, 48, 0, 43, 47, 0,
	38, 39, 44, 45, 31, 0, 0, 30, 33, 0,
	34, 0, 32, 0, 35, 41, 36, 0, 119, 0,
	0, 42, 46, 0, 0, 37, 0, 40, 0, 48,
	0, 43, 47, 0, 38, 39, 44, 45, 31, 0,
	0, 30, 33, 0, 34, 0, 32, 0, 35, 41,
	36, 0, 0, 0, 0, 42, 46, 0, 0, 37,
	102, 40, 0, 48, 0, 43, 47, 0, 38, 39,
	44, 45, 31, 0, 0, 30, 33, 0, 34, 0,
	32, 0, 35, 41, 36, 0, 0, 0, 0, 42,
	46, 0, 0, 37, 0, 40, 0, 48, 116, 43,
	47, 0, 38, 39, 44, 45, 31, 0, 0, 30,
	33, 0, 34, 0, 32, 0, 35, 41, 36, 0,
	0, 0, 0, 42, 46, 0, 0, 37, 0, 40,
	0, 48, 0, 43, 47, 0, 38, 39, 44, 45,
	31, 0, 0, 30, 33, 0, 34, 0, 32, 0,
	35, 41, 36, 0, 0, 0, 0, 42, 46, 0,
	0, 37, 0, 40, 0, 0, 0, 43, 47, 0,
	38, 39, 44, 45, 31, 0, 0, 30, 33, 0,
	34, 0, 32, 0, 35, 0, 36, 0, 0, 0,
	0, 0, 46, 0, 0, 37, 0, 40, 0, 48,
	0, 43, 47, 0, 38, 39, 44, 45, 31, 0,
	0, 30, 33, 0, 34, 0, 32, 0, 35, 0,
	36, 0, 0, 0, 0, 0, 46, 0, 0, 37,
	0, 40, 21, 0, 17, 43, 47, 23, 38, 39,
	44, 45, 0, 0, 20, 0, 22, 0, 0, 0,
	0, 0, 0, 0, 0, 24, 0, 0, 0, 0,
	15, 0, 0, 16, 0, 13, 26, 31, 12, 0,
	30, 33, 0, 34, 0, 32, 0, 35, 0, 36,
	0, 0, 0, 31, 0, 46, 30, 33, 37, 34,
	40, 32, 0, 35, 0, 36, 0, 38, 39, 31,
	45, 46, 30, 33, 37, 34, 40, 32, 0, 35,
	0, 36, 0, 38, 39, 0, 0, 0, 0, 0,
	37, 0, 40, 0, 0, 0, 0, 0, 0, 38,
	39,
}
var yyPact = [...]int{

	-1000, -1000, 167, -1000, -1000, -1000, -34, -1000, -1000, -1000,
	-4, 342, 49, -1000, 78, 49, 49, 49, -1000, -25,
	49, 49, 49, 123, -1000, -1000, -1000, -1000, -39, 103,
	49, 49, 49, 49, 49, 49, 49, 49, 49, 49,
	49, 49, 49, 49, -16, 49, 49, 49, 49, 342,
	-8, 49, 49, 93, 342, -1000, -1000, -1000, 53, 274,
	83, 274, 95, 1, 43, 10, 206, 77, -1000, -35,
	467, 49, -1000, -1000, -1000, 131, 131, 189, 189, 189,
	189, 189, 189, 410, 410, 503, 49, 519, 535, 503,
	308, -1000, 94, 274, 240, 68, 49, 49, -1000, 39,
	-1000, -1000, 123, 49, -1000, 54, -1000, 35, -1000, -1000,
	49, 49, -1000, -1000, 88, 503, 49, -1000, 19, -1000,
	49, 444, 342, -1000, -1000, -10, 42, 78, -1000, -1000,
	342, -1000, 206, 63, 444, -1000, 57, 342, 49, 123,
	49, 86, -1000, 49, 376, 78, 376, -1000, -30, -1000,
	44, -1000, -1000, 142, -1000,
}
var yyPgo = [...]int{

	0, 156, 0, 2, 26, 74, 6, 155, 154, 7,
	5, 150, 148, 1, 137, 136, 3, 8, 134, 4,
	132, 130, 64, 120, 118, 110, 109,
}
var yyR1 = [...]int{

	0, 25, 24, 24, 13, 13, 13, 13, 14, 14,
	15, 16, 18, 18, 17, 17, 17, 17, 26, 26,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 3,
	3, 1, 1, 19, 21, 21, 20, 20, 5, 5,
	6, 6, 7, 7, 22, 23, 23, 11, 8, 9,
	10, 10, 12, 12,
}
var yyR2 = [...]int{

	0, 2, 4, 1, 0, 2, 2, 3, 1, 1,
	7, 4, 0, 3, 1, 2, 1, 1, 0, 1,
	1, 3, 4, 4, 6, 8, 5, 1, 3, 4,
	4, 4, 3, 3, 3, 2, 1, 4, 2, 2,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 4, 3, 3, 3, 5, 0,
	1, 0, 1, 3, 1, 3, 1, 2, 1, 3,
	0, 2, 1, 3, 1, 1, 2, 1, 4, 2,
	1, 2, 0, 3,
}
var yyChk = [...]int{

	-1000, -25, -13, 24, -14, 45, 23, -16, -15, -17,
	40, -2, 41, 38, -4, 33, 36, 7, -11, -23,
	17, 5, 19, 10, 28, -22, 39, 45, -18, 28,
	7, 4, 12, 8, 10, 14, 16, 25, 34, 35,
	27, 15, 21, 31, 36, 37, 22, 32, 29, -2,
	11, 5, 17, -5, -2, -2, -2, -22, -6, -2,
	-5, -2, -6, -19, -20, -6, -2, -21, -4, -26,
	48, 5, -2, -2, -2, -2, -2, -2, -2, -2,
	-2, -2, -2, -2, -2, -2, 31, -2, -2, -2,
	-2, 28, -6, -2, -2, -3, 13, 9, 18, -10,
	-9, -8, 26, 9, -1, -10, 6, -10, 20, 20,
	13, 9, 45, -17, -6, -2, 30, 6, -10, 18,
	13, -2, -2, 18, -9, -12, -7, -4, 6, 20,
	-2, -19, -2, 6, -2, 6, -3, -2, 29, 9,
	31, 13, 18, 13, -2, -4, -2, -24, 45, -16,
	-3, 42, 18, -13, 43,
}
var yyDef = [...]int{

	4, -2, 0, 1, 5, 6, 0, 8, 9, 12,
	0, 14, 16, 17, 36, 0, 0, 0, 20, 27,
	70, 70, 70, 0, 77, 75, 74, 7, 18, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 15,
	0, 70, 59, 0, 68, 38, 39, 76, 0, 68,
	61, 68, 0, 64, 0, 0, 68, 66, 35, 0,
	19, 70, 40, 41, 42, 43, 44, 45, 46, 47,
	48, 49, 50, 51, 52, 53, 0, 55, 56, 57,
	0, 21, 0, 68, 60, 0, 0, 0, 28, 0,
	80, 82, 0, 62, 71, 0, 34, 0, 32, 33,
	0, 67, 11, 13, 0, 54, 0, 22, 0, 23,
	59, 37, 69, 29, 81, 79, 0, 72, 30, 31,
	63, 65, 0, 0, 58, 26, 0, 60, 0, 0,
	0, 0, 24, 59, 83, 73, 78, 10, 0, 3,
	0, 4, 25, 0, 2,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	45, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 3, 3,
	5, 6, 7, 8, 9, 10, 11, 12, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 48,
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
	39, 40, 41, 42, 43, 44, 46, 47,
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
		//line build/parse.y:164
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:171
		{
			yyVAL.block = CodeBlock{
				Start:      yyDollar[2].pos,
				Statements: yyDollar[3].exprs,
				End:        End{Pos: yyDollar[4].pos},
			}
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:179
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
		//line build/parse.y:191
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:196
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
		//line build/parse.y:227
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:233
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
		//line build/parse.y:247
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:251
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 10:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line build/parse.y:257
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
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:272
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastRule = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 12:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:278
		{
			yyVAL.exprs = []Expr{}
		}
	case 13:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:282
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 15:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:289
		{
			_, end := yyDollar[2].expr.Span()
			yyVAL.expr = &ReturnExpr{
				X:   yyDollar[2].expr,
				End: end,
			}
		}
	case 16:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:297
		{
			yyVAL.expr = &ReturnExpr{End: yyDollar[1].pos}
		}
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:301
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:311
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 22:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:320
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
	case 23:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:331
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 24:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:340
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
	case 25:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:351
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
	case 26:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:364
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
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:381
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
	case 28:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:393
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 29:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:403
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
	case 30:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:415
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
	case 31:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:427
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
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:439
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:449
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:459
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
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:479
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 37:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:484
		{
			yyVAL.expr = &LambdaExpr{
				Lambda: yyDollar[1].pos,
				Var:    yyDollar[2].exprs,
				Colon:  yyDollar[3].pos,
				Expr:   yyDollar[4].expr,
			}
		}
	case 38:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:492
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 39:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:493
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:494
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 41:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:495
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 42:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:496
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:497
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 44:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:498
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 45:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:499
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 46:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:500
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 47:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:501
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 48:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:502
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:503
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 50:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:504
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 51:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:505
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 52:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:506
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 53:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:507
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 54:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:508
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:509
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 56:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:510
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 57:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:512
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 58:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:520
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 59:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:531
		{
			yyVAL.expr = nil
		}
	case 61:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:541
		{
			yyVAL.pos = Position{}
		}
	case 63:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:547
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:557
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 65:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:561
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 66:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:567
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 67:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:571
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 68:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:577
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:581
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 70:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:586
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:590
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:596
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 73:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:600
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:606
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:618
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 76:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:622
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:628
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 78:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:634
		{
			yyVAL.forc = &ForClause{
				For:  yyDollar[1].pos,
				Var:  yyDollar[2].exprs,
				In:   yyDollar[3].pos,
				Expr: yyDollar[4].expr,
			}
		}
	case 79:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:644
		{
			yyVAL.forifs = &ForClauseWithIfClausesOpt{
				For: yyDollar[1].forc,
				Ifs: yyDollar[2].ifs,
			}
		}
	case 80:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:653
		{
			yyVAL.forsifs = []*ForClauseWithIfClausesOpt{yyDollar[1].forifs}
		}
	case 81:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:656
		{
			yyVAL.forsifs = append(yyDollar[1].forsifs, yyDollar[2].forifs)
		}
	case 82:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:661
		{
			yyVAL.ifs = nil
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:665
		{
			yyVAL.ifs = append(yyDollar[1].ifs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	}
	goto yystack /* stack new state and value */
}
