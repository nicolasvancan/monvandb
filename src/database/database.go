package database

import (
	"errors"
	"fmt"
	"os"

	files "github.com/nicolasvancan/monvandb/src/files"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

func CreateDatabase(name string) (*Database, error) {
	// Create a new database
	database := &Database{
		Name:       name,
		Tables:     make(map[string]*Table),
		TablePaths: make(map[string]string),
		Path:       utils.GetPath("databases") + utils.SEPARATOR + name,
	}

	// Create the database folder
	err := utils.CreateFolder(database.Path, os.ModePerm)

	if err != nil {
		return nil, fmt.Errorf("error creating database folder: %v", err)
	}

	// Create the database metadata file
	_, err = utils.CreateFile(database.Path + utils.SEPARATOR + utils.METDATA_FILE)

	if err != nil {
		return nil, fmt.Errorf("error creating database metadata file: %v", err)
	}

	// Serialize the database
	json, err := utils.ToJson(database)

	if err != nil {
		return nil, err
	}

	// Write the database to the file
	err = utils.WriteToFile(database.Path+utils.SEPARATOR+utils.METDATA_FILE, json)

	if err != nil {
		return nil, fmt.Errorf("could not write to database metadata file: %v", err)
	}

	return database, nil
}

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

func (t *Table) getPrimaryColumns() []Column {
	var primaryColumns []Column

	for _, column := range t.Columns {
		if column.Primary {
			primaryColumns = append(primaryColumns, column)
		}
	}

	return primaryColumns
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
		Indexes: make(map[string]*Index),
	}
	// Create new table files
	err := createNewTableFiles(*newTable, tablePath)

	if err != nil {
		return err
	}

	// Get all Primary columns
	primaryColumns := newTable.getPrimaryColumns()
	if len(primaryColumns) == 0 {
		return errors.New("table must have at least one primary column")
	}

	newTable.PrimaryKey = &primaryColumns[0]
	newTable.CompositeKey = nil

	if len(primaryColumns) > 1 {
		newTable.PrimaryKey = nil
		newTable.CompositeKey = primaryColumns
	}

	// Add table to database
	d.TablePaths[tableName] = tablePath

	// Update Database file
	json, err := utils.ToJson(d)

	if err != nil {
		return err
	}

	err = utils.WriteToFile(d.Path+string(os.PathSeparator)+"metadata.json", json)

	if err != nil {
		return fmt.Errorf("could not write to database metadata.json file: %v", err)
	}

	d.Tables[tableName], err = LoadTable(tablePath)

	if err != nil {
		return err
	}

	return nil
}

// TODO: Create test
func (d *Database) CreateIndex(tableName string, indexedColumn string, indexName string) error {
	table, err := d.GetTable(tableName)

	if err != nil {
		return err
	}

	// Check if the column exists
	column := table.GetColumnByName(indexedColumn)

	if column == nil {
		return fmt.Errorf("column %s does not exist in table %s", indexedColumn, tableName)
	}
	// Create new pointer to DataFile for index
	indexPath := table.Path + utils.SEPARATOR + indexName + ".index.db"
	indexDataFile, err := files.OpenDataFile(indexPath)

	if err != nil {
		return err
	}
	// Create the index
	index := Index{
		Name:      indexName,
		Column:    indexedColumn,
		Path:      indexPath,
		PDataFile: indexDataFile,
	}

	// Add the index to the table
	table.Indexes[indexedColumn] = &index

	// Update the table metadata file
	json, err := utils.ToJson(table)

	if err != nil {
		return err
	}

	err = utils.WriteToFile(table.Path+string(os.PathSeparator)+utils.METDATA_FILE, json)

	if err != nil {
		return fmt.Errorf("could not write to table metadata file: %v", err)
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
