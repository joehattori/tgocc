package main

import (
	"fmt"
	"strings"
)

func consume(toks []*Token, str string) ([]*Token, bool) {
	curTok := toks[0]
	if curTok.kind != tkReserved || curTok.length != len(str) || !strings.HasPrefix(curTok.str, str) {
		return toks, false
	}
	return toks[1:], false
}

func expect(toks []*Token, str string) error {
	curTok := toks[0]
	if curTok.kind != tkReserved || curTok.length != len(str) || !strings.HasPrefix(curTok.str, str) {
		return fmt.Errorf("%s was expected but got %s", str, curTok.str)
	}
	return nil
}

func expectNum(toks []*Token) (int, error) {
	curTok := toks[0]
	if curTok.kind != tkNum {
		return 0, fmt.Errorf("Number was expected")
	}
	return curTok.val, nil
}

type nodeKind int

const (
	ndAdd = iota
	ndSub
	ndMul
	ndDiv
	ndNum
	ndEq
	ndNeq
	ndLt
	ndLeq
	ndGt
	ndGeq
)

// Node represents each node in ast
type Node struct {
	kind nodeKind
	lhs  *Node
	rhs  *Node
	val  int
}

func newNode(kind nodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{kind: kind, lhs: lhs, rhs: rhs}
}

func newNodeNum(val int) *Node {
	return &Node{kind: ndNum, val: val}
}

// expr       = equality
// equality   = relational ("==" relational | "!=" relational)*
// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
// add        = mul ("+" mul | "-" mul)*
// mul        = unary ("*" unary | "/" unary)*
// unary      = ("+" | "-")? primary
// primary    = num | "(" expr ")"

type opKind struct {
	str  string
	kind nodeKind
}

// Expr parses the expression shown above
func Expr(toks []*Token) ([]*Token, *Node) {
	return equality(toks)
}

func equality(toks []*Token) ([]*Token, *Node) {
	toks, node := relational(toks)
	ops := []opKind{
		opKind{str: "==", kind: ndEq},
		opKind{str: "!=", kind: ndNeq},
	}
	for {
		matched := false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, relNode := relational(newToks)
				toks, node = nxtToks, newNode(op.kind, node, relNode)
				matched = true
			}
		}
		if matched {
			continue
		}
		break
	}
	return toks, node
}

func relational(toks []*Token) ([]*Token, *Node) {
	toks, node := addSub(toks)
	ops := []opKind{
		opKind{str: "<=", kind: ndLeq},
		opKind{str: ">=", kind: ndGeq},
		opKind{str: "<", kind: ndLt},
		opKind{str: ">", kind: ndGt},
	}
	for {
		matched := false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, addSubNode := addSub(newToks)
				toks, node = nxtToks, newNode(op.kind, node, addSubNode)
				matched = true
			}
		}
		if matched {
			continue
		}
		break
	}
	return toks, node
}

func addSub(toks []*Token) ([]*Token, *Node) {
	toks, node := mulDiv(toks)
	ops := []opKind{
		opKind{str: "+", kind: ndAdd},
		opKind{str: "-", kind: ndSub},
	}
	for {
		matched := false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, mulDivNode := mulDiv(newToks)
				toks, node = nxtToks, newNode(op.kind, node, mulDivNode)
				matched = true
			}
		}
		if matched {
			continue
		}
		break
	}
	return toks, node
}

func mulDiv(toks []*Token) ([]*Token, *Node) {
	toks, node := unary(toks)
	ops := []opKind{
		opKind{str: "*", kind: ndMul},
		opKind{str: "/", kind: ndDiv},
	}
	for {
		matched := false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, unaryNode := unary(newToks)
				toks, node = nxtToks, newNode(op.kind, node, unaryNode)
				matched = true
			}
		}
		if matched {
			continue
		}
		break
	}
	return toks, node
}

func unary(toks []*Token) ([]*Token, *Node) {
	if newToks, isPlus := consume(toks, "+"); isPlus {
		return primary(newToks)
	} else if newToks, isMinus := consume(toks, "-"); isMinus {
		newToks, node := primary(newToks)
		return newToks, newNode(ndSub, newNodeNum(0), node)
	}
	return primary(toks)
}

func primary(toks []*Token) ([]*Token, *Node) {
	if newToks, isPar := consume(toks, "("); isPar {
		newToks, node := Expr(newToks)
		toks = newToks
		err := expect(newToks, ")")
		if err != nil {
			panic("\")\" needed")
		}
		return toks[1:], node
	}
	n, err := expectNum(toks)
	if err != nil {
		panic("Number expected")
	}
	return toks[1:], newNodeNum(n)
}
