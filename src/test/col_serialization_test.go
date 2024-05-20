package main

import (
	"os"
	"testing"

	table "github.com/nicolasvancan/monvandb/src/database"
	files "github.com/nicolasvancan/monvandb/src/files"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

const error_v_string string = "Error: %v != %v"

func TestSerialization(t *testing.T) {
	// Create a new column
	tmp := &table.ColumnValue{Col: 1, Value: uint8(10)}

	res, err := utils.Serialize(*tmp)

	if err != nil {
		t.Errorf("error encoding struct: %s", err)
	}
	tmp2 := new(table.ColumnValue)
	err = utils.Deserialize(res, tmp2)

	if err != nil {
		t.Errorf("error decoding struct: %s", err)
	}

	if tmp.Col != tmp2.Col || tmp.Value != tmp2.Value {
		t.Errorf(error_v_string, tmp, tmp2)
	}

	// Test for typing
	switch tmp2.Value.(type) {
	case uint8:
		break
	default:
		t.Errorf(error_v_string, tmp, tmp2)
	}
}

func TestSerializationWithDataFile(t *testing.T) {
	// Create tmp dir path
	tmpDirPath := t.TempDir()

	// Get final temporary path for DataFile file
	dbFilePath := tmpDirPath + string(os.PathSeparator) + "test.db"

	// Create a mock row with 4 columns, each with a different type
	row := make(table.RowValues, 5)

	row[0] = table.ColumnValue{Col: 1, Value: uint8(10)}
	row[1] = table.ColumnValue{Col: 2, Value: "Hello"}
	row[2] = table.ColumnValue{Col: 3, Value: 0.8}
	row[3] = table.ColumnValue{Col: 4, Value: 123456}
	row[4] = table.ColumnValue{Col: 4, Value: nil}

	// Serialize them
	serializedRow, err := utils.Serialize(row)

	if err != nil {
		t.Errorf("Error encoding struct: %s", err)
	}

	// Open the DataFile (Create new if does not exist)
	dFile, err := files.OpenDataFile(dbFilePath)

	if err != nil {
		t.Errorf("Error opening data file: %s", err)
	}

	// Insert one key
	dFile.Insert([]byte("1"), serializedRow)
	dFile.ForceSync()

	dFile.Close()

	// Reopen recent closed DataFile
	dFile, err = files.OpenDataFile(dbFilePath)

	if err != nil {
		t.Errorf("Error opening data file: %s", err)
	}

	// Create new variable to hold value returned from DataFile
	serializedRow2 := dFile.Get([]byte("1"))

	if serializedRow2 == nil {
		t.Errorf("Error getting data from data file")
	}

	var row2 table.RowValues
	// Deserialize it
	err = utils.Deserialize(serializedRow2[0].Value, &row2)

	if err != nil {
		t.Errorf("Error decoding struct: %s", err)
	}

	if len(row) != len(row2) {
		t.Errorf(error_v_string, row, row2)
	}

	for i := 0; i < len(row); i++ {
		if row[i].Col != row2[i].Col || row[i].Value != row2[i].Value {
			t.Errorf(error_v_string, row, row2)
		}
	}
}

func TestSerializationWithColumnRawRows(t *testing.T) {
	myTable := new(table.Table)
	myTable.Columns = make([]table.Column, 5)

	myTable.Columns[0] = table.Column{Name: "test", Type: table.COL_TYPE_INT, Default: nil, Nullable: false, AutoIncrement: false, Primary: false, Unique: false, Foreign: false}
	myTable.Columns[1] = table.Column{Name: "test2", Type: table.COL_TYPE_STRING, Default: nil, Nullable: false, AutoIncrement: false, Primary: false, Unique: false, Foreign: false}
	myTable.Columns[2] = table.Column{Name: "test3", Type: table.COL_TYPE_FLOAT, Default: nil, Nullable: false, AutoIncrement: false, Primary: false, Unique: false, Foreign: false}
	myTable.Columns[3] = table.Column{Name: "test4", Type: table.COL_TYPE_BIG_INT, Default: nil, Nullable: false, AutoIncrement: false, Primary: false, Unique: false, Foreign: false}
	myTable.Columns[4] = table.Column{Name: "test5", Type: table.COL_TYPE_INT, Default: nil, Nullable: true, AutoIncrement: false, Primary: false, Unique: false, Foreign: false}

	// Create a new column
	tmp := make(table.RawRow)
	tmp["test"] = 10
	tmp["test2"] = "Hello"
	tmp["test3"] = 0.8
	tmp["test4"] = 123456
	tmp["test5"] = nil

	res, err := utils.Serialize(myTable.MapRowToColumnValues(tmp))

	if err != nil {
		t.Errorf("Error encoding struct: %s", err)
	}
	var tmp2 []table.ColumnValue = nil
	err = utils.Deserialize(res, &tmp2)
	deserialized := myTable.FromColumnValuesToRow(tmp2)

	if err != nil {
		t.Errorf("Error decoding struct: %s", err)
	}

	if len(tmp) != len(tmp2) {
		t.Errorf(error_v_string, tmp, tmp2)
	}

	for key, value := range deserialized {
		if tmp[key] != value {
			t.Errorf(error_v_string, tmp, deserialized)
		}
	}

}
