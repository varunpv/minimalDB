package btree

type BTree struct {
	root uint64

	// func as first class member !!/??
	get func(uint64) BNode // dereference a pointer to a page
	new func(BNode) uint64 // allocate a new page
	del func(uint64)       // deallocate page
}

const HEADER = 4
const BTREE_PAGE_SIZE = 4096
const BTREE_MAX_KEY_SIZE = 1000
const BTREE_MAX_VALUE_SIZE = 3000

func init() {
	node1max := HEADER + 8 + 2 + 4 + BTREE_MAX_KEY_SIZE + BTREE_MAX_KEY_SIZE + BTREE_MAX_VALUE_SIZE
	if node1max < BTREE_PAGE_SIZE {
		panic("node size is too large")
	}
}

func assert(b bool) {
	if !b {
		panic("assertion failed")
	}
}
