package main

/*
This test file is specific for range purposes. We will test the table Range function specifically. Asserting values
related to the RangeOptions provided by some query.

A lot of different queries will be part of this test. The idea is to test the range function with different queries
*/
import (
	"fmt"
	"testing"

	database "github.com/nicolasvancan/monvandb/src/database"
	helper "github.com/nicolasvancan/monvandb/src/test/helper"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

/*
Query
SELECT * FROM users WHERE id > 10 AND id < 20
*/

func TestMergeOperationsCaseOne(t *testing.T) {
	//k Create a table
	table := helper.GetMocktableReadyForTesting(t)
	// Add columns
	rangeOptions := database.MergeOperationsBasedOnIndexedColumnsAndReturnRangeOptions(table, helper.QueryOne)

	fmt.Printf("%d\n", rangeOptions.To)
	var from int64
	var to int64
	utils.Deserialize(rangeOptions.From, &from)
	utils.Deserialize(rangeOptions.To, &to)

	if from != 10 {
		t.Errorf("expected 10 as result, got %v", from)
	}

	if to != 20 {
		t.Errorf("expected 20 as result, got %v", to)
	}

}

/*
SELECT * FROM users WHERE id > 10 AND id < 20 AND name = 'Joana'
*/

func TestMergeOperationsCaseTwo(t *testing.T) {
	// Create a table
	table := helper.GetMocktableReadyForTesting(t)
	// Add columns
	rangeOptions := database.MergeOperationsBasedOnIndexedColumnsAndReturnRangeOptions(table, helper.QueryTwo)

	var from int64
	var to int64

	utils.Deserialize(rangeOptions.From, &from)
	utils.Deserialize(rangeOptions.To, &to)

	if from != 10 {
		t.Errorf("expected 10 as result, got %v", from)
	}

	if to != 20 {
		t.Errorf("expected 20 as result, got %v", to)
	}
}

/*
SELECT * FROM users u
 INNER JOIN fk_table f ON u.fk = f.id
WHERE users.id in (1,2,3)
*/

func TestMergeOperationsCaseThree(t *testing.T) {
	// Create a table
	table := helper.GetMocktableReadyForTesting(t)
	// Add columns
	rangeOptions := database.MergeOperationsBasedOnIndexedColumnsAndReturnRangeOptions(table, helper.QueryThree)

	if rangeOptions.From != nil {
		t.Error("expected nil as result")
	}

	if rangeOptions.To != nil {
		t.Error("expected nil as result")
	}

}

/*
To indexed columns

SELECT * FROM users WHERE id > 10 AND email < 'Joana'
*/

func TestMergeOperationsCaseFour(t *testing.T) {
	// Create a table
	table := helper.GetMocktableReadyForTesting(t)
	// Add columns
	rangeOptions := database.MergeOperationsBasedOnIndexedColumnsAndReturnRangeOptions(table, helper.QueryFour)

	var from int64
	var to int64

	if rangeOptions.From != nil {
		utils.Deserialize(rangeOptions.From, &from)
	}

	if rangeOptions.To != nil {
		utils.Deserialize(rangeOptions.To, &to)
	}

	if from != 10 {
		t.Errorf("expected 10 as result, got %d", from)
	}

	if rangeOptions.To != nil {
		t.Errorf("expected value != than nil as result, got nil")
	}

}
