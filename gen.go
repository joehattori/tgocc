package main

import "fmt"

// Gen generates assembly code for ast
func Gen(node *Node) {
	if node.kind == ndNum {
		fmt.Printf("	push %d\n", node.val)
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
	}
	fmt.Println("	push rax")
}
