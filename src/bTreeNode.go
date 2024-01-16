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

   - | type| 2B - A Node can be two diferent types: "Node" or "Leaf"
   - | nItems | 2B - Number of items, either  that the node holds
   - | freeBytes | 2B - Number of free bytes in the node
   - | pParent | 8B - Pointer to parent node
   - | offset | 2B - Position of the first freeByte (Used for adding and replacing items)
   --------- Case Node -----------
    \
   	     - | keyLen | 2B - Length of the key for cases where key is not an integer
   n*	 - | pNChildKey | n * 8B - Each Node has one unique key
   	     - | pNChild | n * 8B
    /
   --------- Case Leaf -----------
   - | hasSeq  | 2B - If it has a sequence to complete the row case data exceeds max page value
   - | vSeq | 8B - Pointer to the sequence of the value in case it exceeds the maximum number of bytes (Todo)
    \
         - | kLen | 2B - Len in bytes of the index
         - | ValLen | 8B - idx Values Len (Total size in bytes of a database structure to be saved)
    n*   - | key | n * B - Indexes one after other
         - | Val | n * B - Values concatenated
    /
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

/* Leaft offsets */
const (
	LEAF_HAS_SEQ_OFFSET   = NODE_OFFSET_OFFSET + NODE_OFFSET_LEN
	LEAF_SEQ_P_OFFSET     = LEAF_HAS_SEQ_OFFSET + LEAF_HAS_SEQ_LEN
	LEAF_KEY_LEN_OFFSET   = LEAF_SEQ_P_OFFSET + LEAF_SEQ_P_LEN
	LEAF_VAL_LEN_OFFSET   = LEAF_KEY_LEN_OFFSET + LEAF_KEY_LEN_LEN
	LEAF_VAL_START_OFFSET = LEAF_SEQ_P_OFFSET + LEAF_SEQ_P_LEN
)

/* Basic types declaration */
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

func LoadTreeNode(data []byte) *TreeNode {
	treeNode := &TreeNode{data: data}

	return treeNode
}

/* Basic Getters and Setters for header */

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

/*
Creates a new and Empty NodeNode structure
*/
func NewNodeNode() *TreeNode {
	// Create a pointer to a new Node Structure
	nodeNode := &TreeNode{data: make([]byte, PAGE_SIZE)}
	// Set headers
	setType(nodeNode, TREE_NODE)
	setNItens(nodeNode, 0)
	setNodeOffset(nodeNode, 16)
	setFreeBytes(nodeNode, PAGE_SIZE-NODE_FREE_BYTES_LEN-NODE_TYPE_LEN-NODE_N_ITENS_LEN-NODE_PARENT_ADDR-NODE_OFFSET_LEN)
	return nodeNode
}

func (n *TreeNode) ResetNode() {
	setType(n, TREE_NODE)
	setNItens(n, 0)
	setNodeOffset(n, 16)
	setFreeBytes(n, PAGE_SIZE-NODE_FREE_BYTES_LEN-NODE_TYPE_LEN-NODE_N_ITENS_LEN-NODE_PARENT_ADDR-NODE_OFFSET_LEN)
}

/*
Get all NodeKeyAddresses from a Node Node
*/
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

func (n *TreeNode) DeleteNodeChildrenByAddress(addr uint64) {
	allNodeKeyAddr := getAllNodeKeyAddr(n)
	fmt.Printf("DEBUG::DeleteNodeChildrenByAddress > AllNodeKeyAddr len = %d\n", len(allNodeKeyAddr))
	// Reset Node
	n.ResetNode()
	fmt.Println("DEBUG::DeleteNodeChildrenByAddress > Reset Node")
	// Sort Array
	sortNodeChildren(allNodeKeyAddr)
	for i := 0; i < len(allNodeKeyAddr); i++ {
		if allNodeKeyAddr[i].addr != addr {
			fmt.Println("DELETE::Inserting new one")
			n.PutNodeNewChild(allNodeKeyAddr[i].key, allNodeKeyAddr[i].addr)
		}
	}
}

func (n *TreeNode) DeleteNodeChildrenByKey(key []byte) {
	allNodeKeyAddr := getAllNodeKeyAddr(n)
	// Reset Node
	n.ResetNode()
	// Sort Array
	sortNodeChildren(allNodeKeyAddr)
	for i := 0; i < len(allNodeKeyAddr); i++ {
		if bytes.Compare(allNodeKeyAddr[i].key, key) != 0 {
			n.PutNodeNewChild(allNodeKeyAddr[i].key, allNodeKeyAddr[i].addr)
		}
	}
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

func (n *TreeNode) GetNodeChildByPage(page uint64) *NodeKeyAddr {
	var r *NodeKeyAddr = nil
	unsortedChildren := getAllNodeKeyAddr(n)
	for j := 0; j < len(unsortedChildren); j++ {
		if unsortedChildren[j].addr == page {
			r = &unsortedChildren[j]
			break
		}
	}

	return r
}

func (n *TreeNode) GetNodeChildByKey(key []byte) *NodeKeyAddr {
	var r *NodeKeyAddr = nil
	unsortedChildren := getAllNodeKeyAddr(n)
	for j := 0; j < len(unsortedChildren); j++ {
		if bytes.Compare(unsortedChildren[j].key, key) == 0 {
			r = &unsortedChildren[j]
			break
		}
	}

	return r
}

/*Calculate whether or not should split Node*/
func ShouldSplitNode(node TreeNode, keyLen int, valueLen int) bool {
	freeBytes := getFreeBytes(&node)

	totalNewBytes := keyLen + valueLen
	if node.GetType() == TREE_NODE {
		totalNewBytes = keyLen
	}

	return int(freeBytes)+totalNewBytes+10 > PAGE_SIZE
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
	keyLen := uint16(len(key))
	valLen := uint64(len(value))
	// takes offset
	offset := getNodeOffset(n)
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

func (n *TreeNode) SplitLeaf(key []byte, value []byte) []TreeNode {
	/* Some comments here. When a Leaf is split, we must ensure that it is saved with the right
	   members in each leaf. Since we don't sort in the insertion, we must check between two leaves
	   where our new value will be inserted. The left leaf will return filled with all possible data
	   trying to use the most of it space, whereas the second one, will have just the remaining data
	*/

	// Get all leaf members
	allLeafMembers := getAllLeafKeyValues(n)
	// Append new member
	allLeafMembers = append(
		allLeafMembers,
		LeafKeyValue{
			key:         key,
			keyLength:   uint16(len(key)),
			value:       value,
			valueLength: uint64(len(value)),
		},
	)
	// Sort them
	sortLeafKeyValues(allLeafMembers)
	// create two new Leaves
	newLeaves := []TreeNode{*NewNodeLeaf(), *NewNodeLeaf()}
	fmt.Printf("DEBUG > Free bytes from first %d\n", getFreeBytes(n))
	fmt.Printf("DEBUG > Tamanho novo %d \n", len(allLeafMembers))
	// For every member of leaf, including new one we insert until it reaches the possible max
	activeLeaf := 0
	for i := 0; i < len(allLeafMembers); i++ {
		freeBytes := getFreeBytes(&newLeaves[activeLeaf])

		member := allLeafMembers[i]
		if 10+member.keyLength+uint16(member.valueLength) > freeBytes {
			activeLeaf = 1
		}
		newLeaves[activeLeaf].PutLeafNewKeyValue(member.key, member.value)
	}

	return newLeaves
}

func (n *TreeNode) SplitNode(key []byte, addr uint64) []TreeNode {
	// Get all leaf members
	allNodeMembers := getAllNodeKeyAddr(n)
	allNodeMembers = append(
		allNodeMembers,
		NodeKeyAddr{
			keyLen: uint16(len(key)),
			key:    key,
			addr:   addr,
		},
	)

	// Sort them
	sortNodeChildren(allNodeMembers)

	// create two new Leaves
	newNodes := []TreeNode{*NewNodeNode(), *NewNodeNode()}

	// For every member of leaf, including new one we insert until it reaches the possible max
	activeNode := 0
	for i := 0; i < len(allNodeMembers); i++ {
		freeBytes := getFreeBytes(&newNodes[activeNode])
		member := allNodeMembers[i]
		if 10+member.keyLen > freeBytes {
			activeNode = 1
		}
		newNodes[activeNode].PutNodeNewChild(key, addr)
	}

	return newNodes
}
