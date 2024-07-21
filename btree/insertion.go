package btree

import (
	"bytes"
	"sort"
)

func nodeLookupLE(node BNode, key []byte) uint16 {
	nkeys := node.nkeys()
	found := uint16(0)

	for i:= uint16(1); i <= nkeys; i++ {
		if bytes.Compare(key, node.getKey(i)) >= 0 {
			found = i
		}
	}
	sort.Search(nkeys, func(i int) bool {
}