package main

import (
	"fmt"
	"testing"

	"github.com/nicolasvancan/monvandb/src/database"
	"github.com/nicolasvancan/monvandb/src/test/helper"
	"github.com/nicolasvancan/monvandb/src/utils"
)

func TestLoadDatabase(t *testing.T) {
	// Create a new database and store it to file
	helper.CreateBasePaths(t)
	helper.CreateDatabaseFileAndSetFile(t)

	// Load the database
	db, err := database.LoadDatabase(utils.GetPath("databases") + utils.SEPARATOR + "mock")

	if err != nil {
		t.Errorf("error loading database: %v", err)
	}

	if db.Name != "db_teste" {
		t.Errorf("expected db_teste, got %v", db.Name)
	}
}

func TestCreateTable(t *testing.T) {
	// Create a new database and store it to file
	helper.CreateBasePaths(t)
	helper.CreateDatabaseFileAndSetFile(t)

	// Load the database
	db, err := database.LoadDatabase(utils.GetPath("databases") + utils.SEPARATOR + "mock")

	if err != nil {
		t.Errorf("error loading database: %v", err)
	}

	// Create a new table
	err = db.CreateTable("table_teste", []database.Column{
		{
			Name: "id",
			Type: database.COL_TYPE_INT,
		},
		{
			Name: "name",
			Type: database.COL_TYPE_STRING,
		},
	})

	if err != nil {
		t.Errorf("error creating table: %v", err)
	}

	table, err := db.GetTable("table_teste")

	if err != nil {
		t.Errorf("error getting table: %v", err)
	}

	if table.Name != "table_teste" {
		t.Errorf("expected table_teste, got %v", table.Name)
	}
}

func TestCreateIndex(t *testing.T) {
	// Create a new database and store it to file
	helper.CreateBasePaths(t)
	helper.CreateDatabaseFileAndSetFile(t)

	// Load the database
	db, err := database.LoadDatabase(utils.GetPath("databases") + utils.SEPARATOR + "mock")

	if err != nil {
		t.Errorf("error loading database: %v", err)
	}

	// Create a new table
	err = db.CreateTable("table_teste", []database.Column{
		{
			Name:    "id",
			Type:    database.COL_TYPE_INT,
			Primary: true,
		},
		{
			Name: "name",
			Type: database.COL_TYPE_STRING,
		},
	})

	if err != nil {
		t.Errorf("error creating table: %v", err)
	}

	// Create a new index
	err = db.CreateIndex("table_teste", "name", "name_index")

	if err != nil {
		t.Errorf("error creating index: %v", err)
	}

	table, err := db.GetTable("table_teste")
	fmt.Printf("table: %v\n", table)

	if err != nil {
		t.Errorf("error getting table: %v", err)
	}

	if table.Indexes["name"].Column != "name" {
		t.Errorf("expected id, got %v", table.Indexes["name"].Column)
	}
}
