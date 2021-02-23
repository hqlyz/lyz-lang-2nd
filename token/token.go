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
	IDENT  = "IDENT" // foo, bar, x, y...
	INT    = "INT"   // 5, 10...
	STRING = "STRING"

	// Operators
	ASSIGN   = "="  // 赋值
	PLUS     = "+"  // 加法
	MINUS    = "-"  // 减法
	ASTERISK = "*"  // 乘法
	SLASH    = "/"  // 除法
	EQ       = "==" // 相等
	NOT_EQ   = "!=" // 不相等
	LT       = "<"  // 小于
	GT       = ">"  // 大于
	BANG     = "!"  // 取反

	// Special characters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
	COLON     = ":"

	// Keywords
	LET      = "LET"
	FUNCTION = "FUNCTION"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

var keywords = map[string]TokenType{
	"let":    LET,
	"fn":     FUNCTION,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	if v, ok := keywords[ident]; ok {
		return v
	}
	return IDENT
}
