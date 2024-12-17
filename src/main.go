package main

import (
	"fmt"
	"reflect"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

func main() {
	query := "SELECT DISTINCT tab1.col + 2 FROM tab1 WHERE tab1.col = 1 GROUP BY tab1.col"
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		panic(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		statement := stmt.SelectExprs[0].(*sqlparser.AliasedExpr).Expr.(*sqlparser.BinaryExpr).Left.(*sqlparser.ColName).Name
		//statement := stmt.From[0].(*sqlparser.JoinTableExpr).On.(*sqlparser.AndExpr).Right.(*sqlparser.IsExpr).Expr
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Printf("Value %v\n", statement)
	default:
		fmt.Println("Unsupported statement")
	}

}
