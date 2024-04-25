package btree

import (
	"bytes"
	"encoding/binary"
)

const PAGE_SIZE = 4096

/*
   B-Tree Basic structure

   - | root | 8B BTree has a pointer to the root address
   - | pageMaxSize | 2B Has maximum bytes per page
   - | minItens | 2B Has minimum itens per Node
   - | maxItens | 2B Has maximum itens per node
*/

// Btree sizes in bytes declaration
const (
	BTREE_ROOT_SIZE     = 16
	BTREE_NAME_SIZE     = 400
	BTREE_MIN_NODE_SIZE = 4
	BTREE_MAX_NODE_SIZE = 4
)

// Btree Page offset
const (
	BTREE_OFFSET_ROOT          = 0
	BTREE_OFFSET_NAME          = BTREE_OFFSET_ROOT + BTREE_ROOT_SIZE
	BTREE_OFFSET_MIN_NODE_SIZE = BTREE_OFFSET_NAME + BTREE_NAME_SIZE
	BTREE_OFFSET_MAX_NODE_SIZE = BTREE_OFFSET_MIN_NODE_SIZE + BTREE_MIN_NODE_SIZE
)

/*
Base Btree Implementation
*/
type BTree struct {
	data      []byte                // Page header
	pageSize  uint32                // Page Size. It's still hardcoded
	root      uint64                // Indicates Where the root page starts
	SetHeader func(BTree)           // Update header whenever needed
	Get       func(uint64) TreeNode // Returns a Tree Node
	New       func(TreeNode) uint64 // Allocate a new Page
	Del       func(uint64)
	Set       func(TreeNode, uint64) bool // Del a new page
}

type TreeNodePage struct {
	node TreeNode
	page uint64
}

type BTreeKeyValue struct {
	Key   []byte
	Value []byte
}

func NewTree(pageSize int) *BTree {
	// Returns a pointer to the new BTree in Memory
	nBTree := &BTree{
		data:     make([]byte, pageSize),
		pageSize: uint32(pageSize),
	}

	// For new Trees root will be zero, meaning that there is no page
	nBTree.root = 0
	nBTree.SetRoot(0)

	return nBTree
}

func LoadTree(bTree []byte, pageSize uint32) *BTree {
	tree := &BTree{
		data: bTree[:pageSize-1],
	}

	tree.pageSize = pageSize
	tree.root = tree.GetRoot()

	return tree
}

func (b *BTree) GetBytes() []byte {
	return []byte(b.data)
}

func (b *BTree) SetRoot(root uint64) {
	// Insert value into data structure
	binary.LittleEndian.PutUint64(b.data[BTREE_OFFSET_ROOT:BTREE_ROOT_SIZE], root)
}

func (b *BTree) SetName(name string) {
	bString := []byte(name)
	if len(name) >= BTREE_NAME_SIZE {
		bString = bString[:BTREE_NAME_SIZE]
	}

	copy(b.data[BTREE_OFFSET_NAME:BTREE_NAME_SIZE], bString)
}

func (b *BTree) GetName() string {
	return string(b.data[BTREE_OFFSET_NAME:BTREE_NAME_SIZE])
}

func (b *BTree) GetRoot() uint64 {
	return uint64(binary.LittleEndian.Uint64(b.data[BTREE_OFFSET_ROOT:BTREE_ROOT_SIZE]))
}

/*
BTreeInsert
Main function to insert a key value to a bTree

This function will do almost all the hard work regarding bTreeInsertion.
It will look for the right place to insert the new key value to a Node, If necessary
*/
func BTreeInsert(bTree *BTree, key []byte, value []byte) {
	// Get root
	root := bTree.GetRoot()
	// Root = 0 means that there is no item in it
	if root == 0 {
		// Than we start a new bTree
		startsNewBTree(bTree, key, value)
		return
	}

	// Get node from page
	rootNode := bTree.Get(root)

	// Find leaf to insert value
	leafToInsert, history := findLeaf(bTree, rootNode, key, root, make([]TreeNodePage, 0))

	if leafToInsert == nil {
		/* Didn't find any leaf, that means the key to be inserted is smaller than the smallest key in bytes

		So we find first node and insert the key to it
		*/
		leafToInsert, history = findFirstLeaf(bTree)
	}

	// Verify whether leaf must be splitted
	insertAndReorderTree(bTree, *leafToInsert, history, key, value)
}

/*
BTreeDelete - Base function to delete a key value from bTree
*/
func BTreeDelete(bTree *BTree, key []byte) {
	// Load root page
	rootAddr := bTree.GetRoot()
	rootPage := bTree.Get(rootAddr)
	// Lookup tree to find
	leaf, history := findLeaf(bTree, rootPage, key, rootAddr, make([]TreeNodePage, 0))
	if leaf == nil {
		return
	}

	DeleteKeyValueInLeafAndUpdateNodesRecursivelly(bTree, key, *leaf, history)
}

// TODO: Refactor this function to let it be more optimal in terms of pages operations
func BTreeUpdate(bTree *BTree, key []byte, value []byte) {
	// Load root page
	rootAddr := bTree.GetRoot()
	rootPage := bTree.Get(rootAddr)
	// Lookup tree to find
	leaf, history := findLeaf(bTree, rootPage, key, rootAddr, make([]TreeNodePage, 0))

	if leaf == nil {
		return
	}

	// Delete key value
	DeleteKeyValueInLeafAndUpdateNodesRecursivelly(bTree, key, *leaf, history)

	// Insert new key value
	BTreeInsert(bTree, key, value)
}

/*
BTreeGetOne - Base function to get first Key Values stored in bTree
It will return nil if no key value is found. Otherwise, returns a BTreeKeyValue pointer
which will contain information about the first found key
*/

func BTreeGetOne(bTree *BTree, key []byte) *BTreeKeyValue {
	// Create empty array
	var keyValue *BTreeKeyValue = nil
	// Load root page
	rootAddr := bTree.GetRoot()
	rootPage := bTree.Get(rootAddr)
	// Lookup tree to find
	leaf, _ := findLeaf(bTree, rootPage, key, rootAddr, make([]TreeNodePage, 0))
	if leaf == nil {
		return nil
	}

	allLeafValues := getAllLeafKeyValues(&leaf.node)

	for i := 0; i < len(allLeafValues); i++ {
		if bytes.Equal(allLeafValues[i].key, key) {
			// Found key
			keyValue = new(BTreeKeyValue)
			keyValue.Key = allLeafValues[i].GetKey()
			keyValue.Value = allLeafValues[i].GetValue()
			// In case it has no sequence
			if leaf.node.GetLeafHasSeq() == uint16(1) {
				keyValue.Value = getAllBytesFromSequences(bTree, leaf.node)
			}
		}
	}

	return keyValue
}

/*
BTreeGet - Base function to get all Key Values stored in bTree, returning
an array of BTreeKeyValue
*/

func BTreeGet(bTree *BTree, key []byte) []BTreeKeyValue {
	// Create empty array
	var keyValues []BTreeKeyValue = make([]BTreeKeyValue, 0)
	// Load root page
	rootAddr := bTree.GetRoot()
	rootPage := bTree.Get(rootAddr)
	// Lookup tree to find many keys values if they exist
	leavesFound, history := findLeaves(bTree, rootPage, key, rootAddr, make([]TreeNodePage, 0))
	/* if leavesFound is empty, and history is also empty, means that the root page is a Leaf.
	Therefore, we induce that there might be a key in root page
	*/
	if len(leavesFound) == 0 && len(history) == 0 {
		// Get all key values from root page
		leavesFound = append(leavesFound, TreeNodePage{node: rootPage, page: rootAddr})
	}

	if leavesFound != nil {
		for i := 0; i < len(leavesFound); i++ {
			allLeafValues := getAllLeafKeyValues(&leavesFound[i].node)

			for j := 0; j < len(allLeafValues); j++ {
				if bytes.Equal(allLeafValues[j].key, key) {
					// Found key
					keyValue := new(BTreeKeyValue)
					keyValue.Key = allLeafValues[j].GetKey()
					keyValue.Value = allLeafValues[j].GetValue()
					// In case it has no sequence
					if leavesFound[i].node.GetLeafHasSeq() == uint16(1) {
						keyValue.Value = getAllBytesFromSequences(bTree, leavesFound[j].node)
					}
					keyValues = append(keyValues, *keyValue)
				}
			}
		}

	}
	return keyValues
}
