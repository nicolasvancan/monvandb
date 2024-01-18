package main

import (
	"fmt"
	"testing"

	helper "github.com/nicolasvancan/monvandb/src/test/helper"
)

/*
 * Tests for bTree basic methodes. The first version will be simple.
 */

// Basic setup for testing
func setupTests(t *testing.T) string {
	// Create tmp file path
	tmpFilePath := t.TempDir()
	t.Logf("Created tmpFile path %s\n", tmpFilePath)
	// Create a new bTree
	return helper.CreateBtreeFileAndSetFile(t, tmpFilePath)
}

func cleanUp() {
	// We close Fp after test is concluded
	fmt.Println("Cleaning up test")
	defer helper.Fp.Close()
}

func TestSimpleBTreeLoad(t *testing.T) {
	dbFilePath := setupTests(t)
	t.Logf("Tmp db FIlename is %s\n", dbFilePath)

	// Load bTree
	t.Log("Loading bTree to be used")

	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	btreeName := tree.GetName()
	bTreeRoot := tree.GetRoot()

	if bTreeRoot != 0 {
		t.Errorf("Root should be 0, found %d\n", bTreeRoot)
	}

	if btreeName != "db_teste" {
		t.Errorf("Btree name should be db_teste, found %s\n", btreeName)
	}
	t.Logf("Tmp db name is %s\n", btreeName)
	t.Logf("Tmp db root is %d\n", bTreeRoot)
	t.Cleanup(cleanUp)
}
