package helper

import (
	"os"
	"testing"

	"github.com/nicolasvancan/monvandb/src/database"
	db "github.com/nicolasvancan/monvandb/src/database"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

func CreateBasePaths(t *testing.T) {
	basePath := t.TempDir()
	err := utils.CreateBaseFolders(basePath)

	if err != nil {
		t.Errorf("error creating base folders: %v", err)
	}
}

func CreateDatabaseFileAndSetFile(t *testing.T) {
	databaseFolderName := utils.GetPath("databases") + utils.SEPARATOR + "mock"

	// Create database File
	err := utils.CreateFolder(databaseFolderName, os.ModePerm)

	if err != nil {
		t.Errorf("error creating database folder: %v", err)
	}

	// Create a new database

	database := new(db.Database)
	database.Name = "db_teste"
	database.Path = databaseFolderName
	database.TablePaths = make(map[string]string)
	database.Tables = make(map[string]*db.Table)
	// We first create a db at the tmp folder
	_, err = utils.CreateFile(
		databaseFolderName +
			utils.SEPARATOR +
			utils.METDATA_FILE,
	)

	if err != nil {
		t.Errorf("error creating database file: %v", err)
	}

	// Serialize the database
	res, err := utils.ToJson(*database)

	if err != nil {
		t.Errorf("error encoding struct: %s", err)
	}

	// Write the database to the file
	err = utils.WriteToFile(
		databaseFolderName+
			utils.SEPARATOR+
			utils.METDATA_FILE,
		res,
	)

	if err != nil {
		t.Errorf("error writing to database file: %v", err)
	}
}

func CreateMockTable(t *testing.T) *database.Table {
	// Create a new database and store it to file
	CreateBasePaths(t)
	CreateDatabaseFileAndSetFile(t)

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

	return table
}
