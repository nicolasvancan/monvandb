package main

import (
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

type BTree struct {
	data     []byte
	pageSize uint32
}

func NewTree() *BTree {
	// Returns a pointer to the new BTree in Memory
	return &BTree{
		data: make([]byte, PAGE_SIZE),
	}
}

func LoadTree(bTree []byte) *BTree {
	return &BTree{
		data: bTree[:PAGE_SIZE],
	}
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
