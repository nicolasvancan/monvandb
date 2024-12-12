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
	JoinAlias  string
	LeftAlias  string
	RightAlias string
	How        string
	On         CompExpr
}

type CompExpr struct {
	LeftType   string // column, value, function
	LeftAlias  string
	LeftValue  interface{}
	RightType  string
	RightAlias string
	RightValue interface{}
	Operator   string
}

type ColFunction struct {
	Alias  string
	Column string
	Func   string
	Args   []interface{}
}

type AnalyzedQueryData interface {
	String() string
	Error() error
}

type AnalyzedQuerySelect struct {
	DatabaseName            string
	TablesAlias             map[string]string
	Subqueries              map[string]*AnalyzedQuerySelect
	TablesColumnComparsions map[string][]database.ColumnComparsion
	TablesFilters           map[string]dataframe.Filters
	Joins                   map[string]JoinAnalysis
	From                    FromAnalysis
	JoinsFilters            map[string]dataframe.Filters
	err                     error
}

func (a AnalyzedQuerySelect) String() string {
	return fmt.Sprintf("AnalyzedQuerySelect{DatabaseName: %s\n From %v\n TablesAlias: %v\n Subqueries: %v\n TablesColumnComparsions: %v\n TablesFilters: %v\n Joins: %v\n JoinsFilters: %v\n Error: %v\n", a.DatabaseName, a.From, a.TablesAlias, a.Subqueries, a.TablesColumnComparsions, a.TablesFilters, a.Joins, a.JoinsFilters, a.Error())
}

func NewAnalyzedQuerySelect() *AnalyzedQuerySelect {
	return &AnalyzedQuerySelect{
		TablesAlias:             make(map[string]string),
		Subqueries:              make(map[string]*AnalyzedQuerySelect),
		TablesColumnComparsions: make(map[string][]database.ColumnComparsion),
		TablesFilters:           make(map[string]dataframe.Filters),
		Joins:                   make(map[string]JoinAnalysis),
		From:                    FromAnalysis{},
		JoinsFilters:            make(map[string]dataframe.Filters),
		err:                     nil,
	}
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

	// Analyze Where
	if stmt.Where != nil {
		err = analyzeWhere(
			stmt.Where,
			&analyzedQuerySelect.TablesAlias,
			&analyzedQuerySelect.TablesColumnComparsions,
			&analyzedQuerySelect.TablesFilters,
		)
		if err != nil {
			analyzedQuerySelect.err = err
		}
	}

	return analyzedQuerySelect
}

func getDatabase(databaseName string) (*database.Database, error) {
	databasePath := utils.GetPath("databases")
	db, err := database.LoadDatabase(databasePath + "/" + strings.ToLower(databaseName))
	if err != nil {
		return nil, fmt.Errorf("database %s does not exist", databaseName)
	}
	return db, nil
}
