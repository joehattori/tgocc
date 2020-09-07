package ast

import (
	"log"

	"github.com/joehattori/tgocc/types"
	"github.com/joehattori/tgocc/vars"
)

type (
	// Node represents ast Node
	Node interface {
		gen()
		LoadType() types.Type
	}

	AddressableNode interface {
		Node
		genAddr()
	}

	AddrNode struct {
		Var AddressableNode
		ty  types.Type
	}

	AssignNode struct {
		lhs AddressableNode
		rhs Node
		ty  types.Type
	}

	BinaryNode struct {
		op  nodeKind
		lhs Node
		rhs Node
		ty  types.Type
	}

	BitNotNode struct {
		body Node
	}

	BlkNode struct {
		Body []Node
	}

	BreakNode struct{}

	CaseNode struct {
		cmp  int
		body []Node
		idx  int
	}

	CastNode struct {
		base Node
		toTy types.Type
	}

	ContinueNode struct{}

	DecNode struct {
		body  AddressableNode
		isPre bool
	}

	DerefNode struct {
		ptr Node
		ty  types.Type
	}

	DoWhileNode struct {
		cond Node
		then Node
	}

	ExprNode struct {
		Body Node
	}

	ForNode struct {
		init Node
		cond Node
		inc  Node
		body Node
	}

	FnCallNode struct {
		name   string
		params []Node
		retTy  types.Type
	}

	FnNode struct {
		isStatic  bool
		Params    []*vars.LVar
		Body      []Node
		LVars     []*vars.LVar
		name      string
		StackSize int
		RetTy     types.Type
	}

	IfNode struct {
		cond Node
		then Node
		els  Node
	}

	IncNode struct {
		body  AddressableNode
		isPre bool
	}

	MemberNode struct {
		lhs AddressableNode
		mem *types.Member
	}

	NotNode struct {
		body Node
	}

	NullNode struct{}

	NumNode struct {
		val int64
	}

	RetNode struct {
		rhs    Node
		fnName string
		ty     types.Type
	}

	StmtExprNode struct {
		body []Node
		ty   types.Type
	}

	SwitchNode struct {
		target Node
		cases  []*CaseNode
		dflt   *CaseNode
	}

	TernaryNode struct {
		cond Node
		lhs  Node
		rhs  Node
	}

	VarNode struct {
		Var vars.Var
	}

	WhileNode struct {
		cond Node
		then Node
	}
)

type nodeKind int

const (
	NdAdd nodeKind = iota
	NdSub
	NdMul
	NdDiv
	NdNum
	NdEq
	NdNeq
	NdLt
	NdLeq
	NdGt
	NdGeq
	NdPtrAdd
	NdPtrSub
	NdPtrDiff
	NdAddEq
	NdSubEq
	NdMulEq
	NdDivEq
	NdPtrAddEq
	NdPtrSubEq
	NdBitOr
	NdBitXor
	NdBitAnd
	NdLogOr
	NdLogAnd
	NdShl
	NdShr
	NdShlEq
	NdShrEq
)

func NewAddNode(lhs Node, rhs Node) *BinaryNode {
	l := lhs.LoadType()
	r := rhs.LoadType()
	switch l.(type) {
	case *types.Char, *types.Int, *types.Short, *types.Long, *types.Bool:
		switch r.(type) {
		case *types.Char, *types.Int, *types.Short, *types.Long, *types.Bool:
			return &BinaryNode{op: NdAdd, lhs: lhs, rhs: rhs}
		case *types.Ptr, *types.Arr:
			return &BinaryNode{op: NdPtrAdd, lhs: rhs, rhs: lhs}
		}
	case *types.Ptr, *types.Arr:
		switch r.(type) {
		case *types.Char, *types.Int, *types.Short, *types.Long, *types.Bool:
			return &BinaryNode{op: NdPtrAdd, lhs: lhs, rhs: rhs}
		}
	}
	log.Fatalf("Unexpected types.Typepe for addition: lhs: %T %T, rhs: %T %T", lhs, l, rhs, r)
	return nil
}

func NewAddrNode(v AddressableNode) *AddrNode {
	return &AddrNode{Var: v}
}

func NewAssignNode(lhs AddressableNode, rhs Node) *AssignNode {
	return &AssignNode{lhs: lhs, rhs: rhs}
}

func NewBinaryNode(op nodeKind, lhs Node, rhs Node) *BinaryNode {
	return &BinaryNode{op: op, lhs: lhs, rhs: rhs}
}

func NewBitNotNode(body Node) *BitNotNode {
	return &BitNotNode{body}
}

func NewBlkNode(body []Node) *BlkNode {
	return &BlkNode{body}
}

func NewBreakNode() *BreakNode {
	return &BreakNode{}
}

func NewCaseNode(cmp int, body []Node, idx int) *CaseNode {
	return &CaseNode{cmp, body, idx}
}

func NewCastNode(base Node, t types.Type) *CastNode {
	return &CastNode{base, t}
}

func NewContinueNode() *ContinueNode {
	return &ContinueNode{}
}

func NewDecNode(body AddressableNode, isPre bool) *DecNode {
	return &DecNode{body, isPre}
}

func NewDerefNode(ptr Node) *DerefNode {
	return &DerefNode{ptr: ptr}
}

func NewDoWhileNode(cond Node, then Node) *DoWhileNode {
	return &DoWhileNode{cond, then}
}

func NewExprNode(body Node) *ExprNode {
	return &ExprNode{body}
}

func NewForNode(init Node, cond Node, inc Node, body Node) *ForNode {
	return &ForNode{init, cond, inc, body}
}

func NewFnCallNode(name string, params []Node, retTy types.Type) *FnCallNode {
	return &FnCallNode{name, params, retTy}
}

func NewFnNode(isStatic bool, name string, t types.Type) *FnNode {
	return &FnNode{isStatic: isStatic, name: name, RetTy: t}
}

func NewIfNode(cond Node, then Node, els Node) *IfNode {
	return &IfNode{cond, then, els}
}

func NewIncNode(body AddressableNode, isPre bool) *IncNode {
	return &IncNode{body, isPre}
}

func NewMemberNode(lhs AddressableNode, m *types.Member) *MemberNode {
	return &MemberNode{lhs, m}
}

func NewNotNode(body Node) *NotNode {
	return &NotNode{body}
}

func NewNullNode() *NullNode {
	return &NullNode{}
}

func NewNumNode(val int64) *NumNode {
	return &NumNode{val}
}

func NewRetNode(rhs Node, fnName string) *RetNode {
	return &RetNode{rhs: rhs, fnName: fnName}
}

func NewStmtExprNode(body []Node) *StmtExprNode {
	return &StmtExprNode{body: body}
}

func NewSubNode(lhs Node, rhs Node) *BinaryNode {
	l := lhs.LoadType()
	r := rhs.LoadType()
	switch l.(type) {
	case *types.Char, *types.Int, *types.Long, *types.Short, *types.Bool:
		switch r.(type) {
		case *types.Char, *types.Int, *types.Long, *types.Short, *types.Bool:
			return &BinaryNode{op: NdSub, lhs: lhs, rhs: rhs}
		}
	case *types.Ptr, *types.Arr:
		switch r.(type) {
		case *types.Char, *types.Int, *types.Long, *types.Short, *types.Bool:
			return &BinaryNode{op: NdPtrSub, lhs: lhs, rhs: rhs}
		case *types.Ptr, *types.Arr:
			return &BinaryNode{op: NdPtrDiff, lhs: lhs, rhs: rhs}
		}
	}
	log.Fatalf("Unexpected types.Typepe for subtraction: lhs: %T, rhs: %T", l, r)
	return nil
}

func NewSwitchNode(target Node, cases []*CaseNode, dflt *CaseNode) *SwitchNode {
	return &SwitchNode{target, cases, dflt}
}

func NewTernaryNode(cond Node, lhs Node, rhs Node) *TernaryNode {
	return &TernaryNode{cond, lhs, rhs}
}

func NewVarNode(v vars.Var) *VarNode {
	return &VarNode{v}
}

func NewWhileNode(cond Node, then Node) *WhileNode {
	return &WhileNode{cond: cond, then: then}
}

func (a *AddrNode) LoadType() types.Type {
	t := a.Var.LoadType()
	if arrT, ok := t.(*types.Arr); ok {
		a.ty = types.NewPtr(arrT.Base())
	} else {
		a.ty = types.NewPtr(t)
	}
	return a.ty
}

func (a *AssignNode) LoadType() types.Type {
	if a.ty == nil {
		a.ty = a.lhs.LoadType()
	}
	a.rhs.LoadType()
	return a.ty
}

func (a *BinaryNode) LoadType() types.Type {
	if a.ty == nil {
		a.ty = a.lhs.LoadType()
	}
	a.rhs.LoadType()
	return a.ty
}

func (b *BitNotNode) LoadType() types.Type {
	return b.body.LoadType()
}

func (b *BlkNode) LoadType() types.Type {
	for _, st := range b.Body {
		st.LoadType()
	}
	return types.NewEmpty()
}

func (b *BreakNode) LoadType() types.Type {
	return types.NewEmpty()
}

func (c *CaseNode) LoadType() types.Type {
	return types.NewEmpty()
}

func (c *CastNode) LoadType() types.Type {
	return c.toTy
}

func (c *ContinueNode) LoadType() types.Type {
	return types.NewEmpty()
}

func (d *DecNode) LoadType() types.Type {
	return d.body.LoadType()
}

func (d *DerefNode) LoadType() types.Type {
	switch v := d.ptr.LoadType().(type) {
	case *types.Char, *types.Int, *types.Short, *types.Long:
		d.ty = v
	case *types.Ptr:
		d.ty = v.Base()
	case *types.Arr:
		d.ty = v.Base()
	default:
		log.Fatalf("Cannot dereference types.Typepe %T", d.ptr.LoadType())
	}
	return d.ty
}

func (d *DoWhileNode) LoadType() types.Type {
	return types.NewEmpty()
}

func (e *ExprNode) LoadType() types.Type {
	e.Body.LoadType()
	return types.NewEmpty()
}

func (f *ForNode) LoadType() types.Type {
	if f.init != nil {
		f.init.LoadType()
	}
	if f.cond != nil {
		f.cond.LoadType()
	}
	if f.body != nil {
		f.body.LoadType()
	}
	if f.inc != nil {
		f.inc.LoadType()
	}
	return types.NewEmpty()
}

func (f *FnCallNode) LoadType() types.Type {
	for _, param := range f.params {
		param.LoadType()
	}
	return f.retTy
}

func (f *FnNode) LoadType() types.Type {
	for _, Node := range f.Body {
		Node.LoadType()
	}
	return f.RetTy
}

func (i *IfNode) LoadType() types.Type {
	i.cond.LoadType()
	i.then.LoadType()
	if i.els != nil {
		i.els.LoadType()
	}
	return types.NewEmpty()
}

func (i *IncNode) LoadType() types.Type {
	return i.body.LoadType()
}

func (m *MemberNode) LoadType() types.Type {
	return m.mem.Type
}

func (n *NotNode) LoadType() types.Type {
	return types.NewBool()
}

func (*NullNode) LoadType() types.Type {
	return types.NewEmpty()
}

func (n *NumNode) LoadType() types.Type {
	return types.NewLong()
}

func (r *RetNode) LoadType() types.Type {
	if r.rhs == nil {
		return types.NewEmpty()
	}
	if r.ty == nil {
		r.ty = r.rhs.LoadType()
	}
	return r.ty
}

func (s *StmtExprNode) LoadType() types.Type {
	if s.ty == nil {
		s.ty = s.body[len(s.body)-1].LoadType()
	}
	return s.ty
}

func (s *SwitchNode) LoadType() types.Type {
	return types.NewEmpty()
}

func (t *TernaryNode) LoadType() types.Type {
	return t.lhs.LoadType()
}

func (v *VarNode) LoadType() types.Type {
	return v.Var.Type()
}

func (w *WhileNode) LoadType() types.Type {
	w.cond.LoadType()
	w.then.LoadType()
	return types.NewEmpty()
}

func Eval(n Node) int64 {
	switch n := n.(type) {
	case *BinaryNode:
		switch n.op {
		case NdAdd:
			return Eval(n.lhs) + Eval(n.rhs)
		case NdSub:
			return Eval(n.lhs) - Eval(n.rhs)
		case NdMul:
			return Eval(n.lhs) * Eval(n.rhs)
		case NdDiv:
			return Eval(n.lhs) / Eval(n.rhs)
		case NdBitOr:
			return Eval(n.lhs) | Eval(n.rhs)
		case NdBitXor:
			return Eval(n.lhs) ^ Eval(n.rhs)
		case NdBitAnd:
			return Eval(n.lhs) & Eval(n.rhs)
		case NdShl:
			return Eval(n.lhs) << Eval(n.rhs)
		case NdShr:
			return Eval(n.lhs) >> Eval(n.rhs)
		case NdEq:
			if Eval(n.lhs) == Eval(n.rhs) {
				return 1
			}
			return 0
		case NdNeq:
			if Eval(n.lhs) != Eval(n.rhs) {
				return 1
			}
			return 0
		case NdLt:
			if Eval(n.lhs) < Eval(n.rhs) {
				return 1
			}
			return 0
		case NdLeq:
			if Eval(n.lhs) <= Eval(n.rhs) {
				return 1
			}
			return 0
		case NdGt:
			if Eval(n.lhs) > Eval(n.rhs) {
				return 1
			}
			return 0
		case NdGeq:
			if Eval(n.lhs) <= Eval(n.rhs) {
				return 1
			}
			return 0
		case NdLogAnd:
			return Eval(n.lhs) & Eval(n.rhs)
		case NdLogOr:
			return Eval(n.lhs) | Eval(n.rhs)
		}
	case *BitNotNode:
		return ^Eval(n.body)
	case *NotNode:
		if Eval(n.body) != 0 {
			return 1
		}
		return 0
	case *NumNode:
		return n.val
	case *TernaryNode:
		if Eval(n.cond) == 0 {
			return Eval(n.rhs)
		}
		return Eval(n.lhs)
	}
	log.Fatalf("Not a constant expression.")
	return 0
}
