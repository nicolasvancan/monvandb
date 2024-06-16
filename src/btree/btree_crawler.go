package btree

import (
	"bytes"
	"fmt"
)

type BTreeCrawler struct {
	bTree            *BTree
	Net              []TreeNode
	Cursor           []int
	CurrentKeyValues []LeafKeyValue
}

func newBTreeCrawler(tree *BTree) *BTreeCrawler {
	return &BTreeCrawler{
		bTree:            tree,
		Net:              make([]TreeNode, 0),
		Cursor:           make([]int, 0),
		CurrentKeyValues: make([]LeafKeyValue, 0),
	}
}

func findLeafInPage(page TreeNode, key []byte, crawler *BTreeCrawler) *BTreeCrawler {
	// If leaf has the value we are looking for, we break the loop
	found := -1
	for i := 0; i < int(page.GetNItens()); i++ {
		if bytes.Compare(page.GetLeafKeyValueByIndex(uint16(i)).key, key) >= 0 {
			found = i
			break
		}
	}

	crawler.Net = append(crawler.Net, page)
	crawler.Cursor = append(crawler.Cursor, found)
	// If the key is not found, we will return the next leaf
	if found == -1 {
		crawler = nil
	}

	return crawler
}

func removeLastIdx(crawler *BTreeCrawler) {
	crawler.Net = crawler.Net[:len(crawler.Net)-1]
	crawler.Cursor = crawler.Cursor[:len(crawler.Cursor)-1]
}

/*
Function to go directly to the first list and first item, usefull when dealing with requests
that are not in index and all values must be evaluated
*/
func GoToFirstLeaf(tree *BTree) *BTreeCrawler {
	crawler := newBTreeCrawler(tree)
	rootAddr := tree.GetRoot()
	page := tree.Get(rootAddr)
	// While loop to find the leaf
	for {
		if page.GetType() == TREE_NODE {
			crawler.Net = append(crawler.Net, page)
			crawler.Cursor = append(crawler.Cursor, 0)
			// Get the next page
			page = crawler.bTree.Get(page.GetNodeChildByIndex(0).GetAddr())

		} else {
			crawler.Net = append(crawler.Net, page)
			crawler.Cursor = append(crawler.Cursor, 0)
			crawler.CurrentKeyValues = getAllLeafKeyValues(&page)
			break
		}
	}

	return crawler
}

/*
Function to go directly to the last list and last item, usefull when dealing with requests
that are not in index and all values must be evaluated from backyards
*/
func GoToLastLeaf(tree *BTree) *BTreeCrawler {
	crawler := newBTreeCrawler(tree)
	rootAddr := tree.GetRoot()
	page := tree.Get(rootAddr)
	// While loop to find the leaf
	for {
		if page.GetType() == TREE_NODE {
			cur := int(page.GetNItens() - 1)
			crawler.Net = append(crawler.Net, page)
			crawler.Cursor = append(crawler.Cursor, cur)
			// Get the next page
			page = tree.Get(page.GetNodeChildByIndex(cur).GetAddr())

		} else {
			crawler.Net = append(crawler.Net, page)
			crawler.Cursor = append(crawler.Cursor, int(page.GetNItens()-1))
			crawler.CurrentKeyValues = getAllLeafKeyValues(&page)
			break
		}
	}
	return crawler
}

/*
This function makes the crawler go to the next key value in the bTree if it is leaf
Otherwise it finds the next branch containing leaf and goes after it
*/
func (crawler *BTreeCrawler) Next() error {
	if len(crawler.Net) == 0 {
		return fmt.Errorf("no more keys")
	}

	reachedEnd := true
	// If it has reached maximum
	for i := 0; i < len(crawler.Net); i++ {
		if crawler.Net[i].GetNItens()-1 != uint16(crawler.Cursor[i]) {
			reachedEnd = false
			break
		}
	}

	if reachedEnd {
		return fmt.Errorf("no more keys")
	}

	lastNode := crawler.Net[len(crawler.Net)-1]
	lastNodeIdx := crawler.Cursor[len(crawler.Cursor)-1]
	// If it is a leaf, we need to find the next leaf
	// If it is the last key in leaf we need to remove it from the net and call next again
	if lastNodeIdx == int(lastNode.GetNItens()-1) {

		removeLastIdx(crawler)
		crawler.Next()
	} else {
		// If it is not the last key in leaf, we just increment the Cursor
		crawler.Cursor[len(crawler.Cursor)-1]++
		if lastNode.GetType() == TREE_NODE {
			nextIdx := lastNodeIdx + 1
			nextNode := crawler.bTree.Get(lastNode.GetNodeChildByIndex(nextIdx).GetAddr())
			crawler.Net = append(crawler.Net, nextNode)
			crawler.Cursor = append(crawler.Cursor, 0)

			if nextNode.GetType() == TREE_LEAF {
				crawler.CurrentKeyValues = getAllLeafKeyValues(&nextNode)
			}
		}
	}

	return nil
}

/*
Same as Next() but goes to the previous key value in the bTree
*/
func (crawler *BTreeCrawler) Previous() error {
	if len(crawler.Net) == 0 {
		return fmt.Errorf("no more keys")
	}

	reachedEnd := true
	// If it has reached maximum
	for i := 0; i < len(crawler.Net); i++ {
		if uint16(crawler.Cursor[i]) != 0 {
			reachedEnd = false
			break
		}
	}

	if reachedEnd {
		return fmt.Errorf("no more keys")
	}

	lastNode := crawler.Net[len(crawler.Net)-1]
	lastNodeIdx := crawler.Cursor[len(crawler.Cursor)-1]

	if lastNodeIdx == 0 {
		fmt.Println("Previous lastNodeIdx")
		removeLastIdx(crawler)
		crawler.Previous()
	} else {
		// If it is not the last key in leaf, we just decrement the Cursor
		crawler.Cursor[len(crawler.Cursor)-1]--

		if lastNode.GetType() == TREE_NODE {
			nextIdx := lastNodeIdx - 1
			nextNode := crawler.bTree.Get(lastNode.GetNodeChildByIndex(nextIdx).GetAddr())
			crawler.Net = append(crawler.Net, nextNode)
			crawler.Cursor = append(crawler.Cursor, int(nextNode.GetNItens())-1)

			// Update the current key values
			if nextNode.GetType() == TREE_LEAF {
				crawler.CurrentKeyValues = getAllLeafKeyValues(&nextNode)
			}
		}
	}

	return nil
}

/*
Get actual BTreeKeyValue in the crawler
*/
func (crawler *BTreeCrawler) GetKeyValue() BTreeKeyValue {
	leaf := crawler.Net[len(crawler.Net)-1]

	// It is a linked list
	if leaf.GetLeafHasSeq() > 0 {
		key := leaf.GetLeafKeyValueByIndex(uint16(0)).key
		value := getAllBytesFromSequences(crawler.bTree, leaf)

		return BTreeKeyValue{
			Key:   key,
			Value: value,
		}
	}

	keyValueIdx := crawler.Cursor[len(crawler.Cursor)-1]
	kv := crawler.CurrentKeyValues[keyValueIdx]

	return BTreeKeyValue{
		Key:   kv.key,
		Value: kv.value,
	}
}
