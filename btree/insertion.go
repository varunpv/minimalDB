package btree

import (
	"bytes"
	"encoding/binary"
)

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
