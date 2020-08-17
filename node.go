package main

var labelCount int

var argRegs = [...]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

type (
	// Node represents each node in ast
	Node interface {
		gen()
	}

	// AddrNode represents a node of address
	AddrNode struct {
		v *LVarNode
	}

	// ArithNode represents a node of arithmetic calculation
	ArithNode struct {
		op  nodeKind
		lhs Node
		rhs Node
	}

	// AssignNode represents assignment node
	AssignNode struct {
		lhs Node
		rhs Node
	}

	// BlkNode represents a node of block
	BlkNode struct {
		body []Node
	}

	// DerefNode represents a reference node of pointer
	DerefNode struct {
		pt Node
	}

	// ForNode represents a node of for statement
	ForNode struct {
		init Node
		cond Node
		inc  Node
		body Node
	}

	// FuncCallNode represents a node of function call
	FuncCallNode struct {
		name string
		args []Node
	}

	// FnNode represents a node of function definition
	FnNode struct {
		name  string
		args  []*LVar
		body  []Node
		lvars []*LVar
	}

	// IfNode represents a if statement node
	IfNode struct {
		cond Node
		then Node
		els  Node
	}

	// LVarNode represents a node of local variable
	LVarNode struct {
		offset int
		name   string
	}

	// NumNode represents number node
	NumNode struct {
		val int
	}

	// RetNode represents a return node
	RetNode struct {
		rhs    Node
		fnName string
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
)

// NewAddrNode builds a AddrNode
func NewAddrNode(v *LVarNode) Node {
	return &AddrNode{v: v}
}

// NewArithNode builds a ArithNode
func NewArithNode(op nodeKind, lhs Node, rhs Node) Node {
	return &ArithNode{op: op, lhs: lhs, rhs: rhs}
}

// NewAssignNode builds AssignNode
func NewAssignNode(lhs Node, rhs Node) Node {
	return &AssignNode{lhs, rhs}
}

// NewBlkNode builds a BlkNode
func NewBlkNode(body []Node) Node {
	return &BlkNode{body}
}

// NewDerefNode builds a DerefNode
func NewDerefNode(pt Node) Node {
	return &DerefNode{pt: pt}
}

// NewForNode builds a ForNode
func NewForNode(init Node, cond Node, inc Node, body Node) Node {
	return &ForNode{init: init, cond: cond, inc: inc, body: body}
}

// NewFuncCallNode builds a FuncCallNode
func NewFuncCallNode(name string, args []Node) Node {
	return &FuncCallNode{name: name, args: args}
}

// NewIfNode builds a IfNode
func NewIfNode(cond Node, then Node, els Node) Node {
	return &IfNode{cond, then, els}
}

// NewNumNode builds NumNode
func NewNumNode(val int) Node {
	return &NumNode{val}
}

func (f *FnNode) findLVar(varName string) *LVar {
	for _, v := range f.lvars {
		if v.name == varName {
			return v
		}
	}
	return nil
}

// NewLVarNode builds arNode
func (f *FnNode) NewLVarNode(s string) Node {
	node := &LVarNode{}
	if v := f.findLVar(s); v != nil {
		node.offset = v.offset
	} else {
		offset := 8 * (len(f.lvars) + 1)
		node.offset = offset
		f.lvars = append(f.lvars, &LVar{offset: offset, name: s})
	}
	return node
}

// NewRetNode builds RetNode
func NewRetNode(rhs Node, fnName string) Node {
	return &RetNode{rhs: rhs, fnName: fnName}
}

// NewWhileNode builds a WhileNode
func NewWhileNode(cond Node, then Node) Node {
	return &WhileNode{cond: cond, then: then}
}
