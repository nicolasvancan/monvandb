package parser

/*
	This file contains all table modifications analyzer functions
	such as Create Table, Update Table, Drop Table, Alter Table, and so on
*/
import (
	"errors"
	"strconv"
	"strings"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
	monvan_parser "github.com/nicolasvancan/monvandb/src/dbvengine/parser/custom_parser"
)

func analyzeCreateTable(databaseName string, stmt *sqlparser.CreateTable) *AnalyzedQueryCreateTable {
	analyzedQueryCreateTable := NewAnalyzedQueryCreateTable()
	analyzedQueryCreateTable.DatabaseName = strings.ToLower(databaseName)
	analyzedQueryCreateTable.TableName = strings.ToLower(stmt.NewName.Name.String())
	analyzedQueryCreateTable.Columns = analyzeCreateTableColumns(stmt.Columns)

	return analyzedQueryCreateTable
}

func analyzeCreateTableColumns(colsDef []*sqlparser.ColumnDef) []database.Column {
	columns := make([]database.Column, 0)
	for _, colDef := range colsDef {
		column := database.Column{
			Name: colDef.Name,
		}
		column.AutoIncrement = false
		column.Nullable = true
		column.Primary = false
		column.Default = nil

		// Column type
		splitedType := strings.Split(colDef.Type, "(")
		varType := strings.ToLower(splitedType[0])

		switch varType {
		case "int":
			column.Type = database.COL_TYPE_INT
		case "bigint":
			column.Type = database.COL_TYPE_BIG_INT
		case "smallint":
			column.Type = database.COL_TYPE_SMALL_INT
		case "double":
			column.Type = database.COL_TYPE_DOUBLE
		case "float":
			column.Type = database.COL_TYPE_FLOAT
		case "varchar", "text", "char":
			column.Type = database.COL_TYPE_STRING
		case "tinyint", "bool":
			column.Type = database.COL_TYPE_BOOL
		case "timestamp", "date", "datetime":
			column.Type = database.COL_TYPE_TIMESTAMP
		default:
			column.Type = database.COL_TYPE_STRING
		}

		// Get options
		for _, option := range colDef.Options {
			switch strings.ToLower(option.Type.String()) {
			case "not null":
				column.Nullable = false
			case "null":
				column.Nullable = true
			case "auto_increment":
				column.AutoIncrement = true
			case "primary key":
				column.Primary = true
			case "default":
				// Cast to default value if not null
				if option.Value != "null" {
					switch column.Type {
					case database.COL_TYPE_INT, database.COL_TYPE_BIG_INT, database.COL_TYPE_SMALL_INT:
						column.Default, _ = strconv.Atoi(option.Value)
					case database.COL_TYPE_DOUBLE, database.COL_TYPE_FLOAT:
						column.Default, _ = strconv.ParseFloat(option.Value, 64)
					case database.COL_TYPE_BOOL:
						column.Default, _ = strconv.ParseBool(option.Value)
					default:
						column.Default = option.Value
					}
				}
			}
		}
		columns = append(columns, column)
	}
	return columns
}

func analyzeDropTable(databaseName string, stmt *monvan_parser.TableDrop) *AnalyzedQueryDropTable {

	// Verify if table exists in database
	analyzedQueryCreateDatabase := NewAnalyzedQueryDropTable()

	db, err := database.GetDatabase(databaseName)

	if err != nil {
		analyzedQueryCreateDatabase.err = err
		return analyzedQueryCreateDatabase
	}

	if !database.TableExists(db, strings.ToLower(stmt.Name)) {
		analyzedQueryCreateDatabase.err = errors.New("table " + strings.ToLower(stmt.Name) + " does not exist")
		return analyzedQueryCreateDatabase
	}
	analyzedQueryCreateDatabase.DatabaseName = databaseName
	analyzedQueryCreateDatabase.TableName = strings.ToLower(stmt.Name)
	return analyzedQueryCreateDatabase
}

func analyzeDropIndex(databaseName string, stmt *monvan_parser.IndexDrop) *AnalyzedQueryDropIndex {
	analyzedQueryDropIndex := NewAnalyzedQueryDropIndex()
	analyzedQueryDropIndex.DatabaseName = databaseName
	analyzedQueryDropIndex.TableName = strings.ToLower(stmt.Table)
	analyzedQueryDropIndex.IndexName = strings.ToLower(stmt.Name)

	// Verify if index exists in database
	db, err := database.GetDatabase(databaseName)

	if err != nil {
		analyzedQueryDropIndex.err = err
		return analyzedQueryDropIndex
	}

	if !database.IndexExists(db, analyzedQueryDropIndex.TableName, analyzedQueryDropIndex.IndexName) {
		analyzedQueryDropIndex.err = errors.New("index " + analyzedQueryDropIndex.IndexName + " does not exist")
		return analyzedQueryDropIndex
	}

	return analyzedQueryDropIndex
}

func analyzeAlterTable(databaseName string, stmt *monvan_parser.TableAlter) *AnalyzedQueryAlterTable {
	analyzedQueryAlterTable := NewAnalyzedQueryAlterTable()

	db, err := database.GetDatabase(databaseName)
	if err != nil {
		analyzedQueryAlterTable.err = err
		return analyzedQueryAlterTable
	}

	tableName := strings.ToLower(stmt.Name)
	table, err := db.GetTable(tableName)

	if err != nil {
		analyzedQueryAlterTable.err = errors.New("table " + tableName + " does not exist")
		return analyzedQueryAlterTable
	}

	analyzedQueryAlterTable.DatabaseName = databaseName
	analyzedQueryAlterTable.TableName = strings.ToLower(stmt.Name)
	analyzedQueryAlterTable.Column = *table.GetColumnByName(stmt.Column)
	analyzedQueryAlterTable.AlterType = stmt.Operation
	analyzedQueryAlterTable.ColumnName = stmt.Column
	analyzedQueryAlterTable.Options = stmt.ModifyOptions

	return analyzedQueryAlterTable
}

func analyzeTruncateTable(databaseName string, stmt *sqlparser.TruncateTable) *AnalyzedQueryTruncateTable {
	analyzedQueryTruncateTable := NewAnalyzedQueryTruncateTable()

	db, err := database.GetDatabase(databaseName)
	if err != nil {
		analyzedQueryTruncateTable.err = err
		return analyzedQueryTruncateTable
	}

	tableName := strings.ToLower(stmt.Table.Name.String())
	if !database.TableExists(db, tableName) {
		analyzedQueryTruncateTable.err = errors.New("table " + tableName + " does not exist")
		return analyzedQueryTruncateTable
	}

	analyzedQueryTruncateTable.DatabaseName = databaseName
	analyzedQueryTruncateTable.TableName = strings.ToLower(stmt.Table.Name.String())
	return analyzedQueryTruncateTable
}

func analyzeDropDatabase(stmt *monvan_parser.DatabaseDrop) *AnalyzedQueryDropDatabase {
	analyzedQueryDropDatabase := NewAnalyzedQueryDropDatabase()

	// verify if database exists
	_, err := database.GetDatabase(strings.ToLower(stmt.Name))
	if err != nil {
		analyzedQueryDropDatabase.err = err
		return analyzedQueryDropDatabase
	}
	analyzedQueryDropDatabase.DatabaseName = strings.ToLower(stmt.Name)
	return analyzedQueryDropDatabase
}

func analyzeCreateDatabase(stmt *monvan_parser.DatabaseCreate) *AnalyzedQueryCreateDatabase {
	// Verify if database exists
	analyzedQueryCreateDatabase := NewAnalyzedQueryCreateDatabase()
	_, err := database.GetDatabase(strings.ToLower(stmt.Name))

	if err == nil {
		analyzedQueryCreateDatabase.err = errors.New("database " + strings.ToLower(stmt.Name) + " already exists")
		return analyzedQueryCreateDatabase
	}

	analyzedQueryCreateDatabase.DatabaseName = strings.ToLower(stmt.Name)
	return analyzedQueryCreateDatabase
}

func analyzeDropUser(stmt *monvan_parser.UserDrop) *AnalyzedQueryDropUser {
	analyzedQueryDropUser := NewAnalyzedQueryDropUser()
	analyzedQueryDropUser.Username = strings.ToLower(stmt.Name)
	return analyzedQueryDropUser
}

func analyzeDropRole(stmt *monvan_parser.RoleDrop) *AnalyzedQueryDropRole {
	analyzedQueryDropRole := NewAnalyzedQueryDropRole()
	analyzedQueryDropRole.RoleName = strings.ToLower(stmt.Name)
	return analyzedQueryDropRole
}

func analyzeCreateUser(stmt *monvan_parser.UserCreate) *AnalyzedQueryCreateUser {
	analyzedQueryCreateUser := NewAnalyzedQueryCreateUser()
	analyzedQueryCreateUser.Username = strings.ToLower(stmt.Username)
	analyzedQueryCreateUser.Password = stmt.Password
	return analyzedQueryCreateUser
}

func analyzeCreateRole(stmt *monvan_parser.RoleCreate) *AnalyzedQueryCreateRole {
	analyzedQueryCreateRole := NewAnalyzedQueryCreateRole()
	analyzedQueryCreateRole.RoleName = strings.ToLower(stmt.Name)
	analyzedQueryCreateRole.Options = stmt.Options
	return analyzedQueryCreateRole
}

func analyzeCreateIndex(databaseName string, stmt *monvan_parser.IndexCreate) *AnalyzedQueryCreateIndex {
	// Verify if table exists in database

	analyzedQueryCreateIndex := NewAnalyzedQueryCreateIndex()
	db, err := database.GetDatabase(databaseName)
	if err != nil {
		analyzedQueryCreateIndex.err = err
		return analyzedQueryCreateIndex
	}

	tableName := strings.ToLower(stmt.Table)
	if !database.TableExists(db, tableName) {
		analyzedQueryCreateIndex.err = errors.New("table " + tableName + " does not exist")
		return analyzedQueryCreateIndex
	}

	// analyze if columns exist in table
	table, err := db.GetTable(tableName)
	if err != nil {
		analyzedQueryCreateIndex.err = err
		return analyzedQueryCreateIndex
	}

	for _, column := range stmt.Columns {
		if table.GetColumnByName(column) == nil {
			analyzedQueryCreateIndex.err = errors.New("column " + column + " does not exist in table " + tableName)
		}
	}

	analyzedQueryCreateIndex.DatabaseName = databaseName
	analyzedQueryCreateIndex.TableName = strings.ToLower(stmt.Table)
	analyzedQueryCreateIndex.IndexName = strings.ToLower(stmt.Name)
	analyzedQueryCreateIndex.Columns = stmt.Columns

	return analyzedQueryCreateIndex
}

func analyzeModifyUser(stmt *monvan_parser.UserModify) *AnalyzedQueryModifyUser {
	// Verify if user exists
	analyzedQueryModifyUser := NewAnalyzedQueryModifyUser()
	analyzedQueryModifyUser.Username = strings.ToLower(stmt.Name)
	analyzedQueryModifyUser.Password = stmt.Password
	return analyzedQueryModifyUser
}

func analyzeUseDatabase(stmt *monvan_parser.UseDatabase) *AnalyzedQueryUseDatabase {
	analyzedQueryUseDatabase := NewAnalyzedQueryUseDatabase()
	databaseName := strings.ToLower(stmt.Name)

	// Verify if database exists
	_, err := database.GetDatabase(databaseName)
	if err != nil {
		analyzedQueryUseDatabase.err = err
		return analyzedQueryUseDatabase
	}

	analyzedQueryUseDatabase.DatabaseName = strings.ToLower(stmt.Name)
	return analyzedQueryUseDatabase
}

func analyzeShowTables(databaseName string, stmt *monvan_parser.ShowTables) *AnalyzedQueryShowTables {
	analyzedQueries := NewAnalyzedQueryShowTables()
	// Verify if database exists
	_, err := database.GetDatabase(databaseName)
	if err != nil {
		analyzedQueries.err = err
		return analyzedQueries
	}

	analyzedQueryShowTables := NewAnalyzedQueryShowTables()
	analyzedQueryShowTables.DatabaseName = strings.ToLower(stmt.DatabaseName)
	return analyzedQueryShowTables
}
