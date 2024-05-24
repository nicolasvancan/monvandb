package btree

import (
	"bytes"
)

// This structure is used to map all leaves and history from bTree
type TreeNodeHistoryPages struct {
	TreeNode uint64
	History  []uint64
}

/*
function: startBTree

Starts bTree when it's a new table, or new view.
*/

func startsNewBTree(bTree *BTree, key []byte, value []byte) {
	newLeaf := NewNodeLeaf()
	// Means that it is top page
	setParentAddr(newLeaf, 0)
	newLeaf.PutLeafNewKeyValue(key, value)
	// Calls callback
	pageNumber := bTree.New(*newLeaf)
	bTree.root = pageNumber
	bTree.SetRoot(pageNumber)
	bTree.SetHeader(*bTree)
}

func createRootNodeAndInsertLeaves(bTree *BTree, treeLeaves []TreeNodePage) {
	newRoot := NewNodeNode()
	for i := 0; i < len(treeLeaves); i++ {
		if treeLeaves[i].node.GetType() == TREE_LEAF {
			newRoot.PutNodeNewChild(treeLeaves[i].node.GetLeafKeyValueByIndex(0).GetKey(), treeLeaves[i].page)
		} else {
			newRoot.PutNodeNewChild(treeLeaves[i].node.GetNodeChildByIndex(0).GetKey(), treeLeaves[i].page)
		}

	}

	setParentAddr(newRoot, 0) // Root
	newRootAddr := bTree.New(*newRoot)
	for i := 0; i < len(treeLeaves); i++ {
		setParentAddr(&treeLeaves[i].node, newRootAddr)
		bTree.Set(treeLeaves[i].node, treeLeaves[i].page)
	}

	bTree.SetRoot(newRootAddr)
	bTree.SetHeader(*bTree)
}

/*
function: findFirstLeaf

Find first leaf for given bTree
*/

func findFirstLeaf(bTree *BTree) (*TreeNodePage, []TreeNodePage) {
	rootPage := bTree.GetRoot()
	rootNode := bTree.Get(bTree.GetRoot())
	return findLeafByOrder(bTree, rootNode, rootPage, make([]TreeNodePage, 0), "first")
}

/*
function: findLastLeaf

Find last leaf for given bTree
*/

func findLastLeaf(bTree *BTree) (*TreeNodePage, []TreeNodePage) {
	rootPage := bTree.GetRoot()
	rootNode := bTree.Get(bTree.GetRoot())
	return findLeafByOrder(bTree, rootNode, rootPage, make([]TreeNodePage, 0), "last")
}

/* findLeaf Can be set order input to 'first' or 'last'*/
func findLeafByOrder(bTree *BTree, node TreeNode, page uint64, history []TreeNodePage, order string) (*TreeNodePage, []TreeNodePage) {
	if node.GetType() == TREE_NODE {
		idxToSearch := 0
		if order == "last" {
			idxToSearch = int(node.GetNItens()) - 1
		}
		nodeChild := node.GetNodeChildByIndex(idxToSearch)
		// Read First Node
		foundNode := bTree.Get(nodeChild.addr)
		history = append(history, TreeNodePage{node: foundNode, page: nodeChild.addr})
		return findLeafByOrder(bTree, foundNode, nodeChild.addr, history, order)
	}

	return &TreeNodePage{node: node, page: page}, history
}

/*
This function works recursiverly and returns exactly what Node the data must be inserted
*/
func findLeaf(bTree *BTree, node TreeNode, key []byte, page uint64, history []TreeNodePage) (*TreeNodePage, []TreeNodePage) {
	// Means it has reached some leaf
	if node.GetType() == TREE_NODE {
		// Tries to find Leaf through Internal Node
		if node.GetNItens() == 0 {
			return nil, nil
		}

		if idx := lookupKey(node, key); idx > -1 {
			// Get keyAddress from it
			nodeKeyAddr := node.GetNodeChildByIndex(idx)
			// Reach out the address from bTree file
			tmpNode := bTree.Get(nodeKeyAddr.addr)
			return findLeaf(bTree, tmpNode, key, nodeKeyAddr.addr, append(history, TreeNodePage{
				node: node,
				page: page,
			}))
		} else {
			return nil, nil
		}
	}

	return &TreeNodePage{node: node, page: page}, history
}

func findLeaves(bTree *BTree, node TreeNode, key []byte, page uint64, history []TreeNodePage) ([]TreeNodePage, []TreeNodePage) {
	treeNodes := make([]TreeNodePage, 0)
	// Means it has reached some leaf
	if node.GetType() == TREE_NODE {
		// Tries to find Leaf through Internal Node
		if node.GetNItens() == 0 {
			return nil, nil
		}

		if idx := lookupKeys(node, key); len(idx) > 0 {
			for i := 0; i < len(idx); i++ {
				nodeKeyAddr := node.GetNodeChildByIndex(idx[i])
				// Reach out the address from bTree file
				tmpNode := bTree.Get(nodeKeyAddr.addr)

				leavesFound, _ := findLeaves(bTree, tmpNode, key, nodeKeyAddr.addr, append(history, TreeNodePage{
					node: node,
					page: page,
				}))

				for i := 0; i < len(leavesFound); i++ {
					treeNodes = append(treeNodes, leavesFound[i])
				}
			}
			return treeNodes, history
		} else {
			return nil, nil
		}
	} else {
		return append(treeNodes, TreeNodePage{node: node, page: page}), history
	}
}

/*
mapAllLeavesToArray

Run through all nodes and insert into array all existing leaves through all layers
sorted

for instance, let's suppose we have the following tree

		(1      5      9) page 2
	   /        !       \
	(1 2 3 4) page 3 (5 6 7 8) page 4 (9 10 11 12) page 5

After the map we'll have the following array
[{3, [2]]},{4,[2]},{5,[2]}]
*/

func MapAllLeavesToArray(bTree *BTree) []TreeNodeHistoryPages {
	var mappedLeaves []TreeNodeHistoryPages = make([]TreeNodeHistoryPages, 0)
	root := *new(TreeNodePage)
	// Bind base history to root
	root = TreeNodePage{
		node: bTree.Get(bTree.GetRoot()),
		page: bTree.GetRoot(),
	}

	if root.node.GetType() == TREE_LEAF {
		mappedLeaves = append(mappedLeaves, TreeNodeHistoryPages{
			TreeNode: bTree.GetRoot(),
			History:  make([]uint64, 0),
		})

		return mappedLeaves
	}
	// Get number of Itens
	nItens := root.node.GetNItens()
	for i := 0; i < int(nItens); i++ {

		tmpNode := root.node.GetNodeChildByIndex(i)
		history := make([]uint64, 0)
		history = append(history, root.page)
		mappedLeavesFromNode := getMappedLeafForNode(
			bTree,
			bTree.Get(tmpNode.GetAddr()),
			tmpNode.GetAddr(),
			history,
		)
		// Append every item in order to the final array
		for j := 0; j < len(mappedLeavesFromNode); j++ {
			itemToBeAppended := mappedLeavesFromNode[j]
			mappedLeaves = append(mappedLeaves, itemToBeAppended)
		}
	}
	return mappedLeaves
}

func getMappedLeafForNode(bTree *BTree, node TreeNode, page uint64, history []uint64) []TreeNodeHistoryPages {

	mappedLeaf := make([]TreeNodeHistoryPages, 0)

	if node.GetType() == TREE_NODE {
		// Call this function for each children
		nItens := node.GetNItens()
		for i := 0; i < int(nItens); i++ {

			tmpKeyVal := node.GetNodeChildByIndex(i)
			addr := tmpKeyVal.GetAddr()
			newHistory := append(history, page)
			tmp := getMappedLeafForNode(bTree, bTree.Get(addr), addr, newHistory)
			for j := 0; j < len(tmp); j++ {
				mappedLeaf = append(mappedLeaf, tmp[j])
			}
		}

	} else {
		mappedLeaf = append(mappedLeaf, TreeNodeHistoryPages{
			TreeNode: page,
			History:  history,
		})
	}

	return mappedLeaf
}

/*
Get all bytes from Leaf that has sequence values, concatenating all of them and
returning into on single byte array
*/
func getAllBytesFromSequences(bTree *BTree, node TreeNode) []byte {
	// Sequence infos
	value := make([]byte, 0)
	hasSeq := node.GetLeafHasSeq()
	seqAddr := node.GetLeafSeqPointer()

	if node.GetType() == TREE_LEAF {
		keyVal := node.GetLeafKeyValueByIndex(0)
		value = append(value, keyVal.value...)
	} else {
		val := node.GetLeafSequenceBytes()
		value = append(value, val...)
	}

	if hasSeq == 1 {
		nextNode := bTree.Get(seqAddr)
		value = append(value, getAllBytesFromSequences(bTree, nextNode)...)
	}
	return value
}

func insertOneKeyLeafAndReorderTree(bTree *BTree, tPage TreeNodePage, SeqPage TreeNodePage, history []TreeNodePage) {

	if len(history) == 0 { // Case first node is a leaf node
		createRootNodeAndInsertLeaves(bTree, []TreeNodePage{tPage, SeqPage})
		return
	}

	parentNode := history[len(history)-1]
	// Verify if must split node
	keyLen := SeqPage.node.GetLeafKeyValueByIndex(0).GetKeyLen()

	// If the parent key must be splitted we must create a new node and insert it into parent
	if mustSplitNode(parentNode.node, int(keyLen), 8) {
		splitParentNodeRecursivellyAndReorderTreeIfNeeded(bTree, SeqPage, history)
		return
	}

	// Set parent addr
	parentNode.node.PutNodeNewChild(SeqPage.node.GetLeafKeyValueByIndex(0).GetKey(), SeqPage.page)
	// Set parent into bTree
	bTree.Set(parentNode.node, parentNode.page)
}

func splitParentNodeRecursivellyAndReorderTreeIfNeeded(bTree *BTree, SeqPage TreeNodePage, history []TreeNodePage) {

	var key []byte = nil

	if SeqPage.node.GetType() == TREE_LEAF {
		key = SeqPage.node.GetLeafKeyValueByIndex(0).GetKey()
	} else {
		key = SeqPage.node.GetNodeChildByIndex(0).GetKey()
	}

	parentNode := history[len(history)-1]

	// If the parent key must be splitted we must create a new node and insert it into parent
	splittedNode := parentNode.node.SplitNode(key, SeqPage.page)
	newPage := bTree.New(splittedNode[1])
	// Verify whether parent is root
	if parentNode.page == bTree.GetRoot() {
		createRootNodeAndInsertLeaves(bTree, []TreeNodePage{{node: splittedNode[0], page: parentNode.page}, {node: splittedNode[1], page: newPage}})
		return
	}

	if len(history)-2 > 0 {
		if mustSplitNode(history[len(history)-2].node, len(key), 8) {
			splitParentNodeRecursivellyAndReorderTreeIfNeeded(bTree, TreeNodePage{node: splittedNode[0], page: parentNode.page}, history[:len(history)-1])
			return
		}
	}

	parentNode.node.PutNodeNewChild(key, SeqPage.page)
	bTree.Set(parentNode.node, parentNode.page)
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
	// Means that this leaf is directly linked to the root
	if len(history) == 0 {
		newInternalNode := NewNodeNode()
		setParentAddr(newInternalNode, 0)
		newNodeAddress := bTree.New(*newInternalNode)
		// Set the splitted leaves
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
	For some reason, the node 2 must be splitted into 2 new nodes, at page 10 and 11. Knowing that, we ensure that basically
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

	// Insert into node recursivelly
	insertNodesRecursivelly(bTree, tPage, newLeavesPages, history)
}

func shiftValuesBetweenLeaves(bTree *BTree, tPage TreeNodePage, history []TreeNodePage, key []byte, value []byte) {
	parentAddr := tPage.node.GetParentAddr()
	// check if parent is not zero, indicating that there is a node above it
	if len(history) == 0 {
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

		indexOfLeaf := 0

		isLastIndex := false
		if len(mappedLeaves) <= 1 {
			isLastIndex = true
		}

		// Get next leaf
		nextLeafAddress := mappedLeaves[indexOfLeaf].TreeNode
		nextLeaf := bTree.Get(nextLeafAddress)
		if !isLastIndex {
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
		if !mustSplitNode(nextLeaf, len(keyValuesToBeReinserted[i].GetKey()), len(keyValuesToBeReinserted[i].GetValue())) {
			// Get first key from next leaf
			nextLeafFirstKey := nextLeaf.GetLeafKeyValueByIndex(0).GetKey()
			nextLeaf.PutLeafNewKeyValue(keyValuesToBeReinserted[i].GetKey(), keyValuesToBeReinserted[i].GetValue())

			// Update leaf
			bTree.Set(nextLeaf, nextLeafAddress)

			// Compare if key has changed and update parent
			if !bytes.Equal(nextLeafFirstKey, nextLeaf.GetLeafKeyValueByIndex(0).GetKey()) {
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
Insert value changing the bTree structure whenever it is needed.

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
	if mustSplitNode(tPage.node, keyLen, valueLen) {
		if (keyLen + valueLen + 10) > PAGE_SIZE-26 { // 26 is header size for leaf
			// Special case
			leaf := createLeafAndSequencesForLargeBytes(bTree, key, value)
			insertOneKeyLeafAndReorderTree(bTree, tPage, *leaf, history)
			return
		}
		// Verify whether the leaf if the last one by comparing he page number
		lastLeaf, _ := findLastLeaf(bTree)

		if lastLeaf.page == tPage.page {
			// Last leaf split's backyards Node Recursivelly
			splitBackyardsRecursively(bTree, tPage, history, key, value)
			return
		}
		shiftValuesBetweenLeaves(bTree, tPage, history, key, value)
		return
	}

	tPage.node.PutLeafNewKeyValue(key, value)
	bTree.Set(tPage.node, tPage.page)
}

func getTotalKeyLen(nodes []TreeNode) int {
	totalKeyLen := 0
	for i := 0; i < len(nodes); i++ {
		if nodes[i].GetType() == TREE_LEAF {
			totalKeyLen += len(nodes[i].GetLeafKeyValueByIndex(0).key)
			continue
		}

		totalKeyLen += len(nodes[i].GetNodeChildByIndex(0).key)
	}
	totalKeyLen += 10
	return totalKeyLen
}

func insertNewPagesToNode(bTree *BTree, nodeToInsert TreeNodePage, newPages []TreeNodePage) {
	// Insert new nodes information to parent node
	if newPages[0].node.GetType() == TREE_LEAF {
		nodeToInsert.node.PutNodeNewChild(newPages[0].node.GetLeafKeyValueByIndex(0).key, newPages[0].page)
		nodeToInsert.node.PutNodeNewChild(newPages[1].node.GetLeafKeyValueByIndex(0).key, newPages[1].page)
	} else {
		nodeToInsert.node.PutNodeNewChild(newPages[0].node.GetNodeChildByIndex(0).key, newPages[0].page)
		nodeToInsert.node.PutNodeNewChild(newPages[1].node.GetNodeChildByIndex(0).key, newPages[1].page)
	}

	// Update parent from nodes
	setParentAddr(&newPages[0].node, nodeToInsert.page)
	setParentAddr(&newPages[1].node, nodeToInsert.page)
	// Update pages
	bTree.Set(nodeToInsert.node, nodeToInsert.page)
	bTree.Set(newPages[0].node, newPages[0].page)
	bTree.Set(newPages[1].node, newPages[1].page)
}

func insertNodesRecursivelly(
	bTree *BTree,
	oldPage TreeNodePage,
	newPages []TreeNodePage,
	history []TreeNodePage,
) {
	// Get nodeToInsert, it will never happen when there is an empty history, so we can do it
	nodeToInsert := history[len(history)-1]
	// New pages always come divided into 2 pieces
	newPageOne := newPages[0]
	newPageTwo := newPages[1]

	totalKeyLen := getTotalKeyLen([]TreeNode{newPageOne.node, newPageTwo.node})
	// Remove oldPage reference from node that will receive new pages
	nodeToInsert.node.DeleteNodeChildrenByAddress(oldPage.page)

	// No need to split node
	if !mustSplitNode(nodeToInsert.node, totalKeyLen, 8) {
		insertNewPagesToNode(bTree, nodeToInsert, newPages)
		return
	}

	// We verify if value can be shifted to another existing leaf or we really need to create another leaf
	var splittedNode []TreeNode = nil
	ourInsertion := 1
	if newPageOne.node.GetType() == TREE_LEAF {
		// Leaf case (This is duplicated code unfortunatelly) must remake this
		splittedNode = nodeToInsert.node.SplitNode(newPageOne.node.GetLeafKeyValueByIndex(0).key, newPageOne.page)
		// Where did our key insertion go to the first or second leaf?
		if splittedNode[0].GetNodeChildByKey(newPageOne.node.GetLeafKeyValueByIndex(0).key) != nil {
			ourInsertion = 0
		}

		// Our insertion is sorted, therefore we must insert the second new node into the second splitted node
		splittedNode[1].PutNodeNewChild(newPageTwo.node.GetLeafKeyValueByIndex(0).key, newPageTwo.page)

	} else {
		splittedNode = nodeToInsert.node.SplitNode(newPageOne.node.GetNodeChildByIndex(0).key, newPageOne.page)
		// Where is our insertion?
		if splittedNode[0].GetNodeChildByKey(newPageOne.node.GetNodeChildByIndex(0).key) != nil {
			ourInsertion = 0
		}

		// Our insertion is sorted, therefore we must insert the second new node into the second splitted node
		splittedNode[1].PutNodeNewChild(newPageTwo.node.GetNodeChildByIndex(0).key, newPageTwo.page)
	}

	// Create our new pages
	addr0 := bTree.New(splittedNode[0])
	addr1 := bTree.New(splittedNode[1])

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
				{node: splittedNode[0], page: addr0},
				{node: splittedNode[1], page: addr1},
			},
			history[:len(history)-1])

	} else { // No parent :(
		// Create new root Node
		newRoot := NewNodeNode()
		setParentAddr(newRoot, 0)
		newRootAddress := bTree.New(*newRoot)
		setParentAddr(&splittedNode[0], newRootAddress)
		setParentAddr(&splittedNode[1], newRootAddress)
		newRoot.PutNodeNewChild(splittedNode[0].GetNodeChildByIndex(0).key, addr0)
		newRoot.PutNodeNewChild(splittedNode[1].GetNodeChildByIndex(0).key, addr1)
		bTree.Set(*newRoot, newRootAddress)
		bTree.Set(splittedNode[0], addr0)
		bTree.Set(splittedNode[1], addr1)
		// Set bTree to be redirected to new Root
		bTree.SetRoot(newRootAddress)
		// Commit changes
		bTree.SetHeader(*bTree)
	}
}

func createLeafAndSequencesForLargeBytes(bTree *BTree, key []byte, value []byte) *TreeNodePage {
	// Create in memory all leaves and leaf to be returned
	leaf, sequence := CreateLeafWithSequence(key, value)

	tmpAddr := uint64(0)
	for i := len(sequence) - 1; i >= 0; i-- {
		setLeafHasSeq(&sequence[i], 0)
		if i < len(sequence)-1 {
			setLeafHasSeq(&sequence[i], 1)
			setLeafSeqPointer(&sequence[i], tmpAddr)
		}
		// Create leaf effectivelly
		tmpAddr = bTree.New(sequence[i])
	}

	setLeafHasSeq(leaf, 1)
	setLeafSeqPointer(leaf, tmpAddr)

	r := new(TreeNodePage)
	r.node = *leaf
	r.page = bTree.New(*leaf)

	return r
}

/* Verify whether leaf should be splitted*/
func mustSplitNode(node TreeNode, keyLen int, valueLen int) bool {
	freeBytes := GetFreeBytes(&node)
	totalNewBytes := keyLen + valueLen
	/* If it is TREE_NODE, it has fixed value length of 8 bytes and 2 bytes of
	keylen field, totalizing 10 bytes. If it's a TREE_LAEF, it has key len plus value len
	plus 10 bytes of keylen field and valuelen field
	*/
	if node.GetType() == TREE_NODE {
		totalNewBytes = keyLen
	}

	return freeBytes < uint16(totalNewBytes)+10
}

/*
Implementation of function to lookup key in internal Node, returning page number
*/
func lookupKey(node TreeNode, key []byte) int {
	// Declare found variable initiated with -1 (Case we don't find any)
	var found int = -1
	var allNodeKeyAddr []NodeKeyAddr = nil
	var allLeafKeyValues []LeafKeyValue = nil
	nItens := node.GetNItens()
	// Just in case it is a Internal Node
	if node.GetType() == TREE_NODE {
		allNodeKeyAddr = ([]NodeKeyAddr)(getAllNodeKeyAddr(&node))
		// Iterate over all items to find a corresponding key

		for i := int(nItens) - 1; i >= 0; i-- {
			if bytes.Compare(allNodeKeyAddr[i].key, key) <= 0 {
				found = i
				break
			}
		}

	} else {
		allLeafKeyValues = ([]LeafKeyValue)(getAllLeafKeyValues(&node))
		// Iterate over all items to find a corresponding key
		for i := int(nItens) - 1; i >= 0; i-- {

			if bytes.Compare(allLeafKeyValues[i].key, key) <= 0 {
				found = i
				break
			}
		}
	}
	return found
}

func lookupKeys(node TreeNode, key []byte) []int {
	// Declare found variable initiated with -1 (Case we don't find any)
	var found []int = make([]int, 0)
	var allNodeKeyAddr []NodeKeyAddr = nil
	var allLeafKeyValues []LeafKeyValue = nil
	nItens := node.GetNItens()
	// Just in case it is a Internal Node
	if node.GetType() == TREE_NODE {
		allNodeKeyAddr = ([]NodeKeyAddr)(getAllNodeKeyAddr(&node))
		// Iterate over all items to find a corresponding key

		for i := int(nItens) - 1; i >= 0; i-- {
			comparsion := bytes.Compare(allNodeKeyAddr[i].key, key)
			if comparsion <= 0 {
				found = append(found, i)
				if comparsion < 0 {
					break
				}
			}
		}
	} else {
		allLeafKeyValues = ([]LeafKeyValue)(getAllLeafKeyValues(&node))
		// Iterate over all items to find a corresponding key
		for i := int(nItens) - 1; i >= 0; i-- {
			comparsion := bytes.Compare(allLeafKeyValues[i].key, key)
			if comparsion <= 0 {
				found = append(found, i)
				if comparsion < 0 {
					break
				}
			}
		}
	}
	return found

}

func DeleteKeyValueInLeafAndUpdateNodesRecursivelly(bTree *BTree, key []byte, tPage TreeNodePage, history []TreeNodePage) {
	tmp := tPage.node.GetLeafKeyValueByIndex(0).GetKey()
	firstKeyBeforeDelete := make([]byte, len(tmp))
	copy(firstKeyBeforeDelete, tmp)
	tPage.node.DeleteLeafKeyValueByKey(key)
	nItensAfterDeletion := tPage.node.GetNItens()
	firstkeyAfterDeletion := make([]byte, 0)

	if nItensAfterDeletion > uint16(0) {
		firstkeyAfterDeletion = tPage.node.GetLeafKeyValueByIndex(0).GetKey()
	}

	hasFirstKeyChanged := !bytes.Equal(firstKeyBeforeDelete, firstkeyAfterDeletion)

	tmpPage := tPage.page
	bTree.Set(tPage.node, tPage.page)

	// Single leaf deletion, no need to update parents
	if !hasFirstKeyChanged && nItensAfterDeletion > 0 {
		return
	}

	if tPage.page == bTree.GetRoot() {
		// If it is the root, we must update the root
		if nItensAfterDeletion == 0 {
			bTree.SetRoot(0)
			bTree.SetHeader(*bTree)
		}
		return
	}

	// If it is the first key, we must update all parents
	for i := len(history) - 1; i >= 0; i-- {
		node := history[i].node
		page := history[i].page
		// print first key addr for node
		node.DeleteNodeChildrenByKey(firstKeyBeforeDelete)
		if hasFirstKeyChanged && len(firstkeyAfterDeletion) > 0 {
			node.PutNodeNewChild(firstkeyAfterDeletion, tmpPage)
		}
		bTree.Set(node, page)
		tmpPage = page
	}
}
