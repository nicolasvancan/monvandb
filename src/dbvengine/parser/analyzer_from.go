package parser

import (
	"fmt"
	"strings"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
	dataframe "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
)

func analyzeFrom(
	db *database.Database,
	stmt sqlparser.TableExprs,
	tablesAlias *map[string]string,
	subqueries *map[string]*AnalyzedQuerySelect,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	joinsColumnFilters *map[string]dataframe.Filters,
) (FromAnalysis, map[string]JoinAnalysis, error) {

	fromAnalysis := FromAnalysis{}
	joinsOn := make(map[string]JoinAnalysis)

	i := 0
	// From statements always come with only one element in the array
	if len(stmt) > 0 {
		switch from := stmt[0].(type) {
		case *sqlparser.AliasedTableExpr:

			// Analyze AliasedTableExpr
			analysis, err := analyzeAliasedTableExpr(db, from, tablesAlias, subqueries, tablesColumnComparsions)

			if err != nil {
				return analysis, joinsOn, err
			}

			v, exists := (*tablesColumnComparsions)[analysis.Alias]

			if !exists {
				v = []database.ColumnComparsion{}
			}

			// Full scan
			v = append(v, createColumnComparsionFullScan("*", analysis.Table, analysis.Alias, 0, -1, database.AND, database.AND))

			(*tablesColumnComparsions)[analysis.Alias] = v

			return analysis, joinsOn, nil
		case *sqlparser.JoinTableExpr:
			var err error = nil
			fromAnalysis, err = analyzeJoinExpression(
				db,
				from,
				tablesAlias,
				subqueries,
				tablesColumnComparsions,
				joinsColumnFilters,
				&joinsOn,
				&i,
			)

			if err != nil {
				return fromAnalysis, joinsOn, err
			}

		}
	}

	return fromAnalysis, joinsOn, nil
}

func analyzeJoinExpression(
	db *database.Database,
	expr *sqlparser.JoinTableExpr,
	tablesAlias *map[string]string,
	subqueries *map[string]*AnalyzedQuerySelect,
	tablesColumnComparsions *map[string][]database.ColumnComparsion,
	joinsColumnFilters *map[string]dataframe.Filters,
	joinsOn *map[string]JoinAnalysis,
	i *int,
) (FromAnalysis, error) {
	fromAnalysis := FromAnalysis{}
	joinAnalyzys := JoinAnalysis{}
	// Is this case they will always be two AliasedTableExpr
	leftExpr := expr.LeftExpr
	rightExpr := expr.RightExpr.(*sqlparser.AliasedTableExpr)
	leftAnalysis := FromAnalysis{}
	var err error
	switch leftExpr := leftExpr.(type) {
	case *sqlparser.AliasedTableExpr:
		// Analyze left and right AliasedTableExpr
		leftAnalysis, err = analyzeAliasedTableExpr(db, leftExpr, tablesAlias, subqueries, tablesColumnComparsions)
		if err != nil {
			return FromAnalysis{}, err
		}
	case *sqlparser.JoinTableExpr:
		// Analyze left and right JoinTableExpr
		leftAnalysis, err = analyzeJoinExpression(db, leftExpr, tablesAlias, subqueries, tablesColumnComparsions, joinsColumnFilters, joinsOn, i)

		if err != nil {
			return FromAnalysis{}, err
		}
	}

	rightAnalysis, err := analyzeAliasedTableExpr(db, rightExpr, tablesAlias, subqueries, tablesColumnComparsions)

	if err != nil {
		return FromAnalysis{}, err
	}

	// if i == 0 means that the from clause should be binded to the first table
	if *i == 0 {
		fromAnalysis = leftAnalysis
	}

	// Analyze join
	joinAnalyzys.LeftAlias = leftAnalysis.Alias
	joinAnalyzys.RightAlias = rightAnalysis.Alias
	hash := fmt.Sprintf("%s-%s", leftAnalysis.Alias, rightAnalysis.Alias)
	hash2 := fmt.Sprintf("%s-%s", rightAnalysis.Alias, leftAnalysis.Alias)
	newStr := ""
	newStr = strings.Replace(expr.Join, "join", newStr, -1)
	if strings.Trim(newStr, " ") == "" {
		joinAnalyzys.How = "inner"
	} else {
		joinAnalyzys.How = strings.Trim(newStr, " ")
	}

	// Analyze on
	var joinOn CompExpr
	switch expr.On.(type) {
	case *sqlparser.ComparisonExpr:
		joinOn, err = analyzeComparsionExpr(
			expr.On.(*sqlparser.ComparisonExpr),
			tablesAlias,
			tablesColumnComparsions,
			joinsColumnFilters,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.AND,
			database.AND,
		)
	case *sqlparser.AndExpr:
		joinOn, err = analyzeAndExpr(
			expr.On.(*sqlparser.AndExpr),
			tablesAlias,
			tablesColumnComparsions,
			joinsColumnFilters,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.AND,
			database.AND,
		)
	case *sqlparser.OrExpr:
		joinOn, err = analyzeOrExpr(
			expr.On.(*sqlparser.OrExpr),
			tablesAlias,
			tablesColumnComparsions,
			joinsColumnFilters,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.OR,
			database.AND,
		)
	case *sqlparser.ParenExpr:
		joinOn, err = analyzeParenExpr(
			expr.On.(*sqlparser.ParenExpr),
			tablesAlias,
			tablesColumnComparsions,
			joinsColumnFilters,
			0,  // Indicates that it is the first comparsion
			-1, // No parent,
			database.AND,
			database.AND,
		)
	}

	if err != nil {
		return fromAnalysis, err
	}

	joinAnalyzys.On = joinOn

	// Add backyards as well
	(*joinsOn)[hash] = joinAnalyzys
	(*joinsOn)[hash2] = joinAnalyzys

	return fromAnalysis, nil
}
