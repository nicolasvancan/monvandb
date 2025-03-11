package parser

import (
	"os"
	"testing"

	sqlparser "github.com/blastrain/vitess-sqlparser/sqlparser"
	database "github.com/nicolasvancan/monvandb/src/database"
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
	db.Name = "mock"
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
	}, false, false)

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

	// Create a new table
	db.CreateTable("table_teste2", []database.Column{
		{
			Name:    "id",
			Type:    database.COL_TYPE_INT,
			Primary: true,
		},
		{
			Name: "other_column",
			Type: database.COL_TYPE_STRING,
		},
	}, false, false)

	table, _ = db.GetTable("table_teste2")

	table.Indexes = make(map[string]*database.Index)
	table.PrimaryKey = table.GetColumnByName("id")

	return table
}

func TestFromAnalyzis(t *testing.T) {

	CreateMockTable(t)

	query := "SELECT t.name as name, id as IDZAO FROM table_teste as t INNER JOIN table_teste2 t2 ON t.id = t2.id AND t.id > 20 WHERE t.id = 1"
	stmt, err := sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzed := AnalyzeQuery("mock", stmt)

	if analyzed.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzed.Error())
	}

	switch analyzed := analyzed.(type) {
	case *AnalyzedQuerySelect:
		if analyzed.DatabaseName != "mock" {
			t.Errorf("expected mock, got %v", analyzed.DatabaseName)
		}

		// analyze From
		from := analyzed.From
		if from.Alias != "t" && from.Table != "table_teste" {
			t.Errorf("expected t, got %v", from.Alias)
		}

		// analyze joins
		joins := analyzed.Joins
		if len(joins) != 1 {
			t.Errorf("expected 1, got %v", len(joins))
		}

		if joins["t-t2"].LeftAlias != "t" &&
			joins["t-t2"].RightAlias != "t2" &&
			joins["t-t2"].On.LeftType != "column" &&
			joins["t-t2"].On.RightType != "column" &&
			joins["t-t2"].On.LeftValue.(string) != "id" &&
			joins["t-t2"].On.RightValue.(string) != "id" {
			t.Error("Wrong join")
		}

		// analyze filters
		tableFilters := analyzed.TablesFilters
		if len(tableFilters) != 1 {
			t.Errorf("expected 1, got %v", len(tableFilters))
		}

		filterT := tableFilters["t"]
		resolvedFilters := filterT.Resolve()
		firstFilter := resolvedFilters[0]
		if firstFilter.Values[0].Comparando != 20 &&
			firstFilter.Values[0].Comparando != 1 {
			t.Errorf("expected 1, got %v", firstFilter.Values[0].Comparando)
		}
	default:
		t.Errorf("expected AnalyzedQuerySelect, got %v", analyzed)
	}
}

func TestCreateTableAnalysis(t *testing.T) {
	CreateMockTable(t)

	query := "CREATE TABLE table_teste3 (id INT PRIMARY KEY, name TEXT)"
	stmt, err := sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzed := AnalyzeQuery("mock", stmt)

	if analyzed.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzed.Error())
	}

	switch analyzed := analyzed.(type) {
	case *AnalyzedQueryCreateTable:
		if analyzed.DatabaseName != "mock" {
			t.Errorf("expected mock, got %v", analyzed.DatabaseName)
		}

		if analyzed.TableName != "table_teste3" {
			t.Errorf("expected table_teste3, got %v", analyzed)
		}

		if len(analyzed.Columns) != 2 {
			t.Errorf("expected 2, got %v", len(analyzed.Columns))
		}

		if analyzed.Columns[0].Name != "id" &&
			analyzed.Columns[0].Type != database.COL_TYPE_INT &&
			analyzed.Columns[0].Primary != true {
			t.Errorf("wrong column")
		}

		if analyzed.Columns[1].Name != "name" &&
			analyzed.Columns[1].Type != database.COL_TYPE_STRING &&
			analyzed.Columns[1].Primary != false {
			t.Errorf("wrong column")
		}
	default:
		t.Errorf("expected AnalyzedQueryCreateTable, got %v", analyzed)
	}
}

func TestInsertFunctionStaticValues(t *testing.T) {
	CreateMockTable(t)

	query := "INSERT INTO table_teste (id, name) VALUES (1, 'nicolas')"
	stmt, err := sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzed := AnalyzeQuery("mock", stmt)

	if analyzed.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzed.Error())
	}

	switch analyzed := analyzed.(type) {
	case *AnalyzedQueryInsert:
		if analyzed.DatabaseName != "mock" {
			t.Errorf("expected mock, got %v", analyzed.DatabaseName)
		}

		if analyzed.TableName != "table_teste" {
			t.Errorf("expected table_teste, got %v", analyzed.TableName)
		}

		if len(analyzed.Columns) != 2 {
			t.Errorf("expected 2, got %v", len(analyzed.Columns))
		}

		if analyzed.Columns[0] != "id" &&
			analyzed.Columns[1] != "name" {
			t.Errorf("wrong columns")
		}

		if len(analyzed.Values.([]database.RawRow)) != 1 {
			t.Errorf("expected 1, got %v", len(analyzed.Values.([]database.RawRow)))
		}

		if len(analyzed.Values.([]database.RawRow)[0]) != 2 {
			t.Errorf("expected 2, got %v", len(analyzed.Values.([]database.RawRow)[0]))
		}

		if analyzed.Values.([]database.RawRow)[0]["id"] != 1 &&
			analyzed.Values.([]database.RawRow)[0]["name"] != "nicolas" {
			t.Errorf("wrong values")
		}
	default:
		t.Errorf("expected AnalyzedQueryInsert, got %v", analyzed)
	}
}
