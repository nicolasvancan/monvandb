package main

import (
	"fmt"
	"reflect"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

func main() {
	query := "SELECT * FROM tab1 WHERE tab1.col = 1"
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		panic(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		statement := stmt.Where.Expr
		//statement := stmt.From[0].(*sqlparser.JoinTableExpr).On.(*sqlparser.AndExpr).Right.(*sqlparser.IsExpr).Expr
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Printf("Value %v\n", statement)
	default:
		fmt.Println("Unsupported statement")
	}

}
