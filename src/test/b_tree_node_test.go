package main

/*
BTree node Test cases
*/

import (
	"bytes"
	"strconv"
	"testing"

	bTree "github.com/nicolasvancan/monvandb/src/btree"
)

/*
Teste if node creation returns correct data
*/
func TestNodeCreation(t *testing.T) {
	newNode := bTree.NewNodeNode()

	if newNode.GetType() != bTree.TREE_NODE {
		t.Errorf("Node type should be %d and it is %d\n", bTree.TREE_NODE, newNode.GetType())
	}

	if newNode.GetNItens() != 0 {
		t.Error("Should have 0 items")
	}

	if bTree.GetFreeBytes(newNode) != 4080 {
		t.Error("Should be 4080 free bytes")
	}

}

/*
Teste if Leaf creation returns correct data
*/
func TestLeafCreation(t *testing.T) {
	newLeaf := bTree.NewNodeLeaf()

	if newLeaf.GetType() != bTree.TREE_LEAF {
		t.Errorf("Node type should be %d and it is %d\n", bTree.TREE_LEAF, newLeaf.GetType())
	}

	if newLeaf.GetNItens() != 0 {
		t.Error("Should have 0 items")
	}

	if bTree.GetFreeBytes(newLeaf) != 4070 {
		t.Error("Should be 4070 free bytes")
	}
}

/*
   Section of inserting data into node
*/

func TestInternalNodeInsertion(t *testing.T) {
	newNode := bTree.NewNodeNode()

	key := []byte("1")

	// Insert
	newNode.PutNodeNewChild(key, 1)

	if newNode.GetNItens() != 1 {
		t.Errorf("Number of Items should be 1, found %d\n", newNode.GetNItens())
	}

	if bTree.GetFreeBytes(newNode) != 4069 {
		t.Errorf("Number of FreeBytes shoud be 4069, found %d\n", newNode.GetNItens())
	}

	if bytes.Compare(newNode.GetNodeChildByIndex(0).GetKey(), key) != 0 {
		t.Errorf("Wrong key, found %s\n", newNode.GetNodeChildByIndex(0).GetKey())
	}

	// Insert again but value lower than 1
	key = []byte("0")
	newNode.PutNodeNewChild(key, 2)

	if newNode.GetNItens() != 2 {
		t.Errorf("Number of Items should be 2, found %d\n", newNode.GetNItens())
	}

	if bTree.GetFreeBytes(newNode) != 4058 {
		t.Errorf("Number of FreeBytes shoud be 4058, found %d\n", newNode.GetNItens())
	}

	if bytes.Compare(newNode.GetNodeChildByIndex(0).GetKey(), key) != 0 {
		t.Errorf("Wrong key, found %s\n", newNode.GetNodeChildByIndex(0).GetKey())
	}

	// Fill up Node until it reaches max Limit
	for i := 2; i < 322; i++ {
		newNode.PutNodeNewChild([]byte(strconv.Itoa(i)), uint64(i))
	}

	t.Logf("Free bytes %d\n", bTree.GetFreeBytes(newNode))
	// Try to Put new Log, should return an error
	err := newNode.PutNodeNewChild([]byte("Teste"), 1234)
	if err == nil {
		t.Error("Should have returned an error\n")
	}
}
