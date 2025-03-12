package parser

/*
	This file contains all table modifications analyzer functions
	such as Create Table, Update Table, Drop Table, Alter Table, and so on
*/
import (
	"strconv"
	"strings"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
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

func analyzeDropTable(databaseName string, stmt *sqlparser.DDL) *AnalyzedQueryDropTable {
	analyzedQueryCreateDatabase := NewAnalyzedQueryDropTable()
	analyzedQueryCreateDatabase.DatabaseName = databaseName
	analyzedQueryCreateDatabase.TableName = strings.ToLower(stmt.Table.Name.String())
	return analyzedQueryCreateDatabase
}

// TODO
func analyzeAlterTable(databaseName string, stmt *sqlparser.DDL) *AnalyzedQueryAlterTable {
	analyzedQueryAlterTable := NewAnalyzedQueryAlterTable()
	analyzedQueryAlterTable.DatabaseName = databaseName
	analyzedQueryAlterTable.TableName = strings.ToLower(stmt.Table.Name.String())
	return analyzedQueryAlterTable
}
