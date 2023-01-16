package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT = "IDENT"
	INT   = "INT"

	// Operators
	ASSIGN = "ASSIGN" // `=`
	PLUS   = "PLUS"   // `+`

	// Operators
	MINUS    = "MINUS"    //`-`
	BANG     = "BANG"     // `!`
	ASTERISK = "ASTERISK" // `*`
	SLASH    = "SLASH"    // `/`
	LT       = "LT"       // `>`
	GT       = "GT"       // `<`
	COLON    = "COLON"    // `:`

	// Delimiters
	COMMA     = "COMMA"     // `,`
	SEMICOLON = "SEMICOLON" // `;`

	LPAREN = "LPAREN" // `(`
	RPAREN = "RPAREN" // `)`
	LBRACE = "LBRACE" // `{`
	RBRACE = "RBRACE" // `}`

	LBRACKET = "LBRACKET" // `[`
	RBRACKET = "RBRACKET" // `]`

	EQ     = "EQ"     // `==`
	NOT_EQ = "NOT_EQ" //`!=`

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"

	// String
	STRING = "STRING"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}
