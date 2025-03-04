package executor

import (
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
)

type Operation struct {
	Name string
	Args []interface{}
}

const (
	// Table Operations
	TABLE_FILE_READ   = "TABLE_FILE_READ"
	TABLE_FILE_INSERT = "TABLE_FILE_INSERT"
	TABLE_FILE_UPDATE = "TABLE_FILE_UPDATE"
	TABLE_FILE_DELETE = "TABLE_FILE_DELETE"
	// Dataframe Operations
	SELECT  = "SELECT"
	JOIN    = "JOIN"
	FILTER  = "FILTER"
	GROUPBY = "GROUPBY"
	ORDERBY = "ORDERBY"
	LIMIT   = "LIMIT"
	// Table Operations
	TABLE_CREATE = "TABLE_CREATE"
	TABLE_DROP   = "TABLE_DROP"
	TABLE_ALTER  = "TABLE_ALTER"
	// Database Operations
	DATABASE_CREATE = "DATABASE_CREATE"
	DATABASE_DROP   = "DATABASE_DROP"
	DATABASE_ALTER  = "DATABASE_ALTER"
	// Index Operations
	INDEX_CREATE = "INDEX_CREATE"
	INDEX_DROP   = "INDEX_DROP"
	INDEX_ALTER  = "INDEX_ALTER"
	// User Operations
	USER_CREATE = "USER_CREATE"
	USER_DROP   = "USER_DROP"
	USER_ALTER  = "USER_ALTER"
	// Role Operations
	ROLE_CREATE = "ROLE_CREATE"
	ROLE_DROP   = "ROLE_DROP"
	ROLE_ALTER  = "ROLE_ALTER"
)

var Operations = map[string]func(...interface{}) (df.Dataframe, error){
	// Table Operations
	TABLE_FILE_READ:   TableFileRead,   // Read a table file operation
	TABLE_FILE_INSERT: TableFileInsert, // Insert data into a table file
	TABLE_FILE_UPDATE: TableFileUpdate, // Update data in a table file
	TABLE_FILE_DELETE: TableFileDelete, // Delete data from a table file
	// Dataframe Operations
	SELECT:  DataframeSelect,
	JOIN:    DataframeJoin,
	FILTER:  DataframeFilter,
	GROUPBY: DataframeGroupBy,
	ORDERBY: DataframeOrderBy,
	LIMIT:   DataframeLimit,
	// Table Operations
	TABLE_CREATE: TableFileCreate, // Create a table operation
	TABLE_DROP:   nil,             // Drop a table operation
	TABLE_ALTER:  nil,             // Alter a table operation
	// Database Operations
	DATABASE_CREATE: nil, // Create a database operation
	DATABASE_DROP:   nil, // Drop a database operation
	DATABASE_ALTER:  nil, // Alter a database operation
	// Index Operations
	INDEX_CREATE: nil, // Create an index operation
	INDEX_DROP:   nil, // Drop an index operation
	INDEX_ALTER:  nil, // Alter an index operation
	// User Operations
	USER_CREATE: nil, // Create a user operation
	USER_DROP:   nil, // Drop a user operation
	USER_ALTER:  nil, // Alter a user operation
	// Role Operations
	ROLE_CREATE: nil, // Create a role operation
	ROLE_DROP:   nil, // Drop a role operation
	ROLE_ALTER:  nil, // Alter a role operation
}
