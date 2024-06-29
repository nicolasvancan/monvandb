package main

import (
	"fmt"
	"testing"

	helper "github.com/nicolasvancan/monvandb/src/test/helper"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

func TestBasicScan(t *testing.T) {
	table := helper.GetMocktableReadyForTesting(t)

	key, err := utils.Serialize(int64(1))

	if err != nil {
		t.Errorf("error serializing key: %v", err)
	}

	value := table.PDataFile.Get(key)
	var k int64
	err = utils.Deserialize(value[0].Key, &k)
	if err != nil {
		t.Errorf("error deserializing key: %v", err)
	}

	fmt.Printf("value: %d\n", k)
	if value == nil {
		t.Errorf("error getting value from datafile")
	}

}

func main() {
	TestBasicScan(&testing.T{})
}
