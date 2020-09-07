package ast

import (
	"fmt"
	"log"
	"math"

	"github.com/joehattori/tgocc/types"
	"github.com/joehattori/tgocc/vars"
)

var (
	labelCount  int
	jmpLabelNum int

	paramRegs1 = [...]string{"dil", "sil", "dl", "cl", "r8b", "r9b"}
	paramRegs2 = [...]string{"di", "si", "dx", "cx", "r8w", "r9w"}
	paramRegs4 = [...]string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
	paramRegs8 = [...]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
)

func (a *AddrNode) gen() {
	a.Var.genAddr()
}

func (a *AssignNode) gen() {
	a.lhs.genAddr()
	a.rhs.gen()
	store(a.LoadType())
}

func (b *BinaryNode) gen() {
	lhs, rhs := b.lhs, b.rhs
	switch b.op {
	case NdAddEq, NdSubEq, NdMulEq, NdDivEq, NdPtrAddEq, NdPtrSubEq, NdShlEq, NdShrEq:
		lhs.(AddressableNode).genAddr()
		defer store(lhs.LoadType())
	}

	lhs.gen()
	rhs.gen()

	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")

	switch b.op {
	case NdAdd, NdAddEq:
		fmt.Println("	add rax, rdi")
	case NdSub, NdSubEq:
		fmt.Println("	sub rax, rdi")
	case NdMul, NdMulEq:
		fmt.Println("	imul rax, rdi")
	case NdDiv, NdDivEq:
		fmt.Println("	cqo")
		fmt.Println("	idiv rdi")
	case NdEq:
		fmt.Println("	cmp rax, rdi")
		fmt.Println("	sete al")
		fmt.Println("	movzb rax, al")
	case NdNeq:
		fmt.Println("	cmp rax, rdi")
		fmt.Println("	setne al")
		fmt.Println("	movzb rax, al")
	case NdLt:
		fmt.Println("	cmp rax, rdi")
		fmt.Println("	setl al")
		fmt.Println("	movzb rax, al")
	case NdLeq:
		fmt.Println("	cmp rax, rdi")
		fmt.Println("	setle al")
		fmt.Println("	movzb rax, al")
	case NdGt:
		fmt.Println("	cmp rdi, rax")
		fmt.Println("	setl al")
		fmt.Println("	movzb rax, al")
	case NdGeq:
		fmt.Println("	cmp rdi, rax")
		fmt.Println("	setle al")
		fmt.Println("	movzb rax, al")
	case NdPtrAdd, NdPtrAddEq:
		fmt.Printf("	imul rdi, %d\n", b.LoadType().(types.Pointing).Base().Size())
		fmt.Printf("	add rax, rdi\n")
	case NdPtrSub, NdPtrSubEq:
		fmt.Printf("	imul rdi, %d\n", b.LoadType().(types.Pointing).Base().Size())
		fmt.Printf("	sub rax, rdi\n")
	case NdPtrDiff:
		fmt.Println("	sub rax, rdi")
		fmt.Println("	cqo")
		fmt.Printf("	mov rdi, %d\n", b.lhs.LoadType().(types.Pointing).Base().Size())
		fmt.Println("	idiv rdi")
	case NdBitOr:
		fmt.Println("	or rax, rdi")
	case NdBitXor:
		fmt.Println("	xor rax, rdi")
	case NdBitAnd:
		fmt.Println("	and rax, rdi")
	case NdLogOr:
		c := labelCount
		labelCount++
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	jne .L.true.%d\n", c)
		fmt.Println("	cmp rdi, 0")
		fmt.Printf("	jne .L.true.%d\n", c)
		fmt.Println("	setne al")
		fmt.Printf("	jmp .L.end.%d\n", c)
		fmt.Printf(".L.true.%d:\n", c)
		fmt.Println("	setne al")
		fmt.Printf(".L.end.%d:\n", c)
	case NdLogAnd:
		c := labelCount
		labelCount++
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .L.false.%d\n", c)
		fmt.Println("	cmp rdi, 0")
		fmt.Printf("	je .L.false.%d\n", c)
		fmt.Println("	setne al")
		fmt.Printf("	jmp .L.end.%d\n", c)
		fmt.Printf(".L.false.%d:\n", c)
		fmt.Println("	setne al")
		fmt.Printf(".L.end.%d:\n", c)
	case NdShl, NdShlEq:
		fmt.Println("	mov cl, dil")
		fmt.Println("	sal rax, cl")
	case NdShr, NdShrEq:
		fmt.Println("	mov cl, dil")
		fmt.Println("	sar rax, cl")
	default:
		log.Fatal("Unhandled node kind")
	}

	fmt.Println("	push rax")
}

func (b *BitNotNode) gen() {
	b.body.gen()
	fmt.Println("	pop rax")
	fmt.Println("	not rax")
	fmt.Println("	push rax")
}

func (b *BlkNode) gen() {
	for _, st := range b.Body {
		st.gen()
	}
}

func (b *BreakNode) gen() {
	if jmpLabelNum == 0 {
		log.Fatal("Invalid break statement.")
	}
	fmt.Printf("	jmp .L.break.%d\n", jmpLabelNum)
}

func (c *CaseNode) gen() {
	for _, b := range c.body {
		b.gen()
	}
}

func (c *CastNode) gen() {
	c.base.gen()
	fmt.Println("	pop rax")
	t := c.toTy
	if _, ok := t.(*types.Bool); ok {
		fmt.Println("	cmp rax, 0")
		fmt.Println("	setne al")
	}
	switch t.Size() {
	case 1:
		fmt.Println("	movsx rax, al")
	case 2:
		fmt.Println("	movsx rax, ax")
	case 4:
		fmt.Println("	movsxd rax, eax")
	case 8:
		// rax is 8 bits register
	default:
		log.Fatalf("Unhandled type size: %d", t)
	}
	fmt.Println("	push rax")
}

func (c *ContinueNode) gen() {
	if jmpLabelNum == 0 {
		log.Fatal("invalid continue statement.")
	}
	fmt.Printf("jmp .L.continue.%d\n", jmpLabelNum)
}

func (d *DecNode) gen() {
	body := d.body
	t := body.LoadType()
	var diff int
	if p, ok := body.LoadType().(types.Pointing); ok {
		diff = p.Base().Size()
	} else {
		diff = 1
	}

	body.genAddr()
	fmt.Println("	push [rsp]")
	load(t)
	fmt.Println("	pop rax")
	fmt.Printf("	sub rax, %d\n", diff)
	fmt.Println("	push rax")
	store(t)

	if !d.isPre {
		fmt.Println("	pop rax")
		fmt.Printf("	add rax, %d\n", diff)
		fmt.Println("	push rax")
	}
}

func (d *DerefNode) gen() {
	d.ptr.gen()
	ty := d.LoadType()
	if _, ok := ty.(*types.Arr); !ok {
		load(ty)
	}
}

func (d *DoWhileNode) gen() {
	c := labelCount
	labelCount++
	prev := jmpLabelNum
	jmpLabelNum = c
	fmt.Printf(".L.do.while.%d:\n", c)
	d.then.gen()
	fmt.Printf(".L.continue.%d:\n", jmpLabelNum)
	d.cond.gen()
	fmt.Println("	pop rax")
	fmt.Println("	cmp rax, 0")
	fmt.Printf("	jne .L.do.while.%d\n", c)
	fmt.Printf(".L.break.%d:\n", jmpLabelNum)
	jmpLabelNum = prev
}

func (e *ExprNode) gen() {
	e.Body.gen()
	fmt.Println("	add rsp, 8")
}

func (f *ForNode) gen() {
	c := labelCount
	labelCount++
	prevLoopLabelNum := jmpLabelNum
	jmpLabelNum = c
	if f.init != nil {
		f.init.gen()
	}
	fmt.Printf(".L.begin.%d:\n", c)
	if f.cond != nil {
		f.cond.gen()
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .L.break.%d\n", jmpLabelNum)
	}
	if f.body != nil {
		f.body.gen()
	}
	fmt.Printf(".L.continue.%d:\n", jmpLabelNum)
	if f.inc != nil {
		f.inc.gen()
	}
	fmt.Printf("	jmp .L.begin.%d\n", c)
	fmt.Printf(".L.break.%d:\n", c)
	jmpLabelNum = prevLoopLabelNum
}

func (f *FnCallNode) gen() {
	for _, param := range f.params {
		param.gen()
	}
	for i := len(f.params) - 1; i >= 0; i-- {
		fmt.Printf("	pop %s\n", paramRegs8[i])
	}
	// align rsp to 16 byte boundary
	fmt.Println("	mov rax, rsp")
	fmt.Println("	and rax, 15")
	fmt.Printf("	jz .L.func.call.%d\n", labelCount)
	fmt.Println("	mov rax, 0")
	fmt.Printf("	call %s\n", f.name)
	fmt.Printf("	jmp .L.func.end.%d\n", labelCount)
	fmt.Printf(".L.func.call.%d:\n", labelCount)
	fmt.Println("	sub rsp, 8")
	fmt.Println("	mov rax, 0")
	fmt.Printf("	call %s\n", f.name)
	fmt.Println("	add rsp, 8")
	fmt.Printf(".L.func.end.%d:\n", labelCount)
	fmt.Println("	push rax")
	labelCount++
}

func (f *FnNode) gen() {
	name := f.name
	if !f.isStatic {
		fmt.Printf(".globl %s\n", name)
	}
	fmt.Printf("%s:\n", name)
	fmt.Println("	push rbp")
	fmt.Println("	mov rbp, rsp")
	fmt.Printf("	sub rsp, %d\n", f.StackSize)
	for i, param := range f.Params {
		var r [6]string
		switch param.Type().Size() {
		case 1:
			r = paramRegs1
		case 2:
			r = paramRegs2
		case 4:
			r = paramRegs4
		case 8:
			r = paramRegs8
		default:
			log.Fatalf("Unhandled type size: %d", param.Type().Size())
		}
		fmt.Printf("	mov [rbp-%d], %s\n", param.Offset, r[i])
	}
	for _, node := range f.Body {
		node.gen()
	}
	fmt.Printf(".L.return.%s:\n", name)
	fmt.Println("	mov rsp, rbp")
	fmt.Println("	pop rbp")
	fmt.Println("	ret")
}

func (i *IfNode) gen() {
	c := labelCount
	labelCount++
	if i.els != nil {
		i.cond.gen()
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .L.else.%d\n", c)
		i.then.gen()
		fmt.Printf("	jmp .L.end.%d\n", c)
		fmt.Printf(".L.else.%d:\n", c)
		i.els.gen()
		fmt.Printf(".L.end.%d:\n", c)
	} else {
		i.cond.gen()
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .L.end.%d\n", c)
		i.then.gen()
		fmt.Printf(".L.end.%d:\n", c)
	}
}

func (i *IncNode) gen() {
	body := i.body
	t := body.LoadType()
	var diff int
	if p, ok := body.LoadType().(types.Pointing); ok {
		diff = p.Base().Size()
	} else {
		diff = 1
	}

	body.genAddr()
	fmt.Println("	push [rsp]")
	load(t)
	fmt.Println("	pop rax")
	fmt.Printf("	add rax, %d\n", diff)
	fmt.Println("	push rax")
	store(t)

	if !i.isPre {
		fmt.Println("	pop rax")
		fmt.Printf("	sub rax, %d\n", diff)
		fmt.Println("	push rax")
	}
}

func (m *MemberNode) gen() {
	m.genAddr()
	ty := m.LoadType()
	if _, ok := ty.(*types.Arr); !ok {
		load(ty)
	}
}

func (n *NotNode) gen() {
	n.body.gen()
	fmt.Println("	pop rax")
	fmt.Println("	cmp rax, 0")
	fmt.Println("	sete al")
	fmt.Println("	push rax")
}

func (*NullNode) gen() {}

func (n *NumNode) gen() {
	if n.val > math.MaxInt32 {
		fmt.Printf("	movabs rax, %d\n", n.val)
		fmt.Println("	push rax")
	} else {
		fmt.Printf("	push %d\n", n.val)
	}
}

func (r *RetNode) gen() {
	if r.rhs != nil {
		r.rhs.gen()
		fmt.Println("	pop rax")
	}
	fmt.Printf("	jmp .L.return.%s\n", r.fnName)
}

func (s *StmtExprNode) gen() {
	for _, st := range s.body {
		st.gen()
	}
}

func (s *SwitchNode) gen() {
	c := labelCount
	labelCount++
	prev := jmpLabelNum
	jmpLabelNum = c

	s.target.gen()
	fmt.Println("	pop rax")
	dflt := s.dflt
	for _, cs := range s.cases {
		if dflt != nil && dflt.idx == cs.idx {
			continue
		}
		fmt.Printf("	cmp rax, %d\n", cs.cmp)
		fmt.Printf("	je .L.case.%d.%d\n", c, cs.idx)
	}
	if dflt != nil {
		fmt.Printf("	jmp .L.case.%d.%d\n", c, dflt.idx)
	}
	fmt.Printf("	jmp .L.break.%d\n", jmpLabelNum)
	for _, cs := range s.cases {
		fmt.Printf(".L.case.%d.%d:\n", c, cs.idx)
		cs.gen()
	}
	fmt.Printf(".L.break.%d:\n", jmpLabelNum)
	jmpLabelNum = prev
}

func (t *TernaryNode) gen() {
	c := labelCount
	labelCount++
	t.cond.gen()
	fmt.Println("	pop rax")
	fmt.Println("	cmp rax, 0")
	fmt.Printf("	je .L.ternary.%d.rhs\n", c)
	t.lhs.gen()
	fmt.Printf("	jmp .L.ternary.%d.end\n", c)
	fmt.Printf(".L.ternary.%d.rhs:\n", c)
	t.rhs.gen()
	fmt.Printf(".L.ternary.%d.end:\n", c)
}

func (v *VarNode) gen() {
	v.genAddr()
	ty := v.LoadType()
	if _, ok := ty.(*types.Arr); !ok {
		load(ty)
	}
}

func (w *WhileNode) gen() {
	c := labelCount
	labelCount++
	prevLoopLabelNum := jmpLabelNum
	jmpLabelNum = c
	fmt.Printf(".L.continue.%d:\n", c)
	w.cond.gen()
	fmt.Println("	pop rax")
	fmt.Println("	cmp rax, 0")
	fmt.Printf("	je .L.break.%d\n", c)
	w.then.gen()
	fmt.Printf("	jmp .L.continue.%d\n", c)
	fmt.Printf(".L.break.%d:\n", jmpLabelNum)
	jmpLabelNum = prevLoopLabelNum
}

func (d *DerefNode) genAddr() {
	d.ptr.gen()
}

func (m *MemberNode) genAddr() {
	m.lhs.genAddr()
	fmt.Println("	pop rax")
	fmt.Printf("	add rax, %d\n", m.mem.Offset)
	fmt.Println("	push rax")
}

func (v *VarNode) genAddr() {
	switch v := v.Var.(type) {
	case *vars.GVar:
		fmt.Printf("	push offset %s\n", v.Name())
	case *vars.LVar:
		fmt.Printf("	lea rax, [rbp-%d]\n", v.Offset)
		fmt.Println("	push rax")
	default:
		log.Fatalf("Unhandled case in genAddr()")
	}
}

func load(t types.Type) {
	fmt.Println("	pop rax")
	switch t.Size() {
	case 1:
		fmt.Println("	movsx rax, byte ptr [rax]")
	case 2:
		fmt.Println("	movsx rax, word ptr [rax]")
	case 4:
		fmt.Println("	movsxd rax, dword ptr [rax]")
	case 8:
		fmt.Println("	mov rax, [rax]")
	default:
		log.Fatalf("Unhandled type size: %d", t.Size())
	}
	fmt.Println("	push rax")
}

func store(t types.Type) {
	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")
	if _, ok := t.(*types.Bool); ok {
		fmt.Println("	cmp rdi, 0")
		fmt.Println("	setne dil")
		fmt.Println("	movzb rdi, dil")
	}
	var r string
	switch t.Size() {
	case 1:
		r = "dil"
	case 2:
		r = "di"
	case 4:
		r = "edi"
	case 8:
		r = "rdi"
	default:
		log.Fatalf("Unhandled type size: %d", t.Size())
	}
	fmt.Printf("	mov [rax], %s\n", r)
	fmt.Println("	push rdi")
}
