package btree

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
			fmt.Printf("DEBUG::Didn't find index for byte %d\n", binary.LittleEndian.Uint32(key))
			return nil, nil
		}
	}

	return &TreeNodePage{node: node, page: page}, history
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
	hasSeq := node.GetLeafHasSeq()
	seqAddr := node.GetLeafSeqPointer()

	sequence := bTree.Get(seqAddr)
	fmt.Printf("hasSeq := %d and sequence\n", hasSeq, sequence)
	// Still need to implement for this type of leaf
	return nil
}

/* Verify whether leaf should be splited*/
func mustSplitNode(bTree *BTree, node TreeNode, keyLen int, valueLen int) bool {
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
