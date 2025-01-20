package main

import (
	"fmt"
	"strings"
)

func main() {
	/*query := "SELECT t.col1, t.col2, sum(t.col3) soma FROM tab t GROUP BY t.col1, t.col2 HAVING sum(t.col3) > 10 ORDER BY t.col1"
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		panic(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		statement := stmt.SelectExprs[2].(*sqlparser.AliasedExpr).Expr.(*sqlparser.FuncExpr).Exprs[0]
		//statement := stmt.From[0].(*sqlparser.JoinTableExpr).On.(*sqlparser.AndExpr).Right.(*sqlparser.IsExpr).Expr
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Printf("Value %s\n", statement)
	default:
		fmt.Println("Unsupported statement")
	}*/
	teste := "bang"
	fmt.Println(strings.Split(teste, "."))

}
