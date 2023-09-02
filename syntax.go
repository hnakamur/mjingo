package mjingo

type Syntax struct {
	BlockStart    string
	BlockEnd      string
	VariableStart string
	VariableEnd   string
	CommentStart  string
	CommentEnd    string
}

var DefaultSyntax = Syntax{
	BlockStart:    "{%",
	BlockEnd:      "%}",
	VariableStart: "{{",
	VariableEnd:   "}}",
	CommentStart:  "{#",
	CommentEnd:    "#}",
}

type SyntaxConfig struct {
	Syntax Syntax
}

var DefaultSyntaxConfig = SyntaxConfig{
	Syntax: DefaultSyntax,
}
