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
	return findLeafByOrder(bTree, rootNode, rootPage, make([]TreeNodePage, 0), "last")
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
		mappedLeavesFromNode := getMappedLeafForNode(bTree, bTree.Get(tmpNode.GetAddr()), tmpNode.GetAddr(), make([]uint64, root.page))

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
			newHistory := append(history, addr)
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
