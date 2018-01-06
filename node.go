package wordfilter

import (
	"github.com/streamrail/concurrent-map"
)

type Node struct {
	value        string
	exit         bool
	exitRefCount uint32
	childNodes   cmap.ConcurrentMap
}

func NewNode(word string) *Node {
	return &Node{
		value:        word,
		exit:         false,
		exitRefCount: 0,
		childNodes:   cmap.New(),
	}
}

func (node *Node) IsLeafNode() bool {
	l := node.childNodes.IsEmpty()
	return l
}

func (node *Node) Clear() {
	if node.IsLeafNode() {
		return
	}
	for item := range node.childNodes.Iter() {
		value, ok := item.Val.(*Node)
		if !ok {
			continue
		}
		value.Clear()
	}
	node.childNodes = cmap.New()
}

func (node *Node) Exit() bool {
	return node.exitRefCount > 0
}

func (node *Node) IncExitRefCount() {
	node.exitRefCount++
}

func (node *Node) DecExitRefCount() {
	if node.exitRefCount > 0 {
		node.exitRefCount--
	}
}

func (node *Node) GetChildNode(word string) (*Node, bool) {
	cdnode, ok := node.childNodes.Get(word)
	if ok {
		cnode, cok := cdnode.(*Node)
		if cok {
			return cnode, true
		}
	}
	return nil, false
}

func (node *Node) RemoveNode(word string) bool {
	node.childNodes.Remove(word)
	return true
}

func (node *Node) InsertNode(word string) *Node {
	if _, ok := node.GetChildNode(word); !ok {
		cnode := NewNode(word)
		node.childNodes.Set(word, cnode)
		return cnode
	}
	return nil
}
