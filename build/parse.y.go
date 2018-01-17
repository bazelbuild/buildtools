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
const ShiftInstead = 57365
const _ASSERT = 57366
const _UNARY = 57367

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

//line build/parse.y:588

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
		switch x.(type) {
		case *LiteralExpr, *StringExpr, *UnaryExpr:
			// ok
		default:
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

const yyNprod = 74
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 487

var yyAct = [...]int{

	111, 86, 9, 7, 90, 56, 51, 103, 91, 21,
	77, 47, 48, 49, 22, 126, 52, 54, 59, 82,
	93, 61, 127, 55, 58, 63, 64, 65, 66, 67,
	68, 69, 70, 71, 72, 73, 74, 75, 76, 100,
	78, 79, 80, 81, 128, 84, 85, 123, 112, 118,
	117, 83, 99, 135, 25, 93, 93, 24, 27, 96,
	28, 98, 26, 101, 29, 35, 30, 93, 95, 130,
	93, 36, 40, 89, 129, 31, 20, 34, 104, 42,
	109, 37, 41, 102, 32, 33, 38, 39, 110, 107,
	88, 50, 44, 94, 87, 23, 116, 106, 43, 113,
	97, 62, 119, 121, 45, 113, 122, 113, 120, 53,
	125, 124, 16, 1, 14, 25, 113, 18, 24, 27,
	46, 28, 60, 26, 15, 57, 17, 131, 4, 133,
	132, 125, 134, 25, 2, 19, 24, 27, 114, 28,
	13, 26, 92, 29, 35, 30, 21, 108, 115, 25,
	36, 40, 24, 0, 31, 0, 34, 26, 42, 0,
	37, 41, 0, 32, 33, 38, 39, 25, 0, 0,
	24, 27, 0, 28, 0, 26, 0, 29, 35, 30,
	0, 0, 0, 0, 36, 40, 0, 0, 31, 93,
	34, 0, 42, 0, 37, 41, 0, 32, 33, 38,
	39, 25, 0, 0, 24, 27, 0, 28, 0, 26,
	0, 29, 35, 30, 0, 0, 0, 0, 36, 40,
	0, 0, 31, 0, 34, 0, 42, 105, 37, 41,
	0, 32, 33, 38, 39, 25, 0, 0, 24, 27,
	0, 28, 0, 26, 0, 29, 35, 30, 0, 0,
	0, 0, 36, 40, 0, 0, 31, 0, 34, 0,
	42, 0, 37, 41, 0, 32, 33, 38, 39, 25,
	0, 0, 24, 27, 0, 28, 0, 26, 0, 29,
	35, 30, 0, 0, 0, 0, 36, 40, 0, 0,
	31, 0, 34, 0, 0, 0, 37, 41, 0, 32,
	33, 38, 39, 25, 0, 0, 24, 27, 0, 28,
	0, 26, 0, 29, 0, 30, 0, 0, 0, 0,
	0, 40, 0, 0, 31, 0, 34, 16, 42, 12,
	37, 41, 18, 32, 33, 38, 39, 0, 0, 15,
	0, 17, 0, 0, 0, 6, 3, 0, 0, 0,
	19, 0, 0, 0, 0, 10, 0, 0, 11, 0,
	8, 21, 25, 5, 0, 24, 27, 0, 28, 0,
	26, 0, 29, 0, 30, 0, 0, 0, 0, 0,
	40, 0, 0, 31, 0, 34, 0, 0, 0, 37,
	41, 0, 32, 33, 38, 39, 25, 0, 0, 24,
	27, 0, 28, 0, 26, 0, 29, 0, 30, 0,
	16, 0, 12, 0, 40, 18, 0, 31, 0, 34,
	0, 0, 15, 0, 17, 0, 32, 33, 0, 39,
	0, 0, 0, 19, 0, 0, 0, 0, 10, 25,
	0, 11, 24, 27, 21, 28, 0, 26, 0, 29,
	0, 30, 0, 0, 0, 25, 0, 40, 24, 27,
	31, 28, 34, 26, 0, 29, 0, 30, 0, 32,
	33, 0, 0, 0, 0, 0, 31, 0, 34, 0,
	0, 0, 0, 0, 0, 32, 33,
}
var yyPact = [...]int{

	-1000, -1000, 322, -1000, 86, -1000, -1000, 231, -1000, 87,
	405, 405, 405, -1000, -30, 405, 405, 405, 107, -1000,
	-1000, -1000, -1000, -1000, 405, 405, 405, 405, 405, 405,
	405, 405, 405, 405, 405, 405, 405, 405, -21, 405,
	405, 405, 405, -9, 405, 405, 81, 231, -1000, -1000,
	-1000, 55, 163, 84, 163, 94, -6, 32, 19, 50,
	74, -1000, -37, -1000, -1000, -1000, 145, 145, 111, 111,
	111, 111, 111, 111, 299, 299, 392, 405, 435, 451,
	392, 197, -1000, 91, 163, 129, 67, 405, 405, -1000,
	30, -1000, -1000, 107, 405, -1000, 44, -1000, 29, -1000,
	-1000, 405, 405, -1000, 392, 405, -1000, 41, -1000, 405,
	358, 231, -1000, -1000, -14, 13, 87, -1000, -1000, 231,
	-1000, 50, 358, -1000, 56, 231, 405, 107, 405, -1000,
	405, 265, 87, 265, 35, -1000,
}
var yyPgo = [...]int{

	0, 14, 0, 1, 2, 109, 6, 148, 142, 8,
	4, 140, 138, 134, 128, 5, 125, 122, 76, 114,
	113, 101,
}
var yyR1 = [...]int{

	0, 20, 13, 13, 13, 13, 14, 14, 21, 21,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 3,
	3, 1, 1, 15, 17, 17, 16, 16, 5, 5,
	6, 6, 7, 7, 18, 19, 19, 11, 8, 9,
	10, 10, 12, 12,
}
var yyR2 = [...]int{

	0, 2, 0, 4, 2, 2, 1, 1, 0, 2,
	1, 3, 4, 4, 6, 8, 5, 1, 3, 4,
	4, 4, 3, 3, 3, 2, 1, 4, 2, 2,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 4, 3, 3, 3, 5, 0,
	1, 0, 1, 3, 1, 3, 1, 2, 1, 3,
	0, 2, 1, 3, 1, 1, 2, 1, 4, 2,
	1, 2, 0, 3,
}
var yyChk = [...]int{

	-1000, -20, -13, 24, -14, 41, 23, -2, 38, -4,
	33, 36, 7, -11, -19, 17, 5, 19, 10, 28,
	-18, 39, -1, 9, 7, 4, 12, 8, 10, 14,
	16, 25, 34, 35, 27, 15, 21, 31, 36, 37,
	22, 32, 29, 11, 5, 17, -5, -2, -2, -2,
	-18, -6, -2, -5, -2, -6, -15, -16, -6, -2,
	-17, -4, -21, -2, -2, -2, -2, -2, -2, -2,
	-2, -2, -2, -2, -2, -2, -2, 31, -2, -2,
	-2, -2, 28, -6, -2, -2, -3, 13, 9, 18,
	-10, -9, -8, 26, 9, -1, -10, 6, -10, 20,
	20, 13, 9, 44, -2, 30, 6, -10, 18, 13,
	-2, -2, 18, -9, -12, -7, -4, 6, 20, -2,
	-15, -2, -2, 6, -3, -2, 29, 9, 31, 18,
	13, -2, -4, -2, -3, 18,
}
var yyDef = [...]int{

	2, -2, 0, 1, 51, 4, 5, 6, 7, 26,
	0, 0, 0, 10, 17, 60, 60, 60, 0, 67,
	65, 64, 8, 52, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 60, 49, 0, 58, 28, 29,
	66, 0, 58, 51, 58, 0, 54, 0, 0, 58,
	56, 25, 3, 30, 31, 32, 33, 34, 35, 36,
	37, 38, 39, 40, 41, 42, 43, 0, 45, 46,
	47, 0, 11, 0, 58, 50, 0, 0, 0, 18,
	0, 70, 72, 0, 52, 61, 0, 24, 0, 22,
	23, 0, 57, 9, 44, 0, 12, 0, 13, 49,
	27, 59, 19, 71, 69, 0, 62, 20, 21, 53,
	55, 0, 48, 16, 0, 50, 0, 0, 0, 14,
	49, 73, 63, 68, 0, 15,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	41, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 3, 3,
	5, 6, 7, 8, 9, 10, 11, 12, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 44,
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
	39, 40, 42, 43,
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
		//line build/parse.y:154
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:160
		{
			yyVAL.exprs = nil
			yyVAL.lastRule = nil
		}
	case 3:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:165
		{
			// If this statement follows a comment block,
			// attach the comments to the statement.
			if cb, ok := yyDollar[1].lastRule.(*CommentBlock); ok {
				yyVAL.exprs = yyDollar[1].exprs
				yyVAL.exprs[len(yyDollar[1].exprs)-1] = yyDollar[2].expr
				yyDollar[2].expr.Comment().Before = cb.After
				yyVAL.lastRule = yyDollar[2].expr
				break
			}

			// Otherwise add to list.
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].expr)
			yyVAL.lastRule = yyDollar[2].expr

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
				yyDollar[2].expr.Comment().Before = com.After
				com.After = nil
			}
		}
	case 4:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:196
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastRule = nil
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:202
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
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:217
		{
			yyVAL.expr = &PythonBlock{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 11:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:227
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 12:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:236
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
	case 13:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:247
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 14:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line build/parse.y:256
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
	case 15:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line build/parse.y:267
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
	case 16:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:280
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
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:297
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
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:309
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 19:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:319
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
	case 20:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:331
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
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:343
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
	case 22:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:355
		{
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 23:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:365
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				Comma:          yyDollar[2].comma,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 24:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:375
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
	case 25:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:395
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:400
		{
			yyVAL.expr = &LambdaExpr{
				Lambda: yyDollar[1].pos,
				Var:    yyDollar[2].exprs,
				Colon:  yyDollar[3].pos,
				Expr:   yyDollar[4].expr,
			}
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:408
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 29:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:409
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:410
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:411
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:412
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:413
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:414
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:415
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:416
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 37:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:417
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 38:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:418
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:419
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:420
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 41:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:421
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 42:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:422
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:423
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 44:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:424
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 45:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:425
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 46:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:426
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 47:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:428
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 48:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line build/parse.y:436
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 49:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:447
		{
			yyVAL.expr = nil
		}
	case 51:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:457
		{
			yyVAL.pos = Position{}
		}
	case 53:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:463
		{
			yyVAL.expr = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 54:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:473
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:477
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 56:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:483
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:487
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:493
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:497
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 60:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:502
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 61:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:506
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:512
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 63:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:516
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:522
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:534
		{
			yyVAL.strings = []*StringExpr{yyDollar[1].string}
		}
	case 66:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:538
		{
			yyVAL.strings = append(yyDollar[1].strings, yyDollar[2].string)
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:544
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 68:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line build/parse.y:550
		{
			yyVAL.forc = &ForClause{
				For:  yyDollar[1].pos,
				Var:  yyDollar[2].exprs,
				In:   yyDollar[3].pos,
				Expr: yyDollar[4].expr,
			}
		}
	case 69:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:560
		{
			yyVAL.forifs = &ForClauseWithIfClausesOpt{
				For: yyDollar[1].forc,
				Ifs: yyDollar[2].ifs,
			}
		}
	case 70:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line build/parse.y:569
		{
			yyVAL.forsifs = []*ForClauseWithIfClausesOpt{yyDollar[1].forifs}
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line build/parse.y:572
		{
			yyVAL.forsifs = append(yyDollar[1].forsifs, yyDollar[2].forifs)
		}
	case 72:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line build/parse.y:577
		{
			yyVAL.ifs = nil
		}
	case 73:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line build/parse.y:581
		{
			yyVAL.ifs = append(yyDollar[1].ifs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	}
	goto yystack /* stack new state and value */
}
