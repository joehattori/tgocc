package main

type nodeKind int

type (
	// Node represents each node in ast
	Node interface {
		kind() nodeKind
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

	// FuncDefNode represents a node of function definition
	FuncDefNode struct {
		name      string
		args      []*LVar
		body      []Node
		stackSize int
	}

	// LvarNode represents a node of local variable
	LvarNode struct {
		offset int
		name   string
	}

	// NumNode represents number node
	NumNode struct {
		val int
	}

	// RetNode represents a return node
	RetNode struct {
		rhs Node
	}

	// IfNode represents a if statement node
	IfNode struct {
		cond Node
		then Node
		els  Node
	}

	// WhileNode represents a node of while statement
	WhileNode struct {
		cond Node
		then Node
	}
)

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
	ndLvar
	ndAssign
	ndRet
	ndIf
	ndElse
	ndWhile
	ndFor
	ndBlk
	ndFuncCall
	ndFuncDef
)

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

// NewForNode builds a ForNode
func NewForNode(init Node, cond Node, inc Node, body Node) Node {
	return &ForNode{init: init, cond: cond, inc: inc, body: body}
}

// NewFuncCallNode builds a FuncCallNode
func NewFuncCallNode(name string, args []Node) Node {
	return &FuncCallNode{name: name, args: args}
}

// NewFuncDefNode builds a FuncDefNode
func NewFuncDefNode(name string, args []*LVar, body []Node, stackSize int) Node {
	return &FuncDefNode{name: name, args: args, body: body, stackSize: stackSize}
}

// NewIfNode builds a IfNode
func NewIfNode(cond Node, then Node, els Node) Node {
	return &IfNode{cond, then, els}
}

// NewNumNode builds NumNode
func NewNumNode(val int) Node {
	return &NumNode{val}
}

func findLVar(name string) *LVar {
	for _, v := range LVars {
		if v.name == name {
			return v
		}
	}
	return nil
}

// NewLvarNode builds arNode
func NewLvarNode(s string) Node {
	node := &LvarNode{}
	if v := findLVar(s); v != nil {
		node.offset = v.offset
	} else {
		offset := 8 * (len(LVars) + 1)
		node.offset = offset
		LVars = append(LVars, &LVar{offset: offset, name: s})
	}
	return node
}

// NewRetNode builds RetNode
func NewRetNode(rhs Node) Node {
	return &RetNode{rhs: rhs}
}

// NewWhileNode builds a WhileNode
func NewWhileNode(cond Node, then Node) Node {
	return &WhileNode{cond: cond, then: then}
}

func (a *ArithNode) kind() nodeKind {
	return a.op
}

func (*AssignNode) kind() nodeKind {
	return ndAssign
}

func (*BlkNode) kind() nodeKind {
	return ndBlk
}

func (*ForNode) kind() nodeKind {
	return ndFor
}

func (*FuncCallNode) kind() nodeKind {
	return ndFuncCall
}

func (*FuncDefNode) kind() nodeKind {
	return ndFuncDef
}

func (*LvarNode) kind() nodeKind {
	return ndLvar
}

func (*NumNode) kind() nodeKind {
	return ndNum
}

func (*RetNode) kind() nodeKind {
	return ndRet
}

func (*IfNode) kind() nodeKind {
	return ndIf
}

func (*WhileNode) kind() nodeKind {
	return ndWhile
}
