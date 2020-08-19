package main

import "fmt"

type (
	// Node represents each node in ast
	Node interface {
		gen()
		loadType() Type
	}

	// AddressableNode represents a node whose address can be calculated
	AddressableNode interface {
		Node
		genAddr()
	}

	// AddrNode represents a node in form of &x
	AddrNode struct {
		v  AddressableNode
		ty Type
	}

	// ArithNode represents a node of arithmetic calculation
	ArithNode struct {
		op  nodeKind
		lhs Node
		rhs Node
		ty  Type
	}

	// AssignNode represents assignment node
	AssignNode struct {
		lhs AddressableNode
		rhs Node
		ty  Type
	}

	// BlkNode represents a node of block
	BlkNode struct {
		body []Node
	}

	// DerefNode represents a reference node of pointer
	DerefNode struct {
		ptr Node
		ty  Type
	}

	// ExprNode represents a node of expression
	ExprNode struct {
		body Node
	}

	// ForNode represents a node of for statement
	ForNode struct {
		init Node
		cond Node
		inc  Node
		body Node
	}

	// FnCallNode represents a node of function call
	FnCallNode struct {
		params []Node
		name   string
		ty     Type
	}

	// FnNode represents a node of function definition
	FnNode struct {
		params    []*LVarNode
		body      []Node
		lvars     []*LVarNode
		name      string
		stackSize int
		ty        Type
	}

	// IfNode represents a if statement node
	IfNode struct {
		cond Node
		then Node
		els  Node
	}

	// LVarNode represents a node of local variable
	LVarNode struct {
		name   string
		offset int
		ty     Type
	}

	// NullNode is a node which doesn't emit assembly code
	NullNode struct{}

	// NumNode represents number node
	NumNode struct {
		val int
		ty  Type
	}

	// RetNode represents a return node
	RetNode struct {
		rhs    Node
		fnName string
		ty     Type
	}

	// WhileNode represents a node of while statement
	WhileNode struct {
		cond Node
		then Node
	}
)

type nodeKind int

const (
	ndAdd = iota
	ndSub
	ndMul
	ndDiv
	ndNum
	ndEq
	ndNeq
	ndLt
	ndLeq
	ndGt
	ndGeq
	ndPtrAdd
	ndPtrSub
)

// NewAddNode builds a node for addition
func NewAddNode(lhs Node, rhs Node) Node {
	switch lhs.loadType().(type) {
	case *TyInt:
		switch rhs.loadType().(type) {
		case *TyInt:
			return &ArithNode{op: ndAdd, lhs: lhs, rhs: rhs}
		case *TyPtr:
			return &ArithNode{op: ndPtrAdd, lhs: rhs, rhs: lhs}
		}
	case *TyPtr:
		switch rhs.loadType().(type) {
		case *TyInt, *TyPtr:
			return &ArithNode{op: ndPtrAdd, lhs: lhs, rhs: rhs}
		}
	}
	panic(fmt.Sprintf("Unexpected type: lhs: %T %T, rhs: %T %T", lhs, lhs.loadType(), rhs, rhs.loadType()))
}

// NewAddrNode builds a AddrNode
func NewAddrNode(v *LVarNode) Node {
	return &AddrNode{v: v}
}

// NewArithNode builds a ArithNode
func NewArithNode(op nodeKind, lhs Node, rhs Node) Node {
	return &ArithNode{op: op, lhs: lhs, rhs: rhs}
}

// NewAssignNode builds AssignNode
func NewAssignNode(lhs AddressableNode, rhs Node) Node {
	return &AssignNode{lhs: lhs, rhs: rhs}
}

// NewBlkNode builds a BlkNode
func NewBlkNode(body []Node) Node {
	return &BlkNode{body}
}

// NewDerefNode builds a DerefNode
func NewDerefNode(ptr Node) Node {
	return &DerefNode{ptr: ptr}
}

// NewForNode builds a ForNode
func NewForNode(init Node, cond Node, inc Node, body Node) Node {
	return &ForNode{init, cond, inc, body}
}

// NewFnCallNode builds a FuncCallNode
func NewFnCallNode(name string, params []Node) Node {
	// TODO: change type dynamically
	return &FnCallNode{name: name, params: params, ty: &TyInt{}}
}

// NewIfNode builds a IfNode
func NewIfNode(cond Node, then Node, els Node) Node {
	return &IfNode{cond, then, els}
}

// NewNumNode builds NumNode
func NewNumNode(val int) Node {
	return &NumNode{val: val}
}

// NewSubNode builds a node for subtraction
func NewSubNode(lhs Node, rhs Node) Node {
	switch lhs.loadType().(type) {
	case *TyInt:
		switch rhs.loadType().(type) {
		case *TyInt:
			return &ArithNode{op: ndSub, lhs: lhs, rhs: rhs}
		case *TyPtr:
			return &ArithNode{op: ndPtrSub, lhs: rhs, rhs: lhs}
		}
	case *TyPtr:
		switch rhs.loadType().(type) {
		case *TyInt, *TyPtr:
			return &ArithNode{op: ndPtrSub, lhs: lhs, rhs: rhs}
		}
	}
	panic(fmt.Sprintf("Unexpected type: lhs: %T, rhs: %T", lhs.loadType(), rhs.loadType()))
}

func (f *FnNode) searchLVarNode(varName string) *LVarNode {
	for _, v := range f.lvars {
		if v.name == varName {
			return v
		}
	}
	return nil
}

// FindLVarNode searches LVarNode named s
func (f *FnNode) FindLVarNode(s string) *LVarNode {
	v := f.searchLVarNode(s)
	if v == nil {
		panic(fmt.Sprintf("undefined variable %s", s))
	}
	return v
}

// BuildLVarNode builds LVarNode
func (f *FnNode) BuildLVarNode(s string, ty Type, isArg bool) Node {
	if f.searchLVarNode(s) != nil {
		panic(fmt.Sprintf("variable %s is already defined", s))
	}
	// TODO: think of changing the increment offset dynamically
	offset := f.stackSize + 8
	arg := &LVarNode{name: s, ty: ty, offset: offset}
	f.lvars = append(f.lvars, arg)
	if isArg {
		f.params = append(f.params, arg)
	}
	f.stackSize = offset
	return &NullNode{}
}

// NewRetNode builds RetNode
func NewRetNode(rhs Node, fnName string) Node {
	return &RetNode{rhs: rhs, fnName: fnName}
}

// NewWhileNode builds a WhileNode
func NewWhileNode(cond Node, then Node) Node {
	return &WhileNode{cond: cond, then: then}
}
