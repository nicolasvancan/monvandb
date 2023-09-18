package main

import (
	"encoding/binary"
	"errors"
)

/*
   B-Tree Node structure Fixed Z Size in bytes - for instance 4096 one page

   - | type| 2B - A Node can be two diferent types: Node or Leaf
   - | nItems | 2B - Number of items, either  that the node holds
   - | freeBytes | 2B - Number of free bytes in the node
   --------- Case Node -----------
   - | pNChildKey | n * 8B - Each Node has one unique key
   - | pNChild | n * 8B
   --------- Case Leaf -----------
   - | hasSeq  | 2B - If it has a sequence to complete the row case data exceeds max page value
   - | vSeq | 8B - Pointer to the sequence of the value in case it exceeds the maximum number of bytes (Todo)
   - | kLen | 2B - Len in bytes of the index
   - | ValLen | 8B - idx Values Len (Total size in bytes of a database structure to be saved)

   - | key | n * B - Indexes one after other
 n*
   - | Val | n * B - Values concatenated
*/

/* TreeNode implementation */
const (
	TREE_NODE = iota
	TREE_LEAF
	TREE_LEAF_SEQUENCE
)

const (
	NODE_TYPE_LEN          = 2
	NODE_N_ITENS_LEN       = 2
	NODE_FREE_BYTES_LEN    = 2
	NODE_P_KEY_LEN         = 8
	NODE_P_CHILD_ADD_LEN   = 8
	NODE_P_KEY_ADDRESS_LEN = NODE_P_KEY_LEN + NODE_P_CHILD_ADD_LEN
	NODE_START_OFFSET      = 0
	NODE_N_ITENS_OFFSET    = NODE_START_OFFSET + NODE_TYPE_LEN
	NODE_FREE_BYTES_OFFSET = NODE_N_ITENS_OFFSET + NODE_N_ITENS_LEN
	NODE_P_KEY_ADDR_OFFSET = NODE_FREE_BYTES_OFFSET + NODE_FREE_BYTES_LEN
)

const (
	LEAF_HAS_SEQ_LEN    = 2
	LEAF_HAS_SEQ_OFFSET = NODE_FREE_BYTES_OFFSET + NODE_FREE_BYTES_LEN
	LEAF_SEQ_P_LEN      = 8
	LEAF_SEQ_P_OFFSET   = LEAF_HAS_SEQ_OFFSET + LEAF_HAS_SEQ_LEN
	LEAF_KEY_LEN_LEN    = 2
	LEAF_KEY_LEN_OFFSET = LEAF_SEQ_P_OFFSET + LEAF_SEQ_P_LEN
	LEAF_VAL_LEN_LEN    = 8
	LEAF_VAL_LEN_OFFSET = LEAF_KEY_LEN_OFFSET + LEAF_KEY_LEN_LEN
)

type NodeKeyAddr struct {
	key  uint64
	addr uint64
}

/* Base Node */
type TreeNode struct {
	// This holds bytes to be dumped to the disk
	data []byte
}

func NewNodeNode() *TreeNode {
	nodeNode := &TreeNode{data: make([]byte, PAGE_SIZE)}
	nodeNode.SetType(TREE_NODE)
	setNItens(nodeNode, 0)
	setFreeBytes(nodeNode, PAGE_SIZE-NODE_FREE_BYTES_LEN-NODE_TYPE_LEN-NODE_N_ITENS_LEN)
	return nodeNode
}

func (n *TreeNode) GetType() uint16 {
	return uint16(binary.LittleEndian.Uint16(n.data[NODE_START_OFFSET : NODE_START_OFFSET+NODE_TYPE_LEN]))
}

func (n *TreeNode) SetType(nType uint16) {
	binary.LittleEndian.PutUint16(n.data[NODE_START_OFFSET:NODE_START_OFFSET+NODE_TYPE_LEN], nType)
}

func (n *TreeNode) GetNItens() uint16 {
	return binary.LittleEndian.Uint16(n.data[NODE_N_ITENS_OFFSET : NODE_N_ITENS_OFFSET+NODE_N_ITENS_LEN])
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

func (n *TreeNode) isValidNodeStructure(idx uint16) (bool, string) {
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

func setIdxPointer(n *TreeNode, idx uint16, keyAddr NodeKeyAddr) error {
	// This function can set whichever memory within the range of keys and addresses
	if flag, txt := n.isValidNodeStructure(idx); !flag {
		return errors.New(txt)
	}

	idxBaseOffset := NODE_P_KEY_ADDR_OFFSET + (idx * (NODE_P_KEY_ADDRESS_LEN))
	binary.LittleEndian.PutUint64(
		n.data[idxBaseOffset:idxBaseOffset+(NODE_P_KEY_ADDRESS_LEN)],
		keyAddr.key)
	binary.LittleEndian.PutUint64(
		n.data[idxBaseOffset+NODE_P_KEY_LEN:idxBaseOffset+(NODE_P_KEY_ADDRESS_LEN)+NODE_P_KEY_LEN],
		keyAddr.addr)

	return nil
}

func getIdxPointer(n *TreeNode, idx uint16) (*NodeKeyAddr, error) {
	// Verify if it can be updated
	if flag, txt := n.isValidNodeStructure(idx); !flag {
		return nil, errors.New(txt)
	}

	nItens := n.GetNItens()

	if nItens == 0 {
		return nil, errors.New("No iten available")
	}

	if idx >= n.GetNItens() {
		return nil, errors.New("No existing item")
	}

	// Finds the corresponding offset
	idxBaseOffset := NODE_P_KEY_ADDR_OFFSET + (idx * (NODE_P_KEY_ADDRESS_LEN))
	// Binds values corresponding to addresses
	key := binary.LittleEndian.Uint64(n.data[idxBaseOffset : idxBaseOffset+(NODE_P_KEY_ADDRESS_LEN)])
	addr := binary.LittleEndian.Uint64(n.data[idxBaseOffset+NODE_P_KEY_LEN : idxBaseOffset+(NODE_P_KEY_ADDRESS_LEN)+NODE_P_KEY_LEN])

	return &NodeKeyAddr{
		key:  key,
		addr: addr,
	}, nil
}

func copyPKeyAddress(n *TreeNode, from uint16, to uint16) error {
	if flag, txt := n.isValidNodeStructure(from); !flag {
		return errors.New(txt)
	}

	if flag, txt := n.isValidNodeStructure(to); !flag {
		return errors.New(txt)
	}

	copy, err := getIdxPointer(n, from)

	if err != nil {
		return err
	}

	setIdxPointer(n, to, *copy)
	return nil
}

func (n *TreeNode) InsertNodeNewChild(key uint64, address uint64) error {
	// Keys pointers exist just when the page (Node) refers to NODE_NODE type, therefore we check wheter or not the node is a Node Node
	if n.GetType() != TREE_NODE {
		return errors.New("Is not Node Node")
	}

	nItens := n.GetNItens()
	if NODE_P_KEY_ADDRESS_LEN*nItens+NODE_P_KEY_ADDR_OFFSET > PAGE_SIZE {
		return errors.New("Overflow of Node, you must split node")
	}

	// Retrieves the array of existing items
	posToInsert := 0
	if nItens > 0 {
		for i := 0; i < int(nItens); i++ {
			item, _ := getIdxPointer(n, uint16(i))
			if item.key < key {
				posToInsert = i + 1
			} else {
				break
			}
		}

		if posToInsert < int(nItens) {
			for i := posToInsert; i >= int(nItens-1); i-- {
				copyPKeyAddress(n, uint16(i), uint16(i+1))
			}
		}

		setIdxPointer(n, uint16(posToInsert), NodeKeyAddr{key: key, addr: address})

	} else {
		setIdxPointer(n, 0, NodeKeyAddr{key: key, addr: address})
	}

	setFreeBytes(n, getFreeBytes(n)-NODE_P_KEY_ADDRESS_LEN)
	setNItens(n, nItens+1)

	return nil
}

func (n *TreeNode) RemoveNodeChild(key uint64) error {
	nItens := n.GetNItens()

	if nItens == 0 {
		return errors.New("Empty key")
	}

	// Fin key index
	keyIdx := -1
	//TODO: This is still slow, must implement something better
	for i := 0; i < int(nItens); i++ {
		tmp, _ := getIdxPointer(n, uint16(i))
		if tmp.key == key {
			keyIdx = i
		}

		if keyIdx > -1 {
			if i+1 > int(nItens)-1 {
				break
			}

			copyPKeyAddress(n, uint16(i+1), uint16(i))
		}
	}

	if keyIdx == -1 {
		return errors.New("No key found for key = " + string(rune(key)))
	}

	// Update Node infos
	setFreeBytes(n, getFreeBytes(n)+NODE_P_KEY_ADDRESS_LEN)
	setNItens(n, nItens-1)

	return nil
}

func (n *TreeNode) GetNodeFirstKey() (uint64, error) {
	var fKey uint64

	if n.GetNItens() == 0 {
		return 0, errors.New("No key")
	}

	tmp, _ := getIdxPointer(n, 0)
	fKey = tmp.key
	return fKey, nil
}

func (n *TreeNode) GetNodeLastKey() (uint64, error) {
	var lKey uint64

	if n.GetNItens() == 0 {
		return 0, errors.New("No key")
	}

	tmp, _ := getIdxPointer(n, n.GetNItens()-1)
	lKey = tmp.key
	return lKey, nil
}

/* Leaf Node functions */
func NewNodeLeaf() *TreeNode {
	nodeLeaf := &TreeNode{data: make([]byte, PAGE_SIZE)}
	nodeLeaf.SetType(TREE_LEAF)
	setNItens(nodeLeaf, 0)
	setFreeBytes(nodeLeaf,
		PAGE_SIZE-NODE_FREE_BYTES_LEN-
			NODE_TYPE_LEN-NODE_N_ITENS_LEN-
			LEAF_HAS_SEQ_LEN-LEAF_KEY_LEN_LEN-
			LEAF_SEQ_P_LEN-LEAF_VAL_LEN_LEN)

	return nodeLeaf
}

func setLeafHasSeq(n *TreeNode, hasSeq uint16) {
	binary.LittleEndian.PutUint16(n.data[LEAF_HAS_SEQ_OFFSET:LEAF_HAS_SEQ_OFFSET+LEAF_HAS_SEQ_LEN], hasSeq)
}

func getLeafHasSeq(n *TreeNode) uint16 {
	return binary.LittleEndian.Uint16(n.data[LEAF_HAS_SEQ_OFFSET : LEAF_HAS_SEQ_OFFSET+LEAF_HAS_SEQ_LEN])
}

func (n *TreeNode) SetLeafSeqPointer(p uint64) {
	binary.LittleEndian.PutUint64(n.data[LEAF_SEQ_P_OFFSET:LEAF_SEQ_P_OFFSET+LEAF_SEQ_P_LEN], p)
}

func (n *TreeNode) GetLeafSeqPointer() uint64 {
	return binary.LittleEndian.Uint64(n.data[LEAF_SEQ_P_OFFSET : LEAF_SEQ_P_OFFSET+LEAF_SEQ_P_LEN])
}

func (n *TreeNode) SetLeafKeyLen(p uint64) {
	binary.LittleEndian.PutUint64(n.data[LEAF_KEY_LEN_OFFSET:LEAF_KEY_LEN_OFFSET+LEAF_KEY_LEN_LEN], p)
}

func (n *TreeNode) GetLeafKeyLen() uint16 {
	return binary.LittleEndian.Uint16(n.data[LEAF_KEY_LEN_OFFSET : LEAF_KEY_LEN_OFFSET+LEAF_KEY_LEN_LEN])
}

func (n *TreeNode) SetLeafValueLen(p uint64) {
	binary.LittleEndian.PutUint64(n.data[LEAF_VAL_LEN_OFFSET:LEAF_VAL_LEN_OFFSET+LEAF_VAL_LEN_LEN], p)
}

func (n *TreeNode) GetLeafValueLen() uint64 {
	return binary.LittleEndian.Uint64(n.data[LEAF_VAL_LEN_OFFSET : LEAF_VAL_LEN_OFFSET+LEAF_VAL_LEN_LEN])
}


