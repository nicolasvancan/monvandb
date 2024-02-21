package main

import (
	"bytes"
	"encoding/binary"
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

func fillUpLeafWithNumericValuesUntilItSplits(btree *bTree.BTree) {
	for i := 1; i <= 10000; i++ {
		// I'm going to use little endian 32 bits so 4 bytes
		integerBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(integerBytes, uint32(i))
		bTree.BTreeInsert(btree, integerBytes, []byte(string("teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_"+strconv.Itoa(i))))
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

	// Test for getting all data added data
	for i := 0; i < 282; i++ {
		keyValue := bTree.BTreeGetOne(tree, []byte(strconv.Itoa(i)))

		if keyValue == nil {
			t.Errorf("Should have found the key %s\n", keyValue.Key)
		}
	}
}

func TestInsertMultipleLinesForLargeInt(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := setupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	// Fillup with sequencial bytes
	fillUpLeafWithNumericValuesUntilItSplits(tree)

	for i := 1; i <= 10000; i++ {
		integerBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(integerBytes, uint32(i))
		res := bTree.BTreeGetOne(tree, integerBytes)
		if res == nil {
			tmp := binary.BigEndian.Uint32(res.Key)
			t.Errorf("Did not find a value for %d\n", tmp)
		}
	}
}

func TestFileMapping(t *testing.T) {
	t.Log("Creating basic database to test mapping")
	dbFilePath := setupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	tree = helper.CreateFakeDbPagesForMapping(t, tree)

	mapped := bTree.MapAllLeavesToArray(tree)
	expectedPages := []uint64{4, 5, 6, 7, 8, 9}
	for i := 0; i < len(mapped); i++ {

		if expectedPages[i] != (mapped[i].TreeNode) {
			t.Errorf("Mapped leaves equal %d Should be %d\n", mapped[i].TreeNode, i+4)
		}

		if len(mapped[i].History) != 2 {
			t.Errorf("History should have length of 2 found %d\n", len(mapped[i].History))
		}
	}
}
