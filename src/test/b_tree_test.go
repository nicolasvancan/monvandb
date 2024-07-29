package main

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"testing"

	bTree "github.com/nicolasvancan/monvandb/src/btree"
	helper "github.com/nicolasvancan/monvandb/src/test/helper"
)

/*
 * Tests for bTree basic methodes. The first version will be simple.
 */
const LOADING_B_TREE = "loading bTree to be used for testing"

func TestSimpleBTreeLoad(t *testing.T) {
	dbFilePath := helper.SetupTests(t)
	t.Logf("Tmp db FIlename is %s\n", dbFilePath)

	// Load bTree
	t.Log(LOADING_B_TREE)

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
	t.Cleanup(helper.CleanUp)
}

func TestSimpleBTreeInsertion(t *testing.T) {
	t.Log("starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)

	// Load bTree
	t.Log(LOADING_B_TREE)
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

	if !bytes.Equal(kval.GetKey(), []byte("1")) {
		t.Errorf("Bytes are not equivalent expected string('1'), found string('%s')", kval.GetKey())
	}
}

func TestInsertMultipleLines(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log(LOADING_B_TREE)
	tree := helper.LoadBTreeFromPath(t, dbFilePath)

	// fill Leaf until it splits
	helper.FillUpLeafUntilItSplits(tree)

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
			t.Errorf("Should have found the key %d\n", i)
		}
	}
}

func TestInsertMultipleLinesForLargeInt(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log(LOADING_B_TREE)
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	// Fillup with sequencial bytes
	helper.FillUpLeafWithNumericValuesUntilItSplits(tree, 10000, 0)

	for i := 1; i <= 10000; i++ {
		integerBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(integerBytes, uint32(i))
		res := bTree.BTreeGetOne(tree, integerBytes)
		if res == nil {
			tmp := binary.BigEndian.Uint32(integerBytes)
			t.Errorf("Did not find a value for %d\n", tmp)
		}
	}
}

func TestInsertMultipleLinesWithOneLeafSequence(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	// Fillup with sequencial bytes
	helper.FillUpLeafWithNumericValuesUntilItSplits(tree, 250, 0)
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(251))
	value := helper.CreateValueOf16kLen()

	bTree.BTreeInsert(tree, key, value)
	res := bTree.BTreeGetOne(tree, key)
	if res == nil {
		t.Errorf("Did not find a value for %d\n", 4)
	}
}

func TestInsertMultipleLinesWithMultipleOneLeafSequence(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	// Fillup with sequencial bytes
	helper.FillUpLeafWithNumericValuesUntilItSplits(tree, 250, 0)
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(252))
	helper.FillUpLeavesWith16kvalues(tree, 200, 251)

	res := bTree.BTreeGetOne(tree, key)
	if res == nil {
		t.Errorf("Did not find a value for %s\n", key)
	}
}

func TestDeletionOfAnElementInMiddleOfALeaf(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	bTree.BTreeInsert(tree, []byte("1"), []byte("first"))
	bTree.BTreeInsert(tree, []byte("2"), []byte("second"))
	bTree.BTreeInsert(tree, []byte("3"), []byte("third"))

	// Delete the second element
	bTree.BTreeDelete(tree, []byte("2"))

	// Check if the element is still there
	res := bTree.BTreeGetOne(tree, []byte("2"))
	if res != nil {
		t.Errorf("Should not have found a value for %s\n", res.Key)
	}
}

func TestDeletionOfAnElementFromRootLeafAndTryInsertingAnotherOne(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	bTree.BTreeInsert(tree, []byte("1"), []byte("first"))
	bTree.BTreeDelete(tree, []byte("1"))
	t.Log("After Deleting the first element")
	res := bTree.BTreeGetOne(tree, []byte("1"))
	t.Log("After trying to get the first element")
	if res != nil {
		t.Errorf("Should not have found a value for %s\n", res.Key)
	}

	// Insert another element
	bTree.BTreeInsert(tree, []byte("4"), []byte("fourth"))
	res = bTree.BTreeGetOne(tree, []byte("4"))

	if res == nil {
		t.Errorf("Should have found a value for key []byte('4')")
	}

}

func TestDeletionOfFirstElementOfAnLeaf(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	helper.FillUpLeafUntilItSplits(tree)
	// Delete element 267
	bTree.BTreeDelete(tree, []byte("267"))
	firstPage := tree.Get(tree.GetRoot())
	topNodePreviousKey := firstPage.GetNodeChildByKey([]byte("267"))

	if topNodePreviousKey != nil {
		t.Errorf("Should not have found a value for %s\n", topNodePreviousKey.GetKey())
	}

	replacedKey := firstPage.GetNodeChildByIndex(1)

	if !bytes.Equal(replacedKey.GetKey(), []byte("268")) {
		t.Errorf("Should have found a value for %s\n", replacedKey.GetKey())

	}
}

func TestDeletionOfAnEntireLeaf(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	helper.FillUpLeafUntilItSplits(tree)

	pageFour := tree.Get(4)
	allKeys := make([][]byte, 0)
	for i := 0; i < int(pageFour.GetNItens()); i++ {
		tmp := make([]byte, len(pageFour.GetLeafKeyValueByIndex(uint16(i)).GetKey()))
		copy(tmp, pageFour.GetLeafKeyValueByIndex(uint16(i)).GetKey())
		allKeys = append(allKeys, tmp)
	}

	for i := 0; i < len(allKeys); i++ {
		bTree.BTreeDelete(tree, allKeys[i])
	}

	// Print root page
	rootPage := tree.Get(tree.GetRoot())
	nItens := rootPage.GetNItens()
	if nItens != 1 {
		t.Error("Should be only one item")
	}
}

func TestFileMapping(t *testing.T) {
	t.Log("creating basic database to test mapping")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	tree = helper.CreateFakeDbPagesForMapping(t, tree)

	mapped := bTree.MapAllLeavesToArray(tree)
	expectedPages := []uint64{4, 5, 6, 7, 8, 9}
	for i := 0; i < len(mapped); i++ {

		if expectedPages[i] != (mapped[i].TreeNode) {
			t.Errorf("mapped leaves equal %d Should be %d\n", mapped[i].TreeNode, i+4)
		}

		if len(mapped[i].History) != 2 {
			t.Errorf("History should have length of 2 found %d\n", len(mapped[i].History))
		}
	}
}

func TestBTreeGetMultipleItemsWithSameKey(t *testing.T) {
	t.Log("Creating basic database to test mapping")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	// Insert another element
	bTree.BTreeInsert(tree, []byte("4"), []byte("fourth"))
	bTree.BTreeInsert(tree, []byte("4"), []byte("fifth"))
	res := bTree.BTreeGet(tree, []byte("4"))
	if len(res) != 2 {
		t.Errorf("Should have found 2 items for key 4, found %d\n", len(res))
	}

	// Here enters one more key []byte("4")
	helper.FillUpLeafUntilItSplits(tree)
	// Insert again
	bTree.BTreeInsert(tree, []byte("4"), []byte("sixth"))
	// Get all items for key []byte("4")
	res = bTree.BTreeGet(tree, []byte("4"))
	if len(res) != 4 {
		t.Errorf("Should have found 4 items for key 4, found %d\n", len(res))
	}
}

func TestBTreeGetMultipleItemsWithSameKeyAndLargeValues(t *testing.T) {
	t.Log("Creating basic database to test mapping")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	// Insert another element
	bTree.BTreeInsert(tree, []byte("4"), []byte("fourth"))
	bTree.BTreeInsert(tree, []byte("4"), []byte("fifth"))
	res := bTree.BTreeGet(tree, []byte("4"))
	if len(res) != 2 {
		t.Errorf("Should have found 2 items for key 4, found %d\n", len(res))
	}

	// Here enters one more key []byte("4")
	helper.FillUpLeafUntilItSplits(tree)
	// Insert again
	bTree.BTreeInsert(tree, []byte("4"), helper.CreateValueOf16kLen())
	// Get all items for key []byte("4")
	res = bTree.BTreeGet(tree, []byte("4"))
	if len(res) != 4 {
		t.Errorf("Should have found 4 items for key 4, found %d\n", len(res))
	}
}

func TestSingleLeafUpdate(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	helper.FillUpLeafUntilItSplits(tree)

	bTree.BTreeUpdate(tree, []byte("5"), []byte("updated"))
	// Get item for key []byte("5")
	item := bTree.BTreeGetOne(tree, []byte("5"))
	// Compare if the value is equal to updated
	if !bytes.Equal(item.Value, []byte("updated")) {
		t.Errorf("Should have found a value for %s\n", item.Value)
	}
}

func TestCrawlingForBTree(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	helper.FillUpLeafUntilItSplits(tree)

	crawler := tree.FindLeafForCrawling([]byte("5"))

	// Should have found a leaf
	if crawler == nil {
		t.Error("Should have found a leaf")

	}

	// Get the first leaf
	leaf := crawler.Net[0]
	if leaf.GetNItens() != 2 {
		t.Errorf("Should have found 2 items, found %d\n", leaf.GetNItens())
	}

	// Get the next leaf
	leaf = crawler.Net[1]
	if leaf.GetNItens() != 95 {
		t.Errorf("Should have found 95 items, found %d\n", leaf.GetNItens())
	}

	kv := crawler.GetKeyValue()
	if !bytes.Equal(kv.Key, []byte("5")) {
		t.Errorf("Should have found key 5, found %s\n", kv.Key)
	}

	err := crawler.Next()
	if err != nil {
		t.Errorf("Should not have found an error, found %v\n", err)
	}

	kv = crawler.GetKeyValue()
	if !bytes.Equal(kv.Key, []byte("50")) {
		t.Errorf("Should have found key 50, found %s\n", kv.Key)
	}

	_ = crawler.Previous()
	err = crawler.Previous()
	if err != nil {
		t.Errorf("Should not have found an error, found %v\n", err)
	}

	kv = crawler.GetKeyValue()
	if !bytes.Equal(kv.Key, []byte("49")) {
		t.Errorf("Should have found key 49, found %s\n", kv.Key)
	}

	crawler = tree.FindLeafForCrawling([]byte("300"))
	if crawler == nil {
		t.Error("Should have found a leaf")
	}

}

func TestCrawlingForBTreeReachingEndAndBegin(t *testing.T) {
	// We insert multiple lines until it splits into two different leaves
	t.Log("Starting Test simple bTree Insertion")
	dbFilePath := helper.SetupTests(t)
	// Load bTree
	t.Log("Loading bTree to be used")
	tree := helper.LoadBTreeFromPath(t, dbFilePath)
	helper.FillUpLeafUntilItSplits(tree)

	crawler := bTree.GoToFirstLeaf(tree)
	var err error = nil
	for {
		_ = crawler.Next()
		if v := crawler.GetKeyValue(); bytes.Equal(v.Key, []byte("99")) {

			err = crawler.Next()
			break
		}
	}

	if err == nil {
		t.Errorf("Should have found an error, found %v\n", err)
	}

	crawler = bTree.GoToLastLeaf(tree)

	for {
		_ = crawler.Previous()
		if v := crawler.GetKeyValue(); bytes.Equal(v.Key, []byte("0")) {
			err = crawler.Previous()
			break
		}
	}

	if err == nil {
		t.Errorf("Should have found an error, found %v\n", err)
	}

}
