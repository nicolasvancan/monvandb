package main

import (
	"fmt"
	"testing"

	database "github.com/nicolasvancan/monvandb/src/database"
	helper "github.com/nicolasvancan/monvandb/src/test/helper"
)

var err_getting_data = "error getting data: %v"

/*
This function validates the insertion of data into main DataFile and also the indexed ones.
*/
func TestTableAndIndexInsertion(t *testing.T) {
	// get table with index
	table := helper.CreateMockTableAndIndex(t)

	// insert data
	rows := []database.RawRow{
		{
			"id":    int32(1),
			"name":  "John",
			"email": "john@john.com",
		},
		{
			"id":    int32(2),
			"name":  "Maria",
			"email": "maria@john.com",
		},
	}

	n, err := table.Insert(rows)

	// Returns how many items were inserted
	if n != 2 {
		t.Errorf("expected 2 as result, got %v", n)
	}

	// Check if there is any error
	if err != nil {
		t.Errorf("error inserting data: %v", err)
	}

	// Time to check both indexes
	// Check if the primary key index is working
	mainDataFile := table.PDataFile
	// Check if the email index is working
	indexDataFile := table.Indexes["email"].PDataFile

	fmt.Printf("Indexes: %v\n", table.Indexes)

	mainResults, err := database.RangeFromOptions(table, database.RangeOptions{
		From:      nil,
		To:        nil,
		Limit:     -1,
		PDataFile: mainDataFile,
	},
	)

	if err != nil {
		t.Errorf(err_getting_data, err)
	}

	if len(mainResults) != 2 {
		t.Errorf("expected 2 items, got %v", len(mainResults))
	}

	indexResults, err := database.RangeFromOptions(table, database.RangeOptions{
		From:      nil,
		To:        nil,
		Limit:     -1,
		PDataFile: indexDataFile,
	},
	)

	if len(indexResults) != 2 {
		t.Errorf("expected 2, got %v", len(indexResults))
	}

	if err != nil {
		t.Errorf(err_getting_data, err)
	}
}

func TestGetForTable(t *testing.T) {
	// get table with index
	table := helper.CreateMockTableAndIndexWithRows(t)

	// Get data from table
	data, err := table.Get(table.GetColumnByName("id").Name, int32(1))

	if err != nil {
		t.Errorf(err_getting_data, err)
	}

	if data == nil {
		t.Errorf("expected data, got nil")
	}

	if len(data) != 1 {
		t.Errorf("expected 1 item, got %v", len(data))
	}

	if data[0]["id"] != int32(1) {
		t.Errorf("expected 1, got %v", data[0]["id"])
	}

}

func TestDeleteForTable(t *testing.T) {
	// get table with index
	table := helper.CreateMockTableAndIndexWithRows(t)

	// insert data
	rows := []database.RawRow{
		{
			"id":    int32(1),
			"name":  "John",
			"email": "john@john.com",
		},
	}

	// If the error is not nil, it means that the data was deleted correctly
	_, err := table.Delete(rows)

	if err != nil {
		t.Errorf("error deleting data: %v", err)
	}

	// Verify if the deleted key exists
	data, err := table.Get("id", int32(1))

	if err != nil {
		t.Errorf(err_getting_data, err)
	}

	if len(data) != 0 {
		t.Errorf("error deleting data, it should not exist but is still there")
	}
}

func TestUpdateForTable(t *testing.T) {
	// get table with index
	table := helper.CreateMockTableAndIndexWithRows(t)

	// insert data
	rows := []database.RawRow{
		{
			"id":    int32(1),
			"name":  "Abacate Azul",
			"email": "ab@ab.com",
		},
	}

	n, err := table.Update(rows)
	if n != 1 {
		t.Errorf("expected 1 as result, got %v", n)
	}

	if err != nil {
		t.Errorf("error updateing data: %v", err)
	}

	// Verify if the updated key exists
	data, err := table.Get("id", int32(1))

	if err != nil {
		t.Errorf(err_getting_data, err)
	}

	if len(data) != 1 {
		t.Errorf("expected 1 item, got %v", len(data))
	}

	if data[0]["name"] != "Abacate Azul" {
		t.Errorf("expected Abacate Azul, got %v", data[0]["name"])
	}
}
