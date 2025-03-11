package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
)

/*
	Contains Analyzer functions for CRUD operations
	Such as insert, update and delete
*/

func AnalyseSubSelect(
	databaseName string,
	stmt *sqlparser.Select,
	columnComparsions *map[string][]database.ColumnComparsion,
) *AnalyzedQuerySelect {
	return nil
}

func getTable(database *database.Database, tableName string) (*database.Table, error) {
	table, err := database.GetTable(tableName)
	if err != nil {
		return nil, fmt.Errorf("table %s does not exist", tableName)
	}
	return table, nil
}

func analyzeSelect(databaseName string, stmt *sqlparser.Select) *AnalyzedQuerySelect {
	// Iterate over from clause
	analyzedQuerySelect := NewAnalyzedQuerySelect()
	analyzedQuerySelect.DatabaseName = databaseName

	// Verify database
	db, err := database.GetDatabase(databaseName)

	if err != nil {
		analyzedQuerySelect.err = err
		return analyzedQuerySelect
	}

	// Verify from clause
	fromAnalyzis, joinsOn, err := analyzeFrom(
		db,
		stmt.From,
		&analyzedQuerySelect.TablesAlias,
		&analyzedQuerySelect.Subqueries,
		&analyzedQuerySelect.TablesColumnComparsions,
		&analyzedQuerySelect.TablesFilters,
	)

	if err != nil {
		analyzedQuerySelect.err = err
		return analyzedQuerySelect
	}

	analyzedQuerySelect.From = fromAnalyzis
	analyzedQuerySelect.Joins = joinsOn

	// Analyze Where
	if stmt.Where != nil {
		err = analyzeWhere(
			stmt.Where,
			db,
			&analyzedQuerySelect.TablesAlias,
			&analyzedQuerySelect.TablesColumnComparsions,
			&analyzedQuerySelect.TablesFilters,
			&analyzedQuerySelect.Subqueries,
		)
		if err != nil {
			analyzedQuerySelect.err = err
		}
	}

	// Distinct
	analyzedQuerySelect.Distinct = stmt.Distinct != ""

	// Group by
	analyzedQuerySelect.GroupBy = analyzeGroupBy(stmt.GroupBy)

	// Having
	// Limit
	if stmt.Limit != nil {
		if stmt.Limit.Rowcount != nil {
			limitValue, err := strconv.Atoi(string(stmt.Limit.Rowcount.(*sqlparser.SQLVal).Val))
			if err != nil {
				analyzedQuerySelect.err = fmt.Errorf("invalid limit value: %v", err)
				return analyzedQuerySelect
			}
			analyzedQuerySelect.Limit = limitValue
		}
	}

	// Order
	if stmt.OrderBy != nil {
		analyzedQuerySelect.Order, analyzedQuerySelect.Asc, err = analyzeOrderBy(
			db,
			stmt.OrderBy,
			&analyzedQuerySelect.TablesAlias,
		)

		if err != nil {
			analyzedQuerySelect.err = err
			return analyzedQuerySelect
		}
	}

	// Select
	analyzedQuerySelect.Select, err = analyzeSelectedColumns(
		db,
		stmt.SelectExprs,
		&analyzedQuerySelect.TablesAlias,
		&analyzedQuerySelect.Subqueries,
		&analyzedQuerySelect.TablesColumnComparsions,
	)

	if err != nil {
		analyzedQuerySelect.err = err
	}

	return analyzedQuerySelect
}

/*
	Insert statement analyzer
*/

func analyzeInsert(databaseName string, stmt *sqlparser.Insert) *AnalyzedQueryInsert {
	analyzedQueryInsert := NewAnalyzedQueryInsert()
	analyzedQueryInsert.DatabaseName = databaseName

	// Verify database
	db, err := database.GetDatabase(databaseName)

	if err != nil {
		analyzedQueryInsert.err = err
		return analyzedQueryInsert
	}

	// Verify table
	table, err := getTable(db, stmt.Table.Name.String())
	if err != nil {
		analyzedQueryInsert.err = err
		return analyzedQueryInsert
	}

	analyzedQueryInsert.TableName = stmt.Table.Name.String()

	// Columns
	columns, err := analyzeInsertColumns(stmt.Columns, table)
	if err != nil {
		analyzedQueryInsert.err = err
		return analyzedQueryInsert
	}

	analyzedQueryInsert.Columns = columns

	// Values
	values, err := analyzeInsertValues(databaseName, columns, stmt.Rows, table)

	if err != nil {
		analyzedQueryInsert.err = err
		return analyzedQueryInsert
	}

	analyzedQueryInsert.Values = values

	return analyzedQueryInsert
}

func analyzeInsertColumns(columns []sqlparser.ColIdent, table *database.Table) ([]string, error) {
	columnsNames := make([]string, 0)
	for _, col := range columns {
		columnName := col.String()
		if table.GetColumnByName(columnName) == nil {
			return nil, fmt.Errorf("column %s does not exist", columnName)
		}
		columnsNames = append(columnsNames, columnName)
	}
	return columnsNames, nil
}

func analyzeInsertValues(
	databaseName string,
	columns []string,
	rows sqlparser.InsertRows,
	table *database.Table,
) (interface{}, error) {

	switch rowsValues := rows.(type) {
	default:
		values := make([]database.RawRow, 0)

		for _, row := range rowsValues.(sqlparser.Values) {
			valuesRow := make(database.RawRow)
			for i, value := range row {
				columnName := columns[i]
				val, err := analyzeInsertValue(value, table, columnName)

				if err != nil {
					return nil, err
				}

				valuesRow[columnName] = val
			}

			values = append(values, valuesRow)
		}
		return values, nil
	case *sqlparser.Select:
		return analyzeSelect(databaseName, rowsValues), nil
	}
}

// Todo work on it
func analyzeInsertValue(value sqlparser.Expr, table *database.Table, columnName string) (interface{}, error) {
	// Get column type
	column := table.GetColumnByName(columnName)
	if column == nil {
		return nil, fmt.Errorf("column %s does not exist", columnName)
	}

	// Assume all values are from SQLVal
	val := value.(*sqlparser.SQLVal)

	switch column.Type {
	case database.COL_TYPE_INT, database.COL_TYPE_BIG_INT, database.COL_TYPE_SMALL_INT:
		intValue, err := strconv.Atoi(string(val.Val))
		if err != nil {
			return nil, fmt.Errorf("invalid value for column %s: %v", columnName, err)
		}
		return intValue, nil
	case database.COL_TYPE_FLOAT, database.COL_TYPE_DOUBLE:
		floatValue, err := strconv.ParseFloat(string(val.Val), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value for column %s: %v", columnName, err)
		}

		return floatValue, nil
	case database.COL_TYPE_BOOL:
		boolValue, err := strconv.ParseBool(string(val.Val))
		if err != nil {
			return nil, fmt.Errorf("invalid value for column %s: %v", columnName, err)
		}
		return boolValue, nil
	default:
		return string(val.Val), nil
	}
}

func analyzeTableRowsUpdate(databaseName string, stmt *sqlparser.Update) *AnalyzedQueryUpdate {
	analyzedQueryUpdate := NewAnalyzedQueryUpdate()
	analyzedQueryUpdate.DatabaseName = databaseName

	// Verify database
	db, err := database.GetDatabase(databaseName)
	if err != nil {
		analyzedQueryUpdate.err = err
		return analyzedQueryUpdate
	}

	// Verify from clause
	fromAnalyzis, joinsOn, err := analyzeFrom(
		db,
		stmt.TableExprs,
		&analyzedQueryUpdate.TablesAlias,
		&analyzedQueryUpdate.Subqueries,
		&analyzedQueryUpdate.TablesColumnComparsions,
		&analyzedQueryUpdate.TablesFilters,
	)

	if err != nil {
		analyzedQueryUpdate.err = err
		return analyzedQueryUpdate
	}

	analyzedQueryUpdate.TableName = fromAnalyzis
	analyzedQueryUpdate.Joins = joinsOn

	// Analyze Where
	if stmt.Where != nil {
		err = analyzeWhere(
			stmt.Where,
			db,
			&analyzedQueryUpdate.TablesAlias,
			&analyzedQueryUpdate.TablesColumnComparsions,
			&analyzedQueryUpdate.TablesFilters,
			&analyzedQueryUpdate.Subqueries,
		)
		if err != nil {
			analyzedQueryUpdate.err = err
		}
	}

	// Analyze Sets
	set, err := analyzeUpdateSet(
		db,
		stmt.Exprs,
		&analyzedQueryUpdate.TablesAlias,
	)

	analyzedQueryUpdate.Set = set

	return analyzedQueryUpdate
}

func analyzeUpdateSet(
	db *database.Database,
	exprs sqlparser.UpdateExprs,
	tablesAlias *map[string]string,
) ([]UpdateSet, error) {

	sets := make([]UpdateSet, 0)
	var err error

	for _, expr := range exprs {
		updateExpr := expr
		columnName := updateExpr.Name.Name.String()
		columnAlias := updateExpr.Name.Qualifier.Name.String()

		// Verify if column exists in table
		if columnAlias != "" {
			if _, ok := (*tablesAlias)[columnAlias]; !ok {
				return nil, fmt.Errorf("table %s does not exist", columnAlias)
			}

			table, err := getTable(db, (*tablesAlias)[columnAlias])

			if err != nil {
				return nil, err
			}

			column := table.GetColumnByName(strings.ToLower(columnName))

			if column == nil {
				return nil, fmt.Errorf("column %s does not exist in table %s", columnName, columnAlias)
			}
		}

		var val interface{}

		switch exprVal := updateExpr.Expr.(type) {
		case *sqlparser.SQLVal:
			val, err = getValFromSQLVal(exprVal)

			if err != nil {
				return nil, err
			}
		case *sqlparser.ColName:
			val = ColFunction{
				Column: expr.Name.Name.String(),
				Alias:  expr.Name.Qualifier.Name.String(),
			}
		default:
			return nil, fmt.Errorf("invalid value for column %s", columnName)
		}

		set := UpdateSet{
			Column: strings.ToLower(columnName),
			Value:  val,
		}

		sets = append(sets, set)
	}

	return sets, nil
}

func analyzeTableRowsDelete(databaseName string, stmt *sqlparser.Delete) *AnalyzedQueryDelete {
	analyzedQueryDelete := NewAnalyzedQueryDelete()
	analyzedQueryDelete.DatabaseName = databaseName

	// Verify database
	db, err := database.GetDatabase(databaseName)
	if err != nil {
		analyzedQueryDelete.err = err
		return analyzedQueryDelete
	}

	// Verify from clause
	fromAnalyzis, joinsOn, err := analyzeFrom(
		db,
		stmt.TableExprs,
		&analyzedQueryDelete.TablesAlias,
		&analyzedQueryDelete.Subqueries,
		&analyzedQueryDelete.TablesColumnComparsions,
		&analyzedQueryDelete.TablesFilters,
	)

	if err != nil {
		analyzedQueryDelete.err = err
		return analyzedQueryDelete
	}

	analyzedQueryDelete.TableName = fromAnalyzis
	analyzedQueryDelete.Joins = joinsOn

	// Analyze Where
	if stmt.Where != nil {
		err = analyzeWhere(
			stmt.Where,
			db,
			&analyzedQueryDelete.TablesAlias,
			&analyzedQueryDelete.TablesColumnComparsions,
			&analyzedQueryDelete.TablesFilters,
			&analyzedQueryDelete.Subqueries,
		)
		if err != nil {
			analyzedQueryDelete.err = err
		}
	}

	return analyzedQueryDelete
}
