package main

/*
Tests validator from database package
*/

import (
	"testing"

	database "github.com/nicolasvancan/monvandb/src/database"
	helper "github.com/nicolasvancan/monvandb/src/test/helper"
)

func changeColumnProperties(colName string, newProperties database.Column, table *database.Table) {
	for i := 0; i < len(table.Columns); i++ {
		if table.Columns[i].Name == colName {
			table.Columns[i] = newProperties
		}
	}
}

func TestValidatorForNullValues(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	// First test case is for a column that is Nullable and
	// a value inserted is nil, returning non error
	changeColumnProperties("name",
		database.Column{
			Name:          "name",
			Type:          database.COL_TYPE_STRING,
			Default:       nil,
			Nullable:      true,
			AutoIncrement: false,
			Primary:       false,
		}, table)

	row := database.RawRow{
		"name": nil,
		"id":   int32(490),
	}

	err := table.ValidateColumns(&row)

	if err != nil {
		t.Errorf("error validating columns should be nil: %v", err)
	}

	// Change now the nullable to false, and it should return an error
	changeColumnProperties("name",
		database.Column{
			Name:          "name",
			Type:          database.COL_TYPE_STRING,
			Default:       nil,
			Nullable:      false,
			AutoIncrement: false,
			Primary:       false,
		}, table)

	err = table.ValidateColumns(&row)

	if err == nil {
		t.Errorf("error validating columns should not be nil: %v", err)
	}
}

func TestValidatorForAutoIncrement(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	// First test case is for a column that is AutoIncrement and
	// a value inserted is nil, returning non error
	changeColumnProperties("id",
		database.Column{
			Name:          "id",
			Type:          database.COL_TYPE_INT,
			Default:       nil,
			Nullable:      false,
			AutoIncrement: true,
			Primary:       false,
		}, table)

	row := database.RawRow{
		"name": "test",
		"id":   nil,
	}

	err := table.ValidateColumns(&row)

	if err == nil {
		t.Errorf("should have returned an error. To be considered autoincrement, it must also have the primary enabled")
	}

	// Change now the AutoIncrement to false, and it should return an error
	changeColumnProperties("id",
		database.Column{
			Name:          "id",
			Type:          database.COL_TYPE_INT,
			Default:       nil,
			Nullable:      false,
			AutoIncrement: false,
			Primary:       false,
		}, table)

	err = table.ValidateColumns(&row)

	if err == nil {
		t.Errorf("error validating columns should not be nil")
	}

	// Now it should insert a value for id
	changeColumnProperties("id",
		database.Column{
			Name:          "id",
			Type:          database.COL_TYPE_INT,
			Default:       nil,
			Nullable:      false,
			AutoIncrement: true,
			Primary:       true,
		}, table)

	err = table.ValidateColumns(&row)

	if err != nil {
		t.Errorf("error validating columns should be nil: %v", err)
	}

	if row["id"] != int64(450) {
		t.Errorf("error should be 450 and got: %d", row["id"])
	}
}

func TestDefaultInsertion(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	// First test case is for a column that is AutoIncrement and
	// a value inserted is nil, returning non error
	changeColumnProperties("name",
		database.Column{
			Name:          "name",
			Type:          database.COL_TYPE_STRING,
			Default:       "Nicolas",
			Nullable:      true,
			AutoIncrement: false,
			Primary:       false,
		}, table)
	changeColumnProperties("id",
		database.Column{
			Name:          "id",
			Type:          database.COL_TYPE_BIG_INT,
			Default:       nil,
			Nullable:      false,
			AutoIncrement: true,
			Primary:       true,
		}, table)

	row := database.RawRow{
		"id": nil,
	}

	err := table.ValidateColumns(&row)

	if err != nil {
		t.Errorf("should have returned an error. To be considered autoincrement, it must also have the primary enabled")
	}

	if row["name"] != "Nicolas" {
		t.Errorf("error should be Nicolas and got: %s", row["name"])
	}

	if row["id"] != int64(450) {
		t.Errorf("error should be 450 and got: %d", row["id"])
	}
}
