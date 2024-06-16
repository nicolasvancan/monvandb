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
	Name         string           // Table's name
	Path         string           // Where the table configuration is stored
	Columns      []Column         // reference to Columns
	PrimaryKey   Column           // reference to PrimaryKey
	CompositeKey []Column         // Case column is composite
	Indexes      map[string]Index // reference to Indexes
	PDataFile    *files.DataFile  // private Access btree (Simple)
}

type RawRow = map[string]interface{}

const (
	ASC = iota
	DESC
)

const (
	EQ = iota
	GT
	GTE
	LT
	LTE
	NE
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

/*
TableQueryOperation is a struct that represents the literal query conditions that can be used to query a table.
One exameple is the following WHERE statement:

WHERE id > 10 AND (name = 'John' OR age < 30)

This statement can be represented as the following TableQueryOperation struct:

	TableQueryOperation{
		Column: "id",
		Comparator: GT,
		Value: 10,
		And: []TableQueryOperation{
			TableQueryOperation{
				Column: "name",
				Comparator: EQ,
				Value: "John",
				And: nil,
				Or: []TableQueryOperation{
					TableQueryOperation{
						Column: "age",
						Comparator: LT,
						Value: 30,
						And: nil,
						Or: nil,
					}
				}
			},
		},

		Or: nil,
	}

This structure is also used to represent the conditions of a range query, for example:
Knowing that the id is the primary key of the table, we know that the primary key is indexed and the value must be greater than 10
Therefore, it doesn't matter the other columns conditions, the range will not be altered because of them.
*/
type TableQueryOperation struct {
	Column     string
	Comparator int
	Value      interface{}
	And        []TableQueryOperation
	Or         []TableQueryOperation
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
	Unique        bool
}
