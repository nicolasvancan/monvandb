package database

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
