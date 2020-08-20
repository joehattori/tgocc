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
		params    []*LVar
		body      []Node
		lvars     []*LVar
		name      string
		stackSize int
		ty        Type
	}

	// VarNode represents a node of variable
	VarNode struct {
		v Var
	}

	// IfNode represents a if statement node
	IfNode struct {
		cond Node
		then Node
		els  Node
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
	l := lhs.loadType()
	r := rhs.loadType()
	switch l.(type) {
	case *TyInt, *TyChar:
		switch r.(type) {
		case *TyInt:
			return &ArithNode{op: ndAdd, lhs: lhs, rhs: rhs}
		case *TyPtr, *TyArr:
			return &ArithNode{op: ndPtrAdd, lhs: rhs, rhs: lhs}
		}
	case *TyPtr, *TyArr:
		switch r.(type) {
		case *TyInt, *TyChar:
			return &ArithNode{op: ndPtrAdd, lhs: lhs, rhs: rhs}
		}
	}
	panic(fmt.Sprintf("Unexpected type: lhs: %T %T, rhs: %T %T", lhs, l, rhs, r))
}

// NewAddrNode builds a AddrNode
func NewAddrNode(v AddressableNode) Node {
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

// NewExprNode builds a DerefNode
func NewExprNode(body Node) Node {
	return &ExprNode{body}
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

// NewVarNode builds a new LVarNode instance
func NewVarNode(v Var) Node {
	return &VarNode{v}
}

// NewNumNode builds NumNode
func NewNumNode(val int) Node {
	return &NumNode{val: val}
}

// NewSubNode builds a node for subtraction
func NewSubNode(lhs Node, rhs Node) Node {
	l := lhs.loadType()
	r := rhs.loadType()
	switch l.(type) {
	case *TyInt:
		switch r.(type) {
		case *TyInt:
			return &ArithNode{op: ndSub, lhs: lhs, rhs: rhs}
		case *TyPtr, *TyArr:
			return &ArithNode{op: ndPtrSub, lhs: rhs, rhs: lhs}
		}
	case *TyPtr, *TyArr:
		switch r.(type) {
		case *TyInt:
			return &ArithNode{op: ndPtrSub, lhs: lhs, rhs: rhs}
		}
	}
	panic(fmt.Sprintf("Unexpected type: lhs: %T, rhs: %T", l, r))
}

func (t *Tokenized) searchVar(varName string) Var {
	for _, lv := range t.curFn.lvars {
		if lv.name == varName {
			return lv
		}
	}
	for _, g := range t.res.gvars {
		if g.name == varName {
			return g
		}
	}
	return nil
}

// FindVar searches Var named s
func (t *Tokenized) FindVar(s string) Var {
	v := t.searchVar(s)
	if v == nil {
		panic(fmt.Sprintf("undefined variable %s", s))
	}
	return v
}

// BuildLVarNode builds LVarNode
func (t *Tokenized) BuildLVarNode(s string, ty Type, isArg bool) Node {
	if _, isLVar := t.searchVar(s).(*LVar); isLVar {
		panic(fmt.Sprintf("variable %s is already defined", s))
	}
	f := t.curFn
	offset := f.stackSize + ty.size()
	arg := &LVar{name: s, ty: ty, offset: offset}
	f.lvars = append(f.lvars, arg)
	if isArg {
		f.params = append(f.params, arg)
	}
	// TODO: align
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
