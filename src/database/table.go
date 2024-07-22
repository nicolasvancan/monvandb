package database

import (
	btree "github.com/nicolasvancan/monvandb/src/btree"
	"github.com/nicolasvancan/monvandb/src/files"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

/*
Table interface

The idea of this interface is to provide the most basic and atomic table methodes for the system to be used, namelly, the base CRUD.
When the system receives a query from an user, it parses it using the parser, translate all parsed commands into structured commands,
that are in part conversible to minimal and atomic table operations. The table interface is the one that provides these operations.

BASIC ACCESS METHODES

- Get: Get one or more rows for a given key
- Insert: Insert a row into the table
- Update: Update a row in the table
- Delete: Delete a row from the table
- Range: Get all the rows in a given range

To explain the main idea of this interface, some examples are shown below:

Get:
SELECT * FROM table WHERE indexedColumn = "value"

This query would result in a call of the Get method, since there is a Where clause with indexed column.
If we alter this SQL statement to another type, such as the one below:

SELECT * FROM table WHERE notIndexedColumn = "value" AND indexedColumn > 10

It is easy to note that two conditions are evaluated. One indexed and the other one not. Knowing that there is a not indexed
column needing comparison, there must be a scan through the datafile to get the data.

If the query is altered again, to the one below:

SELECT * FROM table WHERE indexedColumn > 10

Than the Range method is called. This method will return only keys greater than 10.

For other methodes such as, UPDATE and DELETE, the same principle is applied. The main idea is to provide the most basic and atomic
table operations to be used by the system.

For larger and complex queries the same principle is applyed, but after que table queries return the desired data, other operations
are applied to the data, such as joins, unions, and so on.
*/

/*
Get

Get one or more rows for a given key, returning the rows as a slice of RawRow.

This method is now really simple, but later, depending on the implementation of the database, it could be more complex.
For instance, let's say that we have a table with a primary key, and we want to get the row with the primary key. In this case,
the function uses one *PDataFile to access the data.

When running for indexed columns,
*/
func (t *Table) Get(column string, value any) ([]RawRow, error) {

	row := make([]RawRow, 0)
	// Get the pk column
	pk := t.PrimaryKey
	serializedValue, err := utils.Serialize(value)

	if err != nil {
		return nil, err
	}

	// If the column is the primary key, we can use the datafile to get the data
	if pk.Name == column {
		// Serialize the value

		row = append(row, t.FromKeyValueToRawRow(t.PDataFile.Get(serializedValue))...)
		return row, nil
	}

	// We try to find indexed DataFiles
	index, ok := t.Indexes[column]
	if ok {
		// Get the index DataFile
		indexDataFile := index.PDataFile
		// Get the value from the index
		indexValue := indexDataFile.Get(serializedValue)
		row = append(row, t.FromKeyValueToRawRow(indexValue)...)
	} else {
		// Worst case, there must be a scan through the datafile
		// Create a crawler at the beginning of the file
		crawler := btree.GoToFirstLeaf(t.PDataFile.GetBTree())
		// Loop through the datafile
		keyValues := make([]btree.BTreeKeyValue, 0)

		for {
			kv := crawler.GetKeyValue()
			keyValues = append(keyValues, *kv)

			// If the crawler is at the end of the file, break the loop
			if err = crawler.Next(); err != nil {
				break
			}
		}
		row = append(row, t.FromKeyValueToRawRow(keyValues)...)
	}

	return row, nil
}

/*
Insert a row into the table for given []RawRow.
*/

func (t *Table) Insert(rows []RawRow) (int, error) {
	// First step - Validate rows

	validatedRows, err := t.ValidateRawRows(rows)

	// Returns case there is an error on rows validation process
	if err != nil {
		return 0, err
	}

	// We have to insert it into the base table and also into the indexes
	for i, row := range validatedRows {
		// Insert the row into the base table
		serializedRow := t.FromRawRowToKeyValue(row)
		t.PDataFile.Insert(serializedRow.Key, serializedRow.Value)
		// If there is indexed tables, we have to insert the row into the indexed tables
		for _, index := range t.Indexes {
			// Get the index DataFile
			indexDataFile := index.PDataFile
			// get the column name to be serialized
			indexKey := row[index.Column]
			// Serialize the value
			serializedIndexKey, err := utils.Serialize(indexKey)

			if err != nil {
				return i, err
			}
			// Insert the row into the index table
			indexDataFile.Insert(serializedIndexKey, serializedRow.Value)
		}
	}

	return len(rows), nil
}

/*
To update rows we use the principle of deleting and reinserting it.
*/

func (t *Table) Update(rows []RawRow) (int, error) {
	// Delete rows
	_, err := t.Delete(rows)

	if err != nil {
		return 0, err
	}
	// Insert rows
	i, err := t.Insert(rows)
	return i, err
}

func reinsertKeys(
	t *Table,
	indexDataFile *files.DataFile,
	existingKeys []RawRow,
	row RawRow,
	serializedIndexKey []byte,
) {
	musReinsert := make([]RawRow, 0)
	for _, iRow := range existingKeys {
		indexDataFile.Delete(serializedIndexKey)
		if iRow[t.PrimaryKey.Name] != row[t.PrimaryKey.Name] {
			musReinsert = append(musReinsert, iRow)
		}
	}

	// Reinsert
	for _, iRow := range musReinsert {
		serializedIndexRow := t.FromRawRowToKeyValue(iRow)
		indexDataFile.Insert(serializedIndexKey, serializedIndexRow.Value)
	}
}

/*
The delete function is mainly called when the SQL DELETE Statement is used. In general, the statement is used with the
the were condition, indicating that a query must be made before the deletion. Knowing that, It is known that the query will
return the keys of the rows to be deleted. Therefore, the input for this method is an array of byte array

*/

func (t *Table) Delete(rows []RawRow) (int, error) {
	// Iterate over the input
	for _, row := range rows {
		// Delete the row from the base table
		serializedRow := t.FromRawRowToKeyValue(row)

		t.PDataFile.Delete(serializedRow.Key)
		// If there is indexed tables, we have to delete the row from the indexed tables
		for _, index := range t.Indexes {
			// Get the index DataFile
			indexDataFile := index.PDataFile
			// Delete the row from the index table
			indexColumn := row[index.Column]
			serializedIndexKey, err := utils.Serialize(indexColumn)

			if err != nil {
				return 0, err
			}
			// Get the keys from the index
			existingKeys := indexDataFile.Get(serializedIndexKey)

			// Indicates that there is more than one key for the same value
			// All keys must be deleted and the keys that are different from
			// the one that is being deleted must be inserted again
			if len(existingKeys) > 1 {
				deserializedExistingKeys := t.FromKeyValueToRawRow(existingKeys)
				reinsertKeys(t, indexDataFile, deserializedExistingKeys, row, serializedIndexKey)
			} else {
				indexDataFile.Delete(serializedIndexKey)
			}
		}

	}
	return len(rows), nil
}

/*
Range:

Range is one of the main basic table methodes. It is composed of an input of query option, indicating
some details of query that must be taken in consideration, such as the limit, the offset, the order, and the where
clauses. For instance, let's say a query statement comes as written below:

SELECT * FROM table WHERE notIndexedColumn = value and indexedColumn > 10 ORDER BY column LIMIT 10

In this case, the query parser (that is not implemented yet) will parse the query and return a QueryOptions struct
which can be used to execute the range for table or indexes

	It receives an array of column operations related to the table. All columns operation in the array are not

ordered, and don't belong to the same select statement if it is a compound select, rather to the table itself.
For example, let's say we have a simple table select and join with another table, followed by a filter in the Where
clause. In this case, the columns operations are related to the table, and not to the select statement.

SELECT

	*

FROM table t

	INNER JOIN table2 t2 on t2.x = t.x

WHERE t.x > 10

We know that the column x for the first table is indexed, and we also know that the column x for the second table is not indexed.
It will result in two range operations for both tables, as follows:

For the first table (t):

  - input []ColumnOperation will be

    []ColumnOperation{
    {
    Operation: COL_COMP,
    ColumnName: "x",
    TableName: "table",
    Condition: 0, // Equal
    Alias: "",
    Value: ColumnOperationValue{
    IsOtherColumn true
    IsOtherTable  true
    ColumnName    x
    TableHash     "123456789"
    Value         nil
    }
    Transformation: nil,
    TransformationParams: nil
    },
    {
    Operation: COL_COMP,
    ColumnName: "x",
    TableName: "table",
    Condition: 1, // Greater than
    Alias: "",
    Value: ColumnOperationValue{
    IsOtherColumn false
    IsOtherTable  false
    ColumnName    nil
    TableHash     nil
    Value         10
    }
    Transformation: nil,
    TransformationParams: nil
    }
    }

For the second table (t2):

	[]ColumnOperation{{
	    Operation: COL_COMP,
	    ColumnName: "x",
	    TableName: "table2",
	    Condition: 1, // Greater than
	    Alias: "",
	    Value: ColumnOperationValue{
	    IsOtherColumn true
	    IsOtherTable  true
	    ColumnName    x
	    TableHash     98765412
	    Value         nil
	    }
	    Transformation: nil,
	    TransformationParams: nil
	}}

When the inputs field for table is given to the Range query, it will evaluate and fetch only
necessary columns, in this case, it will create a RangeOptions struct optimized for the demand
and return the result.
*/
func (t *Table) Range(input []ColumnComparsion, limit int, order int) []RawRow {
	// create temporary variable for holding values of RangeOptions
	rangeOperation := MergeOperationsBasedOnIndexedColumnsAndReturnRangeOptions(t, input)
	rangeOperation.Limit = limit
	rangeOperation.Order = order

	// Invert the order of from and to
	if order == DESC {
		reverseAscToDesc(&rangeOperation)
	}

	// Get the range
	rows, err := Range(t, rangeOperation)

	if err != nil {
		return nil
	}

	return rows
}

func (t *Table) getLastItem() RawRow {
	lastLeafCrawler := btree.GoToLastLeaf(t.PDataFile.GetBTree())
	if len(lastLeafCrawler.Net) > 0 {
		lastLeaf := lastLeafCrawler.Net[len(lastLeafCrawler.Net)-1]
		lastLeafNItens := lastLeaf.GetNItens()
		lastItem := lastLeaf.GetLeafKeyValueByIndex(uint16(lastLeafNItens - 1))
		return t.FromKeyValueToRawRow([]btree.BTreeKeyValue{{Key: lastItem.GetKey(), Value: lastItem.GetValue()}})[0]
	}

	return nil
}

func (t *Table) GetColumnByName(colName string) *Column {
	for _, col := range t.Columns {
		if col.Name == colName {
			return &col
		}
	}
	return nil
}

func (t *Table) isColumnIndexed(colName string) bool {

	if t.IsComposedKeyTable() {
		return false
	}

	if t.PrimaryKey.Name == colName {
		return true
	}

	for _, index := range t.Indexes {
		if index.Column == colName {
			return true
		}
	}

	return false
}

func (t *Table) IsComposedKeyTable() bool {
	return len(t.CompositeKey) > 0
}
