package parser

import (
	"fmt"
	"reflect"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
	dataframe "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
)

func analyzeWhere(
	stmt *sqlparser.Where,
	db *database.Database,
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tablesColumnFilters *map[string]dataframe.Filters,
	subqueries *map[string]*AnalyzedQuerySelect,
) error {
	jOn := false
	var err error = nil
	switch expr := stmt.Expr.(type) {
	case *sqlparser.ComparisonExpr:

		_, err = analyzeComparsionExpr(
			expr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tablesColumnFilters,
			subqueries,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.AND,
			database.AND,
			&jOn,
		)
	case *sqlparser.AndExpr:
		_, err = analyzeAndExpr(
			expr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tablesColumnFilters,
			subqueries,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.AND,
			database.AND,
			&jOn,
		)
	case *sqlparser.OrExpr:
		_, err = analyzeOrExpr(
			expr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tablesColumnFilters,
			subqueries,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.OR,
			database.AND,
			&jOn,
		)
	case *sqlparser.ParenExpr:
		_, err = analyzeParenExpr(
			expr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tablesColumnFilters,
			subqueries,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.AND,
			database.AND,
			&jOn,
		)
	case *sqlparser.IsExpr:
		_, err = analyzeIsExpr(
			expr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tablesColumnFilters,
			subqueries,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.AND,
			database.AND,
		)
	default:
		fmt.Printf("%s\n", reflect.TypeOf(expr))
		fmt.Printf("Value %v\n", expr)
	}

	if err != nil {
		return err
	}

	return nil
}
