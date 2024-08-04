package main

/*import (
    "fmt"
    "testing"
    parser"github.com/nicolasvancan/monvandb/src/dbvengine/parser"
)*/

import (
	"testing"

	p "github.com/nicolasvancan/monvandb/src/dbvengine/parser"
)

func TestGetTokenFromSimpleQuery(t *testing.T) {

	query := "SELECT * FROM table WHERE id = 1"
	tokenized := p.TokenizeQuery(query)

	if len(tokenized) != 8 {
		t.Errorf("Excpected 8 tokens, got %d", len(tokenized))
	}

}

func TestTokenizeFormultipleTabsAndEnters(t *testing.T) {

	query := "SELECT *\t\t\t\t\t\t\t\t\n\n\n\n\n\n FROM\n\n\n\n table\n\n\n\n                 WHERE id = 1"
	tokenized := p.TokenizeQuery(query)

	if len(tokenized) != 8 {
		t.Errorf("Excpected 8 tokens, got %d", len(tokenized))
	}

}
