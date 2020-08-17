package main

import "fmt"

func (a *Ast) gen() {
	fmt.Println(".intel_syntax noprefix")

	for _, node := range a.nodes {
		node.gen()
	}
}

func (a *AddrNode) gen() {
	a.v.genLVarAddr()
}

func (a *ArithNode) gen() {
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
	default:
		panic("Unhandled node kind")
	}
	fmt.Println("	push rax")
}

func (a *AssignNode) gen() {
	switch l := a.lhs.(type) {
	case *DerefNode:
		l.ptr.gen()
		fmt.Println("	pop rax")
	case *LVarNode:
		l.genLVarAddr()
	default:
		panic(fmt.Sprintf("Unhandled type in assignment %T", l))
	}
	a.rhs.gen()
	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")
	fmt.Println("	mov [rax], rdi")
	fmt.Println("	push rdi")
}

func (b *BlkNode) gen() {
	for _, st := range b.body {
		st.gen()
	}
}

func (d *DerefNode) gen() {
	d.ptr.gen()
	fmt.Println("	pop rax")
	fmt.Println("	mov rax, [rax]")
	fmt.Println("	push rax")
}

func (f *ForNode) gen() {
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

func (f *FuncCallNode) gen() {
	for i, arg := range f.args {
		arg.gen()
		fmt.Printf("	pop %s\n", argRegs[i])
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
	fmt.Printf(".globl %s\n", name)
	fmt.Printf("%s:\n", name)
	fmt.Println("	push rbp")
	fmt.Println("	mov rbp, rsp")
	fmt.Printf("	sub rsp, %d\n", 8*(len(f.lvars)+1))
	for i, arg := range f.args {
		fmt.Printf("	mov [rbp-%d], %s\n", arg.offset, argRegs[i])
	}
	for _, node := range f.body {
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
		fmt.Printf("	je .L.end.%d\n", c)
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
	fmt.Println("	push rax")
}

func (l *LVarNode) gen() {
	l.genLVarAddr()
	fmt.Println("	pop rax")
	fmt.Println("	mov rax, [rax]")
	fmt.Println("	push rax")
}

func (l *LVarNode) genLVarAddr() {
	fmt.Println("   mov rax, rbp")
	fmt.Printf("    sub rax, %d\n", l.offset)
	fmt.Println("   push rax")
}

func (n *NumNode) gen() {
	fmt.Printf("	push %d\n", n.val)
}

func (r *RetNode) gen() {
	r.rhs.gen()
	fmt.Println("	pop rax")
	fmt.Printf("	jmp .L.return.%s\n", r.fnName)
}

func (w *WhileNode) gen() {
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
