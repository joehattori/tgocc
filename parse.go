package main

import (
	"fmt"
	"strings"
)

func consume(toks []*Token, str string) ([]*Token, bool) {
	curTok := toks[0]
	if curTok.kind != tkReserved || !strings.HasPrefix(curTok.str, str) {
		return toks, false
	}
	return toks[1:], false
}

func expect(toks []*Token, str string) error {
	curTok := toks[0]
	if curTok.kind != tkReserved || !strings.HasPrefix(curTok.str, str) {
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

/* expr    = mulDiv ("+" mulDiv | "-" mulDiv)*
   mulDiv  = unary ("*" unary | "/" unary)*
   unary   = ("+" | "-")? primary
   primary = num | "(" expr ")" */

// Expr parses the expression shown above
func Expr(toks []*Token) ([]*Token, *Node) {
	toks, node := mulDiv(toks)
	for {
		if newToks, isPlus := consume(toks, "+"); isPlus {
			nxtToks, mulDivNode := mulDiv(newToks)
			toks, node = nxtToks, newNode(ndAdd, node, mulDivNode)
			continue
		}
		if newToks, isMinus := consume(toks, "-"); isMinus {
			nxtToks, mulDivNode := mulDiv(newToks)
			toks, node = nxtToks, newNode(ndDiv, node, mulDivNode)
			continue
		}
		break
	}
	return toks, node
}

func mulDiv(toks []*Token) ([]*Token, *Node) {
	toks, node := unary(toks)
	for {
		if newToks, isMul := consume(toks, "*"); isMul {
			nxtToks, unaryNode := unary(newToks)
			toks, node = nxtToks, newNode(ndMul, node, unaryNode)
			continue
		}
		if newToks, isDiv := consume(toks, "/"); isDiv {
			nxtToks, unaryNode := unary(newToks)
			toks, node = nxtToks, newNode(ndDiv, node, unaryNode)
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
