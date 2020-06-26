package main

import "fmt"

// genLVar pushes the address of loval variable to stack
func genLVar(node *Node) {
	if node.kind != ndLvar {
		panic("This node should be local variable")
	}
	fmt.Println("	mov rax, rbp")
	fmt.Printf("	sub rax, %d\n", node.offset)
	fmt.Println("	push rax")
}

// Gen generates assembly code for given node. This emulates stack machine.
func Gen(node *Node) {
	switch node.kind {
	case ndNum:
		fmt.Printf("	push %d\n", node.val)
		return
	case ndAssign:
		genLVar(node.lhs)
		Gen(node.rhs)
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

	Gen(node.lhs)
	Gen(node.rhs)

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
