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
	return toks[1:], true
}

func consumeID(toks []*Token) ([]*Token, string, bool) {
	curTok := toks[0]
	r := []rune(curTok.str)
	if curTok.kind != tkID || !('a' <= r[0] && r[0] <= 'z') {
		return toks, "", false
	}
	return toks[1:], string(r[0:1]), true
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
	ndLvar
	ndAssign
)

// Node represents each node in ast
type Node struct {
	kind   nodeKind
	lhs    *Node
	rhs    *Node
	val    int
	offset int
}

func newNode(kind nodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{kind: kind, lhs: lhs, rhs: rhs}
}

func newNodeNum(val int) *Node {
	return &Node{kind: ndNum, val: val}
}

func newNodeVar(s string) *Node {
	return &Node{kind: ndLvar, offset: int([]rune(s)[0]-'a'+1) * 8}
}

// program    = stmt*
// stmt       = expr ";"
// expr       = assign
// assign     = equality ("=" assign) ?
// equality   = relational ("==" relational | "!=" relational)*
// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
// add        = mul ("+" mul | "-" mul)*
// mul        = unary ("*" unary | "/" unary)*
// unary      = ("+" | "-")? primary
// primary    = num | ident | "(" expr ")"

type opKind struct {
	str  string
	kind nodeKind
}

// Code is an array of nodes which represents whole program input
var Code []*Node

// Program parses tokens and fills up Code
func Program(toks []*Token) {
	for toks[0].kind != tkEOF {
		newToks, node := stmt(toks)
		toks = newToks
		Code = append(Code, node)
	}
}

func stmt(toks []*Token) ([]*Token, *Node) {
	newToks, node := expr(toks)
	err := expect(newToks, ";")
	if err != nil {
		panic(err)
	}
	return newToks[1:], node
}

func expr(toks []*Token) ([]*Token, *Node) {
	return assign(toks)
}

func assign(toks []*Token) ([]*Token, *Node) {
	toks, node := equality(toks)
	if newToks, isAssign := consume(toks, "="); isAssign {
		nxtToks, assignNode := assign(newToks)
		toks, node = nxtToks, newNode(ndAssign, node, assignNode)
	}
	return toks, node
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
	matched := true
	for matched {
		matched = false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, addSubNode := addSub(newToks)
				toks, node = nxtToks, newNode(op.kind, node, addSubNode)
				matched = true
			}
		}
	}
	return toks, node
}

func addSub(toks []*Token) ([]*Token, *Node) {
	toks, node := mulDiv(toks)
	ops := []opKind{
		opKind{str: "+", kind: ndAdd},
		opKind{str: "-", kind: ndSub},
	}
	matched := true
	for matched {
		matched = false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, mulDivNode := mulDiv(newToks)
				toks, node = nxtToks, newNode(op.kind, node, mulDivNode)
				matched = true
			}
		}
	}
	return toks, node
}

func mulDiv(toks []*Token) ([]*Token, *Node) {
	toks, node := unary(toks)
	ops := []opKind{
		opKind{str: "*", kind: ndMul},
		opKind{str: "/", kind: ndDiv},
	}
	matched := true
	for matched {
		matched = false
		for _, op := range ops {
			if newToks, isOp := consume(toks, op.str); isOp {
				nxtToks, unaryNode := unary(newToks)
				toks, node = nxtToks, newNode(op.kind, node, unaryNode)
				matched = true
			}
		}
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
		newToks, node := expr(newToks)
		toks = newToks
		err := expect(newToks, ")")
		if err != nil {
			panic(err)
		}
		return toks[1:], node
	}

	if newToks, id, isID := consumeID(toks); isID {
		return newToks, newNodeVar(id)
	}

	n, err := expectNum(toks)
	if err != nil {
		panic(err)
	}
	return toks[1:], newNodeNum(n)
}
