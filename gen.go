package main

import "fmt"

var labelCount int

var argRegs = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

func genLVar(node *Node) {
	if node.kind != ndLvar {
		panic("This node should be local variable")
	}
	fmt.Println("	mov rax, rbp")
	fmt.Printf("	sub rax, %d\n", node.offset)
	fmt.Println("	push rax")
}

func genNode(node *Node, funcName string) {
	switch node.kind {
	case ndNum:
		fmt.Printf("	push %d\n", node.val)
		return
	case ndAssign:
		genLVar(node.lhs)
		genNode(node.rhs, funcName)
		fmt.Println("	pop rdi")
		fmt.Println("	pop rax")
		fmt.Println("	mov [rax], rdi")
		fmt.Println("	push rdi")
		return
	case ndLvar:
		genLVar(node)
		fmt.Println("	pop rax")
		fmt.Println("	mov rax, [rax]")
		fmt.Println("	push rax")
		return
	case ndReturn:
		genNode(node.rhs, funcName)
		fmt.Println("	pop rax")
		fmt.Printf("	jmp .L.return.%s\n", funcName)
		return
	case ndIf:
		c := labelCount
		labelCount++
		if node.els != nil {
			genNode(node.cond, funcName)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .Lelse%d\n", c)
			genNode(node.then, funcName)
			fmt.Printf("	je .Lend%d\n", c)
			fmt.Printf(".Lelse%d:\n", c)
			genNode(node.els, funcName)
			fmt.Printf(".Lend%d:\n", c)
		} else {
			genNode(node.cond, funcName)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .Lend%d\n", c)
			genNode(node.then, funcName)
			fmt.Printf(".Lend%d:\n", c)
		}
		return
	case ndWhile:
		c := labelCount
		labelCount++
		fmt.Printf(".Lbegin%d:\n", c)
		genNode(node.cond, funcName)
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .Lend%d\n", c)
		genNode(node.then, funcName)
		fmt.Printf("	jmp .Lbegin%d\n", c)
		fmt.Printf(".Lend%d:\n", c)
		return
	case ndFor:
		c := labelCount
		labelCount++
		if node.forInit != nil {
			genNode(node.forInit, funcName)
		}
		fmt.Printf(".Lbegin%d:\n", c)
		if node.cond != nil {
			genNode(node.cond, funcName)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .Lend%d\n", c)
		}
		if node.then != nil {
			genNode(node.then, funcName)
		}
		if node.forInc != nil {
			genNode(node.forInc, funcName)
		}
		fmt.Printf("	jmp .Lbegin%d\n", c)
		fmt.Printf(".Lend%d:\n", c)
		return
	case ndBlk:
		for _, st := range node.blkStmts {
			genNode(st, funcName)
		}
		return
	case ndFuncCall:
		for i, arg := range node.funcArgs {
			genNode(arg, funcName)
			fmt.Printf("	pop %s\n", argRegs[i])
		}
		// align rsp to 16 byte boundary
		fmt.Println("	mov rax, rsp")
		fmt.Println("	and rax, 15")
		fmt.Printf("	jz .L.func.call%d\n", labelCount)
		fmt.Println("	mov rax, 0")
		fmt.Printf("	call %s\n", node.funcName)
		fmt.Printf("	jmp .L.func.end%d\n", labelCount)
		fmt.Printf(".L.func.call%d:\n", labelCount)
		fmt.Println("	sub rsp, 8")
		fmt.Println("	mov rax, 0")
		fmt.Printf("	call %s\n", node.funcName)
		fmt.Println("	add rsp, 8")
		fmt.Printf(".L.func.end%d:\n", labelCount)
		fmt.Println("	push rax")
		labelCount++
		return
	}

	genNode(node.lhs, funcName)
	genNode(node.rhs, funcName)

	fmt.Println("	pop rdi")
	fmt.Println("	pop rax")

	switch node.kind {
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
		funcName := code.funcName
		fmt.Printf(".globl %s\n", funcName)
		fmt.Printf("%s:\n", funcName)
		fmt.Println("	push rbp")
		fmt.Println("	mov rbp, rsp")
		fmt.Printf("	sub rsp, %d\n", code.stackSize)
		for _, node := range code.body {
			genNode(node, funcName)
		}
		fmt.Printf(".L.return.%s:\n", funcName)
		fmt.Println("	mov rsp, rbp")
		fmt.Println("	pop rbp")
		fmt.Println("	ret")
	}
}
