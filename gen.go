package main

import "fmt"

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
	}
	fmt.Println("	push rax")
}

// Gode generates assembly for the whole program. This emulates stack machine.
func Gen() {
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	stackSize := 8 * len(LVars)
	// extend stack for local variables (named a to z for now)
	fmt.Println("	push rbp")
	fmt.Println("	mov rbp, rsp")
	fmt.Printf("	sub rsp, %d\n", stackSize)

	for _, code := range Code {
		genNode(code)
		fmt.Println("	pop rax")
	}

	fmt.Println("	mov rsp, rbp")
	fmt.Println("	pop rbp")
	fmt.Println("	ret")
}
