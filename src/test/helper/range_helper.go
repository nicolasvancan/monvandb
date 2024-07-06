package helper

import (
	"testing"

	"strconv"

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
	table := CreateMockTableAndIndex(t)
	return table
}

func createMockTableData() []db.RawRow {
	// que current millis from time.Now()
	data := make([]db.RawRow, 0)
	for i := 1; i < 450; i++ {
		data = append(data, map[string]interface{}{
			"id":    int64(i),
			"name":  arr_mock[int(i%len(arr_mock))],
			"email": "teste@teste.com_" + strconv.Itoa(int(i)),
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

var QueryOne = []db.ColumnComparsion{
	{
		ColumnName: "id",
		TableName:  "users",
		Condition:  db.GT,
		Alias:      "",
		Value: db.ColumnConditionValue{
			IsOtherColumn:        false,
			IsOtherTable:         false,
			ColumnName:           "",
			TableHash:            "",
			Value:                10,
			Transformation:       nil,
			TransformationParams: nil,
		},
		ParentId:        -1,
		ParentLogicalOp: db.AND,
		LayerLogicalOp:  db.AND,
		Id:              0,
	},
	{
		ColumnName: "id",
		TableName:  "users",
		Condition:  db.LT,
		Alias:      "",
		Value: db.ColumnConditionValue{
			IsOtherColumn:        false,
			IsOtherTable:         false,
			ColumnName:           "",
			TableHash:            "",
			Value:                20,
			Transformation:       nil,
			TransformationParams: nil,
		},
		ParentId:        -1,
		ParentLogicalOp: db.AND,
		LayerLogicalOp:  db.AND,
		Id:              0,
	},
}

var QueryTwo = []db.ColumnComparsion{
	QueryOne[0],
	QueryOne[1],
	{
		ColumnName: "name",
		TableName:  "users",
		Condition:  db.EQ,
		Alias:      "",
		Value: db.ColumnConditionValue{
			IsOtherColumn:        false,
			IsOtherTable:         false,
			ColumnName:           "",
			TableHash:            "",
			Value:                "Joana",
			Transformation:       nil,
			TransformationParams: nil,
		},
		ParentId:        -1,
		ParentLogicalOp: db.AND,
		LayerLogicalOp:  db.AND,
		Id:              0,
	},
}

/*
SELECT * FROM users u
 INNER JOIN fk_table f ON u.fk = f.id
WHERE users.id in (1,2,3)
*/

var QueryThree = []db.ColumnComparsion{
	{
		ColumnName: "id",
		TableName:  "users",
		Condition:  db.EQ,
		Alias:      "",
		Value: db.ColumnConditionValue{
			IsOtherColumn:        true,
			IsOtherTable:         true,
			ColumnName:           "fk",
			TableHash:            "1234",
			Value:                nil,
			Transformation:       nil,
			TransformationParams: nil,
		},
		ParentId:        -1,
		ParentLogicalOp: db.AND,
		LayerLogicalOp:  db.AND,
		Id:              0,
	},
	{
		ColumnName: "id",
		TableName:  "users",
		Condition:  db.IN,
		Alias:      "",
		Value: db.ColumnConditionValue{
			IsOtherColumn:        false,
			IsOtherTable:         false,
			ColumnName:           "",
			TableHash:            "",
			Value:                []int64{1, 2, 3},
			Transformation:       nil,
			TransformationParams: nil,
		},
		ParentId:        -1,
		ParentLogicalOp: db.AND,
		LayerLogicalOp:  db.AND,
		Id:              0,
	},
}

/*
SELECT * FROM users WHERE id > 10 AND email < 'Joana'
*/

var QueryFour = []db.ColumnComparsion{
	{
		ColumnName: "id",
		TableName:  "users",
		Condition:  db.GT,
		Alias:      "",
		Value: db.ColumnConditionValue{
			IsOtherColumn:        false,
			IsOtherTable:         false,
			ColumnName:           "",
			TableHash:            "",
			Value:                10,
			Transformation:       nil,
			TransformationParams: nil,
		},
		ParentId:        -1,
		ParentLogicalOp: db.AND,
		LayerLogicalOp:  db.AND,
		Id:              0,
	},
	{
		ColumnName: "email",
		TableName:  "users",
		Condition:  db.LT,
		Alias:      "",
		Value: db.ColumnConditionValue{
			IsOtherColumn:        false,
			IsOtherTable:         false,
			ColumnName:           "",
			TableHash:            "",
			Value:                "Joana",
			Transformation:       nil,
			TransformationParams: nil,
		},
		ParentId:        -1,
		ParentLogicalOp: db.AND,
		LayerLogicalOp:  db.AND,
		Id:              0,
	},
}
