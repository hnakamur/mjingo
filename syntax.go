package mjingo

// Syntax is the delimiter configuration for the environment and the parser.
//
// mjingo allows you to override the syntax configuration for
// templates by setting different delimiters.  The end markers can
// be shared, but the start markers need to be distinct.  It would
// thus not be valid to configure `{{` to be the marker for both
// variables and blocks.
type Syntax struct {
	BlockStart    string
	BlockEnd      string
	VariableStart string
	VariableEnd   string
	CommentStart  string
	CommentEnd    string
}

// DefaultSyntax is the default delimiter configuration for the environment and the parser.
var DefaultSyntax = Syntax{
	BlockStart:    "{%",
	BlockEnd:      "%}",
	VariableStart: "{{",
	VariableEnd:   "}}",
	CommentStart:  "{#",
	CommentEnd:    "#}",
}

type syntaxConfig struct {
	Syntax Syntax
}

var defaultSyntaxConfig = syntaxConfig{
	Syntax: DefaultSyntax,
}
