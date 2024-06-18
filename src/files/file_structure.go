package files

import (
	"os"

	"github.com/nicolasvancan/monvandb/src/btree"
)

type DataFile struct {
	path  string
	bTree *btree.BTree
	fp    *os.File
}

// Comparators

const (
	EQ  = iota // Equal
	GT         // Greater than
	GTE        // Greater than or equal
	LT         // Less than
	LTE        // Less than or equal
	NE         // Not equal
)

/*
Range options is the most simplified version that I could think of to make the range function
It will go from
*/
type RangeOptions struct {
	From        []byte
	To          []byte
	FComparator int
	TComparator int
}

func (p *DataFile) GetBTree() *btree.BTree {
	return p.bTree
}

func OpenDataFile(path string) (*DataFile, error) {
	p := DataFile{
		path:  path,
		bTree: nil,
		fp:    nil,
	}

	p.path = path
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)

	if err != nil {
		return nil, err
	}

	stat, _ := file.Stat()

	// If the file is empty, create a new tree
	if stat.Size() == 0 {
		bTree := btree.NewTree(btree.PAGE_SIZE)
		bTree.SetName("bTree")
		p.bTree = bTree
		// Write the tree to the file
		file.WriteAt(bTree.GetBytes(), 0)
		defer file.Sync()
	}

	treeHeader := make([]byte, btree.PAGE_SIZE)
	file.ReadAt(treeHeader, 0)
	p.fp = file

	p.bTree = btree.LoadTree(treeHeader, btree.PAGE_SIZE)

	err = p.loadCallbacks()

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// Get retrieves a value from the BTree
func (p *DataFile) Get(key []byte) []btree.BTreeKeyValue {
	return btree.BTreeGet(p.bTree, key)
}

// Insert inserts a key-value pair into the BTree
func (p *DataFile) Insert(key []byte, value []byte) {
	btree.BTreeInsert(p.bTree, key, value)
}

// Delete removes a key-value pair from the BTree
func (p *DataFile) Delete(key []byte) {
	btree.BTreeDelete(p.bTree, key)
}

// Update updates a key-value pair in the BTree
func (p *DataFile) Update(key []byte, value []byte) {
	btree.BTreeUpdate(p.bTree, key, value)
}

// Get iterator
func (p *DataFile) GetIterator(key []byte) *btree.BTreeCrawler {
	return p.bTree.FindLeafForCrawling(key)
}

// ForceSync forces the os to flush the file to disk
func (p *DataFile) ForceSync() {
	defer p.fp.Sync()
}

// Close closes the file
func (p *DataFile) Close() {
	defer p.fp.Close()
}

// loadCallbacks sets the callbacks for the BTree
func (p *DataFile) loadCallbacks() error {
	// Set callbacks
	p.bTree.Set = func(node btree.TreeNode, page uint64) bool {
		_, err := p.fp.WriteAt(node.GetBytes(), int64(page*btree.PAGE_SIZE))

		if err != nil {
			return false
		}

		return true
	}

	p.bTree.SetHeader = func(bTree btree.BTree) {
		p.fp.WriteAt(bTree.GetBytes(), 0)
	}

	p.bTree.Get = func(page uint64) btree.TreeNode {
		data := make([]byte, btree.PAGE_SIZE)
		_, err := p.fp.ReadAt(data, int64(page*btree.PAGE_SIZE))

		if err != nil {
			panic(err)
		}

		return *btree.LoadTreeNode(data)
	}

	p.bTree.Del = nil

	p.bTree.New = func(node btree.TreeNode) uint64 {

		// get Stat from file
		fileInfo, err := p.fp.Stat()

		if err != nil {
			panic(err)
		}

		// Without header
		lastPage := (fileInfo.Size() / btree.PAGE_SIZE)

		p.fp.WriteAt(node.GetBytes(), fileInfo.Size())

		return uint64(lastPage)
	}

	return nil
}
