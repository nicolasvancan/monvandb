package main

/*
BTree node Test cases
*/

import (
	"bytes"
	"testing"

	bTree "github.com/nicolasvancan/monvandb/src/btree"
	helper "github.com/nicolasvancan/monvandb/src/test/helper"
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

	if !bytes.Equal(newNode.GetNodeChildByIndex(0).GetKey(), key) {
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

	if !bytes.Equal(newNode.GetNodeChildByIndex(0).GetKey(), key) {
		t.Errorf("Wrong key, found %s\n", newNode.GetNodeChildByIndex(0).GetKey())
	}

	// Fill up Node until it reaches max Limit
	newNode = helper.FillUpNode("node")

	t.Logf("Free bytes %d\n", bTree.GetFreeBytes(newNode))
	// Try to Put new Log, should return an error
	err := newNode.PutNodeNewChild([]byte("Teste"), 1234)
	if err == nil {
		t.Error("Should have returned an error\n")
	}
}

func TestInternalLeafInsertion(t *testing.T) {
	newNode := bTree.NewNodeLeaf()

	key := []byte("1")
	value := []byte("first")

	// Insert
	newNode.PutLeafNewKeyValue(key, value)

	if newNode.GetNItens() != 1 {
		t.Errorf("Number of Items should be 1, found %d\n", newNode.GetNItens())
	}

	if bTree.GetFreeBytes(newNode) != 4054 {
		t.Errorf("Number of FreeBytes shoud be 4054, found %d\n", newNode.GetNItens())
	}

	if !bytes.Equal(newNode.GetLeafKeyValueByIndex(0).GetKey(), key) {
		t.Errorf("Wrong key, found %s\n", newNode.GetNodeChildByIndex(0).GetKey())
	}

	// Insert again but value lower than 1
	key = []byte("0")
	value = []byte("second")

	// Insert again
	newNode.PutLeafNewKeyValue(key, value)

	if newNode.GetNItens() != 2 {
		t.Errorf("Number of Items should be 2, found %d\n", newNode.GetNItens())
	}

	if bTree.GetFreeBytes(newNode) != 4037 {
		t.Errorf("Number of FreeBytes shoud be 4037, found %d\n", bTree.GetFreeBytes(newNode))
	}

	if !bytes.Equal(newNode.GetLeafKeyValueByIndex(0).GetKey(), key) {
		t.Errorf("Wrong key, found %s\n", newNode.GetLeafKeyValueByIndex(0).GetKey())
	}

	// Fill up Node until it reaches max Limit
	newNode = helper.FillUpNode("leaf")

	t.Logf("Free bytes %d\n", bTree.GetFreeBytes(newNode))
	// Try to Put new Log, should return an error
	err := newNode.PutNodeNewChild([]byte("Teste"), 1234)
	if err == nil {
		t.Error("Should have returned an error\n")
	}
}

func TestNodeSplit(t *testing.T) {
	// We fill up the node firstly
	node := helper.FillUpNode("node")
	// It has only 4 free bytes. We should try to split it into two new Nodes
	// We do that first with a sequence key, such as 400 greater than all the previous keys
	key1 := []byte("991")
	addr1 := 991

	splittedNodeCaseOne := node.SplitNode(key1, uint64(addr1))
	// We should see practically the same node as previously for the position 0
	if bTree.GetFreeBytes(&splittedNodeCaseOne[0]) != 4 {
		t.Logf("It should have only 4 bytes free, found %d\n", bTree.GetFreeBytes(&splittedNodeCaseOne[0]))
	}

	// Just to check, we ensure that the first Node doesn't have the key 400, inserted at the function
	// SplitNode
	if splittedNodeCaseOne[0].GetNodeChildByKey([]byte("400")) != nil {
		t.Log("It should not have the key 400 after splitted Node")
	}

	// Check for the second Node
	if !bytes.Equal(splittedNodeCaseOne[1].GetNodeChildByIndex(0).GetKey(), key1) {
		t.Logf("The first item in the array should be %s\n", splittedNodeCaseOne[1].GetNodeChildByIndex(0).GetKey())
	}

	// Case we insert some key that will be smaller than the last key (remembering that the keys within our range)
	// are merelly strings. So higher values such as 40000000 will not be higher than 9, for example. If we choose, lets say
	// a key 1000, we would have a smaller key than the last one which is 99

	key2 := []byte("1000")
	value := 245 // Can be any

	// I think we would remove the key 99 from the node if we insert and split the key 1000. And bytes will be 2 less than before
	// namelly 2

	newSplitedNode := node.SplitNode(key2, uint64(value))

	// for the left side, the free bytes should be 2
	if bTree.GetFreeBytes(&newSplitedNode[0]) != 2 {
		t.Logf("It should have only 2 bytes free, found %d\n", bTree.GetFreeBytes(&newSplitedNode[0]))
	}

	// The first byte of the second leaf will be 99
	if string(newSplitedNode[1].GetNodeChildByIndex(0).GetKey()) != "99" {
		t.Logf("The first key should be 99, found %s\n", newSplitedNode[1].GetNodeChildByIndex(0).GetKey())
	}

	// The final test is related to insertion of a large key, such as 10000000000, resulting in the second leaf
	// to have more than one item, more preciselly 2 items

	key3 := []byte("10000000")
	value3 := 7897
	lastSplit := node.SplitNode(key3, uint64(value3))

	if lastSplit[1].GetNItens() != 2 {
		t.Logf("The second node should have 2 items, found %d\n", lastSplit[1].GetNItens())
	}

	// Probably the first item will be 98, and second 99, lets see
	firstItemSecondNode := lastSplit[1].GetNodeChildByIndex(0).GetKey()
	secondItemSecondNode := lastSplit[1].GetNodeChildByIndex(1).GetKey()

	if string(firstItemSecondNode) != "98" {
		t.Logf("Should be 98")
	}

	if string(secondItemSecondNode) != "98" {
		t.Logf("Should be 99")
	}
}

// The leaf insertion will be skiped, because the implementation is exactly the same as for node
func TestNodeDeletionByAddress(t *testing.T) {
	// Create Node
	node := bTree.NewNodeNode()
	// Put values
	node.PutNodeNewChild([]byte("0"), 4)
	node.PutNodeNewChild([]byte("1"), 6)
	//Delete Value
	node.DeleteNodeChildrenByAddress(4)
	// Get values to evaluate
	nItens := node.GetNItens()

	if nItens > 1 {
		t.Logf("Should be one item")
	}

	firstItem := node.GetNodeChildByIndex(0)

	if firstItem.GetAddr() != 6 {
		t.Logf("Should be 6")
	}

	node.DeleteNodeChildrenByAddress(6)
	nItens = node.GetNItens()

	if nItens > 0 {
		t.Logf("Should be 0")
	}
}

func TestCreateLeafSequence(t *testing.T) {
	// Create leaf sequence
	initialBytes := helper.CreateValueOf16kLen()
	leaf, sequences := bTree.CreateLeafWithSequence([]byte("10"), initialBytes)
	keyVal := leaf.GetLeafKeyValueByIndex(0)

	finalBytes := keyVal.GetValue()
	for i := 0; i < len(sequences); i++ {
		te := sequences[i].GetLeafSequenceBytes()
		finalBytes = append(finalBytes, te...)
	}

	if !bytes.Equal(finalBytes, initialBytes) {
		t.Error("Bytes should be the same")
	}

	if len(finalBytes) != len(initialBytes) {
		t.Error("Bytes len should be the same")
	}

}

func TestLeafDeletionByKey(t *testing.T) {
	// Create Node
	node := helper.FillUpNode("leaf")
	firstItem := node.GetLeafKeyValueByIndex(0)
	if !bytes.Equal(firstItem.GetKey(), []byte("0")) {
		t.Logf("Should be 0")
	}

	//Delete Value
	node.DeleteLeafKeyValueByKey([]byte("0"))
	// Get values to evaluate

	firstItem = node.GetLeafKeyValueByIndex(0)
	if !bytes.Equal(firstItem.GetValue(), []byte("1teste")) {
		t.Error("Should be 6")
	}

	node.DeleteLeafKeyValueByKey([]byte("1"))
}
