package table

import (
	files "github.com/nicolasvancan/monvandb/diary/en/2_table_structures/files"
)

// Column represents a column in a table
const (
	COL_TYPE_INT = iota
	COL_TYPE_SMALL_INT
	COL_TYPE_BIG_INT
	COL_TYPE_STRING
	COL_TYPE_FLOAT
	COL_TYPE_DOUBLE
	COL_TYPE_BOOL
	COL_TYPE_TIMESTAMP
	COL_TYPE_BLOB
)

type ColumnValue struct {
	Value interface{} // Value of the column
	Col   uint16      // Refers to the respective column of a table for example:
	// Table X has column Y and column Z, which are stored in a Table struct as an array of Columns struct
	// Each position of this array represents a column in the table. Whenever a column is serialized,
	// the position of the column is stored in the Col field
}

// Alias for []SerializedColumnValue
type RowValues = []ColumnValue

type Column struct {
	Name          string
	Type          int
	Default       interface{}
	Nullable      bool
	AutoIncrement bool
	Primary       bool
	Unique        bool
	Foreign       bool
}

type Table struct {
	Name      string          // Table's name
	Path      string          // Where the table configuration is stored
	Columns   []Column        // reference to Columns
	pDataFile *files.DataFile // private Access btree (Simple)
}
