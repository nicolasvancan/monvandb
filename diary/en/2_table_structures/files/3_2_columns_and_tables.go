package table_structures

import (
	files "github.com/nicolasvancan/monvandb/src/files"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

type Column struct {
	Name          string
	Type          int
	Default       interface{}
	Nullable      bool
	AutoIncrement bool
	Primary       bool
	Unique        bool
	Foreign       bool
}

type Table struct {
	Name      string          // Table's name
	Path      string          // Where the table configuration is stored
	Columns   []Column        // reference to Columns
	PDataFile *files.DataFile // Access btree (Simple)
}

// Prototype for Loading an existing Table
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
