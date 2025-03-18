package main

import (
	"fmt"
	"reflect"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

type bang struct {
	foo string
}

type bang2 struct {
	bang
	Teste string
}

func main() {
	query := "TRUNCATE TABLE teste"
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
	case *sqlparser.CreateTable:
		statement := stmt.NewName.ToViewName().Name
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Println(statement)
	case *sqlparser.Insert:
		statement := stmt.Rows
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Println(statement)
	case *sqlparser.Update:
		statement := stmt.Exprs[0].Expr
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Println(statement)
	case *sqlparser.Delete:
		statement := stmt.TableExprs
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Println(statement)
	case *sqlparser.DDL:
		statement := stmt
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Println(statement)
	default:
		statement := stmt
		fmt.Printf("%s\n", reflect.TypeOf(statement))
		fmt.Println(statement)
	}
}
