package monvan_parser

import "testing"

func TestTokenizer(t *testing.T) {
	query := "SELECT * FROM table"
	tokenizer := NewTokenizer(query)

	token := tokenizer.NextToken()
	if token.Type != TokenKeyword {
		t.Errorf("Expected token type KEYWORD, got %s", token.Type)
	}
	if token.Value != "SELECT" {
		t.Errorf("Expected token value SELECT, got %s", token.Value)
	}

	token = tokenizer.NextToken()
	if token.Type != TokenOperator {
		t.Errorf("Expected token type OPERATOR, got %s", token.Type)
	}

	if token.Value != "*" {
		t.Errorf("Expected token value *, got %v", len(token.Value))
	}

	token = tokenizer.NextToken()
	if token.Type != TokenKeyword {
		t.Errorf("Expected token type KEYWORD, got %s", token.Type)
	}
	if token.Value != "FROM" {
		t.Errorf("Expected token value FROM, got %s", token.Value)
	}

	token = tokenizer.NextToken()
	if token.Type != TokenKeyword {
		t.Errorf("Expected token type KEYWORD, got %s", token.Type)
	}
	if token.Value != "table" {
		t.Errorf("Expected token value table, got %s", token.Value)
	}

	token = tokenizer.NextToken()
	if token.Type != TokenEOF {
		t.Errorf("Expected token type EOF, got %s", token.Type)
	}
}
