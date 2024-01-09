package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

/*
   B-Tree Node structure Fixed Z Size in bytes - for instance 4096 one page

   - | type| 2B - A Node can be two diferent types: Node or Leaf
   - | nItems | 2B - Number of items, either  that the node holds
   - | freeBytes | 2B - Number of free bytes in the node
   - | pParent | 8B - Pointer to parent node
   - | offset | 2B - Position of the first freeByte (Used for adding and replacing items)
   --------- Case Node -----------
   - | pNChildKey | n * 8B - Each Node has one unique key
   - | pNChild | n * 8B
   --------- Case Leaf -----------
   - | hasSeq  | 2B - If it has a sequence to complete the row case data exceeds max page value
   - | vSeq | 8B - Pointer to the sequence of the value in case it exceeds the maximum number of bytes (Todo)
   ////////////////////// N *
   - | kLen | 2B - Len in bytes of the index
   - | ValLen | 8B - idx Values Len (Total size in bytes of a database structure to be saved)

   - | key | n * B - Indexes one after other

   - | Val | n * B - Values concatenated
*/

/* TreeNode implementation */
const (
	TREE_NODE = iota
	TREE_LEAF
	TREE_LEAF_SEQUENCE
)

/* Lens */
const (
	NODE_TYPE_LEN          = 2
	NODE_OFFSET_LEN        = 2
	NODE_N_ITENS_LEN       = 2
	NODE_FREE_BYTES_LEN    = 2
	NODE_PARENT_ADDR       = 8
	NODE_P_KEY_LEN         = 8
	NODE_P_CHILD_ADD_LEN   = 8
	NODE_P_KEY_ADDRESS_LEN = NODE_P_KEY_LEN + NODE_P_CHILD_ADD_LEN
	LEAF_HAS_SEQ_LEN       = 2
	LEAF_SEQ_P_LEN         = 8
	LEAF_KEY_LEN_LEN       = 2
	LEAF_VAL_LEN_LEN       = 8
)

/* Offsets Header */
const (
	NODE_TYPE_OFFSET        = 0
	NODE_N_ITENS_OFFSET     = NODE_TYPE_OFFSET + NODE_TYPE_LEN
	NODE_FREE_BYTES_OFFSET  = NODE_N_ITENS_OFFSET + NODE_N_ITENS_LEN
	NODE_PARENT_ADDR_OFFSET = NODE_FREE_BYTES_OFFSET + NODE_FREE_BYTES_LEN
	NODE_OFFSET_OFFSET      = NODE_PARENT_ADDR_OFFSET + NODE_PARENT_ADDR
	NODE_P_KEY_ADDR_OFFSET  = NODE_OFFSET_OFFSET + NODE_OFFSET_LEN
)

const (
	LEAF_HAS_SEQ_OFFSET   = NODE_OFFSET_OFFSET + NODE_OFFSET_LEN
	LEAF_SEQ_P_OFFSET     = LEAF_HAS_SEQ_OFFSET + LEAF_HAS_SEQ_LEN
	LEAF_KEY_LEN_OFFSET   = LEAF_SEQ_P_OFFSET + LEAF_SEQ_P_LEN
	LEAF_VAL_LEN_OFFSET   = LEAF_KEY_LEN_OFFSET + LEAF_KEY_LEN_LEN
	LEAF_VAL_START_OFFSET = LEAF_SEQ_P_OFFSET + LEAF_SEQ_P_LEN
)

type NodeKeyAddr struct {
	keyLen uint16
	key    []byte
	addr   uint64
}

type LeafKeyValue struct {
	keyLength   uint16
	valueLength uint64
	key         []byte
	value       []byte
}

/* Base Node */
type TreeNode struct {
	// This holds bytes to be dumped to the disk
	data []byte
}

func (n *TreeNode) GetType() uint16 {
	return uint16(binary.LittleEndian.Uint16(n.data[NODE_TYPE_OFFSET : NODE_TYPE_OFFSET+NODE_TYPE_LEN]))
}

func setNodeOffset(n *TreeNode, offset uint16) {
	binary.LittleEndian.PutUint16(n.data[NODE_OFFSET_OFFSET:NODE_OFFSET_OFFSET+NODE_OFFSET_LEN], offset)
}

func getNodeOffset(n *TreeNode) uint16 {
	return binary.LittleEndian.Uint16(n.data[NODE_OFFSET_OFFSET : NODE_OFFSET_OFFSET+NODE_OFFSET_LEN])
}

func setType(n *TreeNode, nType uint16) {
	binary.LittleEndian.PutUint16(n.data[NODE_TYPE_OFFSET:NODE_TYPE_OFFSET+NODE_TYPE_LEN], nType)
}

func (n *TreeNode) GetNItens() uint16 {
	return binary.LittleEndian.Uint16(n.data[NODE_N_ITENS_OFFSET : NODE_N_ITENS_OFFSET+NODE_N_ITENS_LEN])
}

func (n *TreeNode) GetParentAddr() uint64 {
	return binary.LittleEndian.Uint64(n.data[NODE_PARENT_ADDR_OFFSET : NODE_PARENT_ADDR_OFFSET+NODE_PARENT_ADDR])
}

func setParentAddr(n *TreeNode, addr uint64) {
	binary.LittleEndian.PutUint64(n.data[NODE_PARENT_ADDR_OFFSET:NODE_PARENT_ADDR_OFFSET+NODE_PARENT_ADDR], addr)
}

func setNItens(n *TreeNode, num uint16) {
	binary.LittleEndian.PutUint16(n.data[NODE_N_ITENS_OFFSET:NODE_N_ITENS_OFFSET+NODE_N_ITENS_LEN], num)
}

func getFreeBytes(n *TreeNode) uint16 {
	return binary.LittleEndian.Uint16(n.data[NODE_FREE_BYTES_OFFSET : NODE_FREE_BYTES_OFFSET+NODE_FREE_BYTES_LEN])
}

func setFreeBytes(n *TreeNode, num uint16) {
	binary.LittleEndian.PutUint16(n.data[NODE_FREE_BYTES_OFFSET:NODE_FREE_BYTES_OFFSET+NODE_FREE_BYTES_LEN], num)
}

func (n *TreeNode) ShouldSplitNode(newBytesNum uint32) bool {
	if newBytesNum > uint32(getFreeBytes(n)) {
		return true
	}

	return false
}

func isValidNodeStructure(n *TreeNode, idx uint16) (bool, string) {
	if n.GetType() != TREE_NODE {
		return false, "Is not Node Node"
	}

	if idx < 0 {
		return false, "Index doesn't exist"
	}

	if NODE_N_ITENS_OFFSET+(NODE_P_KEY_ADDRESS_LEN*idx) >= PAGE_SIZE {
		return false, "Overflow Index"
	}

	return true, ""
}

func NewNodeNode() *TreeNode {
	nodeNode := &TreeNode{data: make([]byte, PAGE_SIZE)}
	setType(nodeNode, TREE_NODE)
	setNItens(nodeNode, 0)
	setNodeOffset(nodeNode, 16)
	setFreeBytes(nodeNode, PAGE_SIZE-NODE_FREE_BYTES_LEN-NODE_TYPE_LEN-NODE_N_ITENS_LEN-NODE_PARENT_ADDR-NODE_OFFSET_LEN)
	return nodeNode
}

func getAllNodeKeyAddr(n *TreeNode) []NodeKeyAddr {
	// Get number of items
	nItens := n.GetNItens()
	// Initiate return array
	r := make([]NodeKeyAddr, nItens)
	// Start always at the very beginning
	lastStart := NODE_P_KEY_ADDR_OFFSET
	for i := 0; i < int(nItens); i++ {
		// Get key Length
		kLen := binary.LittleEndian.Uint16(n.data[lastStart : lastStart+2])
		// Get key value in []bytes
		key := n.data[lastStart+2 : lastStart+2+int(kLen)]
		// Get key address
		addr := binary.LittleEndian.Uint64(n.data[lastStart+2+int(kLen) : lastStart+2+int(kLen)+8])
		r[i] = NodeKeyAddr{
			keyLen: kLen,
			key:    key,
			addr:   addr,
		}
		lastStart += 2 + int(kLen) + 8
	}
	return r
}

func (n *TreeNode) PutNodeNewChild(key []byte, addr uint64) error {

	// Verify whether it will exceed total bytes
	aditionalLength := len(key) + 2 + 8
	if int(getFreeBytes(n))-(aditionalLength) < 0 {
		return errors.New("Exceeds total bytes")
	}
	keyLen := uint16(len(key))
	// takes offset
	offset := getNodeOffset(n)

	/*
		2B - Len of key
		Len of Key B - Key
		8B - Address
		Example:
		key = ["a","t","o","m","i","c"]
		addr = 157

		keyLen = 6 - Therefore the size will be 2B + 6B + 8B = 16B
	*/

	// Write len 2B
	binary.LittleEndian.PutUint16(n.data[offset:offset+2], keyLen)
	// Write Key (variable)
	copy(n.data[offset+2:offset+2+keyLen], key)
	// Write Address 8B
	binary.LittleEndian.PutUint64(n.data[offset+2+keyLen:offset+2+keyLen+8], addr)

	// Set new offset
	setNodeOffset(n, offset+2+keyLen+8)
	// Set new Free Bytes
	setFreeBytes(n, getFreeBytes(n)-(2+8+keyLen))
	// Set NItems
	setNItens(n, n.GetNItens()+1)

	return nil
}

func sortNodeChildren(c []NodeKeyAddr) {
	sort.Slice(c, func(i, j int) bool {
		return bytes.Compare(c[i].key, c[j].key) <= 0
	})
}

func (n *TreeNode) GetNodeChildByIndex(idx int) *NodeKeyAddr {
	// Get all keys

	if idx > int(n.GetNItens())-1 {
		return nil
	}

	unsortedChildren := getAllNodeKeyAddr(n)
	// Sort children
	sortNodeChildren(unsortedChildren)
	return &unsortedChildren[idx]
}

func (n *TreeNode) GetNodeChildByKey(key []byte) *NodeKeyAddr {
	var r *NodeKeyAddr = nil
	unsortedChildren := getAllNodeKeyAddr(n)
	for j := 0; j < len(unsortedChildren); j++ {
		if bytes.Compare(unsortedChildren[j].key, key) == 0 {
			r = &unsortedChildren[j]
		}
	}

	return r
}

/* Leaf Node functions */
func NewNodeLeaf() *TreeNode {
	nodeLeaf := &TreeNode{data: make([]byte, PAGE_SIZE)}
	setType(nodeLeaf, TREE_LEAF)
	setLeafHasSeq(nodeLeaf, 0)
	setLeafSeqPointer(nodeLeaf, 0)
	setNItens(nodeLeaf, 0)
	setNodeOffset(nodeLeaf, 26)
	setFreeBytes(nodeLeaf,
		PAGE_SIZE-NODE_FREE_BYTES_LEN-
			NODE_TYPE_LEN-NODE_N_ITENS_LEN-
			NODE_PARENT_ADDR-LEAF_HAS_SEQ_LEN-NODE_OFFSET_LEN-
			LEAF_SEQ_P_LEN)

	return nodeLeaf
}

func setLeafHasSeq(n *TreeNode, hasSeq uint16) {
	binary.LittleEndian.PutUint16(n.data[LEAF_HAS_SEQ_OFFSET:LEAF_HAS_SEQ_OFFSET+LEAF_HAS_SEQ_LEN], hasSeq)
}

func (n *TreeNode) GetLeafHasSeq() uint16 {
	return binary.LittleEndian.Uint16(n.data[LEAF_HAS_SEQ_OFFSET : LEAF_HAS_SEQ_OFFSET+LEAF_HAS_SEQ_LEN])
}

func setLeafSeqPointer(n *TreeNode, p uint64) {
	binary.LittleEndian.PutUint64(n.data[LEAF_SEQ_P_OFFSET:LEAF_SEQ_P_OFFSET+LEAF_SEQ_P_LEN], p)
}

func (n *TreeNode) GetLeafSeqPointer() uint64 {
	return binary.LittleEndian.Uint64(n.data[LEAF_SEQ_P_OFFSET : LEAF_SEQ_P_OFFSET+LEAF_SEQ_P_LEN])
}

func getAllLeafKeyValues(n *TreeNode) []LeafKeyValue {
	nItems := n.GetNItens()
	r := make([]LeafKeyValue, nItems)
	baseOffset := LEAF_VAL_START_OFFSET
	for i := 0; i < int(nItems); i++ {
		kLen := binary.LittleEndian.Uint16(n.data[baseOffset : baseOffset+2])
		vLen := binary.LittleEndian.Uint64(n.data[baseOffset+2 : baseOffset+2+8])
		key := n.data[baseOffset+10 : baseOffset+10+int(kLen)]
		value := n.data[baseOffset+10+int(kLen) : baseOffset+10+int(kLen)+int(vLen)]
		r[i] = LeafKeyValue{
			keyLength:   kLen,
			valueLength: vLen,
			key:         key,
			value:       value,
		}
		// Add values to offset
		baseOffset += (int(kLen)) + int(vLen) + 2 + 8
	}

	return r
}

func (n *TreeNode) PutLeafNewKeyValue(key []byte, value []byte) error {

	aditionalLength := len(key) + 2 + 8 + len(value)

	if int(getFreeBytes(n))-(aditionalLength) < 0 {
		return errors.New("Exceeds total bytes")
	}
	fmt.Println("Startgin")
	keyLen := uint16(len(key))
	valLen := uint64(len(value))
	// takes offset
	offset := getNodeOffset(n)
	fmt.Printf("Offset = %d\n", offset)
	/*
		2B - Len of key
		8B - Len of value
		Len of Key B - Key
		8B - Value
		Example:
		key = ["a","t","o","m","i","c"]
		value = []byte("some value inserted in here")

		keyLen = 6 - Therefore the size will be 2B + 8B + 6B + 27B = 43B
	*/

	// Write keylen 2B
	binary.LittleEndian.PutUint16(n.data[offset:offset+2], keyLen)
	// Write valuelen 8B
	binary.LittleEndian.PutUint64(n.data[offset+2:offset+2+8], valLen)
	// Write Key (variable)
	copy(n.data[offset+10:offset+10+keyLen], key)
	// Write Address 8B
	copy(n.data[offset+10+keyLen:offset+10+keyLen+uint16(valLen)], value)

	// Set new offset
	setNodeOffset(n, offset+10+keyLen+uint16(valLen))
	// Set new Free Bytes
	setFreeBytes(n, getFreeBytes(n)-(10+keyLen+uint16(valLen)))
	// Set NItems
	setNItens(n, n.GetNItens()+1)

	return nil
}

func sortLeafKeyValues(c []LeafKeyValue) {
	sort.Slice(c, func(i, j int) bool {
		return bytes.Compare(c[i].key, c[j].key) <= 0
	})
}

func (n *TreeNode) GetLeafKeyValueByIndex(idx uint16) *LeafKeyValue {
	if idx >= n.GetNItens() {
		return nil
	}

	allKeyValues := getAllLeafKeyValues(n)
	sortLeafKeyValues(allKeyValues)

	return &allKeyValues[idx]
}

func (n *TreeNode) GetLeafKeyValueByKey(key []byte) *LeafKeyValue {
	var r *LeafKeyValue = nil
	allKeyValues := getAllLeafKeyValues(n)
	for i := 0; i < len(allKeyValues); i++ {
		if bytes.Compare(allKeyValues[i].key, key) == 0 {
			r = &allKeyValues[i]
		}
	}

	return r
}

func getSplitParameters(n *TreeNode, times int) (int, int) {
	nItems := n.GetNItens()

	var quantityPerTime int
	var lastAdditional = 0
	if times >= int(nItems) {
		quantityPerTime = 1
	} else {
		quantityPerTime = int(nItems) / times
		if nItems%uint16(times) != 0 {
			lastAdditional = int(nItems) - (times * quantityPerTime)
		}
	}

	return quantityPerTime, lastAdditional
}

func (n *TreeNode) SplitLeaf(times int) []TreeNode {

	quantityPerTime, lastAdditional := getSplitParameters(n, times)
	r := make([]TreeNode, times)
	allLeafKeyValues := getAllLeafKeyValues(n)
	for i := 0; i < times; i++ {
		yLimit := quantityPerTime
		r[i] = *NewNodeLeaf()
		setParentAddr(&r[i], n.GetParentAddr())
		if i < times-1 {
			yLimit = quantityPerTime + lastAdditional
		}
		for y := 0; y < yLimit; y++ {
			tmp := allLeafKeyValues[uint16((i*quantityPerTime)+y)]
			r[i].PutLeafNewKeyValue(tmp.key, tmp.value)
		}
	}

	return r
}

func (n *TreeNode) SplitNode(times int) []TreeNode {

	quantityPerTime, lastAdditional := getSplitParameters(n, times)
	r := make([]TreeNode, times)
	allNodeChildren := getAllNodeKeyAddr(n)
	for i := 0; i < times; i++ {
		yLimit := quantityPerTime
		r[i] = *NewNodeNode()
		setParentAddr(&r[i], n.GetParentAddr())
		if i < times-1 {
			yLimit = quantityPerTime + lastAdditional
		}
		for y := 0; y < yLimit; y++ {
			tmp := allNodeChildren[(i*quantityPerTime)+y]
			r[i].PutNodeNewChild(tmp.key, tmp.addr)
		}
	}

	return r
}
