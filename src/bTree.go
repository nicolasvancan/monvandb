package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
Implementation of function to lookup key in internal Node, returning page number
*/
func lookupKey(node TreeNode, key []byte) int {
	// Declare found variable initiated with nil (Case we don't find any)
	var found int = -1
	var allNodeKeyAddr []NodeKeyAddr = nil
	var allLeafKeyValues []LeafKeyValue = nil

	// Just in case it is a Internal Node
	if node.GetType() == TREE_NODE {
		allNodeKeyAddr = ([]NodeKeyAddr)(getAllNodeKeyAddr(&node))
		// Iterate over all items to find a corresponding key
		for i := 0; i < int(node.GetNItens()); i++ {

			if bytes.Compare(allNodeKeyAddr[i].key, key) <= 0 {
				found = i
			} else {
				break
			}
		}

	} else {
		allLeafKeyValues = ([]LeafKeyValue)(getAllLeafKeyValues(&node))
		// Iterate over all items to find a corresponding key
		for i := 0; i < int(node.GetNItens()); i++ {

			if bytes.Compare(allLeafKeyValues[i].key, key) <= 0 {
				found = i
			} else {
				break
			}
		}
	}

	return found
}

/*
This function works recursiverly and returns exactly what Node the data must be inserted
*/
func findLeafToInsert(bTree *BTree, node TreeNode, key []byte, page uint64, history []TreeNodePage) (TreeNodePage, []TreeNodePage) {
	// Means it has reached some leaf
	if node.GetType() == TREE_NODE {
		// Tries to find Leaf through Internal Node
		if idx := lookupKey(node, key); idx > -1 {
			// Get keyAddress from it
			nodeKeyAddr := node.GetNodeChildByIndex(idx)
			// Reach out the address from bTree file
			fmt.Printf("DEBUG::findLeafToInsert > Getting node at address = %d\n", nodeKeyAddr.addr)
			tmpNode := bTree.Get(nodeKeyAddr.addr)
			fmt.Printf("DEBUG::findLeafToInsert > Node type = %d\n", tmpNode.GetType())
			return findLeafToInsert(bTree, tmpNode, key, nodeKeyAddr.addr, append(history, TreeNodePage{
				node: node,
				page: page,
			}))
		}
	}

	return TreeNodePage{node: node, page: page}, history
}

func insertAndSplitIfNeeded(bTree *BTree, tPage TreeNodePage, history []TreeNodePage, key []byte, value []byte) {
	keyLen := len(key)
	valueLen := len(value)
	if mustSplitNode(bTree, tPage.node, keyLen, valueLen) {
		if (keyLen + valueLen + 10) > PAGE_SIZE-26 {
			// Special case
			fmt.Println("DEBUG::insertAndSplitIfNeeded > Special case where there must be a linked list to suport one unique item")
		} else {
			// Generic case
			fmt.Println("DEBUG::insertAndSplitIfNeeded > Generic insertion")
			splitBackyardsRecursively(bTree, tPage, history, key, value)
		}

	} else {
		fmt.Println("DEBUG::insertAndSplitIfNeeded > Didn't need to split leaf")
		tPage.node.PutLeafNewKeyValue(key, value)
		bTree.Set(tPage.node, tPage.page)
	}
}

func insertNodesRecursivelly(
	bTree *BTree,
	oldPage TreeNodePage,
	newPages []TreeNodePage,
	history []TreeNodePage,
) {
	// Get nodeToInsert, it will never happen when there is an empty history, so we can do it
	nodeToInsert := history[len(history)-1]
	fmt.Printf("DEBUG::insertNodesRecursivelly > History Lenght %d\n", len(history))
	fmt.Printf("DEBUG::insertNodesRecursivelly > Node to insert type %d\n", nodeToInsert.node.GetType())
	// New pages always come divided into 2 pieces
	newPageOne := newPages[0]
	newPageTwo := newPages[1]

	var totalKeyLen int

	// Remove oldPage reference from node that will receive new pages
	fmt.Println("DEBUG::insertNodesRecursivelly > Delete Node Children")
	nodeToInsert.node.DeleteNodeChildrenByAddress(oldPage.page)

	if newPageOne.node.GetType() == TREE_LEAF {
		totalKeyLen = (len(newPageOne.node.GetLeafKeyValueByIndex(0).key) +
			len(newPageTwo.node.GetLeafKeyValueByIndex(0).key) + 10)
	} else {
		totalKeyLen = (len(newPageOne.node.GetNodeChildByIndex(0).key) +
			len(newPageTwo.node.GetNodeChildByIndex(0).key) + 10)
	}
	// verify whether or not the node must be splitted
	if mustSplitNode(bTree, nodeToInsert.node, totalKeyLen, 8) {
		// We verify for both keys if the node must be splitted, if yes
		// we split it for the first page

		splitedNode := nodeToInsert.node.SplitNode(newPageOne.node.GetLeafKeyValueByIndex(0).key, newPageOne.page)
		ourInsertion := 1
		// Where is our insertion?
		if splitedNode[0].GetLeafKeyValueByKey(newPageOne.node.GetLeafKeyValueByIndex(0).key) != nil {
			ourInsertion = 0
		}

		// Our insertion is sorted, therefore we must isert the second new node into the second splited node
		splitedNode[1].PutNodeNewChild(newPageTwo.node.GetLeafKeyValueByIndex(0).key, newPageTwo.page)

		// Create our new pages
		addr0 := bTree.New(splitedNode[0])
		addr1 := bTree.New(splitedNode[1])

		// Set the parent addresses
		if ourInsertion == 0 {
			setParentAddr(&newPageOne.node, addr0)
		} else {
			setParentAddr(&newPageOne.node, addr1)
		}

		setParentAddr(&newPageTwo.node, addr1)

		// Update those pages
		bTree.Set(newPageOne.node, newPageOne.page)
		bTree.Set(newPageTwo.node, newPageTwo.page)

		// Has parent
		if len(history) > 1 {
			// Call the stack recursivelly

			insertNodesRecursivelly(
				bTree,
				nodeToInsert,
				[]TreeNodePage{
					{node: splitedNode[0], page: addr0},
					{node: splitedNode[1], page: addr1},
				},
				history[:len(history)-1])

		} else { // No parent :(
			// Create new root Node
			newRoot := NewNodeNode()
			setParentAddr(newRoot, 0)
			newRootAddress := bTree.New(*newRoot)
			setParentAddr(&splitedNode[0], newRootAddress)
			setParentAddr(&splitedNode[1], newRootAddress)
			newRoot.PutNodeNewChild(splitedNode[0].GetNodeChildByIndex(0).key, addr0)
			newRoot.PutNodeNewChild(splitedNode[1].GetNodeChildByIndex(0).key, addr1)
			bTree.Set(*newRoot, newRootAddress)
			bTree.Set(splitedNode[0], addr0)
			bTree.Set(splitedNode[1], addr1)
			// Set bTree to be redirected to new Root
			bTree.SetRoot(newRootAddress)
			// Commit changes
			bTree.SetHeader(*bTree)
		}

	} else {
		// Insert new nodes
		nodeToInsert.node.PutNodeNewChild(newPageOne.node.GetLeafKeyValueByIndex(0).key, newPageOne.page)
		nodeToInsert.node.PutNodeNewChild(newPageTwo.node.GetLeafKeyValueByIndex(0).key, newPageTwo.page)
		// Update parent from nodes
		setParentAddr(&newPageOne.node, nodeToInsert.page)
		setParentAddr(&newPageTwo.node, nodeToInsert.page)
		// Update pages
		bTree.Set(nodeToInsert.node, nodeToInsert.page)
		bTree.Set(newPageOne.node, newPageOne.page)
		bTree.Set(newPageTwo.node, newPageTwo.page)
	}
}

/*
This function splits all needed leaves and internal nodes backyards
returning the leaf that must be updated with the new value
*/
func splitBackyardsRecursively(
	bTree *BTree,
	tPage TreeNodePage,
	history []TreeNodePage,
	key []byte,
	value []byte,
) {
	// If it came to here, at least a leaf must be split
	splittedLeaf := tPage.node.SplitLeaf(key, value)
	fmt.Printf("DEBUG::splitBackyardsRecursively > splittedLeaf\n")
	// Means that this leaf is directly linked to the root
	if len(history) == 0 {
		fmt.Printf("DEBUG::splitBackyardsRecursively > No Root\n")
		// Create a new Internal Node
		newInternalNode := NewNodeNode()

		// Set parent as root
		setParentAddr(newInternalNode, 0)
		// Create the new Node in the file
		newNodeAddress := bTree.New(*newInternalNode)
		fmt.Printf("DEBUG::splitBackyardsRecursively > Created Root at page %d\n", newNodeAddress)
		// Set the splited leaves
		for i := 0; i < len(splittedLeaf); i++ {
			newLeaf := splittedLeaf[i]
			setParentAddr(&newLeaf, newNodeAddress)
			// First item sorted
			keyVal := newLeaf.GetLeafKeyValueByIndex(0)
			// Create new leaf
			newLeafAddr := bTree.New(newLeaf)
			// Insert them into parent node
			newInternalNode.PutNodeNewChild(keyVal.key, newLeafAddr)
		}
		bTree.Set(*newInternalNode, newNodeAddress)
		bTree.SetRoot(newNodeAddress)
		bTree.SetHeader(*bTree)
		return
	}

	// Every time a node splits, two new nodes emerge. The key point is that we must replace
	// nodes page from previous page. For instance, we have a node at page 1 following by a node at page 2.
	// For some reason, the node 2 must be split into 2 new nodes, at page 10 and 11. Knowing that, we ensure that basically
	// the page 2 now has turned into two pages 10 and 11, therefore, the page two, that is placed within the values from
	// node in page one must also be replaced, because we don't use the page two anymore. So, the page one would

	// Create both new leaves
	newLeavesPages := make([]TreeNodePage, len(splittedLeaf))

	fmt.Println("DEBUG::newLeavesPages > Created new Leaves pages")
	for i := 0; i < len(newLeavesPages); i++ {
		newLeavesPages[i] = TreeNodePage{
			node: splittedLeaf[i],
			page: bTree.New(splittedLeaf[i]),
		}
	}
	fmt.Println("DEBUG::newLeavesPages > After leaves creation")

	// Insert into node recursivelly
	insertNodesRecursivelly(bTree, tPage, newLeavesPages, history)
}

/*
Main function to insert a key value to a bTree

This function will do almost all the hard work regarding bTreeInsertion.
It will look for the right place to insert the new key value to a Node, If necessary
*/
func BTreeInsert(bTree *BTree, key []byte, value []byte) {
	// Get root
	root := bTree.GetRoot()
	// Root = 0 means that there is no item in it
	if root == 0 {
		fmt.Println("DEBUG::BTreeInsert > Root = 0")
		newLeaf := NewNodeLeaf()
		fmt.Println("DEBUG::BTreeInsert > Created new Leaf")
		// Means that it is top page
		setParentAddr(newLeaf, 0)
		fmt.Println("DEBUG::BTreeInsert > Set Parent to zero")
		newLeaf.PutLeafNewKeyValue(key, value)
		fmt.Println("DEBUG::BTreeInsert > Inserted new Key Value")
		// Calls callback
		pageNumber := bTree.New(*newLeaf)
		fmt.Printf("DEBUG::BTreeInsert > Created new page with number %d\n", pageNumber)
		bTree.root = pageNumber
		bTree.SetRoot(pageNumber)
		bTree.SetHeader(*bTree)
		return
	}

	// Look for where to store this key value
	// Get node from page
	node := bTree.Get(root)
	fmt.Printf("DEBUG::BTreeInsert > N Itens %d\n", node.GetNItens())
	fmt.Printf("Free::BTreeInsert Bytes %d\n", getFreeBytes(&node))
	// Find leaf to insert value
	leafToInsert, history := findLeafToInsert(bTree, node, key, root, make([]TreeNodePage, 0))
	// Verify whether leaf must be splited
	insertAndSplitIfNeeded(bTree, leafToInsert, history, key, value)
}

/* Verify whether leaf should be splited*/
func mustSplitNode(bTree *BTree, node TreeNode, keyLen int, valueLen int) bool {
	freeBytes := getFreeBytes(&node)
	totalNewBytes := keyLen + valueLen
	if node.GetType() == TREE_NODE {
		totalNewBytes = keyLen
	}

	return freeBytes < uint16(totalNewBytes)+10
}
