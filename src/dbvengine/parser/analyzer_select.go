package parser

import (
	"fmt"

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

func analyzeOrderBy(db *database.Database, stmt sqlparser.OrderBy, tablesAlias *map[string]string) ([]ColFunction, error) {
	analyzedQuerySelect := make([]ColFunction, 0)

	// Iterate over order by columns
	for _, orderByExpr := range stmt {
		expr := orderByExpr.Expr.(*sqlparser.ColName)

		tmp := ColFunction{}
		// Verify if column exists in the database
		// Validate if table exists in context
		tabRef := expr.Qualifier.Name.String()

		if tabRef == "" {
			foundCol := false
			// The column verifying proccess when there it no tabRef is done by getting all columns
			// from all tables and checking whether or not they exist
			for _, tabName := range *tablesAlias {
				if columnExists(db, tabName, expr.Name.String()) {
					foundCol = true
				}
			}

			if !foundCol {
				return analyzedQuerySelect, fmt.Errorf("column %s does not exist in query context", expr.Name.String())
			}
		}

		tmp.Column = expr.Name.String()
		tmp.Alias = expr.Name.String()
		tmp.Func = ""
		tmp.Args = nil

		analyzedQuerySelect = append(analyzedQuerySelect, tmp)
	}

	return analyzedQuerySelect, nil
}
