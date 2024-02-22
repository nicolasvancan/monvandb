package btree

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

func shiftValuesBetweenLeaves(bTree *BTree, tPage TreeNodePage, history []TreeNodePage, key []byte, value []byte) {
	parentAddr := tPage.node.GetParentAddr()
	// check if parent is not zero, indicating that there is a node above it
	if len(history) == 0 {
		fmt.Println("DEBUG::Shifting key parentAddr = 0")
		splitBackyardsRecursively(
			bTree,
			tPage,
			history,
			key,
			value,
		)
		return
	}
	// If it came to here, at least a leaf must be split
	splittedLeaf := tPage.node.SplitLeaf(key, value)
	setParentAddr(&splittedLeaf[0], parentAddr)
	// Get first key from splittedLeaf
	firstKey := splittedLeaf[0].GetLeafKeyValueByIndex(0)
	// Parent node directly related to tPage
	parentNode := history[len(history)-1]
	// Delete it by address
	parentNode.node.DeleteNodeChildrenByAddress(tPage.page)
	// Insert again the new address with new key
	parentNode.node.PutNodeNewChild(firstKey.key, tPage.page)
	// We don't update it yet, we check for the right part of the split
	rightLeaf := splittedLeaf[1]
	// We can update the value already in page using left side splitted leaf and parent
	bTree.Set(splittedLeaf[0], tPage.page)
	bTree.Set(parentNode.node, parentNode.page)

	// For every new value we check if it is possible to insert into the next one
	keyValuesToBeReinserted := getAllLeafKeyValues(&rightLeaf)

	parentKeyAddr := getAllNodeKeyAddr(&parentNode.node)

	mappedLeaves := MapAllLeavesToArray(bTree)

	// Find where is the start leaf from tPage
	startIndex := 0
	for i := 0; i < len(mappedLeaves); i++ {
		if mappedLeaves[i].TreeNode == tPage.page {
			startIndex = i + 1 // Should never be the last
			break
		}
	}

	mappedLeaves = mappedLeaves[startIndex:]

	for i := 0; i < len(keyValuesToBeReinserted); i++ { // Iterates over keyValuesLeftBehind to be added
		/* For every value from the right leaf, we must add it to another one, so we find
		the index of the left leaf at parentKeyAddr

		TODO: Explain better*/

		fmt.Printf("DEBUG::Right leaf value = %s\n", keyValuesToBeReinserted[i].GetKey())
		indexOfLeaf := 0

		isLastIndex := false
		if len(mappedLeaves) <= 1 {
			isLastIndex = true
		}

		fmt.Printf("DEBUG::Next leaf index = %d\n", indexOfLeaf)
		// Get next leaf
		nextLeafAddress := mappedLeaves[indexOfLeaf].TreeNode
		nextLeaf := bTree.Get(nextLeafAddress)
		fmt.Printf("DEBUG::Next leaf Addrress = %d\n", nextLeafAddress)
		if !isLastIndex {
			fmt.Println("DEBUG::Not last index")
			// Call this function recursivelly for the next ordered leaf
			shiftValuesBetweenLeaves(
				bTree,
				TreeNodePage{
					node: nextLeaf,
					page: nextLeafAddress},
				history, keyValuesToBeReinserted[i].GetKey(),
				keyValuesToBeReinserted[i].GetValue())
			continue
		}
		// If it's the last index call the other one
		fmt.Println("DEBUG::LAST index")
		if !mustSplitNode(bTree, nextLeaf, len(keyValuesToBeReinserted[i].GetKey()), len(keyValuesToBeReinserted[i].GetValue())) {
			// Get first key from next leaf
			nextLeafFirstKey := nextLeaf.GetLeafKeyValueByIndex(0).GetKey()
			nextLeaf.PutLeafNewKeyValue(keyValuesToBeReinserted[i].GetKey(), keyValuesToBeReinserted[i].GetValue())

			// Update leaf
			bTree.Set(nextLeaf, nextLeafAddress)

			// Compare if key has changed and update parent
			if bytes.Compare(nextLeafFirstKey, nextLeaf.GetLeafKeyValueByIndex(0).GetKey()) != 0 {
				// Delete address
				parentNode.node.DeleteNodeChildrenByAddress(nextLeafAddress)
				// Put next Leaf with key and address
				parentNode.node.PutNodeNewChild(nextLeaf.GetLeafKeyValueByIndex(0).GetKey(), nextLeafAddress)
				// Update Parent
				bTree.Set(parentNode.node, parentNode.page)
			}
			continue
		}

		splitBackyardsRecursively(bTree,
			TreeNodePage{
				node: nextLeaf,
				page: parentKeyAddr[indexOfLeaf].GetAddr()},
			history, keyValuesToBeReinserted[i].GetKey(),
			keyValuesToBeReinserted[i].GetValue())
	}
}

/*
Insert value changin the bTree structure whenever it is needed.

# The logic flow is relativelly simple

If the found leaf that will receive key value is not full and must not be splitted, the key value is
inserted imediatelly.

If not, there can happen two cases, the value will be inserted at the last leaf, or it will be inserted
in the middle of a previous leaf (Probably a full one) and a lot of values shiffting must occour.

So, lets map all availables leaves and compare if the last page found is equal to the given page
*/
func insertAndReorderTree(bTree *BTree, tPage TreeNodePage, history []TreeNodePage, key []byte, value []byte) {
	keyLen := len(key)
	valueLen := len(value)

	// There is overflow in page, must split it or shfit values (when inserting unsorted data)
	if mustSplitNode(bTree, tPage.node, keyLen, valueLen) {
		if (keyLen + valueLen + 10) > PAGE_SIZE-26 { // 26 is header size for leaf
			// Special case
			fmt.Println("DEBUG::insertAndReorderTree > Special case where there must be a linked list to suport one unique item")
			leaf := createLeafAndSequencesForLargeBytes(bTree, key, value)
			insertOneKeyLeafAndReorderTree(bTree, tPage, *leaf, history)
		} else {
			// Verify whether the leaf if the last one by comparing he page number
			lastLeaf, _ := findLastLeaf(bTree)

			if lastLeaf.page == tPage.page {
				fmt.Println("DEBUG::insertAndReorderTree > LAST LEAF")
				// Last leaf split's backyards Node Recursivelly
				splitBackyardsRecursively(bTree, tPage, history, key, value)
				return
			}
			fmt.Println("DEBUG::insertAndReorderTree > Shift Values")
			shiftValuesBetweenLeaves(bTree, tPage, history, key, value)
		}
		return
	}

	fmt.Printf("DEBUG::Inserted at Leaf of page %d\n", tPage.page)
	tPage.node.PutLeafNewKeyValue(key, value)
	bTree.Set(tPage.node, tPage.page)
}

func insertNodesRecursivelly(
	bTree *BTree,
	oldPage TreeNodePage,
	newPages []TreeNodePage,
	history []TreeNodePage,
) {
	// Get nodeToInsert, it will never happen when there is an empty history, so we can do it
	nodeToInsert := history[len(history)-1]
	for o := 0; o < len(history); o++ {
		fmt.Printf("DEBUG::insertNodesRecursivelly::History Page %d index %d\n", history[o].page, o)
	}
	// New pages always come divided into 2 pieces
	newPageOne := newPages[0]
	newPageTwo := newPages[1]

	var totalKeyLen int
	fmt.Printf("DEBUG::insertNodesRecursivelly::Deleting Page %d reference from node\n", oldPage.page)
	// Remove oldPage reference from node that will receive new pages
	nodeToInsert.node.DeleteNodeChildrenByAddress(oldPage.page)
	fmt.Printf("DEBUG::insertNodesRecursivelly::Deleted page %d\n", oldPage.page)

	// Get total key length that will be added to nodeToInsert
	if newPageOne.node.GetType() == TREE_LEAF {
		fmt.Println("DEBUG::insertNodesRecursivelly::TREE_LEAF")
		totalKeyLen = (len(newPageOne.node.GetLeafKeyValueByIndex(0).key) +
			len(newPageTwo.node.GetLeafKeyValueByIndex(0).key) + 10) // Added 10 to fit two new records into mustSplitNode
	} else {
		fmt.Println("DEBUG::insertNodesRecursivelly::TREE_NODE")
		totalKeyLen = (len(newPageOne.node.GetNodeChildByIndex(0).key) +
			len(newPageTwo.node.GetNodeChildByIndex(0).key) + 10)
	}

	// verify whether or not the node must be splitted
	if mustSplitNode(bTree, nodeToInsert.node, totalKeyLen, 8) {
		fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode")
		// We verify if value can be shifted to another existing leaf or we really need to create another leaf
		splitedNode := make([]TreeNode, 2)
		ourInsertion := 1
		if newPageOne.node.GetType() == TREE_LEAF {
			// Leaf case (This is duplicated code unfortunatelly) must remake this
			fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode::TREE_LEAF")
			splitedNode = nodeToInsert.node.SplitNode(newPageOne.node.GetLeafKeyValueByIndex(0).key, newPageOne.page)
			fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode::TREE_LEAF:1")
			// Where is our insertion?
			if splitedNode[0].GetNodeChildByKey(newPageOne.node.GetLeafKeyValueByIndex(0).key) != nil {
				fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode::TREE_LEAF:2")
				ourInsertion = 0
			}

			fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode::TREE_LEAF:3")
			// Our insertion is sorted, therefore we must insert the second new node into the second splited node
			splitedNode[1].PutNodeNewChild(newPageTwo.node.GetLeafKeyValueByIndex(0).key, newPageTwo.page)

		} else {
			fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode::TREE_NODE")
			splitedNode = nodeToInsert.node.SplitNode(newPageOne.node.GetNodeChildByIndex(0).key, newPageOne.page)
			fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode::TREE_NODE:1")
			// Where is our insertion?
			if splitedNode[0].GetNodeChildByKey(newPageOne.node.GetNodeChildByIndex(0).key) != nil {
				fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode::TREE_NODE:2")
				ourInsertion = 0
			}

			fmt.Println("DEBUG::insertNodesRecursivelly::mustSplitNode::TREE_NODE:2")
			// Our insertion is sorted, therefore we must insert the second new node into the second splited node
			splitedNode[1].PutNodeNewChild(newPageTwo.node.GetNodeChildByIndex(0).key, newPageTwo.page)
		}

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

		if newPageOne.node.GetType() == TREE_LEAF {
			fmt.Println("DEBUG::Did not need to split Node node")
			fmt.Printf("DEBUG::New node with key %d and page %d\n", binary.LittleEndian.Uint32(newPageOne.node.GetLeafKeyValueByIndex(0).key), newPageOne.page)
			fmt.Printf("DEBUG::New node with key %d and page %d\n", binary.LittleEndian.Uint32(newPageTwo.node.GetLeafKeyValueByIndex(0).key), newPageTwo.page)
			nodeToInsert.node.PutNodeNewChild(newPageOne.node.GetLeafKeyValueByIndex(0).key, newPageOne.page)
			nodeToInsert.node.PutNodeNewChild(newPageTwo.node.GetLeafKeyValueByIndex(0).key, newPageTwo.page)
		} else {

			fmt.Println("DEBUG::Did not need to split Node node")
			fmt.Printf("DEBUG::New node with key %d and page %d\n", binary.LittleEndian.Uint32(newPageOne.node.GetNodeChildByIndex(0).key), newPageOne.page)
			fmt.Printf("DEBUG::New node with key %d and page %d\n", binary.LittleEndian.Uint32(newPageTwo.node.GetNodeChildByIndex(0).key), newPageTwo.page)
			nodeToInsert.node.PutNodeNewChild(newPageOne.node.GetNodeChildByIndex(0).key, newPageOne.page)
			nodeToInsert.node.PutNodeNewChild(newPageTwo.node.GetNodeChildByIndex(0).key, newPageTwo.page)
		}

		// Insert new nodes

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
	fmt.Println("DEBUG::splitBackyardsRecursively")
	// If it came to here, at least a leaf must be split
	splittedLeaf := tPage.node.SplitLeaf(key, value)
	// Means that this leaf is directly linked to the root
	if len(history) == 0 {
		fmt.Println("DEBUG::splitBackyardsRecursively history length = 0")
		// Create a new Internal Node
		newInternalNode := NewNodeNode()
		// Set parent as root
		setParentAddr(newInternalNode, 0)
		// Create the new Node in the file
		newNodeAddress := bTree.New(*newInternalNode)
		// Set the splited leaves
		for i := 0; i < len(splittedLeaf); i++ {
			newLeaf := splittedLeaf[i]
			setParentAddr(&newLeaf, newNodeAddress)
			// First item sorted
			keyVal := newLeaf.GetLeafKeyValueByIndex(0)
			// Create new leaf
			newLeafAddr := bTree.New(newLeaf)
			// Insert them into parent node
			newInternalNode.PutNodeNewChild(keyVal.GetKey(), newLeafAddr)
		}
		bTree.Set(*newInternalNode, newNodeAddress)
		bTree.SetRoot(newNodeAddress)
		bTree.SetHeader(*bTree)
		return
	}

	/* Every time a node splits, two new nodes emerge. The key point is that we must replace
	nodes page from previous page. For instance, we have a node at page 1 followed by a node at page 2.
	For some reason, the node 2 must be split into 2 new nodes, at page 10 and 11. Knowing that, we ensure that basically
	the page 2 now has turned into two pages 10 and 11, therefore, the page two, that is placed within the values range from
	node in page one must also be replaced, because we don't use the page two anymore. So, the top node would lose the reference
	to page 2, and gain two new for page 10 and 11
	*/

	// Create both new leaves
	newLeavesPages := make([]TreeNodePage, len(splittedLeaf))
	for i := 0; i < len(newLeavesPages); i++ {
		newLeavesPages[i] = TreeNodePage{
			node: splittedLeaf[i],
			page: bTree.New(splittedLeaf[i]),
		}
	}

	fmt.Println("DEBUG:: About to insertNodeRecursivelly")
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
		// Than we start a new bTree
		startsNewBTree(bTree, key, value)
		return
	}

	// Get node from page
	rootNode := bTree.Get(root)
	fmt.Printf("DEBUG::BTreeInsert::ROOT > N Itens %d and page %d\n", rootNode.GetNItens(), root)
	fmt.Printf("DEBUG::BTreeInsert::ROOT Bytes %d\n", GetFreeBytes(&rootNode))
	fmt.Printf("DEBUG::BTreeInsert::ROOT Type %d\n\n\n", rootNode.GetType())

	// Find leaf to insert value
	leafToInsert, history := findLeaf(bTree, rootNode, key, root, make([]TreeNodePage, 0))

	if leafToInsert == nil {
		/* Didn't find any leaf, that means the key to be inserted is smaller than the smallest key in bytes

		So we find first node and insert the key to it
		*/
		leafToInsert, history = findFirstLeaf(bTree)
	}

	fmt.Printf("DEBUG::LEAF FOUND AT PAGE %d with %d Items \n", leafToInsert.page, leafToInsert.node.GetNItens())
	// Verify whether leaf must be splited
	insertAndReorderTree(bTree, *leafToInsert, history, key, value)
}

func BTreeDelete(bTree *BTree, key []byte) {
	// Find correspondent Node's (for repeated key)

}

/*
BTreeGet - Base function to get first Key Values stored in bTree
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
	allLeafValues := getAllLeafKeyValues(&leaf.node)

	for i := 0; i < len(allLeafValues); i++ {
		if bytes.Compare(allLeafValues[i].key, key) == 0 {
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
