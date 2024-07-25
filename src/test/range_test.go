package main

import (
	"testing"

	database "github.com/nicolasvancan/monvandb/src/database"
	helper "github.com/nicolasvancan/monvandb/src/test/helper"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

func TestBasicScan(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	// Create a range from 1 to 3
	from, _ := utils.Serialize(int64(1))
	to, _ := utils.Serialize(int64(3))

	rangeOf := database.RangeOptions{
		From:        from,
		To:          to,
		FComparator: database.GTE,
		TComparator: database.GT,
		Order:       database.ASC,
		Limit:       -1,
		PDataFile:   table.PDataFile,
	}

	rangeData, err := database.RangeFromOptions(table, rangeOf)

	if err != nil {
		t.Errorf("error getting range: %v", err)
	}

	if len(rangeData) != 3 {
		t.Errorf("error getting value from datafile %d\n", len(rangeData))
	}
}

func TestBasicScanWholeData(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	rangeOf := database.RangeOptions{
		From:        nil,
		To:          nil,
		FComparator: database.GTE, // FComparator delimitters the condition to start range
		TComparator: database.GTE, // TComparator delimitters the condition to stop range
		Order:       database.ASC,
		Limit:       -1,
		PDataFile:   table.PDataFile,
	}

	rangeData, err := database.RangeFromOptions(table, rangeOf)

	if err != nil {
		t.Errorf("error getting range: %v", err)
	}

	if len(rangeData) != 449 {
		t.Errorf("error getting value from datafile %d\n", len(rangeData))
	}
}

func TestFullScanForNilFromField(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	to, _ := utils.Serialize(int64(3))

	rangeOf := database.RangeOptions{
		From:        nil,
		To:          to,
		FComparator: database.GTE, // FComparator delimitters the condition to start range
		TComparator: database.GT,  // TComparator delimitters the condition to stop range
		Order:       database.ASC,
		Limit:       -1,
		PDataFile:   table.PDataFile,
	}

	rangeData, err := database.RangeFromOptions(table, rangeOf)

	if err != nil {
		t.Errorf("error getting range: %v", err)
	}

	if len(rangeData) != 3 {
		t.Errorf("error getting value from datafile %d\n", len(rangeData))
	}
}

func TestFullScanForNilToField(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	from, _ := utils.Serialize(int64(447))

	rangeOf := database.RangeOptions{
		From:        from,
		To:          nil,
		FComparator: database.GTE, // FComparator delimitters the condition to start range
		TComparator: database.GT,  // TComparator delimitters the condition to stop range
		Order:       database.ASC,
		Limit:       -1,
		PDataFile:   table.PDataFile,
	}

	rangeData, err := database.RangeFromOptions(table, rangeOf)

	if err != nil {
		t.Errorf("error getting range: %v", err)
	}

	if len(rangeData) != 2 {
		t.Errorf("error getting value from datafile %d\n", len(rangeData))
	}
}

func TestScanForLimit(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	rangeOf := database.RangeOptions{
		From:        nil,
		To:          nil,
		FComparator: database.GTE, // FComparator delimitters the condition to start range
		TComparator: database.GT,  // TComparator delimitters the condition to stop range
		Order:       database.ASC,
		Limit:       10,
		PDataFile:   table.PDataFile,
	}

	rangeData, err := database.RangeFromOptions(table, rangeOf)

	if err != nil {
		t.Errorf("error getting range: %v", err)
	}

	if len(rangeData) != 10 {
		t.Errorf("error getting value from datafile %d\n", len(rangeData))
	}

	if rangeData[0]["id"] != int64(1) {
		t.Errorf("error getting value from datafile %d\n", rangeData[0]["id"])
	}
}
