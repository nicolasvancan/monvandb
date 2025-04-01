package monvan_parser

import (
	"fmt"
	"testing"
)

// CREATE DATABASE
// CREATE INDEX
// CREATE ROLE
// CREATE USER
// CREATE
func TestLexer(t *testing.T) {
	query := "REVOKE ROLE x FROM y"
	lexer := NewLexer(query)

	bang, err, shouldPass := lexer.Tokenize()
	fmt.Println(bang, err, shouldPass)
	t.Error("Test failed")
}
