package helper

import (
	"testing"

	db "github.com/nicolasvancan/monvandb/src/database"
)

var arr_mock = []string{
	"Joana",
	"Albert",
	"John",
	"Maria",
	"Peter",
	"James",
	"Alice",
}

// Create environment
func createTestTable(t *testing.T) *db.Table {
	table := CreateMockTable(t)
	return table
}

func createMockTableData() []db.RawRow {
	// que current millis from time.Now()
	data := make([]db.RawRow, 0)
	for i := 1; i < 450; i++ {
		data = append(data, map[string]interface{}{
			"id":   int64(i),
			"name": arr_mock[int(i%len(arr_mock))],
		})
	}
	return data
}

func GetMocktableReadyForTesting(t *testing.T) *db.Table {
	table := createTestTable(t)
	mockData := createMockTableData()
	for _, data := range mockData {
		keyValue := table.FromRawRowToKeyValue(data)
		table.PDataFile.Insert(keyValue.Key, keyValue.Value)
	}
	table.PDataFile.ForceSync()

	return table
}
