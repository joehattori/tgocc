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

func genNode(node *Node) {
	switch node.kind {
	case ndNum:
		fmt.Printf("	push %d\n", node.val)
		return
	case ndAssign:
		genLVar(node.lhs)
		genNode(node.rhs)
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
		genNode(node.rhs)
		fmt.Println("	pop rax")
		fmt.Println("	mov rsp, rbp")
		fmt.Println("	pop rbp")
		fmt.Println("	ret")
		return
	case ndIf:
		c := labelCount
		labelCount++
		if node.els != nil {
			genNode(node.cond)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .Lelse%d\n", c)
			genNode(node.then)
			fmt.Printf("	je .Lend%d\n", c)
			fmt.Printf(".Lelse%d:\n", c)
			genNode(node.els)
			fmt.Printf(".Lend%d:\n", c)
		} else {
			genNode(node.cond)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .Lend%d\n", c)
			genNode(node.then)
			fmt.Printf(".Lend%d:\n", c)
		}
		return
	case ndWhile:
		c := labelCount
		labelCount++
		fmt.Printf(".Lbegin%d:\n", c)
		genNode(node.cond)
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .Lend%d\n", c)
		genNode(node.then)
		fmt.Printf("	jmp .Lbegin%d\n", c)
		fmt.Printf(".Lend%d:\n", c)
		return
	case ndFor:
		c := labelCount
		labelCount++
		if node.forInit != nil {
			genNode(node.forInit)
		}
		fmt.Printf(".Lbegin%d:\n", c)
		if node.cond != nil {
			genNode(node.cond)
			fmt.Println("	pop rax")
			fmt.Println("	cmp rax, 0")
			fmt.Printf("	je .Lend%d\n", c)
		}
		if node.then != nil {
			genNode(node.then)
		}
		if node.forInc != nil {
			genNode(node.forInc)
		}
		fmt.Printf("	jmp .Lbegin%d\n", c)
		fmt.Printf(".Lend%d:\n", c)
		return
	case ndBlk:
		for _, st := range node.blkStmts {
			genNode(st)
		}
		return
	case ndFuncCall:
		for i, arg := range node.funcArgs {
			genNode(arg)
			fmt.Printf("	pop %s\n", argRegs[i])
		}
		fmt.Printf("	call %s\n", node.funcName)
		fmt.Println("	push rax")
		return
	}

	genNode(node.lhs)
	genNode(node.rhs)

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
	fmt.Println(".globl main")
	fmt.Println("main:")

	// extend stack for local variables
	fmt.Println("	push rbp")
	fmt.Println("	mov rbp, rsp")
	fmt.Printf("	sub rsp, %d\n", 8*len(LVars))

	for _, code := range Code {
		genNode(code)
		fmt.Println("	pop rax")
	}

	fmt.Println("	mov rsp, rbp")
	fmt.Println("	pop rbp")
	fmt.Println("	ret")
}
