package database

import (
	btree "github.com/nicolasvancan/monvandb/src/btree"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

/*
Table interface

The table is the entry point to access the data stored within one Datafile (BTree) and holds also all methodes related to the file itself,
whether it is just a simple get function, or even a range query, or scan. All those methodes are called within a function.

The basic struct is composed of some table definition fields, such as the name of the table, the path to the table file, and the columns
of the table. On the otherhand, the table also has other struct that are used to manipulate data or files.

For this version of the database, the table struct is composed of the following fields:
- Name: The name of the table
- Path: The path to the table file
- Columns: The columns of the table
- PDataFile: A pointer to the DataFile struct, which is used to access the data stored in the table file

### BASIC ACCESS METHODES ###

- Get: Get one or more rows for a given key
- Insert: Insert a row into the table
- Update: Update a row in the table
- Delete: Delete a row from the table
- Range: Get all the rows in a given range

### INDEX METHODES ###

- CreateIndex: Create an index on a given column
- DropIndex: Drop an index on a given column
- GetIndex: Get the index on a given column
- GetIndexValue: Get the value of the index on a given column
- InsertIndexValue: Put the value of the index on a given column
- UpdateIndexValue: Update the value of the index on a given column
- DeleteIndexValue: Delete the value of the index on a given column

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
	/*
		Reserved for further implementation, evaluate if the column is indexed or not, and then load the appropriate
		binary tree to get the data. Otherwise, use the default which is the datafile.
	*/
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
			keyValues = append(keyValues, kv)

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

func (t *Table) Insert([]RawRow) int {
	return 0
}

/*
Update a row into the table for given []RawRow
*/

func (t *Table) Update(RawRow) int {
	return 0
}

func (t *Table) Delete([]RawRow) int {
	return 0
}

func (t *Table) isColumnIndexed(column string) bool {
	_, ok := t.Indexes[column]
	return ok
}

func (t *Table) GetColumnByName(colName string) *Column {
	for _, col := range t.Columns {
		if col.Name == colName {
			return &col
		}
	}
	return nil
}
