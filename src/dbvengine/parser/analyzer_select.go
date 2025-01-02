package parser

import (
	"github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
)

func analyzeSelectedColumns(
	db *database.Database,
	stmt sqlparser.SelectExprs,
	tablesAlias *map[string]string,
	subqueries *map[string]*AnalyzedQuerySelect,
	columnComparsions *map[string][]database.ColumnComparsion,
) ([]ColFunction, error) {

	analyzedQuerySelect := make([]ColFunction, 0)

	// Iterate over select columns
	for _, selectExpr := range stmt {
		switch expr := selectExpr.(type) {
		case *sqlparser.AliasedExpr:
			colFunction, err := analyzeAliasedExpr(expr, db, tablesAlias, columnComparsions, subqueries)

			if err != nil {
				return analyzedQuerySelect, err
			}

			analyzedQuerySelect = append(analyzedQuerySelect, colFunction)
		case *sqlparser.StarExpr:
			alias := expr.TableName.Name.String()

			// Verify if alias exist in either subqueries or tablesAlias

			conFunction := ColFunction{
				Alias:  alias,
				Column: "*",
				Func:   "",
				Args:   nil,
			}
			analyzedQuerySelect = append(analyzedQuerySelect, conFunction)
		}
	}

	return analyzedQuerySelect, nil
}
