package database

import (
	files "github.com/nicolasvancan/monvandb/src/files"
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

type Database struct {
	Name       string            // Database's name
	Tables     map[string]*Table // reference to Tables
	TablePaths map[string]string // Paths to the tables
	Path       string            // Path to the database dir
}

type Table struct {
	Name         string            // Table's name
	Path         string            // Where the table configuration is stored
	Columns      []Column          // reference to Columns
	PrimaryKey   *Column           // reference to PrimaryKey
	CompositeKey []Column          // Case column is composite
	Indexes      map[string]*Index // reference to Indexes
	PDataFile    *files.DataFile   // private Access btree (Simple)
}

type RawRow = map[string]interface{}

const (
	ASC = iota
	DESC
)

const (
	EQ    = iota // Equal
	GT           // Greater than
	GTE          // Greater than or equal
	LT           // Less than
	LTE          // Less than or equal
	NE           // Not equal
	IN           // In
	NIN          // Not in
	LIKE         // Like
	NLIKE        // Not Like
)

type RangeOptions struct {
	From        []byte
	To          []byte
	FComparator int             // From Comparator, indicates what type of comparation should be done with the From value ex: GTE or GT
	TComparator int             // To Comparator, indicates what type of comparation should be done with the To value
	Order       int             // Order of the range wheter is ASC os DESC
	Limit       int             // Limit of the range
	PDataFile   *files.DataFile // Pointer to the data file to be used
}

type Index struct {
	Name      string
	Column    string
	Path      string
	PDataFile *files.DataFile
}

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
}

/*
For a table there are a lot of different operations that can be done. Basically, those are done either in the columns
spectre or in the tables spectre. Let's say, I can join two tables, that is a table operation. The select statement
is also done in the table spectre.

When it comes to columns, it reaches another spectre that is responsible for dealling with columns general operations.
For example, A columns can be used for comparsions, or a column can be transformed or a column can be both used for comparsion
after a transformation is done with it.

In a general way, the table specre is wider than the column spectre.
*/

// Col Operations
const (
	COL_COMP       = iota // Comparsion
	COL_TRANSF            // Transformation
	COL_COMP_TANSF        // Comparsion and Transformation
	COL_NONE              // No operation
)

type ColumnOperation struct {
	Operation      int    // Col Operations
	ColumnName     string // Column name
	TableName      string //Table Name
	Condition      int
	Value          ColumnConditionValue
	Transformation func(interface{}) interface{}
}

type ColumnConditionValue struct {
	IsOtherColumn bool        // Indicates if the value is a column comparsion
	IsOtherTable  bool        // Indicates if the value is a column from another table
	ColumnName    string      // Column name
	TableName     string      // Table name
	Value         interface{} // Value is only used if isOther columns and isOtherTable are both false
}
