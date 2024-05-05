package table_structures

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

type Indices struct {
	Primary []string
	Unique  []string
	Foreign []string
}

type Table struct {
	Name    string    // Table's name
	Path    string    // Where the table configuration is stored
	Columns []Column  // reference to Columns
	dFile   *DataFile // Access btree (Simple)
}

// Prototype for Loading an existing Table
func LoadTable(path string) (*Table, error) {
	return nil, nil
}

// Prototype for Creating a new Table
func CreateTable(name string, columns []Column) (*Table, error) {
	return nil, nil
}

func UpdateTable(table *Table, fields interface{}) error {
	return nil
}

func DeleteTable(table *Table) error {
	return nil
}
