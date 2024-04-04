package trie

import "strings"

type Node struct {
	Pattern  string
	Part     string
	Children []*Node
	IsWhild  bool
}

func New() *Node {
	return &Node{}
}

func (n *Node) Insert(Pattern string, parts []string, height int) {
	if len(parts) == height {
		n.Pattern = Pattern
		return
	}

	part := parts[height]
	child := n.MatchChild(part)
	if child == nil {
		child = &Node{Part: part, IsWhild: part[0] == ':' || part[0] == '*'}
		n.Children = append(n.Children, child)
	}
	child.Insert(Pattern, parts, height+1)
}

func (n *Node) MatchChild(part string) *Node {
	for _, child := range n.Children {
		if child.Part == part || child.IsWhild {
			return child
		}
	}
	return nil
}

func (n *Node) MatchChildren(part string) []*Node {
	nodes := make([]*Node, 0)
	for _, child := range n.Children {
		if child.Part == part || child.IsWhild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *Node) Search(parts []string, height int) *Node {
	if len(parts) == height || strings.HasPrefix(n.Part, "*") {
		if n.Pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.MatchChildren(part)

	for _, child := range children {
		result := child.Search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}
