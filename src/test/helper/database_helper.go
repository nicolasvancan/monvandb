package helper

import (
	"os"
	"testing"

	"github.com/nicolasvancan/monvandb/src/database"
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

	db := new(database.Database)
	db.Name = "db_teste"
	db.Path = databaseFolderName
	db.TablePaths = make(map[string]string)
	db.Tables = make(map[string]*database.Table)
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
	res, err := utils.ToJson(*db)

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

	table, err := db.GetTable("table_teste")

	if err != nil {
		t.Errorf("error getting table: %v", err)
	}

	if table.Name != "table_teste" {
		t.Errorf("expected table_teste, got %v", table.Name)
	}

	table.Indexes = make(map[string]database.Index)
	table.PrimaryKey = table.GetColumnByName("id")

	return table
}
