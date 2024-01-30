package main

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	bTree "github.com/nicolasvancan/monvandb/src/btree"
	helper "github.com/nicolasvancan/monvandb/src/test/helper"
)

/*
 * Tests for bTree basic methodes. The first version will be simple.
 */

func fillUpLeafUntilItSplits(btree *bTree.BTree) {
	// Fill up leaf until it splits
	for i := 0; i < 282; i++ {
		bTree.BTreeInsert(btree, []byte(strconv.Itoa(i)), []byte(string("teste_"+strconv.Itoa(i))))
	}
}

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
	btreeName := tree.GetName()[:8]
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

func TestSimpleBTreeInsertion(t *testing.T) {
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := setupTests(t)

	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)

	// Read file stat from Fp
	fs, _ := helper.Fp.Stat()

	if fs.Size() != 4096 {
		t.Errorf("File size should be 4096, found %d\n", fs.Size())
	}

	key := []byte("1")
	value := []byte("first")
	// Insert key value to bTree
	bTree.BTreeInsert(tree, key, value)

	// Here we should see the size of the document
	// It must have increased by 4096 bytes, which is the size of a page
	fs, _ = helper.Fp.Stat()

	if fs.Size() != 4096*2 {
		t.Errorf("File size should be 9192, found %d\n", fs.Size())
	}

	// Get first page and check whether the element is there.
	node := tree.Get(1)
	kval := node.GetLeafKeyValueByIndex(0)

	if bytes.Compare(kval.GetKey(), []byte("1")) != 0 {
		t.Errorf("Bytes are not equivalent expected string('1'), found string('%s')", kval.GetKey())
	}
}

func TestInsertMultipleLines(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := setupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)

	// fill Leaf until it splits
	fillUpLeafUntilItSplits(tree)

	// Check file stat, it should have Header page = 4096 + three leaves = 3*4096 + one internal node = 4096
	// It should have in total 5 * 4096 = 20480 bytes

	fs, _ := helper.Fp.Stat()

	if fs.Size() != 20480 {
		t.Errorf("File size should be 20480, received %d\n", fs.Size())
	}

}
