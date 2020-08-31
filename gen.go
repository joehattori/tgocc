package main

import (
	"fmt"
	"log"
	"strings"
)

var (
	labelCount  int
	jmpLabelNum int

	paramRegs1 = [...]string{"dil", "sil", "dl", "cl", "r8b", "r9b"}
	paramRegs2 = [...]string{"di", "si", "dx", "cx", "r8w", "r9w"}
	paramRegs4 = [...]string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
	paramRegs8 = [...]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
)

func (a *ast) gen() {
	fmt.Println(".intel_syntax noprefix")
	a.genData()
	a.genText()
}

func (a *ast) genData() {
	fmt.Println(".data")
	for _, g := range a.gVars {
		fmt.Printf("%s:\n", g.name)
		genDataVar(g.init, g.ty)
	}
}

func genDataVar(init gVarInit, ty ty) {
	if init == nil {
		fmt.Printf("	.zero %d\n", ty.size())
	} else {
		switch init := init.(type) {
		case *gVarInitArr:
			switch ty := ty.(type) {
			case *tyArr:
				for _, e := range init.body {
					genDataVar(e, ty.of)
				}
			case *tyStruct:
				for i, e := range init.body {
					genDataVar(e, ty.members[i].ty)
				}
			default:
				for _, e := range init.body {
					genDataVar(e, ty)
				}
			}
		case *gVarInitLabel:
			fmt.Printf("	.quad %s\n", init.label)
		case *gVarInitStr:
			fmt.Printf("	.string \"%s\"\n", strings.TrimSuffix(init.content, string('\000')))
		case *gVarInitInt:
			switch init.sz {
			case 1:
				fmt.Printf("	.byte %d\n", init.val)
			case 2:
				fmt.Printf("	.value %d\n", init.val)
			case 4:
				fmt.Printf("	.long %d\n", init.val)
			case 8:
				fmt.Printf("	.quad %d\n", init.val)
			default:
				log.Fatalf("unhandled type size %d on global variable initialization.", init.sz)
			}
		default:
			log.Fatalf("unexpected type on gVar content: %T", init)
		}
	}
}

func (a *ast) genText() {
	fmt.Println(".text")
	for _, f := range a.fns {
		f.gen()
	}
}

func (a *addrNode) gen() {
	a.v.genAddr()
}

func (a *arithNode) gen() {
	switch a.op {
	case ndAddEq, ndSubEq, ndMulEq, ndDivEq, ndPtrAddEq, ndPtrSubEq, ndShlEq, ndShrEq:
		a.lhs.(addressableNode).genAddr()
		defer store(a.lhs.loadType())
	}

	a.lhs.gen()
	a.rhs.gen()

	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")

	switch a.op {
	case ndAdd, ndAddEq:
		fmt.Println("	add rax, rdi")
	case ndSub, ndSubEq:
		fmt.Println("	sub rax, rdi")
	case ndMul, ndMulEq:
		fmt.Println("	imul rax, rdi")
	case ndDiv, ndDivEq:
		fmt.Println("	cqo")
		fmt.Println("	idiv rdi")
	case ndEq:
		fmt.Println("	cmp rax, rdi")
		fmt.Println("	sete al")
		fmt.Println("	movzb rax, al")
	case ndNeq:
		fmt.Println("	cmp rax, rdi")
		fmt.Println("	setne al")
		fmt.Println("	movzb rax, al")
	case ndLt:
		fmt.Println("	cmp rax, rdi")
		fmt.Println("	setl al")
		fmt.Println("	movzb rax, al")
	case ndLeq:
		fmt.Println("	cmp rax, rdi")
		fmt.Println("	setle al")
		fmt.Println("	movzb rax, al")
	case ndGt:
		fmt.Println("	cmp rdi, rax")
		fmt.Println("	setl al")
		fmt.Println("	movzb rax, al")
	case ndGeq:
		fmt.Println("	cmp rdi, rax")
		fmt.Println("	setle al")
		fmt.Println("	movzb rax, al")
	case ndPtrAdd, ndPtrAddEq:
		fmt.Printf("	imul rdi, %d\n", a.loadType().(ptr).base().size())
		fmt.Printf("	add rax, rdi\n")
	case ndPtrSub, ndPtrSubEq:
		fmt.Printf("	imul rdi, %d\n", a.loadType().(ptr).base().size())
		fmt.Printf("	sub rax, rdi\n")
	case ndPtrDiff:
		fmt.Println("	sub rax, rdi")
		fmt.Println("	cqo")
		fmt.Printf("	mov rdi, %d\n", a.lhs.loadType().(ptr).base().size())
		fmt.Println("	idiv rdi")
	case ndBitOr:
		fmt.Println("	or rax, rdi")
	case ndBitXor:
		fmt.Println("	xor rax, rdi")
	case ndBitAnd:
		fmt.Println("	and rax, rdi")
	case ndLogOr:
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
	case ndLogAnd:
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
	case ndShl, ndShlEq:
		fmt.Println("	mov cl, dil")
		fmt.Println("	sal rax, cl")
	case ndShr, ndShrEq:
		fmt.Println("	mov cl, dil")
		fmt.Println("	sar rax, cl")
	default:
		log.Fatal("unhandled node kind")
	}

	fmt.Println("	push rax")
}

func (a *assignNode) gen() {
	a.lhs.genAddr()
	a.rhs.gen()
	store(a.loadType())
}

func (b *bitNotNode) gen() {
	b.body.gen()
	fmt.Println("	pop rax")
	fmt.Println("	not rax")
	fmt.Println("	push rax")
}

func (b *blkNode) gen() {
	for _, st := range b.body {
		st.gen()
	}
}

func (b *breakNode) gen() {
	if jmpLabelNum == 0 {
		log.Fatal("invalid break statement.")
	}
	fmt.Printf("	jmp .L.break.%d\n", jmpLabelNum)
}

func (c *caseNode) gen() {
	for _, b := range c.body {
		b.gen()
	}
}

func (c *castNode) gen() {
	c.base.gen()
	fmt.Println("	pop rax")
	t := c.toTy
	if _, ok := t.(*tyBool); ok {
		fmt.Println("	cmp rax, 0")
		fmt.Println("	setne al")
	}
	switch t.size() {
	case 1:
		fmt.Println("	movsx rax, al")
	case 2:
		fmt.Println("	movsx rax, ax")
	case 4:
		fmt.Println("	movsxd rax, eax")
	case 8:
		// rax is 8 bits register
	default:
		log.Fatalf("unhandled type size: %d", t)
	}
	fmt.Println("	push rax")
}

func (c *continueNode) gen() {
	if jmpLabelNum == 0 {
		log.Fatal("invalid continue statement.")
	}
	fmt.Printf("jmp .L.continue.%d\n", jmpLabelNum)
}

func (d *decNode) gen() {
	body := d.body
	t := body.loadType()
	var diff int
	if p, ok := body.loadType().(ptr); ok {
		diff = p.base().size()
	} else {
		diff = 1
	}

	body.genAddr()
	fmt.Println("	push rax")
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

func (d *defaultNode) gen() {
	for _, b := range d.body {
		b.gen()
	}
}

func (d *derefNode) gen() {
	d.ptr.gen()
	ty := d.loadType()
	if _, ok := ty.(*tyArr); !ok {
		load(ty)
	}
}

func (e *exprNode) gen() {
	e.body.gen()
	fmt.Println("	add rsp, 8")
}

func (f *forNode) gen() {
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
	fmt.Printf(".L.continue.%d:\n", c)
	if f.inc != nil {
		f.inc.gen()
	}
	fmt.Printf("	jmp .L.begin.%d\n", c)
	fmt.Printf(".L.break.%d:\n", c)
	jmpLabelNum = prevLoopLabelNum
}

func (f *fnCallNode) gen() {
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

func (f *fnNode) gen() {
	name := f.name
	if !f.isStatic {
		fmt.Printf(".globl %s\n", name)
	}
	fmt.Printf("%s:\n", name)
	fmt.Println("	push rbp")
	fmt.Println("	mov rbp, rsp")
	fmt.Printf("	sub rsp, %d\n", f.stackSize)
	for i, param := range f.params {
		var r [6]string
		switch param.ty.size() {
		case 1:
			r = paramRegs1
		case 2:
			r = paramRegs2
		case 4:
			r = paramRegs4
		case 8:
			r = paramRegs8
		default:
			log.Fatalf("unhandled type size: %d", param.ty.size())
		}
		fmt.Printf("	mov [rbp-%d], %s\n", param.offset, r[i])
	}
	for _, node := range f.body {
		node.gen()
	}
	fmt.Printf(".L.return.%s:\n", name)
	fmt.Println("	mov rsp, rbp")
	fmt.Println("	pop rbp")
	fmt.Println("	ret")
}

func (i *ifNode) gen() {
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

func (i *incNode) gen() {
	body := i.body
	t := body.loadType()
	var diff int
	if p, ok := body.loadType().(ptr); ok {
		diff = p.base().size()
	} else {
		diff = 1
	}

	body.genAddr()
	fmt.Println("	push rax")
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

func (m *memberNode) gen() {
	m.genAddr()
	ty := m.loadType()
	if _, ok := ty.(*tyArr); !ok {
		load(ty)
	}
}

func (n *notNode) gen() {
	n.body.gen()
	fmt.Println("	pop rax")
	fmt.Println("	cmp rax, 0")
	fmt.Println("	sete al")
	fmt.Println("	push rax")
}

func (*nullNode) gen() {}

func (n *numNode) gen() {
	if n.val >= int64(int(n.val)) {
		fmt.Printf("	movabs rax, %d\n", n.val)
		fmt.Println("	push rax")
	} else {
		fmt.Printf("	push %d\n", n.val)
	}
}

func (r *retNode) gen() {
	if r.rhs != nil {
		r.rhs.gen()
		fmt.Println("	pop rax")
	}
	fmt.Printf("	jmp .L.return.%s\n", r.fnName)
}

func (s *stmtExprNode) gen() {
	for _, st := range s.body {
		st.gen()
	}
}

func (s *switchNode) gen() {
	c := labelCount
	labelCount++
	prev := jmpLabelNum
	jmpLabelNum = c
	for _, cs := range s.cases {
		s.target.gen()
		fmt.Println("	pop rax")
		i := cs.idx
		fmt.Printf("	cmp rax, %d\n", cs.cmp)
		fmt.Printf("	jne .L.nxt.%d.%d\n", c, i)
		cs.gen()
		fmt.Printf(".L.nxt.%d.%d:\n", c, i)
	}
	if d := s.dflt; d != nil {
		d.gen()
		for _, cs := range s.cases {
			if cs.idx > d.idx {
				cs.gen()
			}
		}
	}
	fmt.Printf(".L.break.%d:\n", jmpLabelNum)
	jmpLabelNum = prev
}

func (t *ternaryNode) gen() {
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

func (v *varNode) gen() {
	v.genAddr()
	ty := v.loadType()
	if _, ok := ty.(*tyArr); !ok {
		load(ty)
	}
}

func (w *whileNode) gen() {
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

func (d *derefNode) genAddr() {
	d.ptr.gen()
}

func (m *memberNode) genAddr() {
	m.lhs.genAddr()
	fmt.Println("	pop rax")
	fmt.Printf("	add rax, %d\n", m.mem.offset)
	fmt.Println("	push rax")
}

func (v *varNode) genAddr() {
	switch vr := v.v.(type) {
	case *gVar:
		fmt.Printf("	push offset %s\n", vr.name)
	case *lVar:
		fmt.Printf("	lea rax, [rbp-%d]\n", vr.offset)
		fmt.Println("	push rax")
	default:
		log.Fatalf("unhandled case in genAddr()")
	}
}

func load(t ty) {
	fmt.Println("	pop rax")
	switch t.size() {
	case 1:
		fmt.Println("	movsx rax, byte ptr [rax]")
	case 2:
		fmt.Println("	movsx rax, word ptr [rax]")
	case 4:
		fmt.Println("	movsxd rax, dword ptr [rax]")
	case 8:
		fmt.Println("	mov rax, [rax]")
	default:
		log.Fatalf("unhandled type size: %d", t.size())
	}
	fmt.Println("	push rax")
}

func store(t ty) {
	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")
	if _, ok := t.(*tyBool); ok {
		fmt.Println("	cmp rdi, 0")
		fmt.Println("	setne dil")
		fmt.Println("	movzb rdi, dil")
	}
	var r string
	switch t.size() {
	case 1:
		r = "dil"
	case 2:
		r = "di"
	case 4:
		r = "edi"
	case 8:
		r = "rdi"
	default:
		log.Fatalf("unhandled type size: %d", t.size())
	}
	fmt.Printf("	mov [rax], %s\n", r)
	fmt.Println("	push rdi")
}
