package parser

import (
	"fmt"
	"math/rand"
	"strconv"

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
		aliasedExpr.Column = exprr.Name.Lowered()
		aliasedExpr.Alias = expr.As.Lowered()
		aliasedExpr.Func = ""
		aliasedExpr.Args = nil

		// Verify if the column exists
	case *sqlparser.FuncExpr:
		function, err := analyzeFuncExpr(exprr, columnComparsions)

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
					anFun, err := analyzeFuncExpr(aliasedExpr, tablesColumnComparsions)

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
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	id int,
	parentId int,
	_ int,
	parentLogicalLayer int,
) (CompExpr, error) {
	compExpr := CompExpr{}
	left := expr.Left
	right := expr.Right
	var err error
	switch l := left.(type) {
	case *sqlparser.ParenExpr:
		compExpr, err = analyzeParenExpr(
			l,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			database.AND,
		)
	case *sqlparser.ComparisonExpr:
		compExpr, err = analyzeComparsionExpr(
			l,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
		)
	case *sqlparser.IsExpr:
		compExpr, err = analyzeIsExpr(
			l,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
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
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			database.AND,
		)
	case *sqlparser.ComparisonExpr:
		_, err = analyzeComparsionExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
		)
	case *sqlparser.AndExpr:
		_, err = analyzeAndExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
		)
	case *sqlparser.OrExpr:
		_, err = analyzeOrExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
		)
	case *sqlparser.IsExpr:
		_, err = analyzeIsExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
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
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	id int,
	parentId int,
	_ int,
	parentLogicalLayer int,
) (CompExpr, error) {
	compExpr := CompExpr{}
	left := expr.Left
	right := expr.Right

	var err error
	switch l := left.(type) {
	case *sqlparser.ParenExpr:
		compExpr, err = analyzeParenExpr(
			l,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			database.OR,
		)
	case *sqlparser.ComparisonExpr:
		compExpr, err = analyzeComparsionExpr(
			l,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.OR,
			parentLogicalLayer,
		)
	case *sqlparser.IsExpr:
		compExpr, err = analyzeIsExpr(
			l,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
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
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			database.AND,
		)
	case *sqlparser.ComparisonExpr:
		_, err = analyzeComparsionExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.AND,
			parentLogicalLayer,
		)
	case *sqlparser.AndExpr:
		_, err = analyzeAndExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.OR,
			parentLogicalLayer,
		)
	case *sqlparser.OrExpr:
		_, err = analyzeOrExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			database.OR,
			parentLogicalLayer,
		)
	case *sqlparser.IsExpr:
		_, err = analyzeIsExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
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
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	id int,
	parentId int,
	logicalLayer int,
	parentLogicalLayer int,
) (CompExpr, error) {
	compExpr := CompExpr{}
	var err error
	innerExpr := expr.Expr
	randomId := rand.Intn(1000)
	switch innerExpr := innerExpr.(type) {
	case *sqlparser.AndExpr:
		compExpr, err = analyzeAndExpr(
			innerExpr,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			randomId,
			id,
			database.AND,
			parentLogicalLayer,
		)
	case *sqlparser.OrExpr:
		compExpr, err = analyzeOrExpr(
			innerExpr,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			randomId,
			id,
			database.AND,
			parentLogicalLayer,
		)
	case *sqlparser.ParenExpr:
		compExpr, err = analyzeParenExpr(
			innerExpr,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			randomId,
			id,
			database.AND,
			logicalLayer,
		)
	case *sqlparser.ComparisonExpr:
		compExpr, err = analyzeComparsionExpr(
			innerExpr,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			randomId,
			id,
			database.AND,
			parentLogicalLayer,
		)
	case *sqlparser.IsExpr:
		compExpr, err = analyzeIsExpr(
			innerExpr,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
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
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
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
	tablesAlias *map[string]string,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	tableFilters *map[string]dataframe.Filters,
	id int,
	parentId int,
	logicalLayer int,
	parentLogicalLayer int,
) (CompExpr, error) {
	compExpr := CompExpr{}

	compExpr.Operator = expr.Operator

	left := expr.Left
	right := expr.Right

	switch l := left.(type) {
	case *sqlparser.ColName:
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
		function, err := analyzeFuncExpr(l, tablesColumnComparsions)
		compExpr.LeftValue = function
		if err != nil {
			return compExpr, err
		}
	}

	switch r := right.(type) {
	case *sqlparser.ColName:
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
		function, err := analyzeFuncExpr(r, tablesColumnComparsions)

		if err != nil {
			return compExpr, err
		}

		compExpr.RightValue = function
	case *sqlparser.AndExpr:
		cmpExpr, err := analyzeAndExpr(
			r,
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			logicalLayer,
			parentLogicalLayer,
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
			tablesAlias,
			tablesColumnComparsions,
			tableFilters,
			id,
			parentId,
			logicalLayer,
			parentLogicalLayer,
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
