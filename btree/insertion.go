package btree

import (
	"bytes"
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
