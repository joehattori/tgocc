package main

import "fmt"

var labelCount int

var argRegs = [...]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

func genLVar(v *LvarNode) {
	fmt.Println("	mov rax, rbp")
	fmt.Printf("	sub rax, %d\n", v.offset)
	fmt.Println("	push rax")
}

func genNode(node Node, funcName string) {
	switch node.kind() {
	case ndNum:
		fmt.Printf("	push %d\n", node.(*NumNode).val)
		return
	case ndAssign:
		a := node.(*AssignNode)
		genLVar(a.lhs.(*LvarNode))
		genNode(a.rhs, funcName)
		fmt.Println("	pop rdi")
		fmt.Println("	pop rax")
		fmt.Println("	mov [rax], rdi")
		fmt.Println("	push rdi")
		return
	case ndLvar:
		genLVar(node.(*LvarNode))
		fmt.Println("	pop rax")
		fmt.Println("	mov rax, [rax]")
		fmt.Println("	push rax")
		return
	case ndRet:
		r := node.(*RetNode)
		genNode(r.rhs, funcName)
		fmt.Println("	pop rax")
		fmt.Printf("	jmp .L.return.%s\n", funcName)
		return
	case ndIf:
		c := labelCount
		labelCount++
		i := node.(*IfNode)
		if i.els != nil {
			genNode(i.cond, funcName)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .L.else.%d\n", c)
			genNode(i.then, funcName)
			fmt.Printf("	je .L.end.%d\n", c)
			fmt.Printf(".L.else.%d:\n", c)
			genNode(i.els, funcName)
			fmt.Printf(".L.end.%d:\n", c)
		} else {
			genNode(i.cond, funcName)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .L.end.%d\n", c)
			genNode(i.then, funcName)
			fmt.Printf(".L.end.%d:\n", c)
		}
		return
	case ndWhile:
		c := labelCount
		labelCount++
		w := node.(*WhileNode)
		fmt.Printf(".L.begin.%d:\n", c)
		genNode(w.cond, funcName)
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .L.end.%d\n", c)
		genNode(w.then, funcName)
		fmt.Printf("	jmp .L.begin.%d\n", c)
		fmt.Printf(".L.end.%d:\n", c)
		return
	case ndFor:
		c := labelCount
		labelCount++
		f := node.(*ForNode)
		if f.init != nil {
			genNode(f.init, funcName)
		}
		fmt.Printf(".L.begin.%d:\n", c)
		if f.cond != nil {
			genNode(f.cond, funcName)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .L.end.%d\n", c)
		}
		if f.body != nil {
			genNode(f.body, funcName)
		}
		if f.inc != nil {
			genNode(f.inc, funcName)
		}
		fmt.Printf("	jmp .L.begin.%d\n", c)
		fmt.Printf(".L.end.%d:\n", c)
		return
	case ndBlk:
		b := node.(*BlkNode)
		for _, st := range b.body {
			genNode(st, funcName)
		}
		return
	case ndFuncCall:
		f := node.(*FuncCallNode)
		for i, arg := range f.args {
			genNode(arg, funcName)
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
		return
	}

	nd := node.(*ArithNode)

	genNode(nd.lhs, funcName)
	genNode(nd.rhs, funcName)

	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")

	switch nd.kind() {
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

// Gen generates assembly for the whole program. This emulates stack machine.
func Gen() {
	fmt.Println(".intel_syntax noprefix")

	for _, code := range Code {
		fn := code.(*FuncDefNode)
		funcName := fn.name
		fmt.Printf(".globl %s\n", funcName)
		fmt.Printf("%s:\n", funcName)
		fmt.Println("	push rbp")
		fmt.Println("	mov rbp, rsp")
		fmt.Printf("	sub rsp, %d\n", fn.stackSize)
		for i, arg := range fn.args {
			fmt.Printf("	mov [rbp-%d], %s\n", arg.offset, argRegs[i])
		}
		for _, node := range fn.body {
			genNode(node, funcName)
		}
		fmt.Printf(".L.return.%s:\n", funcName)
		fmt.Println("	mov rsp, rbp")
		fmt.Println("	pop rbp")
		fmt.Println("	ret")
	}
}
