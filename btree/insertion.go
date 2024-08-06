package btree

import (
	"bytes"
	"encoding/binary"
)

// TODO: use binary search here
func nodeLookupLE(node BNode, key []byte) uint16 {
	nkeys := node.nkeys()
	found := uint16(0)

	// this code seems to suggests we will always find the key in the list

	for i := uint16(1); i < nkeys; i++ {
		cmp := bytes.Compare(node.getKey(i), key)
		if cmp <= 0 {
			found = i
		} else {
			break
		}

	}
	return found
}

func leafInsert(new BNode, old BNode, idx uint16, key []byte, value []byte) {
	new.setHeader(BNODE_LEAF, old.nkeys()+1)
	nodeAppendRange(new, old, 0, 0, idx)
	nodeAppendKV(new, idx, 0, key, value)
	nodeAppendRange(new, old, idx+1, idx, old.nkeys()-idx)
}

func leafupdate(new BNode, old BNode, idx uint16, key []byte, value []byte) {
	new.setHeader(BNODE_LEAF, old.nkeys())
	nodeAppendRange(new, old, 0, 0, idx)
	nodeAppendKV(new, idx, 0, key, value)
	nodeAppendRange(new, old, idx+1, idx+1, old.nkeys()-(idx+1))
}

//copy multiple key value into position

func nodeAppendRange(new BNode, old BNode, dstNew uint16, srcOld uint16, n uint16) {
	assert(srcOld+n <= old.nkeys())
	assert(dstNew+n <= new.nkeys())
	if n == 0 {
		return
	}

	// copy the pointers

	for i := uint16(0); i < n; i++ {
		new.setPtr(dstNew+i, new.getPtr(srcOld+i))
	}

	// copy the offsets
	dstBegin := new.getOffset(dstNew)
	srcBegin := new.getOffset(srcOld)

	for i := uint16(1); i <= n; i++ {
		offset := dstBegin + old.getOffset(srcOld+i) - srcBegin
		new.setOffset(dstNew+i, offset)
	}

	// KVs
	begin := old.kvPos(srcOld)
	end := old.kvPos(srcBegin + n)
	copy(new.data[new.kvPos(dstNew):], old.data[begin:end])

}

func nodeAppendKV(new BNode, idx uint16, ptr uint64, key []byte, value []byte) {
	//pointers
	new.setPtr(idx, ptr)
	//kvs
	pos := new.kvPos(idx)

	binary.LittleEndian.PutUint16(new.data[pos:], uint16(len(key)))
	binary.LittleEndian.PutUint16(new.data[pos+2:], uint16(len(value)))
	copy(new.data[pos+4:], key)
	copy(new.data[pos+4+uint16(len(key)):], value)

	// the offset of the next key
	new.setOffset(idx+1, new.getOffset(idx)+4+uint16(len(key)+len(value)))
}

func treeInsert(tree *BTree, node BNode, key []byte, value []byte) BNode {
	new := BNode{data: make([]byte, 2*BTREE_PAGE_SIZE)}
	//where do we insert the inputed key?
	idx := nodeLookupLE(node, key)

	switch node.btype() {
	case BNODE_LEAF:
		if bytes.Equal(key, node.getKey(idx)) {
			leafupdate(new, node, idx, key, value)
		} else {
			leafInsert(new, node, idx+1, key, value)
		}
	case BNODE_NODE:
		nodeInsert(tree, new, node, idx, key, value)
	}
	return new
}

func nodeInsert(tree *BTree, new BNode, node BNode, idx uint16, key []byte, value []byte) {
	kptr := node.getPtr(idx)
	knode := tree.get(kptr)
	tree.del(kptr)
	// recursive insertion to the kid node
	knode = treeInsert(tree, knode, key, value)
	// split the result
	nsplit, splited := nodeSplit3(knode)
	// update the kid links
	nodeReplaceKidN(tree, new, node, idx, splited[:nsplit]...)
}

func nodeSplit2(left, right, old BNode) {
	halfSize := old.nbytes() / 2

	left.setHeader(old.btype(), 1)
	nodeAppendRange(left, old, 0, 0, 1)

	i := uint16(1)
	for {
		// 8 ptr; 2 offset; 4 keylen vallen
		nextleftSize := left.nbytes() + 8 + 2 + 4 + uint16(len(old.getKey(i))+len(old.getVal(i)))

		if len(left.data) == 2*BTREE_PAGE_SIZE {
			if left.nbytes() > halfSize {
				break
			}
		} else {
			if nextleftSize > halfSize {
				break
			}
		}

		nodeAppendRange(left, old, i, i, 1)
		i = i + 1
	}

	nodeAppendRange(right, old, 0, i, old.nkeys()-i)
}

func nodeSplit3(old BNode) (uint16, [3]BNode) {
	if old.nbytes() <= BTREE_PAGE_SIZE {
		old.data = old.data[:BTREE_PAGE_SIZE]
		return 1, [3]BNode{old}
	}
	left := BNode{make([]byte, 2*BTREE_PAGE_SIZE)} // might be split later
	right := BNode{make([]byte, BTREE_PAGE_SIZE)}
	nodeSplit2(left, right, old)
	if left.nbytes() <= BTREE_PAGE_SIZE {
		left.data = left.data[:BTREE_PAGE_SIZE]
		return 2, [3]BNode{left, right}
	}
	// the left node is still too large
	leftleft := BNode{make([]byte, BTREE_PAGE_SIZE)}
	middle := BNode{make([]byte, BTREE_PAGE_SIZE)}
	nodeSplit2(leftleft, middle, left)
	assert(leftleft.nbytes() <= BTREE_PAGE_SIZE)
	return 3, [3]BNode{leftleft, middle, right}
}

func nodeReplaceKidN(tree *BTree, new, old BNode, idx uint16, kids ...BNode) {
	inc := uint16(len(kids))
	new.setHeader(BNODE_NODE, old.nkeys()+inc-1)
	nodeAppendRange(new, old, 0, 0, idx)
	for i, node := range kids {
		nodeAppendKV(new, idx+uint16(i), tree.new(node), node.getKey(0), nil)
	}
	nodeAppendRange(new, old, idx+inc, idx+1, old.nkeys()-(idx+1))
}
