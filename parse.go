package main

import (
	"fmt"
	"regexp"
	"strings"
)

func consume(toks []*Token, str string) ([]*Token, bool) {
	cur := toks[0]
	if cur.kind != tkReserved || cur.length != len(str) || !strings.HasPrefix(cur.str, str) {
		return toks, false
	}
	return toks[1:], true
}

func consumeID(toks []*Token) ([]*Token, string, bool) {
	cur := toks[0]
	reg := regexp.MustCompile(`^[a-zA-Z]+[\w_]*`)
	varName := reg.FindString(cur.str)
	if cur.kind != tkID || varName == "" {
		return toks, "", false
	}
	return toks[1:], varName, true
}

func expect(toks []*Token, str string) []*Token {
	cur := toks[0]
	if cur.kind != tkReserved || cur.length != len(str) || !strings.HasPrefix(cur.str, str) {
		panic(fmt.Sprintf("%s was expected but got %s", str, cur.str))
	}
	return toks[1:]
}

func expectNum(toks []*Token) ([]*Token, int) {
	cur := toks[0]
	if cur.kind != tkNum {
		panic("Number was expected")
	}
	return toks[1:], cur.val
}

func findLVar(name string) *LVar {
	for _, v := range LVars {
		if v.name == name {
			return v
		}
	}
	return nil
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
	ndReturn
	ndIf
	ndElse
	ndWhile
	ndFor
)

// Node represents each node in ast
type Node struct {
	kind   nodeKind
	lhs    *Node
	rhs    *Node
	val    int
	offset int
	cond   *Node
	then   *Node
	els    *Node
}

func newNode(kind nodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{kind: kind, lhs: lhs, rhs: rhs}
}

func newNodeNum(val int) *Node {
	return &Node{kind: ndNum, val: val}
}

func newNodeVar(s string) *Node {
	node := &Node{kind: ndLvar}
	v := findLVar(s)
	if v != nil {
		node.offset = v.offset
	} else {
		var offset int
		if len(LVars) > 0 {
			offset = LVars[len(LVars)-1].offset + 8
		} else {
			offset = 8
		}
		v = &LVar{offset: offset, name: s}
		node.offset = offset
		LVars = append(LVars, v)
	}
	return node
}

func newNodeReturn(rhs *Node) *Node {
	return &Node{kind: ndReturn, rhs: rhs}
}

func newNodeIf(cond *Node, then *Node, els *Node) *Node {
	return &Node{kind: ndIf, cond: cond, then: then, els: els}
}

func newNodeWhile(cond *Node, then *Node) *Node {
	return &Node{kind: ndWhile, cond: cond, then: then}
}

// program    = stmt*
// stmt       = expr ";"
//			    | "return" expr ";"
//				| "if" "(" expr ")" stmt ("else" stmt) ?
//				| "while" "(" expr ")" stmt
//				| "for" "(" expr? ";" expr? ";" expr? ")" stmt
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
	var node *Node

	// handle return
	if newToks, isReturn := consume(toks, "return"); isReturn {
		toks, node = expr(newToks)
		node = newNodeReturn(node)
		toks = expect(toks, ";")
		return toks, node
	}

	// handle if
	if newToks, isIf := consume(toks, "if"); isIf {
		var condNode, thenNode, elsNode *Node
		toks = expect(newToks, "(")
		toks, condNode = expr(toks)
		toks = expect(toks, ")")
		toks, thenNode = stmt(toks)

		if newToks, isElse := consume(toks, "else"); isElse {
			toks, node = stmt(newToks)
			elsNode = node
		}

		return toks, newNodeIf(condNode, thenNode, elsNode)
	}

	// handle while
	if newToks, isWhile := consume(toks, "while"); isWhile {
		var condNode, thenNode *Node
		toks = expect(newToks, "(")
		toks, condNode = expr(toks)
		toks = expect(toks, ")")
		toks, thenNode = stmt(toks)
		return toks, newNodeWhile(condNode, thenNode)
	}

	toks, node = expr(toks)
	toks = expect(toks, ";")
	return toks, node
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
		nxtToks, node := expr(newToks)
		expect(nxtToks, ")")
		return nxtToks[1:], node
	}

	if newToks, id, isID := consumeID(toks); isID {
		return newToks, newNodeVar(id)
	}

	toks, n := expectNum(toks)
	return toks, newNodeNum(n)
}
