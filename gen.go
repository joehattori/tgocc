package main

import "fmt"

var ifLabelCount, elseLabelCount, endifLabelCount int
var beginLabelCount, endLabelCount int

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
		genNode(node.cond)
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .Lelse%d\n", elseLabelCount)
		fmt.Printf("	je .Lend%d\n", endifLabelCount)
		genNode(node.then)
		fmt.Printf(".Lelse%d:\n", elseLabelCount)
		elseLabelCount++
		if node.els != nil {
			genNode(node.els)
		}
		fmt.Printf(".Lend%d:\n", endifLabelCount)
		endifLabelCount++
		return
	case ndWhile:
		fmt.Printf(".Lbegin%d:\n", beginLabelCount)
		genNode(node.cond)
		fmt.Println("	pop rax")
		fmt.Println("	cmp rax, 0")
		fmt.Printf("	je .Lend%d\n", endLabelCount)
		genNode(node.then)
		fmt.Printf("	jmp .Lbegin%d\n", beginLabelCount)
		fmt.Printf(".Lend%d:\n", endLabelCount)
		beginLabelCount++
		endLabelCount++
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
