package main

import (
	"fmt"
	"reflect"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

func main() {
	stmt, err := sqlparser.Parse("SELECT tab1.col1 as coluna1 from tab1 as b left join (SELECT * FROM X) as c on CAST(b.col1 AS DATE) = c.col1 AND (b.col2 = c.col2 OR b.col3 = c.col3)")
	if err != nil {
		panic(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		statement := stmt.From[0].(*sqlparser.JoinTableExpr).On.(*sqlparser.AndExpr).Left.(*sqlparser.ComparisonExpr).Left.(*sqlparser.FuncExpr).Exprs[0].(*sqlparser.AliasedExpr).Expr
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Printf("Value %v\n", statement)
	default:
		fmt.Println("Unsupported statement")
	}

}
