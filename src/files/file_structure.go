package files

import (
	"os"
	"sync"

	"github.com/nicolasvancan/monvandb/src/btree"
)

// DataFile is a proxy to access both BTree file and cached pages.
// It also provides mechanisms to read, write and update the BTree structure for multithreaded access.

type DataFile struct {
	path        string
	lastPage    uint64
	bTree       *btree.BTree
	fp          *os.File
	CacheSystem *CacheSystem
	GetRequests map[uint64]int // Used to track Get requests for the BTree
	Mutex       sync.RWMutex   // Mutex to protect access to the BTree and CacheSystem
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

	_, treeHeader, _ := MmapPage(file, 0, uint64(btree.PAGE_SIZE))
	//file.ReadAt(treeHeader, 0)
	p.fp = file

	p.bTree = btree.LoadTree(treeHeader, btree.PAGE_SIZE)

	p.CacheSystem = NewCacheSystem(300, 20*1024*1024) // 5 minutes TTL for cache and 20MB max cache size

	p.lastPage = uint64(p.GetFileSize() / btree.PAGE_SIZE)

	// Start the cache cleanup routine
	go func() {
		ClearTimedOutCache(p.CacheSystem)
	}()

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

func (p *DataFile) GetFileSize() int64 {
	if p.fp == nil {
		panic("DataFile is not initialized")
	}
	fileInfo, err := p.fp.Stat()
	if err != nil {
		panic(err)
	}
	return fileInfo.Size()
}

// loadCallbacks sets the callbacks for the BTree
func (p *DataFile) loadCallbacks() error {
	// Set callbacks

	p.bTree.Get = p.CBGet
	p.bTree.New = p.CBNew
	p.bTree.Set = p.CBSet

	return nil
}

func (p *DataFile) CBGet(page uint64) btree.TreeNode {

	// Try reading the page from the cache first
	if p.CacheSystem != nil {
		if cachedPage, exists := p.CacheSystem.Get(page); exists {
			return cachedPage
		}
	}

	// If not in cache, read from the file
	if p.fp == nil {
		panic("DataFile is not initialized")
	}

	data := make([]byte, btree.PAGE_SIZE)
	p.Mutex.RLock() // Lock for reading
	_, err := p.fp.ReadAt(data, int64(page*btree.PAGE_SIZE))
	defer p.Mutex.RUnlock() // Ensure mutex is unlocked after reading

	if err != nil {
		panic(err)
	}

	if p.CacheSystem != nil {
		// Store the page in the cache
		p.CacheSystem.Set(page, *btree.LoadTreeNode(data))
	}

	return *btree.LoadTreeNode(data)
}

func (p *DataFile) CBSet(node btree.TreeNode, page uint64) bool {

	if p.CacheSystem != nil {
		p.CacheSystem.ToBeUpdated = append(p.CacheSystem.ToBeUpdated, page)
		p.CacheSystem.Set(page, node)
	}

	if p.fp == nil {
		panic("DataFile is not initialized")
	}
	p.Mutex.Lock() // Lock for writing
	_, err := p.fp.WriteAt(node.GetBytes(), int64(page*btree.PAGE_SIZE))
	defer p.Mutex.Unlock() // Unlock after writing

	return err == nil
}

func (p *DataFile) CBNew(node btree.TreeNode) uint64 {

	newPage := p.lastPage + 1

	if p.CacheSystem != nil {
		// Create a new page in the cache
		p.lastPage = newPage
		p.CacheSystem.Create(newPage, node)
	}
	// Without header
	p.Mutex.Lock() // Lock for writing
	p.fp.WriteAt(node.GetBytes(), int64(p.lastPage*btree.PAGE_SIZE))
	defer p.Mutex.Unlock() // Unlock after writing

	return uint64(newPage)
}
