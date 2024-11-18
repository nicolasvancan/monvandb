package parser

import (
	"fmt"
	"strings"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
	dataframe "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

type SqlStatementType int

const (
	Select SqlStatementType = iota
	Insert
	Update
	Delete
	CreateTable
	DropTable
	AlterTable
	CreateDatabase
	DropDatabase
	ShowDatabases
	ShowTables
	ShowColumns
	ShowIndex
)

type FromAnalysis struct {
	Alias      string
	Table      string
	IsSubQuery bool
}

type JoinAnalysis struct {
	LeftAlias  string
	RightAlias string
	How        string
	On         JoinOn
}

type JoinOn struct {
	LeftColumn    string
	LeftAlias     string
	LeftFunction  ColFunction
	RightColumn   string
	RightAlias    string
	RightFunction ColFunction
	Operator      string
}

type ColFunction struct {
	Alias string
	Func  string
	Args  []string
}

type AnalyzedQueryData interface {
	String() string
	Error() error
}

type AnalyzedQuerySelect struct {
	DatabaseName            string
	TablesAlias             map[string]string
	Subqueries              map[string]AnalyzedQueryData
	TablesColumnComparsions map[string][]database.ColumnComparsion
	TablesFilters           map[string]dataframe.Filter
	Joins                   map[string]JoinAnalysis
	From                    FromAnalysis
	JoinsFilers             map[string]dataframe.Filter
	err                     error
}

func NewAnalyzedQuerySelect() *AnalyzedQuerySelect {
	return &AnalyzedQuerySelect{
		TablesAlias:             make(map[string]string),
		Subqueries:              make(map[string]AnalyzedQueryData),
		TablesColumnComparsions: make(map[string][]database.ColumnComparsion),
		TablesFilters:           make(map[string]dataframe.Filter),
		Joins:                   make(map[string]JoinAnalysis),
		From:                    FromAnalysis{},
		JoinsFilers:             make(map[string]dataframe.Filter),
		err:                     nil,
	}
}

func (a *AnalyzedQuerySelect) String() string {
	return fmt.Sprintf("AnalyzedQuerySelect{DatabaseName: %s, TablesAlias: %v, Subqueries: %v, TablesColumnComparsions: %v, TablesFilters: %v, Joins: %v, JoinsFilers: %v, Error: %v}", a.DatabaseName, a.TablesAlias, a.Subqueries, a.TablesColumnComparsions, a.TablesFilters, a.Joins, a.JoinsFilers, a.Error())
}

func (a *AnalyzedQuerySelect) Error() error {
	return a.err
}

func AnalyzeQuery(databaseName string, parsedQuery sqlparser.Statement) AnalyzedQueryData {
	var analyzedData AnalyzedQueryData
	switch stmt := parsedQuery.(type) {
	case *sqlparser.Select:
		analyzedData = analyzeSelect(databaseName, stmt)
	default:
	}
	return analyzedData
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
	db, err := getDatabase(databaseName)

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

	return analyzedQuerySelect
}

func analyzeAliasedTableExpr(
	db *database.Database,
	expr *sqlparser.AliasedTableExpr,
	tablesAlias *map[string]string,
	subqueries *map[string]AnalyzedQueryData,
) (FromAnalysis, error) {

	tableExpr := expr.Expr
	fromAnalysis := FromAnalysis{}

	switch tableExpr := tableExpr.(type) {
	case *sqlparser.Subquery:
		// Analyze subquery
		subquery := AnalyzeQuery(db.Name, tableExpr.Select)

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

func analyzeFrom(
	db *database.Database,
	stmt sqlparser.TableExprs,
	tablesAlias *map[string]string,
	subqueries *map[string]AnalyzedQueryData,
	TablesColumnComparsions *map[string][]database.ColumnComparsion,
	_ *map[string]dataframe.Filter,
) (FromAnalysis, map[string]JoinAnalysis, error) {

	fromAnalysis := FromAnalysis{}
	joinsOn := make(map[string]JoinAnalysis)

	for i, from := range stmt {
		switch from := from.(type) {
		case *sqlparser.AliasedTableExpr:

			// Analyze AliasedTableExpr
			analysis, err := analyzeAliasedTableExpr(db, from, tablesAlias, subqueries)

			if err != nil {
				return analysis, joinsOn, err
			}

			return analysis, joinsOn, nil
		case *sqlparser.JoinTableExpr:
			joinAnalyzys := JoinAnalysis{}
			// Is this case they will always be two AliasedTableExpr
			leftExpr := from.LeftExpr.(*sqlparser.AliasedTableExpr)
			rightExpr := from.RightExpr.(*sqlparser.AliasedTableExpr)

			// Analyze left and right AliasedTableExpr
			leftAnalysis, err := analyzeAliasedTableExpr(db, leftExpr, tablesAlias, subqueries)
			if err != nil {
				return leftAnalysis, joinsOn, err
			}

			rightAnalysis, err := analyzeAliasedTableExpr(db, rightExpr, tablesAlias, subqueries)

			if err != nil {
				return rightAnalysis, joinsOn, err
			}

			// if i == 0 means that the from clause should be binded to the first table
			if i == 0 {
				fromAnalysis = leftAnalysis
			}

			// Analyze join
			joinAnalyzys.LeftAlias = leftAnalysis.Alias
			joinAnalyzys.RightAlias = rightAnalysis.Alias
			hash := fmt.Sprintf("%s-%s", leftAnalysis.Alias, rightAnalysis.Alias)
			newStr := ""
			newStr = strings.Replace(from.Join, "join", newStr, -1)
			if strings.Trim(newStr, " ") == "" {
				joinAnalyzys.How = "inner"
			} else {
				joinAnalyzys.How = strings.Trim(newStr, " ")
			}

			// Analyze on
			joinOn := JoinOn{}

			// TODO: Implementing when there are multiple joins conditions
			switch on := from.On.(type) {
			case *sqlparser.AndExpr:
			case *sqlparser.OrExpr:
			case *sqlparser.ComparisonExpr:
				// This is the normal case where there is only one condition
				joinOn.Operator = on.Operator

				joinOn.LeftColumn = on.Left.(*sqlparser.ColName).Name.String()
				joinOn.LeftAlias = on.Left.(*sqlparser.ColName).Qualifier.Name.String()
				joinOn.RightColumn = on.Right.(*sqlparser.ColName).Name.String()
				joinOn.RightAlias = on.Right.(*sqlparser.ColName).Qualifier.Name.String()

				// Fillup columns comparsions
				v, exists := (*TablesColumnComparsions)[leftAnalysis.Alias]

				if !exists {
					v = []database.ColumnComparsion{}
				}

				v = append(v, database.ColumnComparsion{
					ColumnName: joinOn.LeftColumn,
					TableName:  (*tablesAlias)[joinOn.LeftAlias],
					Alias:      joinOn.LeftAlias,
					Condition:  database.EQ,
					Value: database.ColumnConditionValue{
						IsOtherColumn:        true,
						IsOtherTable:         true,
						ColumnName:           (*tablesAlias)[joinOn.RightAlias],
						TableHash:            joinOn.LeftAlias,
						Value:                nil,
						Transformation:       nil,
						TransformationParams: nil,
					},
					ParentId:        -1,
					ParentLogicalOp: database.AND,
					LayerLogicalOp:  database.AND,
					Id:              0,
				},
				)
				(*TablesColumnComparsions)[leftAnalysis.Alias] = v

				v, exists = (*TablesColumnComparsions)[rightAnalysis.Alias]

				if !exists {
					v = []database.ColumnComparsion{}
				}

				v = append(v, database.ColumnComparsion{
					ColumnName: joinOn.RightColumn,
					TableName:  (*tablesAlias)[joinOn.RightAlias],
					Alias:      joinOn.RightAlias,
					Condition:  database.EQ,
					Value: database.ColumnConditionValue{
						IsOtherColumn:        true,
						IsOtherTable:         true,
						ColumnName:           (*tablesAlias)[joinOn.RightAlias],
						TableHash:            joinOn.RightAlias,
						Value:                nil,
						Transformation:       nil,
						TransformationParams: nil,
					},
					ParentId:        -1,
					ParentLogicalOp: database.AND,
					LayerLogicalOp:  database.AND,
					Id:              0,
				},
				)

				(*TablesColumnComparsions)[rightAnalysis.Alias] = v
			default:
				return fromAnalysis, joinsOn, fmt.Errorf("unsupported join condition")
			}

			joinAnalyzys.On = joinOn
			joinsOn[hash] = joinAnalyzys
		}
	}
	return fromAnalysis, joinsOn, nil
}

func getDatabase(databaseName string) (*database.Database, error) {
	databasePath := utils.GetPath("databases")
	db, err := database.LoadDatabase(databasePath + "/" + strings.ToLower(databaseName))
	if err != nil {
		return nil, fmt.Errorf("database %s does not exist", databaseName)
	}
	return db, nil
}
