package database

import (
	"errors"
	"fmt"
	"os"

	files "github.com/nicolasvancan/monvandb/src/files"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

func LoadDatabase(path string) (*Database, error) {
	// Read the database metadata
	databaseMetadata, err := utils.ReadFromFile(path + utils.SEPARATOR + utils.METDATA_FILE)

	if err != nil {
		return nil, err
	}

	// Deserialize the database metadata
	database := new(Database)
	err = utils.FromJson(databaseMetadata, database)

	if err != nil {
		return nil, err
	}

	// Initialize the tables
	database.Tables = make(map[string]*Table)

	// Load the tables
	for key, value := range database.TablePaths {
		table, err := LoadTable(value)

		if err != nil {
			return nil, err
		}

		database.Tables[key] = table
	}

	return database, nil
}

// Basic Database function Get Table
// Since it returs a pointer to the table struct
// whenever there is a change in the table struct, the change will be reflected in the database struct
func (d *Database) GetTable(tableName string) (*Table, error) {
	if _, ok := d.Tables[tableName]; !ok {
		return nil, errors.New("table not found")
	}

	return d.Tables[tableName], nil
}

// This function is used to create a new table file
// For now, the tables file are simple json files used to store
// metadata related to tables. Other files may also be created in the future
func createNewTableFiles(tableDef Table, path string) error {

	// Serialize tableDef
	json, err := utils.ToJson(tableDef)

	if err != nil {
		return err
	}
	// Create a new table folder
	err = utils.CreateFolder(path, os.ModePerm)

	if err != nil {
		return fmt.Errorf("error creating folder for new table: %v", err)
	}
	metadataPath := path + utils.SEPARATOR + utils.METDATA_FILE
	// Create a new table data file
	_, err = utils.CreateFile(metadataPath)

	if err != nil {
		return fmt.Errorf("error creating metadata file for new table: %v", err)
	}
	// Write the table metadata to the file
	err = utils.WriteToFile(metadataPath, json)

	if err != nil {
		return fmt.Errorf("could not write to table metadata file: %v", err)
	}
	return nil
}

// Basic Database function Create Table
func (d *Database) CreateTable(tableName string, columns []Column) error {
	if _, ok := d.Tables[tableName]; ok {
		return errors.New("table already exists")
	}

	// Create a new table
	tablePath := d.Path + string(os.PathSeparator) + tableName
	// Create new table structure
	newTable := &Table{
		Name:    tableName,
		Columns: columns,
		Path:    tablePath,
	}
	// Create new table files
	err := createNewTableFiles(*newTable, tablePath)

	if err != nil {
		return err
	}

	// Add table to database
	d.Tables[tableName] = newTable

	// Update Database file
	json, err := utils.ToJson(d)

	if err != nil {
		return err
	}

	err = utils.WriteToFile(d.Path+string(os.PathSeparator)+"metadata.json", json)

	if err != nil {
		return fmt.Errorf("could not write to database metadata.json file: %v", err)
	}

	return nil
}

func LoadTable(path string) (*Table, error) {
	// Read the table metadata
	tableMetadata, err := utils.ReadFromFile(path + utils.SEPARATOR + utils.METDATA_FILE)

	if err != nil {
		return nil, err
	}

	// Deserialize the table metadata
	table := new(Table)
	err = utils.FromJson(tableMetadata, table)

	if err != nil {
		return nil, err
	}

	// Initialize the table data file
	table.PDataFile, err = files.OpenDataFile(path + utils.SEPARATOR + utils.TABLE_FILE)

	if err != nil {
		return nil, err
	}

	return table, nil
}
