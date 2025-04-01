package monvan_parser

import (
	"strconv"
	"strings"
)

type TokenType string

var KEYWORDS = map[string]bool{
	"SELECT":     true,
	"FROM":       true,
	"WHERE":      true,
	"AND":        true,
	"OR":         true,
	"NOT":        true,
	"IN":         true,
	"LIKE":       true,
	"IS":         true,
	"NULL":       true,
	"ORDER":      true,
	"BY":         true,
	"ASC":        true,
	"DESC":       true,
	"LIMIT":      true,
	"OFFSET":     true,
	"INSERT":     true,
	"INTO":       true,
	"VALUES":     true,
	"UPDATE":     true,
	"SET":        true,
	"DELETE":     true,
	"CREATE":     true,
	"TABLE":      true,
	"ALTER":      true,
	"DROP":       true,
	"ADD":        true,
	"MODIFY":     true,
	"PRIMARY":    true,
	"KEY":        true,
	"FOREIGN":    true,
	"REFERENCES": true,
	"JOIN":       true,
	"INNER":      true,
	"LEFT":       true,
	"RIGHT":      true,
	"OUTER":      true,
	"UNION":      true,
	"ALL":        true,
	"GROUP":      true,
	"HAVING":     true,
	"COUNT":      true,
	"SUM":        true,
	"AVG":        true,
	"MAX":        true,
	"MIN":        true,
	"AS":         true,
	"ON":         true,
	"CASE":       true,
	"WHEN":       true,
	"THEN":       true,
	"ELSE":       true,
	"END":        true,
	"EXISTS":     true,
	"SHOW":       true,
	"DATABASES":  true,
	"USE":        true,
	"DESCRIBE":   true,
	"INDEX":      true,
	"KEYS":       true,
	"STATUS":     true,
	"RENAME":     true,
	"TO":         true,
	"DATABASE":   true,
	"CHARACTER":  true,
	"DEFAULT":    true,
	"ENGINE":     true,
	"JSON":       true,
	"USER":       true,
	"ROLE":       true,
	"ASSIGN":     true,
	"REVOKE":     true,
	"INDEXES":    true,
	"COLUMN":     true,
	"INT":        true,
	"VARCHAR":    true,
	"TEXT":       true,
	"DATE":       true,
	"DATETIME":   true,
	"TIMESTAMP":  true,
	"TIME":       true,
	"YEAR":       true,
	"FLOAT":      true,
	"DOUBLE":     true,
	"DECIMAL":    true,
	"NUMERIC":    true,
	"BIT":        true,
	"BOOLEAN":    true,
	"SERIAL":     true,
	"TINYINT":    true,
	"SMALLINT":   true,
	"MEDIUMINT":  true,
	"BIGINT":     true,
	"REAL":       true,
}

const (
	TokenKeyword      TokenType = "KEYWORD"
	TokenParenthesis  TokenType = "PARENTHESIS"
	TokenCurlyBracket TokenType = "CURLY_BRACKET"
	TokenComma        TokenType = "COMMA"
	TokenOperator     TokenType = "OPERATOR"
	TokenLiteral      TokenType = "LITERAL"
	TokenIdentifier   TokenType = "IDENTIFIER"
	TokenComment      TokenType = "COMMENT"
	TokenError        TokenType = "ERROR"
	TokenEOF          TokenType = "EOF"
)

type Token struct {
	Type       TokenType
	Value      string
	PosInitial int
	PosFinal   int
	LowerValue string
}

func (t Token) String() string {
	return t.Value
}

func (t *Token) ErrorString() string {
	return "error on " + t.Value + " position " + strconv.Itoa(t.PosInitial) + " to " + strconv.Itoa(t.PosFinal)
}

type Tokenizer struct {
	Query string
	Pos   int
}

func NewTokenizer(query string) *Tokenizer {
	return &Tokenizer{
		Query: query,
		Pos:   0,
	}
}

func (t *Tokenizer) NextToken() *Token {

	if t.Pos >= len(t.Query) {
		return &Token{
			Type: TokenEOF,
		}
	}

	for t.Pos < len(t.Query) &&
		(t.Query[t.Pos] == ' ' ||
			t.Query[t.Pos] == '\t' ||
			t.Query[t.Pos] == '\n' ||
			t.Query[t.Pos] == '\r') {

		t.Pos++
	}

	if t.Pos >= len(t.Query) {
		return &Token{
			Type: TokenEOF,
		}
	}

	initialPos := t.Pos

	if isError(t.Query[t.Pos]) {
		return &Token{
			Type:       TokenError,
			Value:      string(t.Query[t.Pos]),
			PosInitial: t.Pos,
			PosFinal:   t.Pos + 1,
			LowerValue: strings.ToLower(string(t.Query[t.Pos])),
		}
	}

	if t.Pos >= len(t.Query) || t.Query[t.Pos] == ';' {
		return &Token{
			Type: TokenEOF,
		}
	}

	// Parenthesis
	if t.Query[t.Pos] == '(' || t.Query[t.Pos] == ')' {
		t.Pos++
		return &Token{
			Type:       TokenParenthesis,
			Value:      string(t.Query[t.Pos-1]),
			PosInitial: initialPos,
			PosFinal:   t.Pos,
			LowerValue: strings.ToLower(string(t.Query[t.Pos-1])),
		}
	}

	// Bracket
	if t.Query[t.Pos] == '{' || t.Query[t.Pos] == '}' {
		t.Pos++
		return &Token{
			Type:       TokenCurlyBracket,
			Value:      string(t.Query[t.Pos-1]),
			PosInitial: initialPos,
			PosFinal:   t.Pos,
			LowerValue: strings.ToLower(string(t.Query[t.Pos-1])),
		}
	}

	// Comma
	if t.Query[t.Pos] == ',' {
		t.Pos++
		return &Token{
			Type:       TokenComma,
			Value:      ",",
			PosInitial: initialPos,
			PosFinal:   t.Pos,
			LowerValue: ",",
		}
	}

	if isOperator(t.Query[t.Pos]) {
		return t.readOperator(initialPos)
	}

	op := t.readIdentifierOrKeyword(initialPos)

	return op

}

func (t *Tokenizer) readIdentifierOrKeyword(initialPos int) *Token {
	for t.Pos < len(t.Query) && (isLetter(t.Query[t.Pos]) || isDigit(t.Query[t.Pos])) {
		t.Pos++
	}

	var tokenType TokenType = TokenIdentifier

	// Verify if it is a Keyword
	if KEYWORDS[strings.ToUpper(t.Query[initialPos:t.Pos])] {
		tokenType = TokenKeyword
	}

	return &Token{
		Type:       tokenType,
		Value:      t.Query[initialPos:t.Pos],
		PosInitial: initialPos,
		PosFinal:   t.Pos,
		LowerValue: strings.ToLower(t.Query[initialPos:t.Pos]),
	}
}

func (t *Tokenizer) readOperator(initialPos int) *Token {

	// Case it is certainly not a comment
	if t.Query[initialPos] != '-' && t.Query[initialPos] != '/' {
		for t.Pos < len(t.Query) && isOperator(t.Query[t.Pos]) {
			t.Pos++
		}
	} else {
		// Check for comment
		if t.Pos+1 < len(t.Query) && t.Query[t.Pos+1] == '-' {
			for t.Pos < len(t.Query) && t.Query[t.Pos] != '\n' {
				t.Pos++
			}
			return t.NextToken()
		} else if t.Pos+1 < len(t.Query) && t.Query[t.Pos+1] == '*' {
			for t.Pos < len(t.Query) && (t.Query[t.Pos] != '*' || t.Query[t.Pos+1] != '/') {
				t.Pos++
			}
			t.Pos += 2
			return t.NextToken()
		}
	}

	return &Token{
		Type:       TokenOperator,
		Value:      t.Query[initialPos:t.Pos],
		PosInitial: initialPos,
		PosFinal:   t.Pos,
		LowerValue: strings.ToLower(t.Query[initialPos:t.Pos]),
	}
}

func isError(char byte) bool {
	return !(isLetter(char) || isDigit(char) || char == ' ' || char == '\t' || char == '\n' || char == '\r' || char == '(' || char == ')' || char == '{' || char == '}' || char == ',' || isOperator(char))
}

func isLetter(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' || char == '.' || char == '\'' || char == '"' || char == ':'
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isOperator(char byte) bool {
	return char == '=' || char == '<' || char == '>' || char == '!' || char == '+' || char == '-' || char == '*' || char == '/'
}
