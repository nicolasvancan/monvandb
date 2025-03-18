package parser

import (
	"fmt"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
	dataframe "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
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

// ColFunction represents a column
type ColFunction struct {
	Alias  string
	Column string
	Func   string
	Args   []interface{}
}

/*
Interface Types
*/
type AnalyzedQueryData interface {
	String() string
	Error() error
}

type AnalyzedQuerySelect struct {
	DatabaseName            string
	Select                  []ColFunction
	Distinct                bool
	GroupBy                 []ColFunction
	Having                  []CompExpr
	TablesAlias             map[string]string
	Subqueries              map[string]*AnalyzedQuerySelect
	TablesColumnComparsions map[string][]database.ColumnComparsion
	TablesFilters           map[string]dataframe.Filters
	Joins                   map[string]JoinAnalysis
	From                    FromAnalysis
	JoinsFilters            map[string]dataframe.Filters
	Order                   []ColFunction
	Asc                     bool
	Limit                   int
	err                     error
}

// TODO: Create constraints
type AnalyzedQueryCreateTable struct {
	DatabaseName    string
	TableName       string
	Columns         []database.Column
	Truncate        bool
	VerifyExistence bool
	err             error
}

type AnalyzedQueryInsert struct {
	DatabaseName string
	TableName    string
	Columns      []string
	Values       interface{}
	err          error
}

type UpdateSet struct {
	Column string
	Value  interface{}
}

type AnalyzedQueryUpdate struct {
	DatabaseName            string
	Set                     []UpdateSet
	TablesAlias             map[string]string
	Subqueries              map[string]*AnalyzedQuerySelect
	TablesColumnComparsions map[string][]database.ColumnComparsion
	TablesFilters           map[string]dataframe.Filters
	Joins                   map[string]JoinAnalysis
	TableName               FromAnalysis
	JoinsFilters            map[string]dataframe.Filters
	err                     error
}

type AnalyzedQueryDelete struct {
	DatabaseName            string
	TablesAlias             map[string]string
	Subqueries              map[string]*AnalyzedQuerySelect
	TablesColumnComparsions map[string][]database.ColumnComparsion
	TablesFilters           map[string]dataframe.Filters
	Joins                   map[string]JoinAnalysis
	TableName               FromAnalysis
	JoinsFilters            map[string]dataframe.Filters
	err                     error
}

type AnalyzedQueryDropTable struct {
	DatabaseName string
	TableName    string
	err          error
}

type AnalyzedQueryTruncateTable struct {
	DatabaseName string
	TableName    string
	err          error
}

type AnalyzedQueryAlterTable struct {
	DatabaseName string
	TableName    string
	Columns      []database.Column
	err          error
}

type AnalyzedQueryCreateDatabase struct {
	DatabaseName string
	err          error
}

type AnalyzedQueryDropDatabase struct {
	DatabaseName string
	err          error
}

type AnalyzedQueryShowDatabases struct {
	err error
}

type AnalyzedQueryShowTables struct {
	DatabaseName string
	err          error
}

/* End interface Types*/

func NewAnalyzedQuerySelect() *AnalyzedQuerySelect {
	return &AnalyzedQuerySelect{
		TablesAlias:             make(map[string]string),
		Distinct:                false,
		GroupBy:                 make([]ColFunction, 0),
		Having:                  make([]CompExpr, 0),
		Subqueries:              make(map[string]*AnalyzedQuerySelect),
		TablesColumnComparsions: make(map[string][]database.ColumnComparsion),
		TablesFilters:           make(map[string]dataframe.Filters),
		Joins:                   make(map[string]JoinAnalysis),
		From:                    FromAnalysis{},
		JoinsFilters:            make(map[string]dataframe.Filters),
		Order:                   make([]ColFunction, 0),
		Limit:                   -1,
		Asc:                     false,
		err:                     nil,
	}
}

func (a AnalyzedQuerySelect) String() string {
	return fmt.Sprintf("AnalyzedQuerySelect{DatabaseName: Select %v\n %s\n From %v\n TablesAlias: %v\n Subqueries: %v\n TablesColumnComparsions: %v\n TablesFilters: %v\n Joins: %v\n JoinsFilters: %v\n Error: %v\n", a.Select, a.DatabaseName, a.From, a.TablesAlias, a.Subqueries, a.TablesColumnComparsions, a.TablesFilters, a.Joins, a.JoinsFilters, a.Error())
}

func (a *AnalyzedQuerySelect) Error() error {
	return a.err
}

func NewAnalyzedQueryCreateTable() *AnalyzedQueryCreateTable {
	return &AnalyzedQueryCreateTable{
		Columns:         make([]database.Column, 0),
		Truncate:        false,
		VerifyExistence: false,
		err:             nil,
	}
}

func (a AnalyzedQueryCreateTable) String() string {
	return fmt.Sprintf("AnalyzedQueryCreateTable{DatabaseName: %s\n TableName: %s\n Columns: %v\n Truncate: %v\n VerifyExistence: %v\n Error: %v\n", a.DatabaseName, a.TableName, a.Columns, a.Truncate, a.VerifyExistence, a.Error())
}

func (a *AnalyzedQueryCreateTable) Error() error {
	return a.err
}

func NewAnalyzedQueryInsert() *AnalyzedQueryInsert {
	return &AnalyzedQueryInsert{
		Columns: make([]string, 0),
		Values:  nil,
		err:     nil,
	}
}

func (a *AnalyzedQueryInsert) String() string {
	return fmt.Sprintf("AnalyzedQueryInsert{DatabaseName: %s\n TableName: %s\n Columns: %v\n Values: %v\n", a.DatabaseName, a.TableName, a.Columns, a.Values)
}

func (a *AnalyzedQueryInsert) Error() error {
	return a.err
}

func NewAnalyzedQueryUpdate() *AnalyzedQueryUpdate {
	return &AnalyzedQueryUpdate{
		Set:                     make([]UpdateSet, 0),
		TablesAlias:             make(map[string]string),
		Subqueries:              make(map[string]*AnalyzedQuerySelect),
		TablesColumnComparsions: make(map[string][]database.ColumnComparsion),
		TablesFilters:           make(map[string]dataframe.Filters),
		Joins:                   make(map[string]JoinAnalysis),
		TableName:               FromAnalysis{},
		JoinsFilters:            make(map[string]dataframe.Filters),
	}
}

func (a *AnalyzedQueryUpdate) String() string {
	return fmt.Sprintf("AnalyzedQueryUpdate{DatabaseName: %v\n TableName: %v\n Set: %v\n TablesAlias: %v\n Subqueries: %v\n TablesColumnComparsions: %v\n TablesFilters: %v\n Joins: %v\n From: %v\n JoinsFilters: %v\n", a.DatabaseName, a.TableName, a.Set, a.TablesAlias, a.Subqueries, a.TablesColumnComparsions, a.TablesFilters, a.Joins, a.TableName, a.JoinsFilters)
}

func (a *AnalyzedQueryUpdate) Error() error {
	return a.err
}

func NewAnalyzedQueryDelete() *AnalyzedQueryDelete {
	return &AnalyzedQueryDelete{
		TablesAlias:             make(map[string]string),
		Subqueries:              make(map[string]*AnalyzedQuerySelect),
		TablesColumnComparsions: make(map[string][]database.ColumnComparsion),
		TablesFilters:           make(map[string]dataframe.Filters),
		Joins:                   make(map[string]JoinAnalysis),
		TableName:               FromAnalysis{},
		JoinsFilters:            make(map[string]dataframe.Filters),
	}
}

func (a *AnalyzedQueryDelete) String() string {
	return fmt.Sprintf("AnalyzedQueryDelete{DatabaseName: %v\n TableName: %v\n TablesAlias: %v\n Subqueries: %v\n TablesColumnComparsions: %v\n TablesFilters: %v\n Joins: %v\n From: %v\n JoinsFilters: %v\n", a.DatabaseName, a.TableName, a.TablesAlias, a.Subqueries, a.TablesColumnComparsions, a.TablesFilters, a.Joins, a.TableName, a.JoinsFilters)
}

func (a *AnalyzedQueryDelete) Error() error {
	return a.err
}

func NewAnalyzedQueryDropTable() *AnalyzedQueryDropTable {
	return &AnalyzedQueryDropTable{
		err: nil,
	}
}

func (a *AnalyzedQueryDropTable) String() string {
	return fmt.Sprintf("AnalyzedQueryDropTable{DatabaseName: %s\n TableName: %s\n", a.DatabaseName, a.TableName)
}

func (a *AnalyzedQueryDropTable) Error() error {
	return a.err
}

func NewAnalyzedQueryAlterTable() *AnalyzedQueryAlterTable {
	return &AnalyzedQueryAlterTable{
		Columns: make([]database.Column, 0),
		err:     nil,
	}
}

func (a *AnalyzedQueryAlterTable) String() string {
	return fmt.Sprintf("AnalyzedQueryAlterTable{DatabaseName: %s\n TableName: %s\n Columns: %v\n", a.DatabaseName, a.TableName, a.Columns)
}

func (a *AnalyzedQueryAlterTable) Error() error {
	return a.err
}

func NewAnalyzedQueryCreateDatabase() *AnalyzedQueryCreateDatabase {
	return &AnalyzedQueryCreateDatabase{
		err: nil,
	}
}

func (a *AnalyzedQueryCreateDatabase) String() string {
	return fmt.Sprintf("AnalyzedQueryCreateDatabase{DatabaseName: %s\n", a.DatabaseName)
}

func (a *AnalyzedQueryCreateDatabase) Error() error {
	return a.err
}

func NewAnalyzedQueryDropDatabase() *AnalyzedQueryDropDatabase {
	return &AnalyzedQueryDropDatabase{
		err: nil,
	}
}

func (a *AnalyzedQueryDropDatabase) String() string {
	return fmt.Sprintf("AnalyzedQueryDropDatabase{DatabaseName: %s\n", a.DatabaseName)
}

func (a *AnalyzedQueryDropDatabase) Error() error {
	return a.err
}

func NewAnalyzedQueryTruncateTable() *AnalyzedQueryTruncateTable {
	return &AnalyzedQueryTruncateTable{
		err: nil,
	}
}

func (a *AnalyzedQueryTruncateTable) String() string {
	return fmt.Sprintf("AnalyzedQueryTruncateTable{DatabaseName: %s\n TableName: %s\n", a.DatabaseName, a.TableName)
}

func (a *AnalyzedQueryTruncateTable) Error() error {
	return a.err
}

func NewAnalyzedQueryShowDatabases() *AnalyzedQueryShowDatabases {
	return &AnalyzedQueryShowDatabases{
		err: nil,
	}
}

func (a *AnalyzedQueryShowDatabases) String() string {
	return "AnalyzedQueryShowDatabases{}"
}

func (a *AnalyzedQueryShowDatabases) Error() error {
	return a.err
}

func NewAnalyzedQueryShowTables() *AnalyzedQueryShowTables {
	return &AnalyzedQueryShowTables{
		err: nil,
	}
}

func (a *AnalyzedQueryShowTables) String() string {
	return fmt.Sprintf("AnalyzedQueryShowTables{DatabaseName: %s\n", a.DatabaseName)
}

func AnalyzeQuery(databaseName string, parsedQuery interface{}) AnalyzedQueryData {
	var analyzedData AnalyzedQueryData
	switch stmt := parsedQuery.(type) {
	case *sqlparser.Select:
		analyzedData = analyzeSelect(databaseName, stmt)
	case *sqlparser.CreateTable:
		analyzedData = analyzeCreateTable(databaseName, stmt)
	case *sqlparser.Insert:
		analyzedData = analyzeInsert(databaseName, stmt)
	case *sqlparser.Update:
		analyzedData = analyzeTableRowsUpdate(databaseName, stmt)
	case *sqlparser.Delete:
		analyzedData = analyzeTableRowsDelete(databaseName, stmt)
	case *sqlparser.TruncateTable:
		analyzedData = analyzeTruncateTable(databaseName, stmt)
	case *sqlparser.DDL:
		switch stmt.Action {
		case sqlparser.DropStr:
			analyzedData = analyzeDropTable(databaseName, stmt)
		case sqlparser.AlterStr:
			analyzedData = analyzeAlterTable(databaseName, stmt)
		default:
			analyzedData = nil
		}

	default:
		analyzedData = nil
	}

	return analyzedData
}
