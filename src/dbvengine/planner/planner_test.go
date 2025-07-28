package planner

import (
	"fmt"
	"os"
	"testing"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"github.com/nicolasvancan/monvandb/src/database"
	"github.com/nicolasvancan/monvandb/src/dbvengine/contexts"
	"github.com/nicolasvancan/monvandb/src/dbvengine/executor"
	parser "github.com/nicolasvancan/monvandb/src/dbvengine/parser"
	"github.com/nicolasvancan/monvandb/src/utils"
)

func TestTableDrop(t *testing.T) {
	TestTableCreate(t)
	mock_db, _ := database.GetDatabase("mock")

	table, _ := mock_db.GetTable("my_table")

	fmt.Println(table.Name)
	query := "DROP TABLE my_table"
	parsedQuery, err := sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzedQuery := parser.AnalyzeQuery("mock", parsedQuery)
	if analyzedQuery.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzedQuery.Error())
	}

	// Create a new execution layer
	execLayer := executor.NewExecutionLayer(contexts.NewExecutionContext())

	// Build the plan
	BuildPlan(analyzedQuery, execLayer)

	execLayer.Start()

	mock_db, _ = database.GetDatabase("mock")

	table, _ = mock_db.GetTable("my_table")

	if table != nil {
		t.Errorf("expected nil, got %v", table)
	}
}

func TestTableCreate(t *testing.T) {
	CreateBasePaths(t)
	db, err := database.CreateDatabase("mock")

	if err != nil {
		t.Errorf("error creating database: %v", err)
	}

	if db.Name != "mock" {
		t.Errorf("expected mock, got %v", db.Name)
	}

	query := "CREATE TABLE my_table (id INT PRIMARY KEY, name TEXT)"
	parsedQuery, err := sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzedQuery := parser.AnalyzeQuery("mock", parsedQuery)
	if analyzedQuery.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzedQuery.Error())
	}

	// Create a new execution layer
	execLayer := executor.NewExecutionLayer(contexts.NewExecutionContext())

	// Build the plan
	BuildPlan(analyzedQuery, execLayer)

	execLayer.Start()

	db, err = database.GetDatabase("mock")

	if err != nil {
		t.Errorf("error getting database: %v", err)
	}

	table, err := db.GetTable("my_table")

	if err != nil {
		t.Errorf("error getting table: %v", err)
	}

	if table.Name != "my_table" {
		t.Errorf("expected my_table, got %v", table.Name)
	}
}

func TestTableInsert(t *testing.T) {
	CreateMockTable(t)

	query := "INSERT INTO table_teste (name) VALUES ('name2'),('name2'),('name2')"
	parsedQuery, err := sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzedQuery := parser.AnalyzeQuery("mock", parsedQuery)
	if analyzedQuery.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzedQuery.Error())
	}

	// Create a new execution layer
	execLayer := executor.NewExecutionLayer(contexts.NewExecutionContext())

	// Build the plan
	BuildPlan(analyzedQuery, execLayer)

	execLayer.Start()

	query = "UPDATE table_teste SET name = 'name3' WHERE id = 1"
	parsedQuery, err = sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzedQuery = parser.AnalyzeQuery("mock", parsedQuery)
	if analyzedQuery.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzedQuery.Error())
	}

	// Create a new execution layer
	execLayer = executor.NewExecutionLayer(contexts.NewExecutionContext())
	// Build the plan
	BuildPlan(analyzedQuery, execLayer)

	execLayer.Start()
	fmt.Println(execLayer.Context.Result)

	// Create a new execution layer
	execLayer = executor.NewExecutionLayer(contexts.NewExecutionContext())

	// Build the plan
	BuildPlan(analyzedQuery, execLayer)

	execLayer.Start()
	fmt.Println("Starting delete")
	query = "DELETE FROM table_teste WHERE id = 1"
	parsedQuery, err = sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzedQuery = parser.AnalyzeQuery("mock", parsedQuery)
	if analyzedQuery.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzedQuery.Error())
	}

	// Create a new execution layer
	execLayer = executor.NewExecutionLayer(contexts.NewExecutionContext())
	// Build the plan
	BuildPlan(analyzedQuery, execLayer)

	execLayer.Start()
	fmt.Println(execLayer.Context.Result)

	fmt.Println("Starting Final select")
	query = "SELECT t.* FROM table_teste t "
	parsedQuery, err = sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzedQuery = parser.AnalyzeQuery("mock", parsedQuery)
	if analyzedQuery.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzedQuery.Error())
	}

	// Create a new execution layer
	execLayer = executor.NewExecutionLayer(contexts.NewExecutionContext())
	// Build the plan
	BuildPlan(analyzedQuery, execLayer)

	execLayer.Start()
	fmt.Println(execLayer.Context.Result)

}

func TestCreateTableQuery(t *testing.T) {
	CreateMockTable(t)

	query := "CREATE TABLE table_teste3 (id INT PRIMARY KEY, name TEXT)"
	parsedQuery, err := sqlparser.Parse(query)

	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzedQuery := parser.AnalyzeQuery("mock", parsedQuery)
	if analyzedQuery.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzedQuery.Error())
	}

	// Create a new execution layer
	execLayer := executor.NewExecutionLayer(contexts.NewExecutionContext())

	// Build the plan
	BuildPlan(analyzedQuery, execLayer)

	execLayer.Start()
	fmt.Println(execLayer.Context.Result)
	t.Error("asd")
}
func TestSimpleQueryBuild(t *testing.T) {
	CreateMockTable(t)

	query := "SELECT t.id, t.name as bang, t2.other_column aiusop FROM table_teste t INNER JOIN table_teste2 t2 ON t.id = t2.id"
	parsedQuery, err := sqlparser.Parse(query)
	if err != nil {
		t.Errorf("error parsing query: %v", err)
	}

	analyzedQuery := parser.AnalyzeQuery("mock", parsedQuery)
	if analyzedQuery.Error() != nil {
		t.Errorf("error analyzing query: %v", analyzedQuery.Error())
	}

	// Create a new execution layer
	execLayer := executor.NewExecutionLayer(contexts.NewExecutionContext())
	fmt.Println(analyzedQuery)
	// Build the plan
	BuildPlan(analyzedQuery, execLayer)
	//fmt.Printf("%v", execLayer.Nodes)
	//t.Error("asd")
	execLayer.Start()
	fmt.Printf("%v\n", execLayer.Context.Result)
	t.Error("asd")
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
			Name:          "id",
			Type:          database.COL_TYPE_INT,
			Primary:       true,
			AutoIncrement: true,
		},
		{
			Name:    "name",
			Type:    database.COL_TYPE_STRING,
			Primary: false,
		},
	}, false, false)

	if err != nil {
		t.Errorf("error creating table: %v", err)
	}

	table1, err := db.GetTable("table_teste")

	if err != nil {
		t.Errorf("error getting table: %v", err)
	}

	if table1.Name != "table_teste" {
		t.Errorf("expected table_teste, got %v", table1.Name)
	}

	table1.Indexes = make(map[string]*database.Index)
	table1.PrimaryKey = table1.GetColumnByName("id")

	table1.Insert([]map[string]interface{}{{"id": 1, "name": "name1"}})

	// Create a new table
	db.CreateTable("table_teste2", []database.Column{
		{
			Name:    "id",
			Type:    database.COL_TYPE_INT,
			Primary: true,
		},
		{
			Name:    "other_column",
			Type:    database.COL_TYPE_STRING,
			Primary: false,
		},
	}, false, false)

	table, _ := db.GetTable("table_teste2")

	table.Indexes = make(map[string]*database.Index)
	table.PrimaryKey = table.GetColumnByName("id")
	table.Insert([]map[string]interface{}{{"id": 1, "other_column": "other_column1"}})
	return table1
}

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
