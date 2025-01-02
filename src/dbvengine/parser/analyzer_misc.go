package parser

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
	dataframe "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
)

func analyzeAliasedExpr(
	expr *sqlparser.AliasedExpr,
	db *database.Database,
	tablesAlias *map[string]string,
	columnComparsions *map[string][]database.ColumnComparsion,
	subqueries *map[string]*AnalyzedQuerySelect,
) (ColFunction, error) {
	aliasedExpr := ColFunction{}

	switch exprr := expr.Expr.(type) {
	case *sqlparser.ColName:
		// Validate if table exists in context
		tabRef := exprr.Qualifier.Name.String()

		if tabRef == "" {
			foundCol := false
			// The column verifying proccess when there it no tabRef is done by getting all columns
			// from all tables and checking whether or not they exist
			for _, tabName := range *tablesAlias {
				if columnExists(db, tabName, exprr.Name.String()) {
					foundCol = true
				}
			}

			if !foundCol {
				return aliasedExpr, fmt.Errorf("column %s does not exist in query context", exprr.Name.String())
			}

		} else {
			tabName := (*tablesAlias)[exprr.Qualifier.Name.String()]
			if !tableOrSubQueryAliasExistInContext(tablesAlias, subqueries, tabRef) {
				return aliasedExpr, fmt.Errorf("table or subquery alias %s does not exist in query context", tabRef)
			}
			if !columnExists(db, tabName, exprr.Name.String()) {
				return aliasedExpr, fmt.Errorf("column %s does not exist in table %s", exprr.Name.String(), tabName)
			}
		}

		aliasedExpr.Column = exprr.Name.Lowered()
		aliasedExpr.Alias = expr.As.Lowered()
		aliasedExpr.Func = ""
		aliasedExpr.Args = nil

		// Verify if the column exists
	case *sqlparser.FuncExpr:
		function, err := analyzeFuncExpr(exprr, db, tablesAlias, subqueries, columnComparsions)

		if err != nil {
			return aliasedExpr, err
		}

		aliasedExpr = function
	case *sqlparser.Subquery:
		// Analyze subquery
		analyzedSubquery := AnalyseSubSelect(
			db.Name,
			exprr.Select.(*sqlparser.Select),
			columnComparsions,
		)

		if analyzedSubquery.Error() != nil {
			return aliasedExpr, analyzedSubquery.Error()
		}

		(*subqueries)[expr.As.String()] = analyzedSubquery
	}

	return aliasedExpr, nil
}

func analyzeAliasedTableExpr(
	db *database.Database,
	expr *sqlparser.AliasedTableExpr,
	tablesAlias *map[string]string,
	subqueries *map[string]*AnalyzedQuerySelect,
	columnComparsions *map[string][]database.ColumnComparsion,
) (FromAnalysis, error) {

	tableExpr := expr.Expr
	fromAnalysis := FromAnalysis{}

	switch tableExpr := tableExpr.(type) {
	case *sqlparser.Subquery:
		// Analyze subquery
		subquery := AnalyseSubSelect(db.Name, tableExpr.Select.(*sqlparser.Select), columnComparsions)

		if subquery.Error() != nil {
			return fromAnalysis, subquery.Error()
		}

		(*subqueries)[expr.As.String()] = subquery
		fromAnalysis.Alias = expr.As.String()
		fromAnalysis.IsSubQuery = true

	case sqlparser.TableName:

		// Table exists?
		tableName := tableExpr.Name.String()
		_, err := getTable(db, tableName)

		if err != nil {
			return fromAnalysis, err
		}

		// Verify if it has an alias
		fromAnalysis.Table = tableName
		fromAnalysis.IsSubQuery = false
		if expr.As.String() == "" {
			(*tablesAlias)[tableName] = tableName
			fromAnalysis.Alias = tableName
		} else {
			(*tablesAlias)[expr.As.String()] = tableName
			fromAnalysis.Alias = expr.As.String()

		}

	}

	return fromAnalysis, nil
}

/*
Instruções próximos passos:

para completar o analyzeComparsion é necessário adicionar aos tableColumnComparsions todos os campos que estão sendo comparados
e para isso, criarei funções que preencham esses campos passando nomes de colunas, valores, etc; para evitar códigos imensos
*/

func getValFromSQLVal(val *sqlparser.SQLVal) (interface{}, error) {
	switch val.Type {
	case sqlparser.StrVal:
		return string(val.Val), nil
	case sqlparser.IntVal:
		intValue, err := strconv.Atoi(string(val.Val))
		if err != nil {
			return nil, err
		}
		return intValue, nil
	case sqlparser.FloatVal:
		floatValue, err := strconv.ParseFloat(string(val.Val), 64)
		if err != nil {
			return nil, err
		}
		return floatValue, nil
	default:
		return nil, fmt.Errorf("unsupported value type")
	}
}

func analyzeFuncExpr(
	expr interface{},
	db *database.Database,
	tablesAlias *map[string]string,
	subqueries *map[string]*AnalyzedQuerySelect,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
) (ColFunction, error) {
	function := ColFunction{}

	switch ex := expr.(type) {
	case *sqlparser.FuncExpr:
		// Analyze arguments
		function.Func = ex.Name.String()

		for _, arg := range ex.Exprs {
			switch arg := arg.(type) {
			case *sqlparser.AliasedExpr:
				switch aliasedExpr := arg.Expr.(type) {
				case *sqlparser.ColName:

					// Validate if table exists in context
					tabRef := aliasedExpr.Qualifier.Name.String()

					if tabRef == "" {
						foundCol := false
						// The column verifying proccess when there it no tabRef is done by getting all columns
						// from all tables and checking whether or not they exist
						for _, tabName := range *tablesAlias {
							if columnExists(db, tabName, aliasedExpr.Name.String()) {
								foundCol = true
							}
						}

						if !foundCol {
							return function, fmt.Errorf("column %s does not exist in query context", aliasedExpr.Name.String())
						}

					} else {
						if !tableOrSubQueryAliasExistInContext(tablesAlias, subqueries, tabRef) {
							return function, fmt.Errorf("table or subquery alias %s does not exist in query context", tabRef)
						}

						tabName := (*tablesAlias)[aliasedExpr.Qualifier.Name.String()]
						if !columnExists(db, tabName, aliasedExpr.Name.String()) {
							return function, fmt.Errorf("column %s does not exist in table %s", aliasedExpr.Name.String(), tabName)
						}
					}

					colName := aliasedExpr.Name.String()
					function.Column = colName
					function.Alias = aliasedExpr.Qualifier.Name.String()
					function.Args = append(function.Args, aliasedExpr.Name.String())
				case *sqlparser.SQLVal:
					val, err := getValFromSQLVal(aliasedExpr)

					if err != nil {
						return function, err
					}

					function.Args = append(function.Args, val)
				case *sqlparser.FuncExpr:
					anFun, err := analyzeFuncExpr(aliasedExpr, db, tablesAlias, subqueries, tablesColumnComparsions)

					if err != nil {
						return function, err
					}

					function.Args = append(function.Args, anFun)
				default:
					function.Args = append(function.Args, arg)
				}
			}
		}
	case *sqlparser.ConvertExpr:

	default:
		return function, fmt.Errorf("unsupported function")
	}

	return function, nil
}

func analyzeAndExpr(
	expr *sqlparser.AndExpr,
	db *database.Database,
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	subqueries *map[string]*AnalyzedQuerySelect,
	id int,
	parentId int,
	_ int,
	parentLogicalLayer int,
	joinOn *bool,
) (CompExpr, error) {
	compExpr := CompExpr{}
	left := expr.Left
	right := expr.Right
	var err error
	switch l := left.(type) {
	case *sqlparser.ParenExpr:
		compExpr, err = analyzeParenExpr(
			l,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			database.AND,
			joinOn,
		)
	case *sqlparser.ComparisonExpr:
		compExpr, err = analyzeComparsionExpr(
			l,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.IsExpr:
		compExpr, err = analyzeIsExpr(
			l,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
		)
		// Only in this case we can add the column comparsion and table filter
	}

	switch r := right.(type) {
	case *sqlparser.ParenExpr:
		_, err = analyzeParenExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			database.AND,
			joinOn,
		)
	case *sqlparser.ComparisonExpr:
		_, err = analyzeComparsionExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.AndExpr:
		_, err = analyzeAndExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.OrExpr:
		_, err = analyzeOrExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.IsExpr:
		_, err = analyzeIsExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
		)
	}

	if err != nil {
		return compExpr, err
	}
	return compExpr, nil
}

func analyzeOrExpr(
	expr *sqlparser.OrExpr,
	db *database.Database,
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	subqueries *map[string]*AnalyzedQuerySelect,
	id int,
	parentId int,
	_ int,
	parentLogicalLayer int,
	joinOn *bool,
) (CompExpr, error) {
	compExpr := CompExpr{}
	left := expr.Left
	right := expr.Right

	var err error
	switch l := left.(type) {
	case *sqlparser.ParenExpr:
		compExpr, err = analyzeParenExpr(
			l,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			database.OR,
			joinOn,
		)
	case *sqlparser.ComparisonExpr:
		compExpr, err = analyzeComparsionExpr(
			l,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.OR,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.IsExpr:
		compExpr, err = analyzeIsExpr(
			l,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.OR,
			parentLogicalLayer,
		)
		// Only in this case we can add the column comparsion and table filter
	}

	switch r := right.(type) {
	case *sqlparser.ParenExpr:
		_, err = analyzeParenExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			database.AND,
			joinOn,
		)
	case *sqlparser.ComparisonExpr:
		_, err = analyzeComparsionExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.AndExpr:
		_, err = analyzeAndExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.OR,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.OrExpr:
		_, err = analyzeOrExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.OR,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.IsExpr:
		_, err = analyzeIsExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.OR,
			parentLogicalLayer,
		)
	}

	if err != nil {
		return compExpr, err
	}
	return compExpr, nil
}

func analyzeParenExpr(
	expr *sqlparser.ParenExpr,
	db *database.Database,
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	subqueries *map[string]*AnalyzedQuerySelect,
	id int,
	parentId int,
	logicalLayer int,
	parentLogicalLayer int,
	joinOn *bool,
) (CompExpr, error) {
	compExpr := CompExpr{}
	var err error
	innerExpr := expr.Expr
	randomId := rand.Intn(1000)
	switch innerExpr := innerExpr.(type) {
	case *sqlparser.AndExpr:
		compExpr, err = analyzeAndExpr(
			innerExpr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			randomId,
			id,
			database.AND,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.OrExpr:
		compExpr, err = analyzeOrExpr(
			innerExpr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			randomId,
			id,
			database.AND,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.ParenExpr:
		compExpr, err = analyzeParenExpr(
			innerExpr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			randomId,
			id,
			database.AND,
			logicalLayer,
			joinOn,
		)
	case *sqlparser.ComparisonExpr:
		compExpr, err = analyzeComparsionExpr(
			innerExpr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			randomId,
			id,
			database.AND,
			parentLogicalLayer,
			joinOn,
		)
	case *sqlparser.IsExpr:
		compExpr, err = analyzeIsExpr(
			innerExpr,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
		)
		// Only in this case we can add the column comparsion and table filter
	}

	if err != nil {
		return compExpr, err
	}

	return compExpr, nil
}

func analyzeIsExpr(
	expr *sqlparser.IsExpr,
	db *database.Database,
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	subqueries *map[string]*AnalyzedQuerySelect,
	id int,
	parentId int,
	logicalLayer int,
	parentLogicalLayer int,
) (CompExpr, error) {

	compExpr := CompExpr{}
	var val interface{}
	var op string

	switch expr.Operator {
	case sqlparser.IsNullStr:
		val = nil
		op = "="
	case sqlparser.IsNotNullStr:
		val = nil
		op = "!="
	case sqlparser.IsFalseStr:
		val = false
		op = "="
	case sqlparser.IsNotFalseStr:
		val = false
		op = "!="
	case sqlparser.IsTrueStr:
		val = true
		op = "="
	case sqlparser.IsNotTrueStr:
		val = true
		op = "!="
	}

	// The only type here is colname
	switch v := expr.Expr.(type) {
	case *sqlparser.ColName:
		compExpr.LeftType = "column"
		compExpr.LeftValue = v.Name.String()
		compExpr.LeftAlias = v.Qualifier.Name.String()
		compExpr.RightType = "value"
		compExpr.RightValue = val
		compExpr.Operator = op
	default:
		return compExpr, fmt.Errorf("unsupported type")
	}

	// Add to the column comparsions
	comparsions := createColumnComparsionFromComprExpr(
		compExpr,
		tablesAlias,
		id,
		parentId,
		logicalLayer,
		parentLogicalLayer,
	)

	(*tablesColumnComparsions)[comparsions[0].Alias] = append((*tablesColumnComparsions)[comparsions[0].Alias], comparsions[0])

	// Create filter node based on expr
	createOrUpdateFilterBasedOnCompExpr(
		compExpr,
		tableFilters,
		id,
		parentId,
		logicalLayer,
		parentLogicalLayer,
	)

	return compExpr, nil
}

func analyzeComparsionExpr(
	expr *sqlparser.ComparisonExpr,
	db *database.Database,
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	subqueries *map[string]*AnalyzedQuerySelect,
	id int,
	parentId int,
	logicalLayer int,
	parentLogicalLayer int,
	joinOn *bool, // Controls insertion of filter for columns after ON, should not insert
) (CompExpr, error) {
	compExpr := CompExpr{}

	compExpr.Operator = expr.Operator

	left := expr.Left
	right := expr.Right

	switch l := left.(type) {
	case *sqlparser.ColName:
		// Validate if table exists in context
		tabRef := l.Qualifier.Name.String()

		if tabRef == "" {
			foundCol := false
			// The column verifying proccess when there it no tabRef is done by getting all columns
			// from all tables and checking whether or not they exist
			for _, tabName := range *tablesAlias {
				if columnExists(db, tabName, l.Name.String()) {
					foundCol = true
				}
			}

			if !foundCol {
				return compExpr, fmt.Errorf("column %s does not exist in query context", l.Name.String())
			}

		} else {
			tabName := (*tablesAlias)[l.Qualifier.Name.String()]
			if !tableOrSubQueryAliasExistInContext(tablesAlias, subqueries, tabRef) {
				return compExpr, fmt.Errorf("table or subquery alias %s does not exist in query context", tabRef)
			}
			if !columnExists(db, tabName, l.Name.String()) {
				return compExpr, fmt.Errorf("column %s does not exist in table %s", l.Name.String(), tabName)
			}
		}

		compExpr.LeftType = "column"
		compExpr.LeftValue = l.Name.String()
		compExpr.LeftAlias = l.Qualifier.Name.String()
	case *sqlparser.SQLVal:
		compExpr.LeftType = "value"
		val, err := getValFromSQLVal(l)

		if err != nil {
			return compExpr, err
		}

		compExpr.LeftValue = val
	case *sqlparser.FuncExpr:
		compExpr.LeftType = "function"
		function, err := analyzeFuncExpr(l, db, tablesAlias, subqueries, tablesColumnComparsions)
		compExpr.LeftValue = function
		if err != nil {
			return compExpr, err
		}
	}

	switch r := right.(type) {
	case *sqlparser.ColName:

		tabRef := r.Qualifier.Name.String()

		if tabRef == "" {
			foundCol := false
			// The column verifying proccess when there it no tabRef is done by getting all columns
			// from all tables and checking whether or not they exist
			for _, tabName := range *tablesAlias {
				if columnExists(db, tabName, r.Name.String()) {
					foundCol = true
				}
			}

			if !foundCol {
				return compExpr, fmt.Errorf("column %s does not exist in query context", r.Name.String())
			}

		} else {
			tabName := (*tablesAlias)[r.Qualifier.Name.String()]
			if !tableOrSubQueryAliasExistInContext(tablesAlias, subqueries, tabRef) {
				return compExpr, fmt.Errorf("table or subquery alias %s does not exist in query context", tabRef)
			}
			if !columnExists(db, tabName, r.Name.String()) {
				return compExpr, fmt.Errorf("column %s does not exist in table %s", r.Name.String(), tabName)
			}
		}

		compExpr.RightType = "column"
		compExpr.RightValue = r.Name.String()
		compExpr.RightAlias = r.Qualifier.Name.String()
	case *sqlparser.SQLVal:
		compExpr.RightType = "value"

		val, err := getValFromSQLVal(r)

		if err != nil {
			return compExpr, err
		}

		compExpr.RightValue = val
	case *sqlparser.FuncExpr:
		compExpr.RightType = "function"
		function, err := analyzeFuncExpr(r, db, tablesAlias, subqueries, tablesColumnComparsions)

		if err != nil {
			return compExpr, err
		}

		compExpr.RightValue = function
	case *sqlparser.AndExpr:
		cmpExpr, err := analyzeAndExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			logicalLayer,
			parentLogicalLayer,
			joinOn,
		)

		if err != nil {
			return compExpr, err
		}
		// It is expected to always be a column
		compExpr.RightType = cmpExpr.LeftType
		compExpr.RightValue = cmpExpr.LeftValue
		compExpr.RightAlias = cmpExpr.LeftAlias

	case *sqlparser.OrExpr:
		cmpExpr, err := analyzeOrExpr(
			r,
			db,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			subqueries,
			id,
			parentId,
			logicalLayer,
			parentLogicalLayer,
			joinOn,
		)

		if err != nil {
			return compExpr, err
		}
		// It is expected to always be a column
		compExpr.RightType = cmpExpr.LeftType
		compExpr.RightValue = cmpExpr.LeftValue
		compExpr.RightAlias = cmpExpr.LeftAlias
	default:
		return compExpr, fmt.Errorf("unsupported right type, it should be a column or a value got expression. Use the expression after the column or in the where condition")
	}

	// Add to the column comparsions
	comparsions := createColumnComparsionFromComprExpr(
		compExpr,
		tablesAlias,
		id,
		parentId,
		logicalLayer,
		parentLogicalLayer,
	)

	// Add columns comparsions
	(*tablesColumnComparsions)[comparsions[0].Alias] = append((*tablesColumnComparsions)[comparsions[0].Alias], comparsions[0])
	if len(comparsions) == 2 {
		(*tablesColumnComparsions)[comparsions[1].Alias] = append((*tablesColumnComparsions)[comparsions[1].Alias], comparsions[1])
	}

	if !(*joinOn) {
		createOrUpdateFilterBasedOnCompExpr(
			compExpr,
			tableFilters,
			id,
			parentId,
			logicalLayer,
			parentLogicalLayer,
		)
	}

	// After first call it must be false for further filters
	*joinOn = false

	return compExpr, nil
}

func analyzeGroupBy(stmt sqlparser.GroupBy) []ColFunction {

	colFunctions := make([]ColFunction, 0)
	// For every group by statement
	for _, by := range stmt {
		// They are all ColName pointer
		colName := by.(*sqlparser.ColName)
		colFunction := ColFunction{}
		colFunction.Column = colName.Name.String()
		colFunction.Alias = colName.Qualifier.Name.String()
		colFunction.Func = ""
		colFunction.Args = nil
		colFunctions = append(colFunctions, colFunction)
	}

	return colFunctions
}

func columnExists(db *database.Database, tableName string, column string) bool {
	table, err := db.GetTable(strings.ToLower(tableName))

	if err != nil {
		return false
	}

	col := table.GetColumnByName(column)

	return col != nil
}

func tableOrSubQueryAliasExistInContext(
	tablesAlias *map[string]string,
	subqueries *map[string]*AnalyzedQuerySelect,
	tableAlias string,
) bool {
	_, subqueryExists := (*subqueries)[tableAlias]
	_, tableAliasExists := (*tablesAlias)[tableAlias]

	return subqueryExists || tableAliasExists
}
