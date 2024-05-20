package main

import (
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
