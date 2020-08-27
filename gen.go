package main

import (
	"fmt"
	"log"
)

var paramRegs1 = [...]string{"dil", "sil", "dl", "cl", "r8b", "r9b"}
var paramRegs2 = [...]string{"di", "si", "dx", "cx", "r8w", "r9w"}
var paramRegs4 = [...]string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
var paramRegs8 = [...]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

var labelCount int

func (a *ast) gen() {
	fmt.Println(".intel_syntax noprefix")
	a.genData()
	a.genText()
}

func (a *ast) genData() {
	fmt.Println(".data")
	for _, g := range a.gVars {
		fmt.Printf("%s:\n", g.name)
		if c := g.content; c == nil {
			fmt.Printf("	.zero %d\n", g.ty.size())
		} else {
			switch c.(type) {
			case string:
				fmt.Printf("	.string \"%s\"\n", c)
			case int:
				// TODO
			default:
				log.Fatalf("unexpected type on gVar content: %T", c)
			}
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
	a.lhs.gen()
	a.rhs.gen()

	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")

	switch a.op {
	case ndAdd:
		fmt.Println("	add rax, rdi")
	case ndSub:
		fmt.Println("	sub rax, rdi")
	case ndMul:
		fmt.Println("	imul rax, rdi")
	case ndDiv:
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
	case ndPtrAdd:
		fmt.Printf("	imul rdi, %d\n", a.loadType().(ptr).base().size())
		fmt.Printf("	add rax, rdi\n")
	case ndPtrSub:
		fmt.Printf("	imul rdi, %d\n", a.loadType().(ptr).base().size())
		fmt.Printf("	sub rax, rdi\n")
	case ndPtrDiff:
		fmt.Printf("	sub rax, rdi\n")
		fmt.Printf("	cqo\n")
		fmt.Printf("	mov rdi, %d\n", a.lhs.loadType().(ptr).base().size())
		fmt.Println("	idiv rdi")
	default:
		log.Fatal("Unhandled node kind")
	}
	fmt.Println("	push rax")
}

func (a *assignNode) gen() {
	a.lhs.genAddr()
	a.rhs.gen()
	store(a.loadType())
}

func (b *blkNode) gen() {
	for _, st := range b.body {
		st.gen()
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
	if f.init != nil {
		f.init.gen()
	}
	fmt.Printf(".L.begin.%d:\n", c)
	if f.cond != nil {
		f.cond.gen()
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .L.end.%d\n", c)
	}
	if f.body != nil {
		f.body.gen()
	}
	if f.inc != nil {
		f.inc.gen()
	}
	fmt.Printf("	jmp .L.begin.%d\n", c)
	fmt.Printf(".L.end.%d:\n", c)
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
	fmt.Printf(".globl %s\n", name)
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

func (m *memberNode) gen() {
	m.genAddr()
	ty := m.loadType()
	if _, ok := ty.(*tyArr); !ok {
		load(ty)
	}
}

func (*nullNode) gen() {}

func (n *numNode) gen() {
	fmt.Printf("	push %d\n", n.val)
}

func (r *retNode) gen() {
	r.rhs.gen()
	fmt.Println("	pop rax")
	fmt.Printf("	jmp .L.return.%s\n", r.fnName)
}

func (s *stmtExprNode) gen() {
	for _, st := range s.body {
		st.gen()
	}
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
	fmt.Printf(".L.begin.%d:\n", c)
	w.cond.gen()
	fmt.Println("	pop rax")
	fmt.Println("	cmp rax, 0")
	fmt.Printf("	je .L.end.%d\n", c)
	w.then.gen()
	fmt.Printf("	jmp .L.begin.%d\n", c)
	fmt.Printf(".L.end.%d:\n", c)
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
	}
}

func load(ty ty) {
	fmt.Println("	pop rax")
	switch ty.size() {
	case 1:
		fmt.Println("	movsx rax, byte ptr [rax]")
	case 2:
		fmt.Println("	movsx rax, word ptr [rax]")
	case 4:
		fmt.Println("	movsxd rax, dword ptr [rax]")
	case 8:
		fmt.Println("	mov rax, [rax]")
	default:
		log.Fatalf("unhandled type size: %d", ty.size())
	}
	fmt.Println("	push rax")
}

func store(ty ty) {
	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")
	if _, ok := ty.(*tyBool); ok {
		fmt.Println("	cmp rdi, 0")
		fmt.Println("	setne dil")
		fmt.Println("	movzb rdi, dil")
	}
	var r string
	switch ty.size() {
	case 1:
		r = "dil"
	case 2:
		r = "di"
	case 4:
		r = "edi"
	case 8:
		r = "rdi"
	default:
		log.Fatalf("unhandled type size: %d", ty.size())
	}
	fmt.Printf("	mov [rax], %s\n", r)
	fmt.Println("	push rdi")
}
