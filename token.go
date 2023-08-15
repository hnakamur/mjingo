package mjingo

import "fmt"

type token interface {
	String() string

	typ() tokenType
}

type templateDataToken struct{ s string }
type variableStartToken struct{}
type variableEndToken struct{}
type blockStartToken struct{}
type blockEndToken struct{}
type identToken struct{ s string }
type stringToken struct{ s string }
type intToken struct{ n int64 }
type floatToken struct{ f float64 }
type plusToken struct{}
type minusToken struct{}
type mulToken struct{}
type divToken struct{}
type floorDivToken struct{}
type powToken struct{}
type modToken struct{}
type bangToken struct{}
type dotToken struct{}
type commaToken struct{}
type colonToken struct{}
type tildeToken struct{}
type assignToken struct{}
type pipeToken struct{}
type eqToken struct{}
type neToken struct{}
type gtToken struct{}
type gteToken struct{}
type ltToken struct{}
type lteToken struct{}
type bracketOpenToken struct{}
type bracketCloseToken struct{}
type parenOpenToken struct{}
type parenCloseToken struct{}
type braceOpenToken struct{}
type braceCloseToken struct{}

var _ = token(templateDataToken{})
var _ = token(variableStartToken{})
var _ = token(variableEndToken{})
var _ = token(blockStartToken{})
var _ = token(blockEndToken{})
var _ = token(identToken{})
var _ = token(stringToken{})
var _ = token(intToken{})
var _ = token(floatToken{})
var _ = token(plusToken{})
var _ = token(minusToken{})
var _ = token(mulToken{})
var _ = token(divToken{})
var _ = token(floorDivToken{})
var _ = token(powToken{})
var _ = token(modToken{})
var _ = token(bangToken{})
var _ = token(dotToken{})
var _ = token(commaToken{})
var _ = token(colonToken{})
var _ = token(tildeToken{})
var _ = token(assignToken{})
var _ = token(pipeToken{})
var _ = token(eqToken{})
var _ = token(neToken{})
var _ = token(gtToken{})
var _ = token(gteToken{})
var _ = token(ltToken{})
var _ = token(lteToken{})
var _ = token(bracketOpenToken{})
var _ = token(bracketCloseToken{})
var _ = token(parenOpenToken{})
var _ = token(parenCloseToken{})
var _ = token(braceOpenToken{})
var _ = token(braceCloseToken{})

func (t templateDataToken) String() string  { return t.typ().String() }
func (t variableStartToken) String() string { return t.typ().String() }
func (t variableEndToken) String() string   { return t.typ().String() }
func (t blockStartToken) String() string    { return t.typ().String() }
func (t blockEndToken) String() string      { return t.typ().String() }
func (t identToken) String() string         { return t.typ().String() }
func (t stringToken) String() string        { return t.typ().String() }
func (t intToken) String() string           { return t.typ().String() }
func (t floatToken) String() string         { return t.typ().String() }
func (t plusToken) String() string          { return t.typ().String() }
func (t minusToken) String() string         { return t.typ().String() }
func (t mulToken) String() string           { return t.typ().String() }
func (t divToken) String() string           { return t.typ().String() }
func (t floorDivToken) String() string      { return t.typ().String() }
func (t powToken) String() string           { return t.typ().String() }
func (t modToken) String() string           { return t.typ().String() }
func (t bangToken) String() string          { return t.typ().String() }
func (t dotToken) String() string           { return t.typ().String() }
func (t commaToken) String() string         { return t.typ().String() }
func (t colonToken) String() string         { return t.typ().String() }
func (t tildeToken) String() string         { return t.typ().String() }
func (t assignToken) String() string        { return t.typ().String() }
func (t pipeToken) String() string          { return t.typ().String() }
func (t eqToken) String() string            { return t.typ().String() }
func (t neToken) String() string            { return t.typ().String() }
func (t gtToken) String() string            { return t.typ().String() }
func (t gteToken) String() string           { return t.typ().String() }
func (t ltToken) String() string            { return t.typ().String() }
func (t lteToken) String() string           { return t.typ().String() }
func (t bracketOpenToken) String() string   { return t.typ().String() }
func (t bracketCloseToken) String() string  { return t.typ().String() }
func (t parenOpenToken) String() string     { return t.typ().String() }
func (t parenCloseToken) String() string    { return t.typ().String() }
func (t braceOpenToken) String() string     { return t.typ().String() }
func (t braceCloseToken) String() string    { return t.typ().String() }

func (templateDataToken) typ() tokenType  { return tokenTypeTemplateData }
func (variableStartToken) typ() tokenType { return tokenTypeVariableStart }
func (variableEndToken) typ() tokenType   { return tokenTypeVariableEnd }
func (blockStartToken) typ() tokenType    { return tokenTypeBlockStart }
func (blockEndToken) typ() tokenType      { return tokenTypeBlockEnd }
func (identToken) typ() tokenType         { return tokenTypeIdent }
func (stringToken) typ() tokenType        { return tokenTypeString }
func (intToken) typ() tokenType           { return tokenTypeInt }
func (floatToken) typ() tokenType         { return tokenTypeFloat }
func (plusToken) typ() tokenType          { return tokenTypePlus }
func (minusToken) typ() tokenType         { return tokenTypeMinus }
func (mulToken) typ() tokenType           { return tokenTypeMul }
func (divToken) typ() tokenType           { return tokenTypeDiv }
func (floorDivToken) typ() tokenType      { return tokenTypeFloorDiv }
func (powToken) typ() tokenType           { return tokenTypePow }
func (modToken) typ() tokenType           { return tokenTypeMod }
func (bangToken) typ() tokenType          { return tokenTypeBang }
func (dotToken) typ() tokenType           { return tokenTypeDot }
func (commaToken) typ() tokenType         { return tokenTypeComma }
func (colonToken) typ() tokenType         { return tokenTypeColon }
func (tildeToken) typ() tokenType         { return tokenTypeTilde }
func (assignToken) typ() tokenType        { return tokenTypeAssign }
func (pipeToken) typ() tokenType          { return tokenTypePipe }
func (eqToken) typ() tokenType            { return tokenTypeEq }
func (neToken) typ() tokenType            { return tokenTypeNe }
func (gtToken) typ() tokenType            { return tokenTypeGt }
func (gteToken) typ() tokenType           { return tokenTypeGte }
func (ltToken) typ() tokenType            { return tokenTypeLt }
func (lteToken) typ() tokenType           { return tokenTypeLte }
func (bracketOpenToken) typ() tokenType   { return tokenTypeBracketOpen }
func (bracketCloseToken) typ() tokenType  { return tokenTypeBracketClose }
func (parenOpenToken) typ() tokenType     { return tokenTypeParenOpen }
func (parenCloseToken) typ() tokenType    { return tokenTypeParenClose }
func (braceOpenToken) typ() tokenType     { return tokenTypeBraceOpen }
func (braceCloseToken) typ() tokenType    { return tokenTypeBraceClose }

func isTokenOfType[T any](tkn token) bool {
	_, ok := tkn.(T)
	return ok
}

type tokenType int

const (
	// Raw template data.
	tokenTypeTemplateData tokenType = iota + 1
	// Variable block start.
	tokenTypeVariableStart
	// Variable block end
	tokenTypeVariableEnd
	// Statement block start
	tokenTypeBlockStart
	// Statement block end
	tokenTypeBlockEnd
	// An identifier.
	tokenTypeIdent
	// A borrowed string.
	// NOTE: not used in mjingo
	tokenTypeStr
	// An allocated string.
	tokenTypeString
	// An integer (limited to i64)
	tokenTypeInt
	// A float
	tokenTypeFloat
	// A plus (`+`) operator.
	tokenTypePlus
	// A plus (`-`) operator.
	tokenTypeMinus
	// A mul (`*`) operator.
	tokenTypeMul
	// A div (`/`) operator.
	tokenTypeDiv
	// A floor division (`//`) operator.
	tokenTypeFloorDiv
	// Power operator (`**`).
	tokenTypePow
	// A mod (`%`) operator.
	tokenTypeMod
	// The bang (`!`) operator.
	tokenTypeBang
	// A dot operator (`.`)
	tokenTypeDot
	// The comma operator (`,`)
	tokenTypeComma
	// The colon operator (`:`)
	tokenTypeColon
	// The tilde operator (`~`)
	tokenTypeTilde
	// The assignment operator (`=`)
	tokenTypeAssign
	// The pipe symbol.
	tokenTypePipe
	// `==` operator
	tokenTypeEq
	// `!=` operator
	tokenTypeNe
	// `>` operator
	tokenTypeGt
	// `>=` operator
	tokenTypeGte
	// `<` operator
	tokenTypeLt
	// `<=` operator
	tokenTypeLte
	// Open Bracket
	tokenTypeBracketOpen
	// Close Bracket
	tokenTypeBracketClose
	// Open Parenthesis
	tokenTypeParenOpen
	// Close Parenthesis
	tokenTypeParenClose
	// Open Brace
	tokenTypeBraceOpen
	// Close Brace
	tokenTypeBraceClose
)

func (k tokenType) String() string {
	switch k {
	case tokenTypeTemplateData:
		return "template-data"
	case tokenTypeVariableStart:
		return "start of variable block"
	case tokenTypeVariableEnd:
		return "end of variable block"
	case tokenTypeBlockStart:
		return "start of block"
	case tokenTypeBlockEnd:
		return "end of block"
	case tokenTypeIdent:
		return "identifier"
	case tokenTypeStr:
		return "string"
	case tokenTypeString:
		return "string"
	case tokenTypeInt:
		return "integer"
	case tokenTypeFloat:
		return "float"
	case tokenTypePlus:
		return "`+`"
	case tokenTypeMinus:
		return "`-`"
	case tokenTypeMul:
		return "`*`"
	case tokenTypeDiv:
		return "`/`"
	case tokenTypeFloorDiv:
		return "`//`"
	case tokenTypePow:
		return "`**`"
	case tokenTypeMod:
		return "`%`"
	case tokenTypeBang:
		return "`!`"
	case tokenTypeDot:
		return "`.`"
	case tokenTypeComma:
		return "`,`"
	case tokenTypeColon:
		return "`:`"
	case tokenTypeTilde:
		return "`~`"
	case tokenTypeAssign:
		return "`=`"
	case tokenTypePipe:
		return "`|`"
	case tokenTypeEq:
		return "`==`"
	case tokenTypeNe:
		return "`!=`"
	case tokenTypeGt:
		return "`>`"
	case tokenTypeGte:
		return "`>=`"
	case tokenTypeLt:
		return "`<`"
	case tokenTypeLte:
		return "`<=`"
	case tokenTypeBracketOpen:
		return "`[`"
	case tokenTypeBracketClose:
		return "`]`"
	case tokenTypeParenOpen:
		return "`(`"
	case tokenTypeParenClose:
		return "`)`"
	case tokenTypeBraceOpen:
		return "`{`"
	case tokenTypeBraceClose:
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
