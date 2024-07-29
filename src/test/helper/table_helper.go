package helper

import (
	"testing"

	database "github.com/nicolasvancan/monvandb/src/database"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

func CreateMockTableAndIndex(t *testing.T) *database.Table {
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
		{
			Name: "email",
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

	table.Indexes = make(map[string]*database.Index)
	table.PrimaryKey = table.GetColumnByName("id")

	// Create index for table
	err = db.CreateIndex("table_teste", "email", "email_index")

	if err != nil {
		t.Errorf("error creating index: %v", err)
	}

	return table
}

func CreateMockTableAndIndexWithRows(t *testing.T) *database.Table {
	table := CreateMockTableAndIndex(t)

	// Insert some rows
	table.Insert([]database.RawRow{
		{
			"id":    int32(1),
			"name":  "Nicolas",
			"email": "nicolas@nicolas.com",
		},
		{
			"id":    int32(2),
			"name":  "Paulo",
			"email": "paulo@paulo.com",
		},
	})

	return table
}
