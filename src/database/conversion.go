package database

import (
	btree "github.com/nicolasvancan/monvandb/src/btree"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

/*
Basic table interface package, which contains basic methodes for the table structure
*/

// Functions that are used to convert a row to column values and vice versa
func (t *Table) MapRowToColumnValues(row RawRow) []ColumnValue {
	// Create a map to hold the column values
	columnValues := make([]ColumnValue, len(t.Columns))

	// Loop through the columns and add them to the map
	for index := range t.Columns {
		// Get the value from the row
		value := row[t.Columns[index].Name]

		// Create a new column value
		columnValues[index] = ColumnValue{
			Col:   uint16(index),
			Value: value,
		}
	}

	return columnValues
}

func (t *Table) FromColumnValuesToRow(columnValues []ColumnValue) RawRow {
	// Create a map to hold the column values
	row := make(RawRow)

	// Loop through the columns and add them to the map
	for index := range t.Columns {
		// Get the column value
		columnValue := columnValues[index]
		// Add the value to the row
		row[t.Columns[index].Name] = columnValue.Value
	}

	return row
}

func (t *Table) FromRawRowToKeyValue(row RawRow) btree.BTreeKeyValue {
	// Create a new column value
	columnValues := t.MapRowToColumnValues(row)

	// Serialize the column values
	serializedColumnValues, _ := utils.Serialize(columnValues)
	key, _ := utils.Serialize(row[t.PrimaryKey.Name])
	return btree.BTreeKeyValue{
		Key:   key,
		Value: serializedColumnValues,
	}
}

func (t *Table) FromKeyValueToRawRow(keyValues []btree.BTreeKeyValue) []RawRow {
	// Create a new column value
	rawRows := make([]RawRow, len(keyValues))

	for i := range keyValues {
		// Create tmp ColumnValue array
		var columnValue []ColumnValue
		// Get the value from the row
		err := utils.Deserialize(keyValues[i].Value, &columnValue)

		if err != nil {
			return nil
		}

		rawRow := t.FromColumnValuesToRow(columnValue)
		// Add the value to the row
		rawRows[i] = rawRow
	}

	return rawRows
}
