package tree

import (
	"strings"
)

// Handler 处理函数接口
type Handler interface {
	Handle(interface{})
}

// Node 路由树节点
type Node struct {
	Pattern  string           // 路由模式
	Part     string           // 路由部分
	Children map[string]*Node // 子节点
	IsWild   bool             // 是否通配符
	Handlers []Handler        // 处理函数
}

// NewNode 创建新节点
func NewNode() *Node {
	return &Node{
		Children: make(map[string]*Node),
	}
}

// Insert 插入路由
func (n *Node) Insert(pattern string, parts []string, height int, handlers []Handler) {
	if len(parts) == height {
		n.Pattern = pattern
		n.Handlers = handlers
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &Node{
			Part:     part,
			Children: make(map[string]*Node),
			IsWild:   part[0] == ':' || part[0] == '*',
		}
		n.Children[part] = child
	}
	child.Insert(pattern, parts, height+1, handlers)
}

// Search 搜索路由
func (n *Node) Search(parts []string, height int) *Node {
	if len(parts) == height || strings.HasPrefix(n.Part, "*") {
		if n.Pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		result := child.Search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}

// matchChild 匹配子节点
func (n *Node) matchChild(part string) *Node {
	if child, ok := n.Children[part]; ok {
		return child
	}
	for _, child := range n.Children {
		if child.IsWild {
			return child
		}
	}
	return nil
}

// matchChildren 匹配所有子节点
func (n *Node) matchChildren(part string) []*Node {
	nodes := make([]*Node, 0)
	for _, child := range n.Children {
		if child.Part == part || child.IsWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// GetParams 获取路由参数
func (n *Node) GetParams(pattern string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(pattern, "/")
	nPattern := n.Pattern
	if nPattern == "" {
		return params
	}
	searchParts := strings.Split(nPattern, "/")

	for index, part := range searchParts {
		if len(part) == 0 || index >= len(parts) {
			continue
		}
		if part[0] == ':' {
			params[part[1:]] = parts[index]
		}
		if part[0] == '*' && len(part) > 1 {
			params[part[1:]] = strings.Join(parts[index:], "/")
			break
		}
	}
	return params
}
