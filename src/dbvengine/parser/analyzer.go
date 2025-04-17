package parser

import (
	"fmt"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
	dataframe "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
	monvan_parser "github.com/nicolasvancan/monvandb/src/dbvengine/parser/custom_parser"
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
	AlterType    string
	ColumnName   string
	Column       database.Column
	Options      string
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

type AnalyzedQueryDropIndex struct {
	DatabaseName string
	IndexName    string
	TableName    string
	err          error
}

type AnalyzedQueryDropUser struct {
	Username string
	err      error
}

type AnalyzedQueryDropRole struct {
	RoleName string
	err      error
}

type AnalyzedQueryCreateUser struct {
	User     string
	Username string
	Password string
	err      error
}

type AnalyzedQueryCreateRole struct {
	RoleName string
	Options  map[string]interface{}
	err      error
}

type AnalyzedQueryCreateIndex struct {
	DatabaseName string
	IndexName    string
	TableName    string
	Columns      []string
	err          error
}

type AnalyzedQueryUseDatabase struct {
	DatabaseName string
	err          error
}

type AnalyzedQueryModifyUser struct {
	Username string
	Password string
	err      error
}

func NewAnalyzedQueryModifyUser() *AnalyzedQueryModifyUser {
	return &AnalyzedQueryModifyUser{
		Username: "",
		Password: "",
		err:      nil,
	}
}

func (a *AnalyzedQueryModifyUser) String() string {
	return fmt.Sprintf("AnalyzedQueryUseDatabase{Username: %s\nPassword: %s\n}", a.Username, a.Password)
}

func (a *AnalyzedQueryModifyUser) Error() error {
	return a.err
}

func NewAnalyzedQueryUseDatabase() *AnalyzedQueryUseDatabase {
	return &AnalyzedQueryUseDatabase{
		DatabaseName: "",
		err:          nil,
	}
}

func (a *AnalyzedQueryUseDatabase) String() string {
	return fmt.Sprintf("AnalyzedQueryUseDatabase{DatabaseName: %s\n", a.DatabaseName)
}

func (a *AnalyzedQueryUseDatabase) Error() error {
	return a.err
}

func NewAnalyzedQueryCreateRole() *AnalyzedQueryCreateRole {
	return &AnalyzedQueryCreateRole{
		RoleName: "",
		Options:  make(map[string]interface{}),
		err:      nil,
	}
}

func (a *AnalyzedQueryCreateRole) String() string {
	return fmt.Sprintf("AnalyzedQueryCreateRole{RoleName: %s\n Options: %v\n", a.RoleName, a.Options)
}

func (a *AnalyzedQueryCreateRole) Error() error {
	return a.err
}

func NewAnalyzedQueryCreateIndex() *AnalyzedQueryCreateIndex {
	return &AnalyzedQueryCreateIndex{
		DatabaseName: "",
		IndexName:    "",
		TableName:    "",
		Columns:      make([]string, 0),
		err:          nil,
	}
}

func (a *AnalyzedQueryCreateIndex) String() string {
	return fmt.Sprintf("AnalyzedQueryCreateIndex{DatabaseName: %s\n IndexName: %s\n TableName: %s\n Columns: %v\n", a.DatabaseName, a.IndexName, a.TableName, a.Columns)
}

func (a *AnalyzedQueryCreateIndex) Error() error {
	return a.err
}

func NewAnalyzedQueryCreateUser() *AnalyzedQueryCreateUser {
	return &AnalyzedQueryCreateUser{
		User:     "",
		Username: "",
		Password: "",
		err:      nil,
	}
}

func (a *AnalyzedQueryCreateUser) String() string {
	return fmt.Sprintf("AnalyzedQueryCreateUser{User: %s\n Username: %s\n Password: %s\n", a.User, a.Username, a.Password)
}

func (a *AnalyzedQueryCreateUser) Error() error {
	return a.err
}

func NewAnalyzedQueryDropRole() *AnalyzedQueryDropRole {
	return &AnalyzedQueryDropRole{
		RoleName: "",
		err:      nil,
	}
}

func (a *AnalyzedQueryDropRole) String() string {
	return fmt.Sprintf("AnalyzedQueryDropRole{Username: %s\n", a.RoleName)
}

func (a *AnalyzedQueryDropRole) Error() error {
	return a.err
}

/* End interface Types*/
func NewAnalyzedQueryDropUser() *AnalyzedQueryDropUser {
	return &AnalyzedQueryDropUser{
		Username: "",
		err:      nil,
	}
}

func (a *AnalyzedQueryDropUser) String() string {
	return fmt.Sprintf("AnalyzedQueryDropUser{Username: %s\n", a.Username)
}

func (a *AnalyzedQueryDropUser) Error() error {
	return a.err
}

func NewAnalyzedQueryDropIndex() *AnalyzedQueryDropIndex {
	return &AnalyzedQueryDropIndex{
		DatabaseName: "",
		IndexName:    "",
		TableName:    "",
		err:          nil,
	}
}

func (a *AnalyzedQueryDropIndex) String() string {
	return fmt.Sprintf("AnalyzedQueryDropIndex{DatabaseName: %s\n IndexName: %s\n TableName: %s\n", a.DatabaseName, a.IndexName, a.TableName)
}

func (a *AnalyzedQueryDropIndex) Error() error {
	return a.err
}

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
		Column: database.Column{},
		err:    nil,
	}
}

func (a *AnalyzedQueryAlterTable) String() string {
	return fmt.Sprintf("AnalyzedQueryAlterTable{DatabaseName: %s\n TableName: %s\n Columns: %v\n", a.DatabaseName, a.TableName, a.Column)
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

func (a *AnalyzedQueryShowTables) Error() error {
	return a.err
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
	case *monvan_parser.DatabaseDrop:
		analyzedData = analyzeDropDatabase(stmt)
	case *monvan_parser.TableDrop:
		analyzedData = analyzeDropTable(databaseName, stmt)
	case *monvan_parser.IndexDrop:
		analyzedData = analyzeDropIndex(databaseName, stmt)
	case *monvan_parser.UserDrop:
		analyzedData = analyzeDropUser(stmt)
	case *monvan_parser.RoleDrop:
		analyzedData = analyzeDropRole(stmt)
	case *monvan_parser.DatabaseCreate:
		analyzedData = analyzeCreateDatabase(stmt)
	case *monvan_parser.UserCreate:
		analyzedData = analyzeCreateUser(stmt)
	case *monvan_parser.RoleCreate:
		analyzedData = analyzeCreateRole(stmt)
	case *monvan_parser.IndexCreate:
		analyzedData = analyzeCreateIndex(databaseName, stmt)
	case *monvan_parser.UserModify:
		analyzedData = analyzeModifyUser(stmt)
	case *monvan_parser.TableAlter:
		analyzedData = analyzeAlterTable(databaseName, stmt)
	case *monvan_parser.UseDatabase:
		analyzedData = analyzeUseDatabase(stmt)
	case *monvan_parser.ShowDatabases:
		analyzedData = NewAnalyzedQueryShowDatabases()
	case *monvan_parser.ShowTables:
		analyzedData = analyzeShowTables(databaseName, stmt)

	default:
		analyzedData = nil
	}

	return analyzedData
}
