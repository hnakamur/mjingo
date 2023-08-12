package mjingo

import "fmt"

type tokenKind int

const (
	// Raw template data.
	tokenKindTemplateData tokenKind = iota
	// Variable block start.
	tokenKindVariableStart
	// Variable block end
	tokenKindVariableEnd
	// Statement block start
	tokenKindBlockStart
	// Statement block end
	tokenKindBlockEnd
	// An identifier.
	tokenKindIdent
	// A borrowed string.
	tokenKindStr
	// An allocated string.
	tokenKindString
	// An integer (limited to i64)
	tokenKindInt
	// A float
	tokenKindFloat
	// A plus (`+`) operator.
	tokenKindPlus
	// A plus (`-`) operator.
	tokenKindMinus
	// A mul (`*`) operator.
	tokenKindMul
	// A div (`/`) operator.
	tokenKindDiv
	// A floor division (`//`) operator.
	tokenKindFloorDiv
	// Power operator (`**`).
	tokenKindPow
	// A mod (`%`) operator.
	tokenKindMod
	// The bang (`!`) operator.
	tokenKindBang
	// A dot operator (`.`)
	tokenKindDot
	// The comma operator (`,`)
	tokenKindComma
	// The colon operator (`:`)
	tokenKindColon
	// The tilde operator (`~`)
	tokenKindTilde
	// The assignment operator (`=`)
	tokenKindAssign
	// The pipe symbol.
	tokenKindPipe
	// `==` operator
	tokenKindEq
	// `!=` operator
	tokenKindNe
	// `>` operator
	tokenKindGt
	// `>=` operator
	tokenKindGte
	// `<` operator
	tokenKindLt
	// `<=` operator
	tokenKindLte
	// Open Bracket
	tokenKindBracketOpen
	// Close Bracket
	tokenKindBracketClose
	// Open Parenthesis
	tokenKindParenOpen
	// Close Parenthesis
	tokenKindParenClose
	// Open Brace
	tokenKindBraceOpen
	// Close Brace
	tokenKindBraceClose
)

type templateDataTokenData string
type identTokenData string
type strTokenData string
type stringTokenData string
type intTokenData int64
type floatTokenData float64

type token struct {
	kind tokenKind
	data any
}

func (t *token) String() string {
	switch t.kind {
	case tokenKindTemplateData:
		return "template-data"
	case tokenKindVariableStart:
		return "start of variable block"
	case tokenKindVariableEnd:
		return "end of variable block"
	case tokenKindBlockStart:
		return "start of block"
	case tokenKindBlockEnd:
		return "end of block"
	case tokenKindIdent:
		return "identifier"
	case tokenKindStr:
		return "string"
	case tokenKindString:
		return "string"
	case tokenKindInt:
		return "integer"
	case tokenKindFloat:
		return "float"
	case tokenKindPlus:
		return "`+`"
	case tokenKindMinus:
		return "`-`"
	case tokenKindMul:
		return "`*`"
	case tokenKindDiv:
		return "`/`"
	case tokenKindFloorDiv:
		return "`//`"
	case tokenKindPow:
		return "`**`"
	case tokenKindMod:
		return "`%`"
	case tokenKindBang:
		return "`!`"
	case tokenKindDot:
		return "`.`"
	case tokenKindComma:
		return "`,`"
	case tokenKindColon:
		return "`:`"
	case tokenKindTilde:
		return "`~`"
	case tokenKindAssign:
		return "`=`"
	case tokenKindPipe:
		return "`|`"
	case tokenKindEq:
		return "`==`"
	case tokenKindNe:
		return "`!=`"
	case tokenKindGt:
		return "`>`"
	case tokenKindGte:
		return "`>=`"
	case tokenKindLt:
		return "`<`"
	case tokenKindLte:
		return "`<=`"
	case tokenKindBracketOpen:
		return "`[`"
	case tokenKindBracketClose:
		return "`]`"
	case tokenKindParenOpen:
		return "`(`"
	case tokenKindParenClose:
		return "`)`"
	case tokenKindBraceOpen:
		return "`{`"
	case tokenKindBraceClose:
		return "`}`"
	default:
		return "unknown"
	}
}

type span struct {
	startLine   uint32
	startCol    uint32
	startOffset uint32
	endLine     uint32
	endCol      uint32
	endOffset   uint32
}

func (s *span) String() string {
	return fmt.Sprintf(" @ %d:%d-%d:%d", s.startLine, s.startCol, s.endLine, s.endCol)
}

type loc struct {
	line   uint32
	col    uint32
	offset uint32
}
