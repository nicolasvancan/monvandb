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
	AND = iota
	OR
	NOT
)

// Conditions
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

/*
ColumnOperation represents a column operation

This struct might change with time. The idea of ColumnOperation is to represent every column operation in SQL statement
as a ColumnOperation struct. Simple example:

SELECT

	name as name,
	age as age,
	age + 10 as age_plus_10

FROM

	users

WHERE

	age > 10

In this example, we have 4 columns operations: name, age, age + 10 and age > 10. The first two are simple columns operations
, as follows

name as name:

	ColumnOperation{
		Operation: COL_NONE,
		ColumnName: "name",
		TableName: "users",
		Condition: -1, // No condition
		Alias: "name",
		Value: nil,
		Transformation: nil
		TransformationParams: nil
	}

The same happens for age as age. For case where we have a transformation, we have the following:

	func sum(row interface{}, value []interface{}) interface{} {
		var r int = row.(int)
		for _, v := range value {
			r += v.(int)
		}
		return r
	}

	ColumnOperation{
		Operation: COL_TRANSF,
		ColumnName: "age",
		TableName: "users",
		Condition: -1, // No condition
		Alias: "age_plus_10",
		Value: nil,
		Transformation: sum
		TransformationParams: [10]
	}

For the last column operation, we have the following:

		ColumnOperation{
			Operation: COL_COMP,
			ColumnName: "age",
			TableName: "users",
			Condition: GT,
			Alias: "",
			Value: ColumnConditionValue{
				IsOtherColumn: false,
				IsOtherTable: false,
				ColumnName: "",
				TableHash: "",
				Value: 10
			},
			Transformation: nil
			TransformationParams: nil
	}

Note that the structure is supposed to be generic so that we can use it not only for comparsion of select statements but also
for the select fields, or even where statements fields.
*/
type ColumnComparsion struct {
	ColumnName string               // Column name
	TableName  string               // Table Name ()
	Condition  int                  // Condition EQ, NEQ, GT, GTE, LT, LTE, IN, NIN, LIKE, NLIKE
	Alias      string               // Alias for the column
	Value      ColumnConditionValue // Value to be used in the comparsion
	Id         int                  // Identifier for comparsion, example WHERE x < 1 AND x > -3, the identifier belongs
	// to the layer of comparsion. Bosh x<1 and x>-3 have the same identifier, meaning that they are in the same layer
	ParentId        int // Parent layer id, if it is base layer, it is -1
	ParentLogicalOp int // Parent AND or OR
	LayerLogicalOp  int // AND or OR
}

type ColumnConditionValue struct {
	IsOtherColumn        bool                                         // Indicates if the value is a column comparsion
	IsOtherTable         bool                                         // Indicates if the value is a column from another table
	ColumnName           string                                       // Column name
	TableHash            string                                       // Table Hash (Used to get the table from the database)
	Value                interface{}                                  // Value is only used if isOther columns and isOtherTable are both false
	Transformation       func(interface{}, []interface{}) interface{} // Still not implemented
	TransformationParams []interface{}                                // Still not implemented
}
