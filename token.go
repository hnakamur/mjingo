package mjingo

import "fmt"

type tokenType int

const (
	// Raw template data.
	tokenTemplateData tokenType = iota
	// Variable block start.
	tokenVariableStart
	// Variable block end
	tokenVariableEnd
	// Statement block start
	tokenBlockStart
	// Statement block end
	tokenBlockEnd
	// An identifier.
	tokenIdent
	// A borrowed string.
	tokenStr
	// An allocated string.
	tokenString
	// An integer (limited to i64)
	tokenInt
	// A float
	tokenFloat
	// A plus (`+`) operator.
	tokenPlus
	// A plus (`-`) operator.
	tokenMinus
	// A mul (`*`) operator.
	tokenMul
	// A div (`/`) operator.
	tokenDiv
	// A floor division (`//`) operator.
	tokenFloorDiv
	// Power operator (`**`).
	tokenPow
	// A mod (`%`) operator.
	tokenMod
	// The bang (`!`) operator.
	tokenBang
	// A dot operator (`.`)
	tokenDot
	// The comma operator (`,`)
	tokenComma
	// The colon operator (`:`)
	tokenColon
	// The tilde operator (`~`)
	tokenTilde
	// The assignment operator (`=`)
	tokenAssign
	// The pipe symbol.
	tokenPipe
	// `==` operator
	tokenEq
	// `!=` operator
	tokenNe
	// `>` operator
	tokenGt
	// `>=` operator
	tokenGte
	// `<` operator
	tokenLt
	// `<=` operator
	tokenLte
	// Open Bracket
	tokenBracketOpen
	// Close Bracket
	tokenBracketClose
	// Open Parenthesis
	tokenParenOpen
	// Close Parenthesis
	tokenParenClose
	// Open Brace
	tokenBraceOpen
	// Close Brace
	tokenBraceClose
)

type token struct {
	Type      tokenType
	StrData   string
	IntData   int64
	FloatData float64
}

func (t *token) String() string {
	switch t.Type {
	case tokenTemplateData:
		return "template-data"
	case tokenVariableStart:
		return "start of variable block"
	case tokenVariableEnd:
		return "end of variable block"
	case tokenBlockStart:
		return "start of block"
	case tokenBlockEnd:
		return "end of block"
	case tokenIdent:
		return "identifier"
	case tokenStr:
		return "string"
	case tokenString:
		return "string"
	case tokenInt:
		return "integer"
	case tokenFloat:
		return "float"
	case tokenPlus:
		return "`+`"
	case tokenMinus:
		return "`-`"
	case tokenMul:
		return "`*`"
	case tokenDiv:
		return "`/`"
	case tokenFloorDiv:
		return "`//`"
	case tokenPow:
		return "`**`"
	case tokenMod:
		return "`%`"
	case tokenBang:
		return "`!`"
	case tokenDot:
		return "`.`"
	case tokenComma:
		return "`,`"
	case tokenColon:
		return "`:`"
	case tokenTilde:
		return "`~`"
	case tokenAssign:
		return "`=`"
	case tokenPipe:
		return "`|`"
	case tokenEq:
		return "`==`"
	case tokenNe:
		return "`!=`"
	case tokenGt:
		return "`>`"
	case tokenGte:
		return "`>=`"
	case tokenLt:
		return "`<`"
	case tokenLte:
		return "`<=`"
	case tokenBracketOpen:
		return "`[`"
	case tokenBracketClose:
		return "`]`"
	case tokenParenOpen:
		return "`(`"
	case tokenParenClose:
		return "`)`"
	case tokenBraceOpen:
		return "`{`"
	case tokenBraceClose:
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
