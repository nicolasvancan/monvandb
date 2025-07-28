package database

import (
	"fmt"

	"github.com/nicolasvancan/monvandb/src/utils"
)

/*

This file holds functions related to columns validation, whether it is to validate incomming data or to validate
duplicates rows where there is a constraint in table. The functions are used to validate the data before inserting
This might turn the insert process slower, but it is necessary to keep the data integrity.

*/

/*
This is the main function where all funcions end up here. It is used to validate the columns of a table and the incomming
values that are going to be inserted. The function receives the columns of the table and the rows that are going to be inserted.

The function returns an error if any, otherwise it returns nil.
*/

func (t *Table) ValidateRawRows(rows []RawRow) ([]RawRow, error) {
	validatedRows := make([]RawRow, 0)
	for i, row := range rows {
		err := t.ValidateColumns(&row, i)
		if err != nil {
			return nil, err
		}
		validatedRows = append(validatedRows, row)
	}
	return validatedRows, nil
}

func (t *Table) ValidateColumns(row *RawRow, cur int) error {
	columns := t.Columns
	for _, column := range columns {
		fillupMissingFields(t, row, column, cur)
		err := validateIfColumnExist(column, t)
		if err != nil {
			return err
		}
		err = validateNull(column, (*row)[column.Name])

		if err != nil {
			return err
		}

		err = validateUnique(t, *row)

		if err != nil {
			return err
		}
	}

	return nil
}

// Validate if column exist, otherwise returns an error indicating that a not existing column was given
// to be inserted into file
func validateIfColumnExist(column Column, t *Table) error {
	if t.GetColumnByName(column.Name) == nil {
		return fmt.Errorf("column %s does not exist in table %s", column.Name, t.Name)
	}

	return nil
}

func fillupMissingFields(t *Table, row *RawRow, column Column, cur int) error {
	// Fill up missing fields with default values
	if _, ok := (*row)[column.Name]; !ok {
		if !column.AutoIncrement && !column.Primary && !ok {
			(*row)[column.Name] = column.Default
			return nil
		}
	}

	// Get last Value, cast it to int64 and add one (This should be only with autoincrement columns)
	// Meaning that the columns is an integer between all possiblities
	// SMALL_INT, BIG_INT, INT, etc
	if column.AutoIncrement && column.Primary {
		if (*row)[column.Name] != nil {
			return nil
		}

		lastKey := t.LastKey
		if lastKey == nil {
			// We must insert the first value according to the type of the column
			switch column.Type {
			case COL_TYPE_BIG_INT:
				(*row)[column.Name] = int64(1)
			case COL_TYPE_INT:
				(*row)[column.Name] = int(1)
			case COL_TYPE_SMALL_INT:
				(*row)[column.Name] = int16(1)
			default:
				return fmt.Errorf("column %s is auto increment but not an integer type", column.Name)
			}
			return nil
		} else {
			switch lastVal := lastKey.(type) {
			case int64:
				(*row)[column.Name] = lastVal + 1 + int64(cur)
			case int:
				(*row)[column.Name] = lastVal + 1 + cur
			case int32:
				(*row)[column.Name] = lastVal + 1 + int32(cur)
			case int16:
				(*row)[column.Name] = lastVal + 1 + int16(cur)
			case int8:
				(*row)[column.Name] = lastVal + 1 + int8(cur)
			}
		}
	}
	return nil
}

func validateNull(column Column, value interface{}) error {
	if !column.Nullable && value == nil {
		return fmt.Errorf("column %s cannot be null", column.Name)
	}

	return nil
}

func validateUnique(table *Table, row RawRow) error {

	// It is a table without any constraints
	if table.PrimaryKey == nil && table.CompositeKey == nil {
		return fmt.Errorf("not indexed table. Cannot validate uniqueness")
	}

	// Uniqueness works only for primary keys or composite keys (Constraints)
	var pk *Column = nil
	var key []byte = make([]byte, 0)
	if table.PrimaryKey != nil {
		pk = table.PrimaryKey
		// Indicates that the PK is present and not composite key
		tmp := row[pk.Name]
		serialized, err := utils.Serialize(tmp)

		if err != nil {
			return err
		}

		key = append(key, serialized...)
	}

	if table.CompositeKey != nil {
		// If composite key is different than nil, then it is a composite key
		for _, column := range table.CompositeKey {
			tmp := row[column.Name]
			serialized, err := utils.Serialize(tmp)

			if err != nil {
				return err
			}

			key = append(key, serialized...)
		}
	}
	if len(table.PDataFile.Get(key)) > 0 {
		return fmt.Errorf("row already exists in table")
	}

	return nil
}
